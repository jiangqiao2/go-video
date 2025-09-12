package app

import (
	"context"
	"fmt"
	"transcode-service/ddd/application/dto"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/domain/vo"
)

// SchedulerApp 调度器应用服务接口
type SchedulerApp interface {
	// CreateTask 创建转码任务
	CreateTask(ctx context.Context, req *dto.CreateTranscodeTaskRequest) (*dto.CreateTranscodeTaskResponse, error)
	
	// GetTask 获取任务详情
	GetTask(ctx context.Context, req *dto.GetTaskRequest) (*dto.GetTaskResponse, error)
	
	// ListTasks 获取任务列表
	ListTasks(ctx context.Context, req *dto.ListTasksRequest) (*dto.ListTasksResponse, error)
	
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(ctx context.Context, req *dto.UpdateTaskStatusRequest) error
	
	// UpdateTaskProgress 更新任务进度
	UpdateTaskProgress(ctx context.Context, req *dto.UpdateTaskProgressRequest) error
	
	// CancelTask 取消任务
	CancelTask(ctx context.Context, req *dto.CancelTaskRequest) error
	
	// RetryTask 重试任务
	RetryTask(ctx context.Context, req *dto.RetryTaskRequest) error
	
	// GetTaskStatistics 获取任务统计
	GetTaskStatistics(ctx context.Context) (*dto.TaskStatisticsResponse, error)
	
	// BatchOperation 批量操作任务
	BatchOperation(ctx context.Context, req *dto.BatchOperationRequest) (*dto.BatchOperationResponse, error)
	
	// AssignTasks 分配任务给Worker
	AssignTasks(ctx context.Context) error
	
	// CleanupExpiredTasks 清理过期任务
	CleanupExpiredTasks(ctx context.Context) (int, error)
	
	// ProcessRetryTasks 处理重试任务
	ProcessRetryTasks(ctx context.Context) (int, error)
}

type schedulerAppImpl struct {
	taskService   service.TranscodeTaskService
	workerService service.WorkerService
}

// NewSchedulerApp 创建调度器应用服务
func NewSchedulerApp(
	taskService service.TranscodeTaskService,
	workerService service.WorkerService,
) SchedulerApp {
	return &schedulerAppImpl{
		taskService:   taskService,
		workerService: workerService,
	}
}

// CreateTask 创建转码任务
func (s *schedulerAppImpl) CreateTask(ctx context.Context, req *dto.CreateTranscodeTaskRequest) (*dto.CreateTranscodeTaskResponse, error) {
	// 设置默认优先级
	if req.Priority == 0 {
		req.Priority = 5
	}
	
	// 创建任务
	task, err := s.taskService.CreateTask(
		ctx,
		req.UserID,
		req.SourceVideoPath,
		req.OutputPath,
		req.Config,
		req.Priority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	
	// 设置元数据
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			task.SetMetadata(k, v)
		}
	}
	
	return &dto.CreateTranscodeTaskResponse{
		TaskID:    task.TaskID(),
		Status:    task.Status().String(),
		Message:   "Task created successfully",
		CreatedAt: dto.FormatTime(task.CreatedAt()),
	}, nil
}

// GetTask 获取任务详情
func (s *schedulerAppImpl) GetTask(ctx context.Context, req *dto.GetTaskRequest) (*dto.GetTaskResponse, error) {
	// 这里需要从仓储获取任务，简化实现
	// 实际项目中需要通过taskService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// ListTasks 获取任务列表
func (s *schedulerAppImpl) ListTasks(ctx context.Context, req *dto.ListTasksRequest) (*dto.ListTasksResponse, error) {
	// 这里需要从仓储获取任务列表，简化实现
	// 实际项目中需要通过taskService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// UpdateTaskStatus 更新任务状态
func (s *schedulerAppImpl) UpdateTaskStatus(ctx context.Context, req *dto.UpdateTaskStatusRequest) error {
	// 验证状态
	status := vo.TaskStatus(req.Status)
	if !status.IsValid() {
		return fmt.Errorf("invalid task status: %s", req.Status)
	}
	
	// 验证状态转换
	err := s.taskService.ValidateTaskTransition(ctx, req.TaskID, status)
	if err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}
	
	// 根据状态执行相应操作
	switch status {
	case vo.TaskStatusProcessing:
		return s.taskService.StartTask(ctx, req.TaskID)
	case vo.TaskStatusCompleted:
		return s.taskService.CompleteTask(ctx, req.TaskID)
	case vo.TaskStatusFailed:
		return s.taskService.FailTask(ctx, req.TaskID, "Task failed")
	case vo.TaskStatusCancelled:
		return s.taskService.CancelTask(ctx, req.TaskID)
	default:
		return fmt.Errorf("unsupported status update: %s", req.Status)
	}
}

// UpdateTaskProgress 更新任务进度
func (s *schedulerAppImpl) UpdateTaskProgress(ctx context.Context, req *dto.UpdateTaskProgressRequest) error {
	return s.taskService.UpdateTaskProgress(ctx, req.TaskID, req.Progress)
}

// CancelTask 取消任务
func (s *schedulerAppImpl) CancelTask(ctx context.Context, req *dto.CancelTaskRequest) error {
	return s.taskService.CancelTask(ctx, req.TaskID)
}

// RetryTask 重试任务
func (s *schedulerAppImpl) RetryTask(ctx context.Context, req *dto.RetryTaskRequest) error {
	return s.taskService.RetryTask(ctx, req.TaskID)
}

// GetTaskStatistics 获取任务统计
func (s *schedulerAppImpl) GetTaskStatistics(ctx context.Context) (*dto.TaskStatisticsResponse, error) {
	// 这里需要从仓储获取统计信息，简化实现
	// 实际项目中需要通过taskService或直接调用repository
	return nil, fmt.Errorf("not implemented")
}

// BatchOperation 批量操作任务
func (s *schedulerAppImpl) BatchOperation(ctx context.Context, req *dto.BatchOperationRequest) (*dto.BatchOperationResponse, error) {
	successCount := 0
	failedCount := 0
	var failedTasks []string
	
	for _, taskID := range req.TaskIDs {
		var err error
		
		switch req.Operation {
		case "cancel":
			err = s.taskService.CancelTask(ctx, taskID)
		case "retry":
			err = s.taskService.RetryTask(ctx, taskID)
		default:
			err = fmt.Errorf("unsupported operation: %s", req.Operation)
		}
		
		if err != nil {
			failedCount++
			failedTasks = append(failedTasks, taskID)
		} else {
			successCount++
		}
	}
	
	return &dto.BatchOperationResponse{
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedTasks:  failedTasks,
		Message:      fmt.Sprintf("Batch %s completed: %d success, %d failed", req.Operation, successCount, failedCount),
	}, nil
}

// AssignTasks 分配任务给Worker
func (s *schedulerAppImpl) AssignTasks(ctx context.Context) error {
	// 获取待处理任务
	pendingTask, err := s.taskService.GetNextPendingTask(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending task: %w", err)
	}
	
	if pendingTask == nil {
		// 没有待处理任务
		return nil
	}
	
	// 获取最佳Worker
	bestWorker, err := s.workerService.GetBestWorkerForTask(ctx)
	if err != nil {
		return fmt.Errorf("failed to get best worker: %w", err)
	}
	
	if bestWorker == nil {
		// No available worker found for task assignment
		return nil
	}
	
	// 分配任务给Worker
	err = s.taskService.AssignTaskToWorker(ctx, pendingTask.TaskID(), bestWorker.WorkerID())
	if err != nil {
		return fmt.Errorf("failed to assign task to worker: %w", err)
	}
	
	// 更新Worker任务计数
	err = s.workerService.AssignTaskToWorker(ctx, bestWorker.WorkerID())
	if err != nil {
		// Failed to update worker task count, but task assignment succeeded
		// 不返回错误，因为任务已经分配成功
	}
	
	// Task assigned to worker successfully
	
	return nil
}

// CleanupExpiredTasks 清理过期任务
func (s *schedulerAppImpl) CleanupExpiredTasks(ctx context.Context) (int, error) {
	cleanedCount, err := s.taskService.CleanupExpiredTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tasks: %w", err)
	}
	
	// Cleaned up expired tasks if any
	
	return cleanedCount, nil
}

// ProcessRetryTasks 处理重试任务
func (s *schedulerAppImpl) ProcessRetryTasks(ctx context.Context) (int, error) {
	retryableTasks, err := s.taskService.GetRetryableTasks(ctx, 10) // 一次最多处理10个
	if err != nil {
		return 0, fmt.Errorf("failed to get retryable tasks: %w", err)
	}
	
	processedCount := 0
	for _, task := range retryableTasks {
		err = s.taskService.RetryTask(ctx, task.TaskID())
		if err != nil {
			// Failed to retry task, continue with next
			continue
		}
		processedCount++
	}
	
	// Processed retry tasks if any
	
	return processedCount, nil
}