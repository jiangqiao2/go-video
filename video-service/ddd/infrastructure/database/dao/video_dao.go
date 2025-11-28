package dao

import (
	"context"
	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
)

type VideoDao struct {
	db *gorm.DB
}

func NewVideoDao() *VideoDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &VideoDao{db: deps.DB}
}

func (d *VideoDao) Create(ctx context.Context, v *po.Video) error {
	return d.db.WithContext(ctx).Model(&po.Video{}).Create(v).Error
}

func (d *VideoDao) Save(ctx context.Context, v *po.Video) error {
	return d.db.WithContext(ctx).Model(&po.Video{}).Save(v).Error
}

func (d *VideoDao) QueryByUUID(ctx context.Context, videoUUID string) (*po.Video, error) {
	var video po.Video
	err := d.db.WithContext(ctx).Model(&po.Video{}).Where("video_uuid = ?", videoUUID).First(&video).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &video, nil
}

func (d *VideoDao) List(ctx context.Context, page, size int) ([]po.Video, int64, error) {
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
	if err := d.db.WithContext(ctx).Model(&po.Video{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []po.Video
	if err := d.db.WithContext(ctx).Model(&po.Video{}).Order("created_at DESC").Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
