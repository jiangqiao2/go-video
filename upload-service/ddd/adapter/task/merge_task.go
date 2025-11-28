package task

import (
	"context"
	"sync"
	"time"

	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
	"upload-service/pkg/logger"
)

const (
	mergeQueueSize  = 128
	mergeJobTimeout = 60 * time.Minute
)

type mergeRequest struct {
	uploadVideoUUID string
}

var (
	mergeChan chan mergeRequest
	mergeOnce sync.Once
)

func StartMergeTask() {
	mergeOnce.Do(func() {
		mergeChan = make(chan mergeRequest, mergeQueueSize)
		go mergeWorker()
		logger.Infof("merge task worker started")
	})
}

func EnqueueMergeTask(uploadVideoUUID string) {
	if uploadVideoUUID == "" {
		return
	}
	if mergeChan == nil {
		StartMergeTask()
	}
	select {
	case mergeChan <- mergeRequest{uploadVideoUUID: uploadVideoUUID}:
	default:
		logger.Warnf("merge queue full upload_video_uuid=%s", uploadVideoUUID)
	}
}

func mergeWorker() {
	svc := rustfsInfra.DefaultRustFSService()
	repo := persistence.NewUploadVideoRepository()
	for req := range mergeChan {
		ctx, cancel := context.WithTimeout(context.Background(), mergeJobTimeout)
		entity, err := repo.QueryUploadVideoByUUID(ctx, req.uploadVideoUUID)
		if err != nil || entity == nil {
			cancel()
			continue
		}
		err = svc.MergeChunk(ctx, vo.NewMergeChunkVo(entity.StoragePath(), entity.ChunkStoragePath(), int64(entity.TotalChunks())))
		if err != nil {
			_ = repo.UpdateUploadVideoStatus(ctx, entity.UploadVideoUUID(), vo.UploadVideoStatusFailed)
			logger.Errorf("merge failed upload_video_uuid=%s error=%v", entity.UploadVideoUUID(), err)
			cancel()
			continue
		}
		if err := repo.UpdateUploadVideoStatus(ctx, entity.UploadVideoUUID(), vo.UploadVideoStatusSuccess); err != nil {
			logger.Errorf("update status failed upload_video_uuid=%s error=%v", entity.UploadVideoUUID(), err)
			cancel()
			continue
		}
		EnqueueChunkCleanup(entity.ChunkStoragePath(), int64(entity.TotalChunks()))
		logger.Infof("merge success upload_video_uuid=%s", entity.UploadVideoUUID())
		cancel()
	}
}
