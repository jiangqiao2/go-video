package repo

import (
	"context"

	"upload-service/ddd/domain/entity"
)

// VideoRepository persists published video metadata.
type VideoRepository interface {
	Create(ctx context.Context, video *entity.VideoEntity) error
	FindByUploadVideoUUID(ctx context.Context, uploadVideoUUID string) (*entity.VideoEntity, error)
	FindByVideoUUID(ctx context.Context, videoUUID string) (*entity.VideoEntity, error)
}
