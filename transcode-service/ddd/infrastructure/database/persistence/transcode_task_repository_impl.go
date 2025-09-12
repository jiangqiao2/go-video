package persistence

import (
	"context"
	"fmt"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/convertor"
	"transcode-service/ddd/infrastructure/database/dao"
	"gorm.io/gorm"
)

// transcodeTaskRepositoryImpl 转码任务仓储实现
type transcodeTaskRepositoryImpl struct {
	taskDao   *dao.TranscodeTaskDao
	convertor *convertor.TranscodeTaskConvertor
}

// NewTranscodeTaskRepository 创建转码任务仓储实现
func NewTranscodeTaskRepository(db *gorm.DB) repo.TranscodeTaskRepository {
	return &transcodeTaskRepositoryImpl{
		taskDao:   dao.NewTranscodeTaskDao(db),
		convertor: convertor.NewTranscodeTaskConvertor(),
	}
}

// CreateTask 创建任务
func (r *transcodeTaskRepositoryImpl) CreateTask(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	po, err := r.convertor.EntityToPO(task)
	if err != nil {
		return fmt.Errorf("failed to convert entity to po: %w", err)
	}
	
	return r.taskDao.Create(ctx, po)
}

// GetTaskByID 根据ID获取任务
func (r *transcodeTaskRepositoryImpl) GetTaskByID(ctx context.Context, taskID string) (*entity.TranscodeTaskEntity, error) {
	po, err := r.taskDao.GetByTaskID(ctx, taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return r.convertor.POToEntity(po)
}

// GetTasksByUserID 根据用户ID获取任务列表
func (r *transcodeTaskRepositoryImpl) GetTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetTasksByStatus 根据状态获取任务列表
func (r *transcodeTaskRepositoryImpl) GetTasksByStatus(ctx context.Context, status vo.TaskStatus, limit, offset int) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetByStatus(ctx, status.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetPendingTasks 获取待处理任务（按优先级排序）
func (r *transcodeTaskRepositoryImpl) GetPendingTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetPendingTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetTasksByWorkerID 根据Worker ID获取任务列表
func (r *transcodeTaskRepositoryImpl) GetTasksByWorkerID(ctx context.Context, workerID string) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetByWorkerID(ctx, workerID)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// UpdateTask 更新任务
func (r *transcodeTaskRepositoryImpl) UpdateTask(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	po, err := r.convertor.EntityToPO(task)
	if err != nil {
		return fmt.Errorf("failed to convert entity to po: %w", err)
	}
	
	return r.taskDao.Update(ctx, po)
}

// UpdateTaskStatus 更新任务状态
func (r *transcodeTaskRepositoryImpl) UpdateTaskStatus(ctx context.Context, taskID string, status vo.TaskStatus) error {
	return r.taskDao.UpdateStatus(ctx, taskID, status.String())
}

// UpdateTaskProgress 更新任务进度
func (r *transcodeTaskRepositoryImpl) UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error {
	return r.taskDao.UpdateProgress(ctx, taskID, progress)
}

// AssignTaskToWorker 分配任务给Worker
func (r *transcodeTaskRepositoryImpl) AssignTaskToWorker(ctx context.Context, taskID, workerID string) error {
	return r.taskDao.AssignToWorker(ctx, taskID, workerID)
}

// GetExpiredTasks 获取过期任务
func (r *transcodeTaskRepositoryImpl) GetExpiredTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetExpiredTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// GetFailedTasksForRetry 获取可重试的失败任务
func (r *transcodeTaskRepositoryImpl) GetFailedTasksForRetry(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error) {
	poList, err := r.taskDao.GetFailedTasksForRetry(ctx, limit)
	if err != nil {
		return nil, err
	}
	
	return r.convertor.POListToEntityList(poList)
}

// DeleteTask 删除任务
func (r *transcodeTaskRepositoryImpl) DeleteTask(ctx context.Context, taskID string) error {
	return r.taskDao.Delete(ctx, taskID)
}

// CountTasksByStatus 统计各状态任务数量
func (r *transcodeTaskRepositoryImpl) CountTasksByStatus(ctx context.Context, status vo.TaskStatus) (int64, error) {
	return r.taskDao.CountByStatus(ctx, status.String())
}

// CountTasksByUserID 统计用户任务数量
func (r *transcodeTaskRepositoryImpl) CountTasksByUserID(ctx context.Context, userID string) (int64, error) {
	return r.taskDao.CountByUserID(ctx, userID)
}

// GetTaskStatistics 获取任务统计信息
func (r *transcodeTaskRepositoryImpl) GetTaskStatistics(ctx context.Context) (*repo.TaskStatistics, error) {
	stats, err := r.taskDao.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}
	
	// 转换统计信息
	taskStats := &repo.TaskStatistics{
		TotalTasks:      getStatValue(stats, "total"),
		PendingTasks:    getStatValue(stats, "pending"),
		ProcessingTasks: getStatValue(stats, "processing"),
		CompletedTasks:  getStatValue(stats, "completed"),
		FailedTasks:     getStatValue(stats, "failed"),
		CancelledTasks:  getStatValue(stats, "cancelled"),
	}
	
	// 计算总任务数
	taskStats.TotalTasks = taskStats.PendingTasks + taskStats.ProcessingTasks + 
		taskStats.CompletedTasks + taskStats.FailedTasks + taskStats.CancelledTasks
	
	return taskStats, nil
}

// getStatValue 获取统计值
func getStatValue(stats map[string]int64, key string) int64 {
	if value, exists := stats[key]; exists {
		return value
	}
	return 0
}