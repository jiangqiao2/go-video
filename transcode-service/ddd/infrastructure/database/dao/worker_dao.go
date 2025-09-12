package dao

import (
	"context"
	"time"
	"transcode-service/ddd/infrastructure/database/po"
	"gorm.io/gorm"
)

// WorkerDao Worker数据访问对象
type WorkerDao struct {
	db *gorm.DB
}

// NewWorkerDao 创建Worker DAO
func NewWorkerDao(db *gorm.DB) *WorkerDao {
	return &WorkerDao{
		db: db,
	}
}

// Create 创建Worker
func (d *WorkerDao) Create(ctx context.Context, worker *po.WorkerPO) error {
	return d.db.WithContext(ctx).Create(worker).Error
}

// GetByID 根据ID获取Worker
func (d *WorkerDao) GetByID(ctx context.Context, id uint) (*po.WorkerPO, error) {
	var worker po.WorkerPO
	err := d.db.WithContext(ctx).First(&worker, id).Error
	if err != nil {
		return nil, err
	}
	return &worker, nil
}

// GetByWorkerID 根据WorkerID获取Worker
func (d *WorkerDao) GetByWorkerID(ctx context.Context, workerID string) (*po.WorkerPO, error) {
	var worker po.WorkerPO
	err := d.db.WithContext(ctx).Where("worker_id = ?", workerID).First(&worker).Error
	if err != nil {
		return nil, err
	}
	return &worker, nil
}

// GetAll 获取所有Worker
func (d *WorkerDao) GetAll(ctx context.Context) ([]*po.WorkerPO, error) {
	var workers []*po.WorkerPO
	err := d.db.WithContext(ctx).Find(&workers).Error
	return workers, err
}

// GetByStatus 根据状态获取Worker列表
func (d *WorkerDao) GetByStatus(ctx context.Context, status string) ([]*po.WorkerPO, error) {
	var workers []*po.WorkerPO
	err := d.db.WithContext(ctx).
		Where("status = ?", status).
		Find(&workers).Error
	return workers, err
}

// GetAvailable 获取可用的Worker（在线且未满负载）
func (d *WorkerDao) GetAvailable(ctx context.Context) ([]*po.WorkerPO, error) {
	var workers []*po.WorkerPO
	err := d.db.WithContext(ctx).
		Where("status IN ? AND current_tasks < max_tasks", []string{"online", "idle"}).
		Order("current_tasks ASC, cpu_usage ASC").
		Find(&workers).Error
	return workers, err
}

// Update 更新Worker
func (d *WorkerDao) Update(ctx context.Context, worker *po.WorkerPO) error {
	worker.UpdatedAt = time.Now()
	return d.db.WithContext(ctx).Save(worker).Error
}

// UpdateStatus 更新Worker状态
func (d *WorkerDao) UpdateStatus(ctx context.Context, workerID, status string) error {
	return d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Where("worker_id = ?", workerID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateHeartbeat 更新Worker心跳
func (d *WorkerDao) UpdateHeartbeat(ctx context.Context, workerID string, heartbeatData map[string]interface{}) error {
	heartbeatData["last_heartbeat_at"] = time.Now()
	heartbeatData["updated_at"] = time.Now()
	
	return d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Where("worker_id = ?", workerID).
		Updates(heartbeatData).Error
}

// Delete 删除Worker
func (d *WorkerDao) Delete(ctx context.Context, workerID string) error {
	return d.db.WithContext(ctx).Where("worker_id = ?", workerID).Delete(&po.WorkerPO{}).Error
}

// GetUnhealthy 获取不健康的Worker
func (d *WorkerDao) GetUnhealthy(ctx context.Context, timeout time.Duration) ([]*po.WorkerPO, error) {
	var workers []*po.WorkerPO
	unhealthyTime := time.Now().Add(-timeout)
	err := d.db.WithContext(ctx).
		Where("status NOT IN ? AND last_heartbeat_at < ?", []string{"offline", "maintenance"}, unhealthyTime).
		Find(&workers).Error
	return workers, err
}

// GetStatistics 获取Worker统计信息
func (d *WorkerDao) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	
	var statusCounts []StatusCount
	err := d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	
	if err != nil {
		return nil, err
	}
	
	// 获取总体统计
	var totalStats struct {
		TotalWorkers       int64   `json:"total_workers"`
		TotalTasks         int64   `json:"total_tasks"`
		AverageCPUUsage    float64 `json:"average_cpu_usage"`
		AverageMemoryUsage float64 `json:"average_memory_usage"`
		AverageLoadFactor  float64 `json:"average_load_factor"`
	}
	
	err = d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Select(`
			COUNT(*) as total_workers,
			SUM(current_tasks) as total_tasks,
			AVG(cpu_usage) as average_cpu_usage,
			AVG(memory_usage) as average_memory_usage,
			AVG(CASE WHEN max_tasks > 0 THEN current_tasks * 1.0 / max_tasks ELSE 0 END) as average_load_factor
		`).
		Scan(&totalStats).Error
	
	if err != nil {
		return nil, err
	}
	
	// 组装结果
	stats := map[string]interface{}{
		"total_workers":        totalStats.TotalWorkers,
		"total_tasks":          totalStats.TotalTasks,
		"average_cpu_usage":    totalStats.AverageCPUUsage,
		"average_memory_usage": totalStats.AverageMemoryUsage,
		"average_load_factor":  totalStats.AverageLoadFactor,
	}
	
	// 添加状态统计
	for _, statusCount := range statusCounts {
		stats[statusCount.Status+"_workers"] = statusCount.Count
	}
	
	return stats, nil
}

// GetBestWorkerForTask 获取最适合执行任务的Worker
func (d *WorkerDao) GetBestWorkerForTask(ctx context.Context) (*po.WorkerPO, error) {
	var worker po.WorkerPO
	err := d.db.WithContext(ctx).
		Where("status IN ? AND current_tasks < max_tasks", []string{"online", "idle"}).
		Order("(current_tasks * 1.0 / max_tasks) ASC, cpu_usage ASC, memory_usage ASC").
		First(&worker).Error
	
	if err != nil {
		return nil, err
	}
	return &worker, nil
}

// IncrementTaskCount 增加Worker任务计数
func (d *WorkerDao) IncrementTaskCount(ctx context.Context, workerID string) error {
	return d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Where("worker_id = ?", workerID).
		Update("current_tasks", gorm.Expr("current_tasks + 1")).Error
}

// DecrementTaskCount 减少Worker任务计数
func (d *WorkerDao) DecrementTaskCount(ctx context.Context, workerID string) error {
	return d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Where("worker_id = ? AND current_tasks > 0", workerID).
		Update("current_tasks", gorm.Expr("current_tasks - 1")).Error
}

// BatchUpdateStatus 批量更新Worker状态
func (d *WorkerDao) BatchUpdateStatus(ctx context.Context, workerIDs []string, status string) error {
	if len(workerIDs) == 0 {
		return nil
	}
	
	return d.db.WithContext(ctx).
		Model(&po.WorkerPO{}).
		Where("worker_id IN ?", workerIDs).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}