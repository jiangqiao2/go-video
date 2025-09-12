package app

import (
	"context"
	"fmt"
	"time"
	"transcode-service/ddd/application/dto"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/domain/vo"
)

// WorkerApp Worker应用服务接口
type WorkerApp interface {
	// RegisterWorker 注册Worker
	RegisterWorker(ctx context.Context, req *dto.RegisterWorkerRequest) (*dto.RegisterWorkerResponse, error)
	
	// GetWorker 获取Worker详情
	GetWorker(ctx context.Context, req *dto.GetWorkerRequest) (*dto.GetWorkerResponse, error)
	
	// ListWorkers 获取Worker列表
	ListWorkers(ctx context.Context, req *dto.ListWorkersRequest) (*dto.ListWorkersResponse, error)
	
	// UpdateWorkerStatus 更新Worker状态
	UpdateWorkerStatus(ctx context.Context, req *dto.UpdateWorkerStatusRequest) error
	
	// ProcessHeartbeat 处理Worker心跳
	ProcessHeartbeat(ctx context.Context, req *dto.WorkerHeartbeatRequest) (*dto.WorkerHeartbeatResponse, error)
	
	// GetWorkerStatistics 获取Worker统计
	GetWorkerStatistics(ctx context.Context) (*dto.WorkerStatisticsResponse, error)
	
	// DeleteWorker 删除Worker
	DeleteWorker(ctx context.Context, req *dto.DeleteWorkerRequest) error
	
	// GetWorkerTasks 获取Worker任务列表
	GetWorkerTasks(ctx context.Context, req *dto.WorkerTasksRequest) (*dto.WorkerTasksResponse, error)
	
	// BatchWorkerOperation 批量Worker操作
	BatchWorkerOperation(ctx context.Context, req *dto.BatchWorkerOperationRequest) (*dto.BatchWorkerOperationResponse, error)
	
	// CheckWorkerHealth 检查Worker健康状态
	CheckWorkerHealth(ctx context.Context, workerID string) (*dto.WorkerHealthCheckResponse, error)
	
	// CleanupUnhealthyWorkers 清理不健康的Worker
	CleanupUnhealthyWorkers(ctx context.Context, timeout time.Duration) (int, error)
}

type workerAppImpl struct {
	workerService service.WorkerService
}

// NewWorkerApp 创建Worker应用服务
func NewWorkerApp(workerService service.WorkerService) WorkerApp {
	return &workerAppImpl{
		workerService: workerService,
	}
}

// RegisterWorker 注册Worker
func (w *workerAppImpl) RegisterWorker(ctx context.Context, req *dto.RegisterWorkerRequest) (*dto.RegisterWorkerResponse, error) {
	worker, err := w.workerService.RegisterWorker(ctx, req.WorkerID, req.Name, req.MaxTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to register worker: %w", err)
	}
	
	return &dto.RegisterWorkerResponse{
		WorkerID:     worker.WorkerID(),
		Status:       worker.Status().String(),
		Message:      "Worker registered successfully",
		RegisteredAt: dto.FormatTime(worker.RegisteredAt()),
	}, nil
}

// GetWorker 获取Worker详情
func (w *workerAppImpl) GetWorker(ctx context.Context, req *dto.GetWorkerRequest) (*dto.GetWorkerResponse, error) {
	// 这里需要从仓储获取Worker，简化实现
	// 实际项目中需要通过workerService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// ListWorkers 获取Worker列表
func (w *workerAppImpl) ListWorkers(ctx context.Context, req *dto.ListWorkersRequest) (*dto.ListWorkersResponse, error) {
	// 这里需要从仓储获取Worker列表，简化实现
	// 实际项目中需要通过workerService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// UpdateWorkerStatus 更新Worker状态
func (w *workerAppImpl) UpdateWorkerStatus(ctx context.Context, req *dto.UpdateWorkerStatusRequest) error {
	// 验证状态
	status := vo.WorkerStatus(req.Status)
	if !status.IsValid() {
		return fmt.Errorf("invalid worker status: %s", req.Status)
	}
	
	// 根据状态执行相应操作
	switch status {
	case vo.WorkerStatusOnline:
		return w.workerService.SetWorkerOnline(ctx, req.WorkerID)
	case vo.WorkerStatusOffline:
		return w.workerService.SetWorkerOffline(ctx, req.WorkerID)
	case vo.WorkerStatusMaintenance:
		return w.workerService.SetWorkerMaintenance(ctx, req.WorkerID)
	default:
		return fmt.Errorf("unsupported status update: %s", req.Status)
	}
}

// ProcessHeartbeat 处理Worker心跳
func (w *workerAppImpl) ProcessHeartbeat(ctx context.Context, req *dto.WorkerHeartbeatRequest) (*dto.WorkerHeartbeatResponse, error) {
	// 创建心跳对象
	heartbeat := &vo.WorkerHeartbeat{
		WorkerID:        req.WorkerID,
		Status:          vo.WorkerStatus(req.Status),
		CurrentTasks:    req.CurrentTasks,
		MaxTasks:        req.MaxTasks,
		CPUUsage:        req.CPUUsage,
		MemoryUsage:     req.MemoryUsage,
		LastHeartbeatAt: time.Now(),
		SystemInfo:      req.SystemInfo,
	}
	
	// 更新心跳
	err := w.workerService.UpdateWorkerHeartbeat(ctx, heartbeat)
	if err != nil {
		return nil, fmt.Errorf("failed to update worker heartbeat: %w", err)
	}
	
	return &dto.WorkerHeartbeatResponse{
		WorkerID:  req.WorkerID,
		Status:    "success",
		Message:   "Heartbeat processed successfully",
		Timestamp: dto.FormatTime(time.Now()),
	}, nil
}

// GetWorkerStatistics 获取Worker统计
func (w *workerAppImpl) GetWorkerStatistics(ctx context.Context) (*dto.WorkerStatisticsResponse, error) {
	stats, err := w.workerService.GetWorkerStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get worker statistics: %w", err)
	}
	
	return &dto.WorkerStatisticsResponse{
		TotalWorkers:       stats.TotalWorkers,
		OnlineWorkers:      stats.OnlineWorkers,
		OfflineWorkers:     stats.OfflineWorkers,
		BusyWorkers:        stats.BusyWorkers,
		IdleWorkers:        stats.IdleWorkers,
		MaintenanceWorkers: stats.MaintenanceWorkers,
		TotalTasks:         stats.TotalTasks,
		AverageCPUUsage:    stats.AverageCPUUsage,
		AverageMemoryUsage: stats.AverageMemoryUsage,
		AverageLoadFactor:  stats.AverageLoadFactor,
	}, nil
}

// DeleteWorker 删除Worker
func (w *workerAppImpl) DeleteWorker(ctx context.Context, req *dto.DeleteWorkerRequest) error {
	if !req.Force {
		// 验证Worker是否可以删除（没有正在处理的任务）
		err := w.workerService.ValidateWorkerCapacity(ctx, req.WorkerID)
		if err == nil {
			// Worker还有任务在处理，不能删除
			return fmt.Errorf("worker has active tasks, use force=true to delete anyway")
		}
	}
	
	// 先设置Worker为离线状态
	err := w.workerService.SetWorkerOffline(ctx, req.WorkerID)
	if err != nil {
		// 忽略错误，继续删除
	}
	
	// 删除Worker（这里需要通过repository实现）
	// return w.workerRepository.DeleteWorker(ctx, req.WorkerID)
	return fmt.Errorf("not implemented")
}

// GetWorkerTasks 获取Worker任务列表
func (w *workerAppImpl) GetWorkerTasks(ctx context.Context, req *dto.WorkerTasksRequest) (*dto.WorkerTasksResponse, error) {
	// 这里需要从任务仓储获取Worker的任务列表，简化实现
	// 实际项目中需要通过taskRepository获取
	return nil, fmt.Errorf("not implemented")
}

// BatchWorkerOperation 批量Worker操作
func (w *workerAppImpl) BatchWorkerOperation(ctx context.Context, req *dto.BatchWorkerOperationRequest) (*dto.BatchWorkerOperationResponse, error) {
	successCount := 0
	failedCount := 0
	var failedWorkers []string
	
	for _, workerID := range req.WorkerIDs {
		var err error
		
		switch req.Operation {
		case "online":
			err = w.workerService.SetWorkerOnline(ctx, workerID)
		case "offline":
			err = w.workerService.SetWorkerOffline(ctx, workerID)
		case "maintenance":
			err = w.workerService.SetWorkerMaintenance(ctx, workerID)
		default:
			err = fmt.Errorf("unsupported operation: %s", req.Operation)
		}
		
		if err != nil {
			failedCount++
			failedWorkers = append(failedWorkers, workerID)
		} else {
			successCount++
		}
	}
	
	return &dto.BatchWorkerOperationResponse{
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		FailedWorkers: failedWorkers,
		Message:       fmt.Sprintf("Batch %s completed: %d success, %d failed", req.Operation, successCount, failedCount),
	}, nil
}

// CheckWorkerHealth 检查Worker健康状态
func (w *workerAppImpl) CheckWorkerHealth(ctx context.Context, workerID string) (*dto.WorkerHealthCheckResponse, error) {
	// 这里需要从仓储获取Worker信息，简化实现
	// 实际项目中需要通过workerService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// CleanupUnhealthyWorkers 清理不健康的Worker
func (w *workerAppImpl) CleanupUnhealthyWorkers(ctx context.Context, timeout time.Duration) (int, error) {
	cleanedCount, err := w.workerService.CleanupUnhealthyWorkers(ctx, timeout)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup unhealthy workers: %w", err)
	}
	
	// Cleaned up unhealthy workers if any
	return cleanedCount, nil
}