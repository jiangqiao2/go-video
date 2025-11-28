package repo

import (
	"context"

	"video-service/ddd/domain/entity"
)

// VideoRepository defines persistence operations for videos.
type VideoRepository interface {
	Create(ctx context.Context, video *entity.Video) error
	Update(ctx context.Context, video *entity.Video) error
	FindByUUID(ctx context.Context, videoUUID string) (*entity.Video, error)
	List(ctx context.Context, page, size int) ([]*entity.Video, int64, error)
}
