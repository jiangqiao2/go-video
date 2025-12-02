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
	rootDao  *dao.CommentRootDao
	replyDao *dao.CommentReplyDao
}

func NewCommentRepository(db *gorm.DB) *CommentRepository { // compatibility
	return &CommentRepository{
		rootDao:  dao.NewCommentRootDao(),
		replyDao: dao.NewCommentReplyDao(),
	}
}

func DefaultCommentRepository() *CommentRepository {
	return &CommentRepository{
		rootDao:  dao.NewCommentRootDao(),
		replyDao: dao.NewCommentReplyDao(),
	}
}

var _ repo.CommentRepository = (*CommentRepository)(nil)

func (r *CommentRepository) CreateRoot(ctx context.Context, comment *entity.Comment) error {
	return r.rootDao.Create(ctx, convertor.ToCommentRootPo(comment))
}

func (r *CommentRepository) CreateReply(ctx context.Context, comment *entity.Comment) error {
	parentPath := comment.Path
	return r.replyDao.Create(ctx, convertor.ToCommentReplyPo(comment), parentPath)
}

func (r *CommentRepository) ListRootsByVideo(ctx context.Context, videoUUID string, sortBy string, page, size int) ([]*entity.Comment, int64, error) {
	list, total, err := r.rootDao.ListByVideo(ctx, videoUUID, sortBy, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Comment, 0, len(list))
	for i := range list {
		res = append(res, convertor.ToCommentRootEntity(&list[i]))
	}
	return res, total, nil
}

func (r *CommentRepository) CountRootsByVideo(ctx context.Context, videoUUID string) (int64, error) {
	return r.rootDao.CountByVideo(ctx, videoUUID)
}

func (r *CommentRepository) FindRootByUUID(ctx context.Context, rootUUID string) (*entity.Comment, error) {
	poComment, err := r.rootDao.FindByUUID(ctx, rootUUID)
	if err != nil || poComment == nil {
		return nil, err
	}
	return convertor.ToCommentRootEntity(poComment), nil
}

func (r *CommentRepository) UpdateRootLikeCount(ctx context.Context, commentUUID string, likeCount int64) error {
	return r.rootDao.UpdateLikeCount(ctx, commentUUID, likeCount)
}

func (r *CommentRepository) IncrementRootReplyCount(ctx context.Context, commentUUID string, delta int64) error {
	return r.rootDao.IncrementReplyCount(ctx, commentUUID, delta)
}

func (r *CommentRepository) ListReplies(ctx context.Context, rootUUID string, parentUUID string, sortBy string, page, size int) ([]*entity.Comment, int64, error) {
	list, total, err := r.replyDao.ListReplies(ctx, rootUUID, parentUUID, sortBy, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Comment, 0, len(list))
	for i := range list {
		res = append(res, convertor.ToCommentReplyEntity(&list[i]))
	}
	return res, total, nil
}

func (r *CommentRepository) FindReplyByUUID(ctx context.Context, commentUUID string) (*entity.Comment, error) {
	poComment, err := r.replyDao.FindByUUID(ctx, commentUUID)
	if err != nil || poComment == nil {
		return nil, err
	}
	return convertor.ToCommentReplyEntity(poComment), nil
}

func (r *CommentRepository) UpdateReplyLikeCount(ctx context.Context, commentUUID string, likeCount int64) error {
	return r.replyDao.UpdateLikeCount(ctx, commentUUID, likeCount)
}

func (r *CommentRepository) IncrementReplyCount(ctx context.Context, commentUUID string, delta int64) error {
	return r.replyDao.IncrementReplyCount(ctx, commentUUID, delta)
}
