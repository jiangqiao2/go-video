package dao

import (
	"context"
	"fmt"
	"time"
	"transcode-service/ddd/infrastructure/database/po"
	"gorm.io/gorm"
)

// TranscodeTaskDao 转码任务数据访问对象
type TranscodeTaskDao struct {
	db *gorm.DB
}

// NewTranscodeTaskDao 创建转码任务DAO
func NewTranscodeTaskDao(db *gorm.DB) *TranscodeTaskDao {
	return &TranscodeTaskDao{
		db: db,
	}
}

// Create 创建任务
func (d *TranscodeTaskDao) Create(ctx context.Context, task *po.TranscodeTaskPO) error {
	return d.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取任务
func (d *TranscodeTaskDao) GetByID(ctx context.Context, id uint) (*po.TranscodeTaskPO, error) {
	var task po.TranscodeTaskPO
	err := d.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByTaskID 根据TaskID获取任务
func (d *TranscodeTaskDao) GetByTaskID(ctx context.Context, taskID string) (*po.TranscodeTaskPO, error) {
	var task po.TranscodeTaskPO
	err := d.db.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUserID 根据用户ID获取任务列表
func (d *TranscodeTaskDao) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

// GetByStatus 根据状态获取任务列表
func (d *TranscodeTaskDao) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	err := d.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

// GetPendingTasks 获取待处理任务（按优先级排序）
func (d *TranscodeTaskDao) GetPendingTasks(ctx context.Context, limit int) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	err := d.db.WithContext(ctx).
		Where("status IN ?", []string{"pending", "retrying"}).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// GetByWorkerID 根据Worker ID获取任务列表
func (d *TranscodeTaskDao) GetByWorkerID(ctx context.Context, workerID string) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	err := d.db.WithContext(ctx).
		Where("worker_id = ?", workerID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// Update 更新任务
func (d *TranscodeTaskDao) Update(ctx context.Context, task *po.TranscodeTaskPO) error {
	task.UpdatedAt = time.Now()
	return d.db.WithContext(ctx).Save(task).Error
}

// UpdateStatus 更新任务状态
func (d *TranscodeTaskDao) UpdateStatus(ctx context.Context, taskID, status string) error {
	return d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateProgress 更新任务进度
func (d *TranscodeTaskDao) UpdateProgress(ctx context.Context, taskID string, progress float64) error {
	return d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"progress":   progress,
			"updated_at": time.Now(),
		}).Error
}

// AssignToWorker 分配任务给Worker
func (d *TranscodeTaskDao) AssignToWorker(ctx context.Context, taskID, workerID string) error {
	return d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("task_id = ? AND status = ?", taskID, "pending").
		Updates(map[string]interface{}{
			"worker_id":  workerID,
			"status":     "assigned",
			"updated_at": time.Now(),
		}).Error
}

// GetExpiredTasks 获取过期任务
func (d *TranscodeTaskDao) GetExpiredTasks(ctx context.Context, limit int) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	expiredTime := time.Now().Add(-24 * time.Hour)
	err := d.db.WithContext(ctx).
		Where("status NOT IN ? AND created_at < ?", []string{"completed", "failed", "cancelled"}, expiredTime).
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// GetFailedTasksForRetry 获取可重试的失败任务
func (d *TranscodeTaskDao) GetFailedTasksForRetry(ctx context.Context, limit int) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	err := d.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retry_count", "failed").
		Order("created_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// Delete 删除任务
func (d *TranscodeTaskDao) Delete(ctx context.Context, taskID string) error {
	return d.db.WithContext(ctx).Where("task_id = ?", taskID).Delete(&po.TranscodeTaskPO{}).Error
}

// CountByStatus 统计各状态任务数量
func (d *TranscodeTaskDao) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

// CountByUserID 统计用户任务数量
func (d *TranscodeTaskDao) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetStatistics 获取任务统计信息
func (d *TranscodeTaskDao) GetStatistics(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	
	var results []StatusCount
	err := d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}
	
	return stats, nil
}

// BatchUpdateStatus 批量更新任务状态
func (d *TranscodeTaskDao) BatchUpdateStatus(ctx context.Context, taskIDs []string, status string) error {
	if len(taskIDs) == 0 {
		return nil
	}
	
	return d.db.WithContext(ctx).
		Model(&po.TranscodeTaskPO{}).
		Where("task_id IN ?", taskIDs).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// GetTasksWithConditions 根据条件查询任务
func (d *TranscodeTaskDao) GetTasksWithConditions(ctx context.Context, conditions map[string]interface{}, limit, offset int, orderBy string) ([]*po.TranscodeTaskPO, error) {
	var tasks []*po.TranscodeTaskPO
	query := d.db.WithContext(ctx)
	
	// 添加查询条件
	for key, value := range conditions {
		switch key {
		case "user_id", "status", "worker_id":
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		case "priority":
			query = query.Where("priority = ?", value)
		case "start_time":
			query = query.Where("created_at >= ?", value)
		case "end_time":
			query = query.Where("created_at <= ?", value)
		case "status_in":
			if statuses, ok := value.([]string); ok {
				query = query.Where("status IN ?", statuses)
			}
		}
	}
	
	// 排序
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at DESC")
	}
	
	// 分页
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&tasks).Error
	return tasks, err
}