package repo

import (
	"context"
	"transcode-service/ddd/domain/entity"
)

// TranscodeTaskRepository 转码任务仓储接口
type TranscodeTaskRepository interface {
	// CreateTranscodeTask 创建转码任务
	CreateTranscodeTask(ctx context.Context, task *entity.TranscodeTaskEntity) error
}
