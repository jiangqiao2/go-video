package repo

import (
	"context"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/vo"
)

// TranscodeTaskRepository 转码任务仓储接口
type TranscodeTaskRepository interface {
	// CreateTask 创建任务
	CreateTask(ctx context.Context, task *entity.TranscodeTaskEntity) error
	
	// GetTaskByID 根据ID获取任务
	GetTaskByID(ctx context.Context, taskID string) (*entity.TranscodeTaskEntity, error)
	
	// GetTasksByUserID 根据用户ID获取任务列表
	GetTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.TranscodeTaskEntity, error)
	
	// GetTasksByStatus 根据状态获取任务列表
	GetTasksByStatus(ctx context.Context, status vo.TaskStatus, limit, offset int) ([]*entity.TranscodeTaskEntity, error)
	
	// GetPendingTasks 获取待处理任务（按优先级排序）
	GetPendingTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error)
	
	// GetTasksByWorkerID 根据Worker ID获取任务列表
	GetTasksByWorkerID(ctx context.Context, workerID string) ([]*entity.TranscodeTaskEntity, error)
	
	// UpdateTask 更新任务
	UpdateTask(ctx context.Context, task *entity.TranscodeTaskEntity) error
	
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(ctx context.Context, taskID string, status vo.TaskStatus) error
	
	// UpdateTaskProgress 更新任务进度
	UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error
	
	// AssignTaskToWorker 分配任务给Worker
	AssignTaskToWorker(ctx context.Context, taskID, workerID string) error
	
	// GetExpiredTasks 获取过期任务
	GetExpiredTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error)
	
	// GetFailedTasksForRetry 获取可重试的失败任务
	GetFailedTasksForRetry(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error)
	
	// DeleteTask 删除任务
	DeleteTask(ctx context.Context, taskID string) error
	
	// CountTasksByStatus 统计各状态任务数量
	CountTasksByStatus(ctx context.Context, status vo.TaskStatus) (int64, error)
	
	// CountTasksByUserID 统计用户任务数量
	CountTasksByUserID(ctx context.Context, userID string) (int64, error)
	
	// GetTaskStatistics 获取任务统计信息
	GetTaskStatistics(ctx context.Context) (*TaskStatistics, error)
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TotalTasks      int64 `json:"total_tasks"`
	PendingTasks    int64 `json:"pending_tasks"`
	ProcessingTasks int64 `json:"processing_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	FailedTasks     int64 `json:"failed_tasks"`
	CancelledTasks  int64 `json:"cancelled_tasks"`
}

// TaskQuery 任务查询条件
type TaskQuery struct {
	UserID     string           `json:"user_id,omitempty"`
	Status     []vo.TaskStatus  `json:"status,omitempty"`
	WorkerID   string           `json:"worker_id,omitempty"`
	Priority   *int             `json:"priority,omitempty"`
	StartTime  *string          `json:"start_time,omitempty"`
	EndTime    *string          `json:"end_time,omitempty"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	OrderBy    string           `json:"order_by,omitempty"`
	OrderDesc  bool             `json:"order_desc,omitempty"`
}