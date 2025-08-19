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
		db: resource.DefaultMysqlResource(),
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
