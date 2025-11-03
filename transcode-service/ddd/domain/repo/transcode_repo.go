package repo

import (
	"context"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/vo"
)

// TranscodeTaskRepository 转码任务仓储接口
type TranscodeTaskRepository interface {
	// CreateTranscodeTask 创建转码任务
	CreateTranscodeTask(ctx context.Context, task *entity.TranscodeTaskEntity) error
	UpdateTranscodeTaskProgress(ctx context.Context, taskUUID string, progress int) error
	UpdateTranscodeTask(ctx context.Context, task *entity.TranscodeTaskEntity) error
	GetTranscodeTask(ctx context.Context, taskUUID string) (*entity.TranscodeTaskEntity, error)
	UpdateTranscodeTaskStatus(ctx context.Context, taskUUID string, status vo.TaskStatus, message, outputPath string, progress int) error
	QueryTranscodeTasksByStatus(ctx context.Context, status vo.TaskStatus, limit int) ([]*entity.TranscodeTaskEntity, error)
	
	// HLS相关方法
	UpdateHLSProgress(ctx context.Context, taskUUID string, progress int) error
	UpdateHLSStatus(ctx context.Context, taskUUID string, status string) error
	UpdateHLSOutputPath(ctx context.Context, taskUUID string, outputPath string) error
	UpdateHLSError(ctx context.Context, taskUUID string, errorMessage string) error
	UpdateHLSCompleted(ctx context.Context, taskUUID string) error
	UpdateHLSFailed(ctx context.Context, taskUUID string, errorMessage string) error
	QueryHLSEnabledTasks(ctx context.Context, status string, limit int) ([]*entity.TranscodeTaskEntity, error)
}
