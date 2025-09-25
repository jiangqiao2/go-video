package service

import (
	"context"
	"fmt"
	"transcode-service/ddd/application/dto"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/domain/vo"
	"transcode-service/pkg/errno"
)

// TranscodeApp 转码应用服务接口
type TranscodeApp interface {
	CreateTranscodeTask(ctx context.Context, userUUID, videoUUID, originalPath, resolution, bitrate string) (*dto.TranscodeTaskDto, error)
	QueryTranscodeTask(ctx context.Context, taskUUID, userUUID string) (*dto.TranscodeTaskDto, error)
	ListTranscodeTasks(ctx context.Context, userUUID, videoUUID, status string, page, size int) (*dto.TranscodeTaskListDto, error)
	UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, userUUID, status string) error
	CancelTranscodeTask(ctx context.Context, taskUUID, userUUID string) error
	GetTranscodeProgress(ctx context.Context, taskUUID, userUUID string) (*dto.TranscodeProgressDto, error)
}

// transcodeAppImpl 转码应用服务实现
type transcodeAppImpl struct {
	transcodeRepo    repo.TranscodeTaskRepository
	transcodeService service.TranscodeService
}

// NewTranscodeApp 创建转码应用服务
func NewTranscodeApp(transcodeRepo repo.TranscodeTaskRepository, transcodeService service.TranscodeService) TranscodeApp {
	return &transcodeAppImpl{
		transcodeRepo:    transcodeRepo,
		transcodeService: transcodeService,
	}
}

// CreateTranscodeTask 创建转码任务
func (app *transcodeAppImpl) CreateTranscodeTask(ctx context.Context, userUUID, videoUUID, originalPath, resolution, bitrate string) (*dto.TranscodeTaskDto, error) {
	// 验证转码参数
	params, err := vo.NewTranscodeParams(resolution, bitrate)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrInvalidParam, err)
	}

	// 验证是否可以创建任务
	err = app.transcodeService.CanCreateTask(ctx, userUUID, videoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrTranscodeTaskExists, err)
	}

	// 创建转码任务实体
	task := entity.DefaultTranscodeTaskEntity(userUUID, videoUUID, originalPath, *params)
	
	// 保存到仓储
	err = app.transcodeRepo.CreateTranscodeTask(ctx, task)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	
	return dto.NewTranscodeTaskDto(task), nil
}

// QueryTranscodeTask 查询转码任务
func (app *transcodeAppImpl) QueryTranscodeTask(ctx context.Context, taskUUID, userUUID string) (*dto.TranscodeTaskDto, error) {
	task, err := app.transcodeRepo.QueryTranscodeTaskByUUID(ctx, taskUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrTranscodeTaskNotFound, err)
	}

	// 验证用户权限
	if task.UserUUID() != userUUID {
		return nil, errno.NewBizError(errno.ErrUnauthorized, fmt.Errorf("user %s not authorized to access task %s", userUUID, taskUUID))
	}

	return dto.NewTranscodeTaskDto(task), nil
}

// ListTranscodeTasks 列表转码任务
func (app *transcodeAppImpl) ListTranscodeTasks(ctx context.Context, userUUID, videoUUID, status string, page, size int) (*dto.TranscodeTaskListDto, error) {
	// 根据用户UUID查询任务
	tasks, err := app.transcodeRepo.QueryTranscodeTasksByUserUUID(ctx, userUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}

	// 过滤结果
	filteredTasks := make([]*entity.TranscodeTaskEntity, 0)
	for _, task := range tasks {
		if videoUUID != "" && task.VideoUUID() != videoUUID {
			continue
		}
		if status != "" && task.Status().String() != status {
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}

	// 分页处理
	total := int64(len(filteredTasks))
	start := (page - 1) * size
	end := start + size
	if start >= len(filteredTasks) {
		filteredTasks = []*entity.TranscodeTaskEntity{}
	} else if end > len(filteredTasks) {
		filteredTasks = filteredTasks[start:]
	} else {
		filteredTasks = filteredTasks[start:end]
	}

	return dto.NewTranscodeTaskListDto(filteredTasks, total, page, size), nil
}

// UpdateTranscodeTaskStatus 更新转码任务状态
func (app *transcodeAppImpl) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, userUUID, status string) error {
	// 查询任务
	task, err := app.transcodeRepo.QueryTranscodeTaskByUUID(ctx, taskUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrTranscodeTaskNotFound, err)
	}

	// 验证用户权限
	if task.UserUUID() != userUUID {
		return errno.NewBizError(errno.ErrUnauthorized, fmt.Errorf("user %s not authorized to update task %s", userUUID, taskUUID))
	}

	// 验证状态转换
	newStatus, err := vo.NewTaskStatusFromString(status)
	if err != nil {
		return errno.NewBizError(errno.ErrInvalidTaskStatus, err)
	}

	if !task.Status().CanTransitionTo(newStatus) {
		return errno.NewBizError(errno.ErrInvalidTaskStatus, fmt.Errorf("cannot transition from %s to %s", task.Status().String(), status))
	}

	// 更新状态
	err = app.transcodeRepo.UpdateTranscodeTaskStatus(ctx, taskUUID, newStatus)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}

	return nil
}

// CancelTranscodeTask 取消转码任务
func (app *transcodeAppImpl) CancelTranscodeTask(ctx context.Context, taskUUID, userUUID string) error {
	// 查询任务
	task, err := app.transcodeRepo.QueryTranscodeTaskByUUID(ctx, taskUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrTranscodeTaskNotFound, err)
	}

	// 验证用户权限
	if task.UserUUID() != userUUID {
		return errno.NewBizError(errno.ErrUnauthorized, fmt.Errorf("user %s not authorized to cancel task %s", userUUID, taskUUID))
	}

	// 验证是否可以取消
	if !task.CanCancel() {
		return errno.NewBizError(errno.ErrInvalidTaskStatus, fmt.Errorf("task %s cannot be cancelled in status %s", taskUUID, task.Status().String()))
	}

	// 更新状态为已取消
	err = app.transcodeRepo.UpdateTranscodeTaskStatus(ctx, taskUUID, vo.TaskStatusCancelled)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}

	return nil
}

// GetTranscodeProgress 获取转码进度
func (app *transcodeAppImpl) GetTranscodeProgress(ctx context.Context, taskUUID, userUUID string) (*dto.TranscodeProgressDto, error) {
	task, err := app.transcodeRepo.QueryTranscodeTaskByUUID(ctx, taskUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrTranscodeTaskNotFound, err)
	}

	// 验证用户权限
	if task.UserUUID() != userUUID {
		return nil, errno.NewBizError(errno.ErrUnauthorized, fmt.Errorf("user %s not authorized to access task %s", userUUID, taskUUID))
	}

	return dto.NewTranscodeProgressDto(task), nil
}