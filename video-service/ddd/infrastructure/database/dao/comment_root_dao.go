package dao

import (
	"context"
	"errors"

	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommentRootDao struct {
	db *gorm.DB
}

func NewCommentRootDao() *CommentRootDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &CommentRootDao{db: deps.DB}
}

func (d *CommentRootDao) Create(ctx context.Context, c *po.VideoCommentRoot) error {
	return d.db.WithContext(ctx).Model(&po.VideoCommentRoot{}).Create(c).Error
}

func (d *CommentRootDao) ListByVideo(ctx context.Context, videoUUID string, sortBy string, page, size int) ([]po.VideoCommentRoot, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	offset := (page - 1) * size
	query := d.db.WithContext(ctx).Model(&po.VideoCommentRoot{}).Where("video_uuid = ? AND is_deleted = 0", videoUUID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if sortBy == "hot" {
		query = query.Order(clause.Expr{SQL: "like_count + reply_count DESC"}).Order("created_at DESC")
	} else {
		query = query.Order("created_at DESC")
	}
	var list []po.VideoCommentRoot
	if err := query.Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *CommentRootDao) CountByVideo(ctx context.Context, videoUUID string) (int64, error) {
	var total int64
	if err := d.db.WithContext(ctx).Model(&po.VideoCommentRoot{}).Where("video_uuid = ? AND is_deleted = 0", videoUUID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (d *CommentRootDao) FindByUUID(ctx context.Context, rootUUID string) (*po.VideoCommentRoot, error) {
	var c po.VideoCommentRoot
	if err := d.db.WithContext(ctx).Where("root_uuid = ? AND is_deleted = 0", rootUUID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (d *CommentRootDao) UpdateLikeCount(ctx context.Context, rootUUID string, likeCount int64) error {
	return d.db.WithContext(ctx).Model(&po.VideoCommentRoot{}).Where("root_uuid = ? AND is_deleted = 0", rootUUID).
		Update("like_count", likeCount).Error
}

func (d *CommentRootDao) IncrementReplyCount(ctx context.Context, rootUUID string, delta int64) error {
	return d.db.WithContext(ctx).Model(&po.VideoCommentRoot{}).Where("root_uuid = ? AND is_deleted = 0", rootUUID).
		UpdateColumn("reply_count", gorm.Expr("GREATEST(reply_count + ?, 0)", delta)).Error
}
