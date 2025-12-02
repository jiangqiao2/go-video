package repo

import (
	"context"

	"video-service/ddd/domain/entity"
)

// CommentRepository manages video comments.
type CommentRepository interface {
	// root comments
	CreateRoot(ctx context.Context, comment *entity.Comment) error
	ListRootsByVideo(ctx context.Context, videoUUID string, sortBy string, page, size int) ([]*entity.Comment, int64, error)
	CountRootsByVideo(ctx context.Context, videoUUID string) (int64, error)
	FindRootByUUID(ctx context.Context, rootUUID string) (*entity.Comment, error)
	UpdateRootLikeCount(ctx context.Context, commentUUID string, likeCount int64) error
	IncrementRootReplyCount(ctx context.Context, rootUUID string, delta int64) error

	// replies
	CreateReply(ctx context.Context, comment *entity.Comment) error
	ListReplies(ctx context.Context, rootUUID string, parentUUID string, sortBy string, page, size int) ([]*entity.Comment, int64, error)
	FindReplyByUUID(ctx context.Context, commentUUID string) (*entity.Comment, error)
	UpdateReplyLikeCount(ctx context.Context, commentUUID string, likeCount int64) error
	IncrementReplyCount(ctx context.Context, commentUUID string, delta int64) error
}
