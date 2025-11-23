package task

import (
	"context"
	"sync"
	"time"

	"upload-service/ddd/domain/gateway"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
	"upload-service/pkg/logger"
)

const (
	cleanupQueueSize     = 128
	cleanupBatchSize     = 16
	cleanupFlushInterval = 3 * time.Second
	cleanupJobTimeout    = 2 * time.Minute
)

type cleanupRequest struct {
	chunkPath   string
	totalChunks int64
}

var (
	cleanupChan chan cleanupRequest
	startOnce   sync.Once
	minioSvc    gateway.MinioService
)

// StartChunkCleanupTask initializes the cleanup worker and must be called once during service startup.
func StartChunkCleanupTask() {
	startOnce.Do(func() {
		minioSvc = rustfsInfra.DefaultRustFSService()
		cleanupChan = make(chan cleanupRequest, cleanupQueueSize)
		go cleanupWorker()
		logger.Info("chunk cleanup task worker started")
	})
}

// EnqueueChunkCleanup pushes a cleanup task into the worker queue.
func EnqueueChunkCleanup(chunkPath string, totalChunks int64) {
	if chunkPath == "" || totalChunks <= 0 {
		return
	}

	StartChunkCleanupTask()

	req := cleanupRequest{
		chunkPath:   chunkPath,
		totalChunks: totalChunks,
	}

	select {
	case cleanupChan <- req:
	default:
		// Queue is full, fallback to asynchronous immediate processing to avoid blocking.
		go processBatch([]cleanupRequest{req})
	}
}

func cleanupWorker() {
	ticker := time.NewTicker(cleanupFlushInterval)
	defer ticker.Stop()

	batch := make([]cleanupRequest, 0, cleanupBatchSize)
	for {
		select {
		case req := <-cleanupChan:
			batch = append(batch, req)
			if len(batch) >= cleanupBatchSize {
				processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func processBatch(batch []cleanupRequest) {
	if len(batch) == 0 {
		return
	}

	// Deduplicate by chunk storage path, keep the largest totalChunks value.
	dedup := make(map[string]int64, len(batch))
	for _, req := range batch {
		if req.chunkPath == "" || req.totalChunks <= 0 {
			continue
		}
		if existing, ok := dedup[req.chunkPath]; !ok || req.totalChunks > existing {
			dedup[req.chunkPath] = req.totalChunks
		}
	}

	for path, total := range dedup {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupJobTimeout)
		err := minioSvc.DeleteChunks(ctx, path, total)
		cancel()

		if err != nil {
			logger.Errorf("chunk cleanup failed %v", map[string]interface{}{
				"path":        path,
				"totalChunks": total,
				"error":       err,
			})
			continue
		}

		logger.Info("chunk cleanup success", map[string]interface{}{
			"path":        path,
			"totalChunks": total,
		})
	}
}
