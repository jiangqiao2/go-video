package queue

import (
	"context"
	"fmt"
	"sync"

	"transcode-service/ddd/domain/entity"
	"transcode-service/pkg/config"
)

type HLSJobQueue interface {
	Enqueue(ctx context.Context, job *entity.HLSJobEntity) error
	Dequeue(ctx context.Context) (*entity.HLSJobEntity, error)
	TryDequeue(ctx context.Context) (*entity.HLSJobEntity, error)
	Size() int
	IsEmpty() bool
	Close() error
	IsClosed() bool
}

type memoryHLSJobQueue struct {
	queue  chan *entity.HLSJobEntity
	closed bool
	mu     sync.RWMutex
}

func NewMemoryHLSJobQueue(capacity int) HLSJobQueue {
	if capacity <= 0 {
		capacity = 100
	}
	return &memoryHLSJobQueue{queue: make(chan *entity.HLSJobEntity, capacity)}
}

func (q *memoryHLSJobQueue) Enqueue(ctx context.Context, job *entity.HLSJobEntity) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return fmt.Errorf("queue is closed")
	}
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	select {
	case q.queue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("queue is full")
	}
}

func (q *memoryHLSJobQueue) Dequeue(ctx context.Context) (*entity.HLSJobEntity, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return nil, fmt.Errorf("queue is closed")
	}
	select {
	case job := <-q.queue:
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (q *memoryHLSJobQueue) TryDequeue(ctx context.Context) (*entity.HLSJobEntity, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return nil, fmt.Errorf("queue is closed")
	}
	select {
	case job := <-q.queue:
		return job, nil
	default:
		return nil, nil
	}
}

func (q *memoryHLSJobQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return 0
	}
	return len(q.queue)
}

func (q *memoryHLSJobQueue) IsEmpty() bool { return q.Size() == 0 }

func (q *memoryHLSJobQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true
	close(q.queue)
	return nil
}

func (q *memoryHLSJobQueue) IsClosed() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.closed
}

var (
	hlsQueueOnce    sync.Once
	defaultHLSQueue HLSJobQueue
)

func DefaultHLSJobQueue() HLSJobQueue {
	hlsQueueOnce.Do(func() {
		capacity := 100
		if cfg := config.GetGlobalConfig(); cfg != nil && cfg.Worker.QueueCapacity > 0 {
			capacity = cfg.Worker.QueueCapacity
		}
		defaultHLSQueue = NewMemoryHLSJobQueue(capacity)
	})
	return defaultHLSQueue
}

func CloseDefaultHLSJobQueue() {
	if defaultHLSQueue != nil {
		_ = defaultHLSQueue.Close()
	}
}
