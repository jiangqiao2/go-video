package service

import (
	"context"
	"fmt"
	"time"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
)

// WorkerService Worker管理领域服务
type WorkerService interface {
	// RegisterWorker 注册Worker
	RegisterWorker(ctx context.Context, workerID, name string, maxTasks int) (*entity.WorkerEntity, error)
	
	// UpdateWorkerHeartbeat 更新Worker心跳
	UpdateWorkerHeartbeat(ctx context.Context, heartbeat *vo.WorkerHeartbeat) error
	
	// GetBestWorkerForTask 获取最适合执行任务的Worker
	GetBestWorkerForTask(ctx context.Context) (*entity.WorkerEntity, error)
	
	// AssignTaskToWorker 分配任务给Worker
	AssignTaskToWorker(ctx context.Context, workerID string) error
	
	// CompleteWorkerTask 完成Worker任务
	CompleteWorkerTask(ctx context.Context, workerID string) error
	
	// SetWorkerMaintenance 设置Worker维护模式
	SetWorkerMaintenance(ctx context.Context, workerID string) error
	
	// SetWorkerOnline 设置Worker在线
	SetWorkerOnline(ctx context.Context, workerID string) error
	
	// SetWorkerOffline 设置Worker离线
	SetWorkerOffline(ctx context.Context, workerID string) error
	
	// CleanupUnhealthyWorkers 清理不健康的Worker
	CleanupUnhealthyWorkers(ctx context.Context, timeout time.Duration) (int, error)
	
	// GetWorkerStatistics 获取Worker统计信息
	GetWorkerStatistics(ctx context.Context) (*repo.WorkerStatistics, error)
	
	// ValidateWorkerCapacity 验证Worker容量
	ValidateWorkerCapacity(ctx context.Context, workerID string) error
}

type workerServiceImpl struct {
	workerRepo repo.WorkerRepository
}

// NewWorkerService 创建Worker管理领域服务
func NewWorkerService(workerRepo repo.WorkerRepository) WorkerService {
	return &workerServiceImpl{
		workerRepo: workerRepo,
	}
}

// RegisterWorker 注册Worker
func (s *workerServiceImpl) RegisterWorker(ctx context.Context, workerID, name string, maxTasks int) (*entity.WorkerEntity, error) {
	// 检查Worker是否已存在
	existingWorker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err == nil && existingWorker != nil {
		return nil, fmt.Errorf("worker %s already exists", workerID)
	}
	
	// 验证参数
	if workerID == "" {
		return nil, fmt.Errorf("worker ID cannot be empty")
	}
	if maxTasks <= 0 {
		return nil, fmt.Errorf("max tasks must be greater than 0")
	}
	
	// 创建Worker实体
	worker := entity.NewWorkerEntity(workerID, name, maxTasks)
	
	// 保存到仓储
	err = s.workerRepo.RegisterWorker(ctx, worker)
	if err != nil {
		return nil, fmt.Errorf("failed to register worker: %w", err)
	}
	
	return worker, nil
}

// UpdateWorkerHeartbeat 更新Worker心跳
func (s *workerServiceImpl) UpdateWorkerHeartbeat(ctx context.Context, heartbeat *vo.WorkerHeartbeat) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, heartbeat.WorkerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 更新心跳
	err = worker.UpdateHeartbeat(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to update worker heartbeat: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// GetBestWorkerForTask 获取最适合执行任务的Worker
func (s *workerServiceImpl) GetBestWorkerForTask(ctx context.Context) (*entity.WorkerEntity, error) {
	// 获取所有可用的Worker
	availableWorkers, err := s.workerRepo.GetAvailableWorkers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available workers: %w", err)
	}
	
	if len(availableWorkers) == 0 {
		return nil, fmt.Errorf("no available workers")
	}
	
	// 选择负载最低的Worker
	var bestWorker *entity.WorkerEntity
	lowestLoad := 1.1 // 大于1.0，确保能找到更好的
	
	for _, worker := range availableWorkers {
		if worker.CanAcceptTask() {
			loadFactor := worker.GetLoadFactor()
			if loadFactor < lowestLoad {
				lowestLoad = loadFactor
				bestWorker = worker
			}
		}
	}
	
	if bestWorker == nil {
		return nil, fmt.Errorf("no suitable worker found")
	}
	
	return bestWorker, nil
}

// AssignTaskToWorker 分配任务给Worker
func (s *workerServiceImpl) AssignTaskToWorker(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 分配任务
	err = worker.AssignTask()
	if err != nil {
		return fmt.Errorf("failed to assign task to worker: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// CompleteWorkerTask 完成Worker任务
func (s *workerServiceImpl) CompleteWorkerTask(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 完成任务
	err = worker.CompleteTask()
	if err != nil {
		return fmt.Errorf("failed to complete worker task: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// SetWorkerMaintenance 设置Worker维护模式
func (s *workerServiceImpl) SetWorkerMaintenance(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 设置维护模式
	err = worker.SetMaintenance()
	if err != nil {
		return fmt.Errorf("failed to set worker maintenance: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// SetWorkerOnline 设置Worker在线
func (s *workerServiceImpl) SetWorkerOnline(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 设置在线
	err = worker.Online()
	if err != nil {
		return fmt.Errorf("failed to set worker online: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// SetWorkerOffline 设置Worker离线
func (s *workerServiceImpl) SetWorkerOffline(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	// 设置离线
	err = worker.Offline()
	if err != nil {
		return fmt.Errorf("failed to set worker offline: %w", err)
	}
	
	// 保存更新
	err = s.workerRepo.UpdateWorker(ctx, worker)
	if err != nil {
		return fmt.Errorf("failed to save worker: %w", err)
	}
	
	return nil
}

// CleanupUnhealthyWorkers 清理不健康的Worker
func (s *workerServiceImpl) CleanupUnhealthyWorkers(ctx context.Context, timeout time.Duration) (int, error) {
	unhealthyWorkers, err := s.workerRepo.GetUnhealthyWorkers(ctx, timeout)
	if err != nil {
		return 0, fmt.Errorf("failed to get unhealthy workers: %w", err)
	}
	
	cleanedCount := 0
	for _, worker := range unhealthyWorkers {
		if !worker.IsHealthy(timeout) {
			// 设置Worker为离线状态
			err = worker.Offline()
			if err != nil {
				continue // 跳过无法设置离线的Worker
			}
			
			// 保存更新
			err = s.workerRepo.UpdateWorker(ctx, worker)
			if err != nil {
				continue // 跳过保存失败的Worker
			}
			
			cleanedCount++
		}
	}
	
	return cleanedCount, nil
}

// GetWorkerStatistics 获取Worker统计信息
func (s *workerServiceImpl) GetWorkerStatistics(ctx context.Context) (*repo.WorkerStatistics, error) {
	return s.workerRepo.GetWorkerStatistics(ctx)
}

// ValidateWorkerCapacity 验证Worker容量
func (s *workerServiceImpl) ValidateWorkerCapacity(ctx context.Context, workerID string) error {
	worker, err := s.workerRepo.GetWorkerByID(ctx, workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	
	if !worker.CanAcceptTask() {
		return fmt.Errorf("worker %s cannot accept new tasks", workerID)
	}
	
	return nil
}