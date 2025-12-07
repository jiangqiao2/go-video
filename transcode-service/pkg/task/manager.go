package task

import (
	"context"
	"sync"
)

// BackgroundTask represents a long-running background process (consumer, worker, cron).
type BackgroundTask interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
}

type manager struct {
	tasks  []BackgroundTask
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	defaultManager = &manager{tasks: make([]BackgroundTask, 0)}
)

// Register adds a background task; should be called during init/assembly before StartAll.
func Register(task BackgroundTask) {
	if task == nil {
		return
	}
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()
	defaultManager.tasks = append(defaultManager.tasks, task)
}

// StartAll starts all registered tasks once.
func StartAll(ctx context.Context) error {
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()
	if defaultManager.cancel != nil {
		return nil
	}
	defaultManager.ctx, defaultManager.cancel = context.WithCancel(ctx)
	for _, t := range defaultManager.tasks {
		if t == nil {
			continue
		}
		if err := t.Start(defaultManager.ctx); err != nil {
			return err
		}
	}
	return nil
}

// StopAll stops all running tasks.
func StopAll() {
	defaultManager.mu.Lock()
	defer defaultManager.mu.Unlock()
	if defaultManager.cancel != nil {
		defaultManager.cancel()
	}
	for i := len(defaultManager.tasks) - 1; i >= 0; i-- {
		if t := defaultManager.tasks[i]; t != nil {
			_ = t.Stop()
		}
	}
	defaultManager.cancel = nil
}
