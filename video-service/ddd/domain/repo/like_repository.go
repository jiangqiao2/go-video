package repo

import "context"

// LikeRepository manages likes.
type LikeRepository interface {
	Add(ctx context.Context, videoUUID, userUUID string) (bool, error)
	Remove(ctx context.Context, videoUUID, userUUID string) error
	CountByVideo(ctx context.Context, videoUUID string) (int64, error)
}
