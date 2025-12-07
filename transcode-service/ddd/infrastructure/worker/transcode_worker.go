package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/queue"
)

// TranscodeWorker 转码工作器接口
type TranscodeWorker interface {
	// Start 启动工作器
	Start(ctx context.Context) error

	// Stop 停止工作器
	Stop() error

	// IsRunning 检查工作器是否运行中
	IsRunning() bool

	// GetStats 获取工作器统计信息
	GetStats() WorkerStats
}

// WorkerStats 工作器统计信息
type WorkerStats struct {
	ProcessedTasks   uint64
	SuccessfulTasks  uint64
	FailedTasks      uint64
	CurrentlyRunning int
	StartTime        time.Time
	LastTaskTime     time.Time
}

// transcodeWorkerImpl 转码工作器实现
type transcodeWorkerImpl struct {
	id               string
	taskQueue        queue.TaskQueue
	transcodeService service.TranscodeService
	taskRepo         repo.TranscodeJobRepository
	workerCount      int
	running          bool
	cancel           context.CancelFunc
	stats            WorkerStats
	mu               sync.RWMutex
	wg               sync.WaitGroup
}

// NewTranscodeWorker 创建转码工作器
func NewTranscodeWorker(
	id string,
	taskQueue queue.TaskQueue,
	transcodeService service.TranscodeService,
	taskRepo repo.TranscodeJobRepository,
	workerCount int,
) TranscodeWorker {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &transcodeWorkerImpl{
		id:               id,
		taskQueue:        taskQueue,
		transcodeService: transcodeService,
		taskRepo:         taskRepo,
		workerCount:      workerCount,
		stats: WorkerStats{
			StartTime: time.Now(),
		},
	}
}

// Start 启动工作器
func (w *transcodeWorkerImpl) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker %s is already running", w.id)
	}

	// 创建可取消的上下文
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.running = true
	w.stats.StartTime = time.Now()

	log.Printf("Starting transcode worker %s with %d goroutines", w.id, w.workerCount)

	// 启动多个工作协程
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.workerLoop(workerCtx, i)
	}

	// 启动任务恢复协程 - 已禁用
	// w.wg.Add(1)
	// go w.taskRecoveryLoop(workerCtx)

	return nil
}

// Stop 停止工作器
func (w *transcodeWorkerImpl) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	log.Printf("Stopping transcode worker %s", w.id)

	// 取消上下文
	if w.cancel != nil {
		w.cancel()
	}

	// 等待所有协程结束
	w.wg.Wait()

	w.running = false
	log.Printf("Transcode worker %s stopped", w.id)

	return nil
}

// IsRunning 检查工作器是否运行中
func (w *transcodeWorkerImpl) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetStats 获取工作器统计信息
func (w *transcodeWorkerImpl) GetStats() WorkerStats {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stats
}

// workerLoop 工作器主循环
func (w *transcodeWorkerImpl) workerLoop(ctx context.Context, workerID int) {
	defer w.wg.Done()

	log.Printf("Worker %s-%d started", w.id, workerID)
	defer log.Printf("Worker %s-%d stopped", w.id, workerID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 从队列中获取任务
			task, err := w.taskQueue.Dequeue(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				log.Printf("Worker %s-%d failed to dequeue task: %v", w.id, workerID, err)
				time.Sleep(time.Second) // 避免忙等待
				continue
			}

			if task == nil {
				continue
			}

			// 处理任务
			w.processTask(ctx, task, workerID)
		}
	}
}

// processTask 处理单个任务
func (w *transcodeWorkerImpl) processTask(ctx context.Context, task *entity.TranscodeTaskEntity, workerID int) {
	log.Printf("Worker %s-%d processing task %s", w.id, workerID, task.TaskUUID())

	// Refresh latest state from repository to avoid stale entity after restart.
	if w.taskRepo != nil {
		if fresh, err := w.taskRepo.GetTranscodeJob(ctx, task.TaskUUID()); err == nil && fresh != nil {
			task = fresh
		}
	}
	if task.IsCompleted() || task.IsFailed() || task.IsCancelled() {
		log.Printf("Worker %s-%d skip terminal task %s status=%s", w.id, workerID, task.TaskUUID(), task.Status().String())
		return
	}

	// 更新统计信息
	w.updateStats(func(stats *WorkerStats) {
		stats.CurrentlyRunning++
		stats.LastTaskTime = time.Now()
	})

	defer func() {
		w.updateStats(func(stats *WorkerStats) {
			stats.CurrentlyRunning--
			stats.ProcessedTasks++
		})
	}()

	// 执行转码
	err := w.transcodeService.ExecuteTranscode(ctx, task)
	if err != nil {
		log.Printf("Worker %s-%d failed to process task %s: %v", w.id, workerID, task.TaskUUID(), err)
		w.updateStats(func(stats *WorkerStats) {
			stats.FailedTasks++
		})
	} else {
		log.Printf("Worker %s-%d successfully processed task %s", w.id, workerID, task.TaskUUID())
		w.updateStats(func(stats *WorkerStats) {
			stats.SuccessfulTasks++
		})
	}
}

// taskRecoveryLoop 任务恢复循环，处理异常中断的任务
func (w *transcodeWorkerImpl) taskRecoveryLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.recoverStuckTasks(ctx)
		}
	}
}

// recoverStuckTasks 恢复卡住的任务
func (w *transcodeWorkerImpl) recoverStuckTasks(ctx context.Context) {
	// 查找处理中但可能卡住的任务（处理时间超过1小时）
	log.Printf("Worker %s checking for stuck tasks", w.id)

	// 查找长时间处于processing状态的任务
	stuckTasks, err := w.taskRepo.QueryTranscodeJobsByStatus(ctx, vo.TaskStatusProcessing, 100)
	if err != nil {
		log.Printf("Worker %s failed to query stuck tasks: %v", w.id, err)
		return
	}

	// 过滤出真正卡住的任务（更新时间超过1小时）
	stuckThreshold := time.Now().Add(-time.Hour)
	for _, task := range stuckTasks {
		if task.UpdatedAt().After(stuckThreshold) {
			continue // 任务还在正常处理中
		}

		log.Printf("Worker %s recovering stuck task %s", w.id, task.TaskUUID())

		// 将任务重新设置为pending状态
		if err := w.taskRepo.UpdateTranscodeJobStatus(ctx, task.TaskUUID(), vo.TaskStatusPending, "", task.OutputPath(), 0); err != nil {
			log.Printf("Worker %s failed to reset stuck task %s: %v", w.id, task.TaskUUID(), err)
			continue
		}

		// 重新加入队列
		if err := w.taskQueue.Enqueue(ctx, task); err != nil {
			log.Printf("Worker %s failed to re-enqueue stuck task %s: %v", w.id, task.TaskUUID(), err)
			continue
		}

		log.Printf("Worker %s successfully recovered stuck task %s", w.id, task.TaskUUID())
	}
}

// updateStats 更新统计信息
func (w *transcodeWorkerImpl) updateStats(updateFunc func(*WorkerStats)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	updateFunc(&w.stats)
}

// WorkerManager 工作器管理器
type WorkerManager struct {
	workers []TranscodeWorker
	mu      sync.RWMutex
}

// NewWorkerManager 创建工作器管理器
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make([]TranscodeWorker, 0),
	}
}

// AddWorker 添加工作器
func (wm *WorkerManager) AddWorker(worker TranscodeWorker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workers = append(wm.workers, worker)
}

// StartAll 启动所有工作器
func (wm *WorkerManager) StartAll(ctx context.Context) error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, worker := range wm.workers {
		if err := worker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start worker: %w", err)
		}
	}

	return nil
}

// StopAll 停止所有工作器
func (wm *WorkerManager) StopAll() error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var errs []error
	for _, worker := range wm.workers {
		if err := worker.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop some workers: %v", errs)
	}

	return nil
}

// GetAllStats 获取所有工作器的统计信息
func (wm *WorkerManager) GetAllStats() map[string]WorkerStats {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := make(map[string]WorkerStats)
	for i, worker := range wm.workers {
		stats[fmt.Sprintf("worker-%d", i)] = worker.GetStats()
	}

	return stats
}
