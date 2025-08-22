package dao

import (
	"context"
	"errors"
	"go-video/ddd/internal/resource"
	"go-video/ddd/video/infrastructure/database/po"
	"gorm.io/gorm"
)

type VideoDao struct {
	db *gorm.DB
}

func NewVideoDao() *VideoDao {
	return &VideoDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

// Create 创建视频
func (v *VideoDao) Create(ctx context.Context, videoPo *po.VideoPo) error {
	return v.db.WithContext(ctx).Create(videoPo).Error
}

func (v *VideoDao) GetByUUID(ctx context.Context, uuid string) (*po.VideoPo, error) {
	var videoPo po.VideoPo
	err := v.db.WithContext(ctx).First(&videoPo, "uuid = ? AND is_deleted = 0", uuid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &videoPo, nil
}
func (v *VideoDao) GetByUserUUID(ctx context.Context, userUUID string) ([]*po.VideoPo, error) {
	var videoPos []*po.VideoPo
	err := v.db.WithContext(ctx).Find(&videoPos, "user_uuid = ? AND is_deleted = 0", userUUID).Error
	if err != nil {
		return nil, err
	}
	return videoPos, nil
}

// CreateVideoAndTask 通过事务插入两条记录，保证原子性
func (d *VideoDao) CreateVideoAndTask(ctx context.Context, video *po.VideoPo, videoUploadTaskPo *po.VideoUploadTaskPo) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(video).Error; err != nil {
			return err
		}
		if err := tx.Create(videoUploadTaskPo).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *VideoDao) UpdateVideoStatus(ctx context.Context, videoUUID string, videoStatus string, videoUploadTaskUUID string, videoTaskStatus string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.VideoPo{}).Where("uuid = ? AND is_deleted = 0 ", videoUUID).Update("status", videoStatus).Error; err != nil {
			return err
		}
		if err := tx.Model(&po.VideoUploadTaskPo{}).Where("uuid = ? AND is_deleted = 0 ", videoUploadTaskUUID).Update("status", videoTaskStatus).Error; err != nil {
			return err
		}
		return nil
	})
}
