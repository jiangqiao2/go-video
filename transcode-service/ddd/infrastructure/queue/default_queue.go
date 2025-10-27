package queue

import (
	"sync"

	"transcode-service/pkg/config"
)

var (
	queueOnce    sync.Once
	defaultQueue TaskQueue
)

// DefaultTaskQueue 获取默认任务队列
func DefaultTaskQueue() TaskQueue {
	queueOnce.Do(func() {
		capacity := 100
		if cfg := config.GetGlobalConfig(); cfg != nil {
			if cfg.Worker.QueueCapacity > 0 {
				capacity = cfg.Worker.QueueCapacity
			}
		}
		defaultQueue = NewMemoryTaskQueue(capacity)
	})
	return defaultQueue
}

// CloseDefaultTaskQueue 关闭默认任务队列
func CloseDefaultTaskQueue() {
	if defaultQueue != nil {
		_ = defaultQueue.Close()
	}
}
