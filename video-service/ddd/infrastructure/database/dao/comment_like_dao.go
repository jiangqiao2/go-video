package dao

import (
	"context"
	"errors"

	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
)

type CommentLikeDao struct {
	db *gorm.DB
}

func NewCommentLikeDao() *CommentLikeDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &CommentLikeDao{db: deps.DB}
}

func (d *CommentLikeDao) Add(ctx context.Context, commentUUID, userUUID string) (bool, error) {
	like := &po.VideoCommentLike{
		CommentUUID: commentUUID,
		UserUUID:    userUUID,
		Status:      "Liked",
	}
	err := d.db.WithContext(ctx).Model(&po.VideoCommentLike{}).Create(like).Error
	if err != nil {
		// duplicate means already liked
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *CommentLikeDao) Remove(ctx context.Context, commentUUID, userUUID string) error {
	return d.db.WithContext(ctx).
		Where("comment_uuid = ? AND user_uuid = ?", commentUUID, userUUID).
		Delete(&po.VideoCommentLike{}).Error
}

func (d *CommentLikeDao) CountByComment(ctx context.Context, commentUUID string) (int64, error) {
	var total int64
	if err := d.db.WithContext(ctx).
		Model(&po.VideoCommentLike{}).
		Where("comment_uuid = ?", commentUUID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (d *CommentLikeDao) Exists(ctx context.Context, commentUUID, userUUID string) (bool, error) {
	var cnt int64
	if err := d.db.WithContext(ctx).
		Model(&po.VideoCommentLike{}).
		Where("comment_uuid = ? AND user_uuid = ?", commentUUID, userUUID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}
