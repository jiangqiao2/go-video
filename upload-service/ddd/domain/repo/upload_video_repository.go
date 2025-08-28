package repo

import (
	"context"
	"upload-service/ddd/domain/entity"
)

type UploadVideoRepository interface {
	CreateUploadVideoAndChunks(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity,
		uploadChunkEntitys []*entity.UploadChunkEntity) error

	QueryUploadVideoByName(ctx context.Context, fileName string, fileSize int, fileHash string) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
}
