package repo

import (
	"context"
	"time"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/vo"
)

// WorkerRepository Worker仓储接口
type WorkerRepository interface {
	// RegisterWorker 注册Worker
	RegisterWorker(ctx context.Context, worker *entity.WorkerEntity) error
	
	// GetWorkerByID 根据ID获取Worker
	GetWorkerByID(ctx context.Context, workerID string) (*entity.WorkerEntity, error)
	
	// GetAllWorkers 获取所有Worker
	GetAllWorkers(ctx context.Context) ([]*entity.WorkerEntity, error)
	
	// GetWorkersByStatus 根据状态获取Worker列表
	GetWorkersByStatus(ctx context.Context, status vo.WorkerStatus) ([]*entity.WorkerEntity, error)
	
	// GetAvailableWorkers 获取可用的Worker（在线且未满负载）
	GetAvailableWorkers(ctx context.Context) ([]*entity.WorkerEntity, error)
	
	// UpdateWorker 更新Worker
	UpdateWorker(ctx context.Context, worker *entity.WorkerEntity) error
	
	// UpdateWorkerStatus 更新Worker状态
	UpdateWorkerStatus(ctx context.Context, workerID string, status vo.WorkerStatus) error
	
	// UpdateWorkerHeartbeat 更新Worker心跳
	UpdateWorkerHeartbeat(ctx context.Context, workerID string, heartbeat *vo.WorkerHeartbeat) error
	
	// DeleteWorker 删除Worker
	DeleteWorker(ctx context.Context, workerID string) error
	
	// GetUnhealthyWorkers 获取不健康的Worker
	GetUnhealthyWorkers(ctx context.Context, timeout time.Duration) ([]*entity.WorkerEntity, error)
	
	// GetWorkerStatistics 获取Worker统计信息
	GetWorkerStatistics(ctx context.Context) (*WorkerStatistics, error)
	
	// GetBestWorkerForTask 获取最适合执行任务的Worker
	GetBestWorkerForTask(ctx context.Context) (*entity.WorkerEntity, error)
}

// WorkerStatistics Worker统计信息
type WorkerStatistics struct {
	TotalWorkers     int64   `json:"total_workers"`
	OnlineWorkers    int64   `json:"online_workers"`
	OfflineWorkers   int64   `json:"offline_workers"`
	BusyWorkers      int64   `json:"busy_workers"`
	IdleWorkers      int64   `json:"idle_workers"`
	MaintenanceWorkers int64 `json:"maintenance_workers"`
	TotalTasks       int64   `json:"total_tasks"`
	AverageCPUUsage  float64 `json:"average_cpu_usage"`
	AverageMemoryUsage float64 `json:"average_memory_usage"`
	AverageLoadFactor float64 `json:"average_load_factor"`
}