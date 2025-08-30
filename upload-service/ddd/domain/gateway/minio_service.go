package gateway

import (
	"context"
	"upload-service/ddd/domain/vo"
)

type MinioService interface {
	GenerateStoragePath(ctx context.Context, genStoPathVo *vo.GenerateStoragePathVO) string
	GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string, chunkIndex int) string
	UploadChunk(ctx context.Context, minIoChunkVo *vo.MinIoUploadChunkVo) error
}
