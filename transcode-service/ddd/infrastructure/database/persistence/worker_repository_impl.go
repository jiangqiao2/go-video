package persistence

import (
	"context"
	"fmt"
	"time"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/convertor"
	"transcode-service/ddd/infrastructure/database/dao"
	"gorm.io/gorm"
)

// workerRepositoryImpl Worker仓储实现
type workerRepositoryImpl struct {
	workerDao *dao.WorkerDao
	convertor *convertor.WorkerConvertor
}

// NewWorkerRepository 创建Worker仓储实现
func NewWorkerRepository(db *gorm.DB) repo.WorkerRepository {
	return &workerRepositoryImpl{
		workerDao: dao.NewWorkerDao(db),
		convertor: convertor.NewWorkerConvertor(),
	}
}

// RegisterWorker 注册Worker
func (r *workerRepositoryImpl) RegisterWorker(ctx context.Context, worker *entity.WorkerEntity) error {
	po, err := r.convertor.EntityToPO(worker)
	if err != nil {
		return fmt.Errorf("failed to convert entity to po: %w", err)
	}
	
	return r.workerDao.Create(ctx, po)
}

// GetWorkerByID 根据ID获取Worker
func (r *workerRepositoryImpl) GetWorkerByID(ctx context.Context, workerID string) (*entity.WorkerEntity, error) {
	po, err := r.workerDao.GetByWorkerID(ctx, workerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.convertor.POToEntity(po)
}

// GetAllWorkers 获取所有Worker
func (r *workerRepositoryImpl) GetAllWorkers(ctx context.Context) ([]*entity.WorkerEntity, error) {
	poList, err := r.workerDao.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetWorkersByStatus 根据状态获取Worker列表
func (r *workerRepositoryImpl) GetWorkersByStatus(ctx context.Context, status vo.WorkerStatus) ([]*entity.WorkerEntity, error) {
	poList, err := r.workerDao.GetByStatus(ctx, status.String())
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetAvailableWorkers 获取可用的Worker（在线且未满负载）
func (r *workerRepositoryImpl) GetAvailableWorkers(ctx context.Context) ([]*entity.WorkerEntity, error) {
	poList, err := r.workerDao.GetAvailable(ctx)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// UpdateWorker 更新Worker
func (r *workerRepositoryImpl) UpdateWorker(ctx context.Context, worker *entity.WorkerEntity) error {
	po, err := r.convertor.EntityToPO(worker)
	if err != nil {
		return fmt.Errorf("failed to convert entity to po: %w", err)
	}
	
	return r.workerDao.Update(ctx, po)
}

// UpdateWorkerStatus 更新Worker状态
func (r *workerRepositoryImpl) UpdateWorkerStatus(ctx context.Context, workerID string, status vo.WorkerStatus) error {
	return r.workerDao.UpdateStatus(ctx, workerID, status.String())
}

// UpdateWorkerHeartbeat 更新Worker心跳
func (r *workerRepositoryImpl) UpdateWorkerHeartbeat(ctx context.Context, workerID string, heartbeat *vo.WorkerHeartbeat) error {
	heartbeatData := map[string]interface{}{
		"status":             heartbeat.Status.String(),
		"current_tasks":      heartbeat.CurrentTasks,
		"cpu_usage":          heartbeat.CPUUsage,
		"memory_usage":       heartbeat.MemoryUsage,
		"last_heartbeat_at":  heartbeat.LastHeartbeatAt,
	}
	
	return r.workerDao.UpdateHeartbeat(ctx, workerID, heartbeatData)
}

// DeleteWorker 删除Worker
func (r *workerRepositoryImpl) DeleteWorker(ctx context.Context, workerID string) error {
	return r.workerDao.Delete(ctx, workerID)
}

// GetUnhealthyWorkers 获取不健康的Worker
func (r *workerRepositoryImpl) GetUnhealthyWorkers(ctx context.Context, timeout time.Duration) ([]*entity.WorkerEntity, error) {
	poList, err := r.workerDao.GetUnhealthy(ctx, timeout)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetWorkerStatistics 获取Worker统计信息
func (r *workerRepositoryImpl) GetWorkerStatistics(ctx context.Context) (*repo.WorkerStatistics, error) {
	stats, err := r.workerDao.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}
	
	// 转换统计信息
	workerStats := &repo.WorkerStatistics{
		TotalWorkers:       getWorkerStatValue(stats, "total_workers"),
		OnlineWorkers:      getWorkerStatValue(stats, "online_workers"),
		OfflineWorkers:     getWorkerStatValue(stats, "offline_workers"),
		BusyWorkers:        getWorkerStatValue(stats, "busy_workers"),
		IdleWorkers:        getWorkerStatValue(stats, "idle_workers"),
		MaintenanceWorkers: getWorkerStatValue(stats, "maintenance_workers"),
		TotalTasks:         getWorkerStatValue(stats, "total_tasks"),
		AverageCPUUsage:    getWorkerStatFloatValue(stats, "average_cpu_usage"),
		AverageMemoryUsage: getWorkerStatFloatValue(stats, "average_memory_usage"),
		AverageLoadFactor:  getWorkerStatFloatValue(stats, "average_load_factor"),
	}
	
	return workerStats, nil
}

// GetBestWorkerForTask 获取最适合执行任务的Worker
func (r *workerRepositoryImpl) GetBestWorkerForTask(ctx context.Context) (*entity.WorkerEntity, error) {
	po, err := r.workerDao.GetBestWorkerForTask(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.convertor.POToEntity(po)
}

// getWorkerStatValue 获取Worker统计值
func getWorkerStatValue(stats map[string]interface{}, key string) int64 {
	if value, exists := stats[key]; exists {
		if intValue, ok := value.(int64); ok {
			return intValue
		}
		if floatValue, ok := value.(float64); ok {
			return int64(floatValue)
		}
	}
	return 0
}

// getWorkerStatFloatValue 获取Worker统计浮点值
func getWorkerStatFloatValue(stats map[string]interface{}, key string) float64 {
	if value, exists := stats[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
		if intValue, ok := value.(int64); ok {
			return float64(intValue)
		}
	}
	return 0.0
}