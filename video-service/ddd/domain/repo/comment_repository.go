package repo

import (
	"context"

	"video-service/ddd/domain/entity"
)

// CommentRepository manages video comments.
type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	ListByVideo(ctx context.Context, videoUUID string, page, size int) ([]*entity.Comment, int64, error)
	CountByVideo(ctx context.Context, videoUUID string) (int64, error)
}
