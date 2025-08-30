package repo

import (
	"context"
	"upload-service/ddd/domain/entity"
)

type UploadVideoRepository interface {
	CreateUploadVideoAndChunks(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity,
		uploadChunkEntitys []*entity.UploadChunkEntity) error

	QueryUploadVideoByName(ctx context.Context, userUUID, fileName string, fileSize int, fileHash string) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
	QueryUploadVideoByUUID(ctx context.Context, uploadVideoUUID string) (*entity.UploadVideoEntity, error)

	QueryUploadVideoByChunkUUID(ctx context.Context, query *UploadChunkCheckQuery) (*entity.UploadChunkEntity, error)
}

type UploadChunkCheckQuery struct {
	UserUUID        string
	ChunkUUID       string
	UploadVideoUUID string
	ChunkIndex      int
}
