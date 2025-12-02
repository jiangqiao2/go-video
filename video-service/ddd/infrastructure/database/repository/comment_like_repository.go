package repository

import (
	"context"

	"video-service/ddd/domain/repo"
	"video-service/ddd/infrastructure/database/dao"

	"gorm.io/gorm"
)

type CommentLikeRepository struct {
	dao *dao.CommentLikeDao
}

func NewCommentLikeRepository(db *gorm.DB) *CommentLikeRepository { // compatibility
	return &CommentLikeRepository{dao: dao.NewCommentLikeDao()}
}

func DefaultCommentLikeRepository() *CommentLikeRepository {
	return &CommentLikeRepository{dao: dao.NewCommentLikeDao()}
}

var _ repo.CommentLikeRepository = (*CommentLikeRepository)(nil)

func (r *CommentLikeRepository) Add(ctx context.Context, commentUUID, userUUID string) (bool, error) {
	return r.dao.Add(ctx, commentUUID, userUUID)
}

func (r *CommentLikeRepository) Remove(ctx context.Context, commentUUID, userUUID string) error {
	return r.dao.Remove(ctx, commentUUID, userUUID)
}

func (r *CommentLikeRepository) CountByComment(ctx context.Context, commentUUID string) (int64, error) {
	return r.dao.CountByComment(ctx, commentUUID)
}

func (r *CommentLikeRepository) Exists(ctx context.Context, commentUUID, userUUID string) (bool, error) {
	return r.dao.Exists(ctx, commentUUID, userUUID)
}
