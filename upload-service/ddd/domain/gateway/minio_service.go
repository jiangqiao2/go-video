package gateway

import (
	"context"
	"time"
	"upload-service/ddd/domain/vo"
)

type MinioService interface {
	GenerateStoragePath(ctx context.Context, genStoPathVo *vo.GenerateStoragePathVO) string
	GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string) string
	UploadChunk(ctx context.Context, minIoChunkVo *vo.MinIoUploadChunkVo) error
	MergeChunk(ctx context.Context, mergeChunkVo *vo.MergeChunkVo) error
	DeleteChunks(ctx context.Context, chunkStoragePath string, totalChunks int64) error
	GenerateImagePath(ctx context.Context, vo *vo.GenerateImagePathVO) string
	PresignPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
	PresignGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
	HeadObject(ctx context.Context, bucket, key string) (int64, error)
}
