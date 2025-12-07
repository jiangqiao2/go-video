package app

import (
	"context"
	"fmt"
	"sync"

	"transcode-service/ddd/application/cqe"
	"transcode-service/ddd/application/dto"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/port"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/persistence"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/errno"
	"transcode-service/pkg/logger"
)

var (
	singleTranscodeApp TranscodeApp
	onceTranscodeApp   sync.Once
)

type TranscodeApp interface {
	// CreateTranscodeTask 创建转码任务
	CreateTranscodeTask(ctx context.Context, req *cqe.TranscodeTaskCqe) (*dto.TranscodeTaskDTO, error)
	// GetTranscodeTask 获取转码任务详情
	GetTranscodeTask(ctx context.Context, taskUUID string) (*dto.TranscodeTaskDTO, error)
	// ListTranscodeTasks 获取转码任务列表
	ListTranscodeTasks(ctx context.Context, userUUID string, page, size int) ([]*dto.TranscodeTaskDTO, int64, error)
	// UpdateTranscodeTaskStatus 更新转码任务状态
	UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, errorMessage string) error
	// CancelTranscodeTask 取消转码任务
	CancelTranscodeTask(ctx context.Context, taskUUID string) error
	// GetTranscodeProgress 获取转码进度
	GetTranscodeProgress(ctx context.Context, taskUUID string) (float64, error)
}

type transcodeAppImpl struct {
	transcodeRepo repo.TranscodeJobRepository
	taskQueue     queue.TaskQueue
	progressSink  port.ProgressSink
	maxRetries    int
}

func DefaultTranscodeApp() TranscodeApp {
	assert.NotCircular()
	onceTranscodeApp.Do(func() {
		singleTranscodeApp = NewTranscodeAppWith(persistence.NewTranscodeRepository(), queue.DefaultTaskQueue(), nil, 3)
	})
	assert.NotNil(singleTranscodeApp)
	return singleTranscodeApp
}

func NewTranscodeAppWith(repo repo.TranscodeJobRepository, q queue.TaskQueue, sink port.ProgressSink, maxRetries int) TranscodeApp {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &transcodeAppImpl{
		transcodeRepo: repo,
		taskQueue:     q,
		progressSink:  sink,
		maxRetries:    maxRetries,
	}
}

func (t *transcodeAppImpl) CreateTranscodeTask(ctx context.Context, req *cqe.TranscodeTaskCqe) (*dto.TranscodeTaskDTO, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 幂等：检查同一视频是否已有未完成任务
	if existing, err := t.findActiveByVideo(ctx, req.VideoUUID); err == nil && existing != nil {
		return dto.NewTranscodeTaskDto(existing), nil
	}

	// 创建转码参数
	params, err := vo.NewTranscodeParams(req.Resolution, req.Bitrate)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrInvalidParam, err)
	}

	// 创建转码任务实体（不再与 HLS 耦合）
	task := entity.DefaultTranscodeTaskEntity(req.UserUUID, req.VideoUUID, req.VideoPushUUID, req.OriginalPath, *params)

	// 保存到仓储
	err = t.transcodeRepo.CreateTranscodeJob(ctx, task)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}

	// 将任务加入队列，触发异步处理
	if err := t.taskQueue.Enqueue(ctx, task); err != nil {
		logger.Errorf("任务入队失败 task_uuid=%s error=%v", task.TaskUUID(), err)
		failErr := fmt.Errorf("enqueue task failed: %w", err)
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(failErr.Error())
		_ = t.transcodeRepo.UpdateTranscodeJobStatus(ctx, task.TaskUUID(), vo.TaskStatusFailed, failErr.Error(), task.OutputPath(), task.Progress())
		return nil, errno.ErrQueueFull
	}

	// 转换为DTO返回
	return dto.NewTranscodeTaskDto(task), nil
}

func (t *transcodeAppImpl) GetTranscodeTask(ctx context.Context, taskUUID string) (*dto.TranscodeTaskDTO, error) {
	if taskUUID == "" {
		return nil, errno.ErrTaskUUIDRequired
	}
	taskEntity, err := t.transcodeRepo.GetTranscodeJob(ctx, taskUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if taskEntity == nil {
		return nil, errno.ErrTranscodeTaskNotFound
	}
	return dto.NewTranscodeTaskDto(taskEntity), nil
}

func (t *transcodeAppImpl) ListTranscodeTasks(ctx context.Context, userUUID string, page, size int) ([]*dto.TranscodeTaskDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}
	statuses := []vo.TaskStatus{vo.TaskStatusProcessing, vo.TaskStatusPending, vo.TaskStatusCompleted, vo.TaskStatusFailed, vo.TaskStatusCancelled}
	var all []*entity.TranscodeTaskEntity
	for _, st := range statuses {
		list, err := t.transcodeRepo.QueryTranscodeJobsByStatus(ctx, st, size*page)
		if err != nil {
			continue
		}
		for _, job := range list {
			if job == nil {
				continue
			}
			if userUUID != "" && job.UserUUID() != userUUID {
				continue
			}
			all = append(all, job)
		}
	}
	total := int64(len(all))
	start := (page - 1) * size
	if start > len(all) {
		return []*dto.TranscodeTaskDTO{}, total, nil
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	slice := all[start:end]
	dtos := make([]*dto.TranscodeTaskDTO, 0, len(slice))
	for _, e := range slice {
		dtos = append(dtos, dto.NewTranscodeTaskDto(e))
	}
	return dtos, total, nil
}

func (t *transcodeAppImpl) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, errorMessage string) error {
	if taskUUID == "" {
		return errno.ErrTaskUUIDRequired
	}
	target, err := vo.NewTaskStatusFromString(status)
	if err != nil {
		return errno.ErrInvalidTaskStatus
	}
	task, err := t.transcodeRepo.GetTranscodeJob(ctx, taskUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	if task == nil {
		return errno.ErrTranscodeTaskNotFound
	}
	if !task.Status().CanTransitionTo(target) {
		return errno.ErrInvalidTaskStatus
	}
	return t.transcodeRepo.UpdateTranscodeJobStatus(ctx, taskUUID, target, errorMessage, task.OutputPath(), task.Progress())
}

func (t *transcodeAppImpl) CancelTranscodeTask(ctx context.Context, taskUUID string) error {
	return t.UpdateTranscodeTaskStatus(ctx, taskUUID, vo.TaskStatusCancelled.String(), "cancelled by user")
}

func (t *transcodeAppImpl) GetTranscodeProgress(ctx context.Context, taskUUID string) (float64, error) {
	task, err := t.transcodeRepo.GetTranscodeJob(ctx, taskUUID)
	if err != nil {
		return 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if task == nil {
		return 0, errno.ErrTranscodeTaskNotFound
	}
	return float64(task.Progress()), nil
}

// findActiveByVideo returns a pending/processing task for the same video if exists.
func (t *transcodeAppImpl) findActiveByVideo(ctx context.Context, videoUUID string) (*entity.TranscodeTaskEntity, error) {
	if videoUUID == "" {
		return nil, nil
	}
	statuses := []vo.TaskStatus{vo.TaskStatusPending, vo.TaskStatusProcessing}
	for _, st := range statuses {
		jobs, err := t.transcodeRepo.QueryTranscodeJobsByStatus(ctx, st, 100)
		if err != nil {
			continue
		}
		for _, job := range jobs {
			if job != nil && job.VideoUUID() == videoUUID {
				return job, nil
			}
		}
	}
	return nil, nil
}
