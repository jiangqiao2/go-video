package repo

import (
	"context"
	goTime "time"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/vo"
	pkgtime "upload-service/pkg/time"
)

type UploadVideoRepository interface {
	CreateUploadVideoAndChunks(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity,
		uploadChunkEntitys []*entity.UploadChunkEntity) error
	QueryUploadVideoByName(ctx context.Context, userUUID, fileName string, fileSize int, fileHash string) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
	QueryUploadVideoByUUID(ctx context.Context, uploadVideoUUID string) (*entity.UploadVideoEntity, error)
	QueryUploadVideoByChunkUUID(ctx context.Context, query *UploadChunkCheckQuery) (*entity.UploadChunkEntity, error)
	QueryByUserAndUUID(ctx context.Context, uploadVideoUUID, userUUID string) (*entity.UploadVideoEntity, error)
	CountChunkByUploadVideoUUID(ctx context.Context, uploadVideoUUID, status string) (int64, error)
	QueryByStoragePath(ctx context.Context, userUUID, chunkUUID string) (string, error)
	UpdateUploadChunkStatus(ctx context.Context, uploadChunkUUID string, uploadChunkStatus vo.UploadChunkStatus) error
	UpdateUploadChunkPresign(ctx context.Context, uploadChunkUUID, putURL string, expiredAt goTime.Time) error
	UpdateUploadVideoStatus(ctx context.Context, uploadVideoUUID string, status vo.UploadVideoStatus) error
	MarkChunkCompleted(ctx context.Context, uploadChunkUUID, chunkHash string, chunkSize int) error
	QueryByUploadVideoFileHash(ctx context.Context, query *UploadVideoHashQuery) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
}

type UploadChunkCheckQuery struct {
	UserUUID        string
	ChunkUUID       string
	UploadVideoUUID string
	ChunkIndex      int
}

type UploadVideoHashQuery struct {
	UserUUID  string
	FileName  string
	FileSize  int
	FileHash  string
	StartTime pkgtime.Time
}
