package repo

import (
	"context"
	"time"

	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/vo"
)

// VideoRepository persists published video metadata.
type VideoRepository interface {
	Create(ctx context.Context, video *entity.VideoEntity) error
	FindByUploadVideoUUID(ctx context.Context, uploadVideoUUID string) (*entity.VideoEntity, error)
	FindByVideoUUID(ctx context.Context, videoUUID string) (*entity.VideoEntity, error)
	UpdateVideoTranscodeInfo(ctx context.Context, videoUUID string, status vo.VideoStatus, videoURL string, transcodeTaskUUID string, errorMessage string, publishedAt *time.Time) error
	ListByUser(ctx context.Context, userUUID string, status string, offset, limit int) ([]*entity.VideoEntity, int64, error)
}
