package repo

import "context"

// CommentLikeRepository manages likes on comments.
type CommentLikeRepository interface {
	Add(ctx context.Context, commentUUID, userUUID string) (bool, error)
	Remove(ctx context.Context, commentUUID, userUUID string) error
	CountByComment(ctx context.Context, commentUUID string) (int64, error)
	Exists(ctx context.Context, commentUUID, userUUID string) (bool, error)
}
