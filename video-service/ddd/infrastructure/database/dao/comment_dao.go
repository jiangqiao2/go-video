package dao

import (
	"context"
	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
)

type CommentDao struct {
	db *gorm.DB
}

func NewCommentDao() *CommentDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &CommentDao{db: deps.DB}
}

func (d *CommentDao) Create(ctx context.Context, c *po.VideoComment) error {
	return d.db.WithContext(ctx).Model(&po.VideoComment{}).Create(c).Error
}

func (d *CommentDao) ListByVideo(ctx context.Context, videoUUID string, page, size int) ([]po.VideoComment, int64, error) {
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
	var total int64
	if err := d.db.WithContext(ctx).Model(&po.VideoComment{}).Where("video_uuid = ?", videoUUID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []po.VideoComment
	if err := d.db.WithContext(ctx).Model(&po.VideoComment{}).Where("video_uuid = ?", videoUUID).Order("created_at DESC").Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *CommentDao) CountByVideo(ctx context.Context, videoUUID string) (int64, error) {
	var total int64
	if err := d.db.WithContext(ctx).Model(&po.VideoComment{}).Where("video_uuid = ?", videoUUID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
