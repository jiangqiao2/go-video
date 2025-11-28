package repository

import (
	"context"
	"video-service/ddd/domain/entity"
	"video-service/ddd/domain/repo"
	"video-service/ddd/infrastructure/database/convertor"
	"video-service/ddd/infrastructure/database/dao"

	"gorm.io/gorm"
)

type CommentRepository struct {
	commentDao *dao.CommentDao
}

func NewCommentRepository(db *gorm.DB) *CommentRepository { // compatibility
	return &CommentRepository{commentDao: dao.NewCommentDao()}
}

func DefaultCommentRepository() *CommentRepository {
	return &CommentRepository{commentDao: dao.NewCommentDao()}
}

var _ repo.CommentRepository = (*CommentRepository)(nil)

func (r *CommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	return r.commentDao.Create(ctx, convertor.ToCommentPo(comment))
}

func (r *CommentRepository) ListByVideo(ctx context.Context, videoUUID string, page, size int) ([]*entity.Comment, int64, error) {
	list, total, err := r.commentDao.ListByVideo(ctx, videoUUID, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Comment, 0, len(list))
	for i := range list {
		res = append(res, convertor.ToCommentEntity(&list[i]))
	}
	return res, total, nil
}

func (r *CommentRepository) CountByVideo(ctx context.Context, videoUUID string) (int64, error) {
	return r.commentDao.CountByVideo(ctx, videoUUID)
}
