package gateway

import (
	"context"
	"upload-service/ddd/domain/vo"
)

type MinioService interface {
	GenerateStoragePath(ctx context.Context, genStoPathVo *vo.GenerateStoragePathVO) string
	GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string) string
	UploadChunk(ctx context.Context, minIoChunkVo *vo.MinIoUploadChunkVo) error
	MergeChunk(ctx context.Context, mergeChunkVo *vo.MergeChunkVo) error
	DeleteChunks(ctx context.Context, chunkStoragePath string, totalChunks int64) error
}
