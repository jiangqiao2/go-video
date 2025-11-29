package app

import (
	"context"
	"fmt"
	"sync"

	"transcode-service/ddd/application/cqe"
	"transcode-service/ddd/application/dto"
	"transcode-service/ddd/domain/entity"
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
}

func DefaultTranscodeApp() TranscodeApp {
	assert.NotCircular()
	onceTranscodeApp.Do(func() {
		singleTranscodeApp = &transcodeAppImpl{
			transcodeRepo: persistence.NewTranscodeRepository(),
			taskQueue:     queue.DefaultTaskQueue(),
		}
	})
	assert.NotNil(singleTranscodeApp)
	return singleTranscodeApp
}

func (t *transcodeAppImpl) CreateTranscodeTask(ctx context.Context, req *cqe.TranscodeTaskCqe) (*dto.TranscodeTaskDTO, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
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
	// TODO: 实现获取转码任务列表逻辑
	return nil, 0, nil
}

func (t *transcodeAppImpl) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, errorMessage string) error {
	// TODO: 实现更新转码任务状态逻辑
	return nil
}

func (t *transcodeAppImpl) CancelTranscodeTask(ctx context.Context, taskUUID string) error {
	// TODO: 实现取消转码任务逻辑
	return nil
}

func (t *transcodeAppImpl) GetTranscodeProgress(ctx context.Context, taskUUID string) (float64, error) {
	// TODO: 实现获取转码进度逻辑
	return 0.0, nil
}
