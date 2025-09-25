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
	
	// QueryTranscodeTaskByUUID 根据任务UUID查询转码任务
	QueryTranscodeTaskByUUID(ctx context.Context, taskUUID string) (*entity.TranscodeTaskEntity, error)
	
	// QueryTranscodeTasksByVideoUUID 根据视频UUID查询转码任务列表
	QueryTranscodeTasksByVideoUUID(ctx context.Context, videoUUID string) ([]*entity.TranscodeTaskEntity, error)
	
	// QueryTranscodeTasksByUserUUID 根据用户UUID查询转码任务列表
	QueryTranscodeTasksByUserUUID(ctx context.Context, userUUID string) ([]*entity.TranscodeTaskEntity, error)
	
	// QueryTranscodeTasksByStatus 根据状态查询转码任务列表
	QueryTranscodeTasksByStatus(ctx context.Context, status vo.TaskStatus) ([]*entity.TranscodeTaskEntity, error)
	
	// UpdateTranscodeTaskStatus 更新转码任务状态
	UpdateTranscodeTaskStatus(ctx context.Context, taskUUID string, status vo.TaskStatus) error
	
	// UpdateTranscodeTaskProgress 更新转码任务进度
	UpdateTranscodeTaskProgress(ctx context.Context, taskUUID string, progress int) error
	
	// UpdateTranscodeTaskOutputPath 更新转码任务输出路径
	UpdateTranscodeTaskOutputPath(ctx context.Context, taskUUID string, outputPath string) error
	
	// UpdateTranscodeTaskError 更新转码任务错误信息
	UpdateTranscodeTaskError(ctx context.Context, taskUUID string, errorMessage string) error
	
	// DeleteTranscodeTask 删除转码任务
	DeleteTranscodeTask(ctx context.Context, taskUUID string) error
	
	// CountTranscodeTasksByStatus 统计指定状态的转码任务数量
	CountTranscodeTasksByStatus(ctx context.Context, status vo.TaskStatus) (int64, error)
}

// TranscodeTaskQuery 转码任务查询条件
type TranscodeTaskQuery struct {
	UserUUID    string
	VideoUUID   string
	TaskUUID    string
	Status      *vo.TaskStatus
	Limit       int
	Offset      int
}
