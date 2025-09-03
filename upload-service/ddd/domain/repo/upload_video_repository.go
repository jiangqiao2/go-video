package repo

import (
	"context"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/vo"
)

type UploadVideoRepository interface {
	CreateUploadVideoAndChunks(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity,
		uploadChunkEntitys []*entity.UploadChunkEntity) error
	QueryUploadVideoByName(ctx context.Context, userUUID, fileName string, fileSize int, fileHash string) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
	QueryUploadVideoByUUID(ctx context.Context, uploadVideoUUID string) (*entity.UploadVideoEntity, error)
	QueryUploadVideoByChunkUUID(ctx context.Context, query *UploadChunkCheckQuery) (*entity.UploadChunkEntity, error)
	QueryByUserAndUUID(ctx context.Context, uploadVideoUUID, userUUID string) (*entity.UploadVideoEntity, error)
	CountChunkByUploadVideoUUID(ctx context.Context, uploadVideoUUID, status string) (int64, error)

	UpdateUploadChunkStatus(ctx context.Context, uploadChunkUUID string, uploadChunkStatus vo.UploadChunkStatus) error
}

type UploadChunkCheckQuery struct {
	UserUUID        string
	ChunkUUID       string
	UploadVideoUUID string
	ChunkIndex      int
}
