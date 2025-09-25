package queue

import (
	"context"
	"fmt"
	"sync"
	
	"transcode-service/ddd/domain/entity"
)

// TaskQueue 任务队列接口
type TaskQueue interface {
	// Enqueue 入队任务
	Enqueue(ctx context.Context, task *entity.TranscodeTask) error
	
	// Dequeue 出队任务（阻塞）
	Dequeue(ctx context.Context) (*entity.TranscodeTask, error)
	
	// TryDequeue 尝试出队任务（非阻塞）
	TryDequeue(ctx context.Context) (*entity.TranscodeTask, error)
	
	// Size 获取队列大小
	Size() int
	
	// IsEmpty 检查队列是否为空
	IsEmpty() bool
	
	// Close 关闭队列
	Close() error
	
	// IsClosed 检查队列是否已关闭
	IsClosed() bool
}

// MemoryTaskQueue 基于内存的任务队列实现
type MemoryTaskQueue struct {
	queue   chan *entity.TranscodeTask
	closed  bool
	mu      sync.RWMutex
	metrics *QueueMetrics
}

// QueueMetrics 队列指标
type QueueMetrics struct {
	EnqueueCount uint64
	DequeueCount uint64
	MaxSize      int
	CurrentSize  int
	mu           sync.RWMutex
}

// NewMemoryTaskQueue 创建内存任务队列
func NewMemoryTaskQueue(capacity int) TaskQueue {
	if capacity <= 0 {
		capacity = 1000 // 默认容量
	}
	
	return &MemoryTaskQueue{
		queue: make(chan *entity.TranscodeTask, capacity),
		metrics: &QueueMetrics{
			MaxSize: capacity,
		},
	}
}

// Enqueue 入队任务
func (q *MemoryTaskQueue) Enqueue(ctx context.Context, task *entity.TranscodeTask) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	if q.closed {
		return fmt.Errorf("queue is closed")
	}
	
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	select {
	case q.queue <- task:
		q.updateEnqueueMetrics()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("queue is full")
	}
}

// Dequeue 出队任务（阻塞）
func (q *MemoryTaskQueue) Dequeue(ctx context.Context) (*entity.TranscodeTask, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	if q.closed {
		return nil, fmt.Errorf("queue is closed")
	}
	
	select {
	case task := <-q.queue:
		q.updateDequeueMetrics()
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TryDequeue 尝试出队任务（非阻塞）
func (q *MemoryTaskQueue) TryDequeue(ctx context.Context) (*entity.TranscodeTask, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	if q.closed {
		return nil, fmt.Errorf("queue is closed")
	}
	
	select {
	case task := <-q.queue:
		q.updateDequeueMetrics()
		return task, nil
	default:
		return nil, nil // 队列为空
	}
}

// Size 获取队列大小
func (q *MemoryTaskQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	if q.closed {
		return 0
	}
	
	return len(q.queue)
}

// IsEmpty 检查队列是否为空
func (q *MemoryTaskQueue) IsEmpty() bool {
	return q.Size() == 0
}

// Close 关闭队列
func (q *MemoryTaskQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.closed {
		return nil
	}
	
	q.closed = true
	close(q.queue)
	return nil
}

// IsClosed 检查队列是否已关闭
func (q *MemoryTaskQueue) IsClosed() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.closed
}

// GetMetrics 获取队列指标
func (q *MemoryTaskQueue) GetMetrics() QueueMetrics {
	q.metrics.mu.RLock()
	defer q.metrics.mu.RUnlock()
	
	metrics := *q.metrics
	metrics.CurrentSize = q.Size()
	return metrics
}

// updateEnqueueMetrics 更新入队指标
func (q *MemoryTaskQueue) updateEnqueueMetrics() {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	q.metrics.EnqueueCount++
}

// updateDequeueMetrics 更新出队指标
func (q *MemoryTaskQueue) updateDequeueMetrics() {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	q.metrics.DequeueCount++
}

// PriorityTaskQueue 优先级任务队列（可扩展）
type PriorityTaskQueue struct {
	highPriorityQueue TaskQueue
	normalPriorityQueue TaskQueue
	lowPriorityQueue TaskQueue
	mu sync.RWMutex
}

// NewPriorityTaskQueue 创建优先级任务队列
func NewPriorityTaskQueue(capacity int) *PriorityTaskQueue {
	return &PriorityTaskQueue{
		highPriorityQueue:   NewMemoryTaskQueue(capacity / 3),
		normalPriorityQueue: NewMemoryTaskQueue(capacity / 3),
		lowPriorityQueue:    NewMemoryTaskQueue(capacity / 3),
	}
}

// EnqueueWithPriority 根据优先级入队
func (pq *PriorityTaskQueue) EnqueueWithPriority(ctx context.Context, task *entity.TranscodeTask, priority int) error {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	
	switch {
	case priority >= 8: // 高优先级
		return pq.highPriorityQueue.Enqueue(ctx, task)
	case priority >= 5: // 普通优先级
		return pq.normalPriorityQueue.Enqueue(ctx, task)
	default: // 低优先级
		return pq.lowPriorityQueue.Enqueue(ctx, task)
	}
}

// DequeueByPriority 按优先级出队
func (pq *PriorityTaskQueue) DequeueByPriority(ctx context.Context) (*entity.TranscodeTask, error) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	
	// 优先处理高优先级任务
	if task, err := pq.highPriorityQueue.TryDequeue(ctx); err == nil && task != nil {
		return task, nil
	}
	
	// 然后处理普通优先级任务
	if task, err := pq.normalPriorityQueue.TryDequeue(ctx); err == nil && task != nil {
		return task, nil
	}
	
	// 最后处理低优先级任务
	return pq.lowPriorityQueue.Dequeue(ctx)
}

// Close 关闭所有队列
func (pq *PriorityTaskQueue) Close() error {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	var errs []error
	if err := pq.highPriorityQueue.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := pq.normalPriorityQueue.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := pq.lowPriorityQueue.Close(); err != nil {
		errs = append(errs, err)
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("failed to close priority queues: %v", errs)
	}
	return nil
}