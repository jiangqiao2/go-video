package service

import (
	"context"
	"fmt"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
)

// TranscodeTaskService 转码任务领域服务
type TranscodeTaskService interface {
	// CreateTask 创建转码任务
	CreateTask(ctx context.Context, userID, sourceVideoPath, outputPath string, 
		config *vo.TranscodeConfig, priority int) (*entity.TranscodeTaskEntity, error)
	
	// AssignTaskToWorker 分配任务给Worker
	AssignTaskToWorker(ctx context.Context, taskID, workerID string) error
	
	// StartTask 开始执行任务
	StartTask(ctx context.Context, taskID string) error
	
	// UpdateTaskProgress 更新任务进度
	UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error
	
	// CompleteTask 完成任务
	CompleteTask(ctx context.Context, taskID string) error
	
	// FailTask 任务失败
	FailTask(ctx context.Context, taskID, errorMessage string) error
	
	// RetryTask 重试任务
	RetryTask(ctx context.Context, taskID string) error
	
	// CancelTask 取消任务
	CancelTask(ctx context.Context, taskID string) error
	
	// GetNextPendingTask 获取下一个待处理任务
	GetNextPendingTask(ctx context.Context) (*entity.TranscodeTaskEntity, error)
	
	// ValidateTaskTransition 验证任务状态转换
	ValidateTaskTransition(ctx context.Context, taskID string, targetStatus vo.TaskStatus) error
	
	// CleanupExpiredTasks 清理过期任务
	CleanupExpiredTasks(ctx context.Context) (int, error)
	
	// GetRetryableTasks 获取可重试的任务
	GetRetryableTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error)
}

type transcodeTaskServiceImpl struct {
	taskRepo repo.TranscodeTaskRepository
}

// NewTranscodeTaskService 创建转码任务领域服务
func NewTranscodeTaskService(taskRepo repo.TranscodeTaskRepository) TranscodeTaskService {
	return &transcodeTaskServiceImpl{
		taskRepo: taskRepo,
	}
}

// CreateTask 创建转码任务
func (s *transcodeTaskServiceImpl) CreateTask(ctx context.Context, userID, sourceVideoPath, outputPath string, 
	config *vo.TranscodeConfig, priority int) (*entity.TranscodeTaskEntity, error) {
	
	// 验证转码配置
	if !config.IsValid() {
		return nil, fmt.Errorf("invalid transcode config")
	}
	
	// 验证优先级
	if priority < 1 || priority > 10 {
		return nil, fmt.Errorf("priority must be between 1 and 10")
	}
	
	// 创建任务实体
	task := entity.NewTranscodeTaskEntity(
		userID,
		sourceVideoPath,
		outputPath,
		config,
		priority,
		3, // 默认最大重试次数
	)
	
	// 保存到仓储
	err := s.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	
	return task, nil
}

// AssignTaskToWorker 分配任务给Worker
func (s *transcodeTaskServiceImpl) AssignTaskToWorker(ctx context.Context, taskID, workerID string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 分配给Worker
	err = task.AssignToWorker(workerID)
	if err != nil {
		return fmt.Errorf("failed to assign task to worker: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// StartTask 开始执行任务
func (s *transcodeTaskServiceImpl) StartTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 开始处理
	err = task.StartProcessing()
	if err != nil {
		return fmt.Errorf("failed to start task processing: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// UpdateTaskProgress 更新任务进度
func (s *transcodeTaskServiceImpl) UpdateTaskProgress(ctx context.Context, taskID string, progress float64) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 更新进度
	err = task.UpdateProgress(progress)
	if err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// CompleteTask 完成任务
func (s *transcodeTaskServiceImpl) CompleteTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 完成任务
	err = task.Complete()
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// FailTask 任务失败
func (s *transcodeTaskServiceImpl) FailTask(ctx context.Context, taskID, errorMessage string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 任务失败
	err = task.Fail(errorMessage)
	if err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// RetryTask 重试任务
func (s *transcodeTaskServiceImpl) RetryTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 重试任务
	err = task.Retry()
	if err != nil {
		return fmt.Errorf("failed to retry task: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// CancelTask 取消任务
func (s *transcodeTaskServiceImpl) CancelTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// 取消任务
	err = task.Cancel()
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	
	// 更新任务
	err = s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	
	return nil
}

// GetNextPendingTask 获取下一个待处理任务
func (s *transcodeTaskServiceImpl) GetNextPendingTask(ctx context.Context) (*entity.TranscodeTaskEntity, error) {
	tasks, err := s.taskRepo.GetPendingTasks(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}
	
	if len(tasks) == 0 {
		return nil, nil // 没有待处理任务
	}
	
	return tasks[0], nil
}

// ValidateTaskTransition 验证任务状态转换
func (s *transcodeTaskServiceImpl) ValidateTaskTransition(ctx context.Context, taskID string, targetStatus vo.TaskStatus) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	if !task.Status().CanTransitionTo(targetStatus) {
		return fmt.Errorf("cannot transition from %s to %s", task.Status(), targetStatus)
	}
	
	return nil
}

// CleanupExpiredTasks 清理过期任务
func (s *transcodeTaskServiceImpl) CleanupExpiredTasks(ctx context.Context) (int, error) {
	expiredTasks, err := s.taskRepo.GetExpiredTasks(ctx, 100) // 一次最多处理100个
	if err != nil {
		return 0, fmt.Errorf("failed to get expired tasks: %w", err)
	}
	
	cleanedCount := 0
	for _, task := range expiredTasks {
		if task.IsExpired() {
			// 取消过期任务
			err = task.Cancel()
			if err != nil {
				continue // 跳过无法取消的任务
			}
			
			// 更新任务状态
			err = s.taskRepo.UpdateTask(ctx, task)
			if err != nil {
				continue // 跳过更新失败的任务
			}
			
			cleanedCount++
		}
	}
	
	return cleanedCount, nil
}

// GetRetryableTasks 获取可重试的任务
func (s *transcodeTaskServiceImpl) GetRetryableTasks(ctx context.Context, limit int) ([]*entity.TranscodeTaskEntity, error) {
	return s.taskRepo.GetFailedTasksForRetry(ctx, limit)
}