package dao

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"upload-service/ddd/infrastructure/database/po"
	"upload-service/internal/resource"
)

// VideoDao encapsulates CRUD operations for video_publish table.
type VideoDao struct {
	db *gorm.DB
}

// NewVideoDao creates a dao backed by the default mysql resource.
func NewVideoDao() *VideoDao {
	return &VideoDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *VideoDao) Create(ctx context.Context, video *po.VideoPo) error {
	return d.db.WithContext(ctx).Model(&po.VideoPo{}).Create(video).Error
}

func (d *VideoDao) QueryByUploadVideoUUID(ctx context.Context, uploadVideoUUID string) (*po.VideoPo, error) {
	var video po.VideoPo
	err := d.db.WithContext(ctx).
		Model(&po.VideoPo{}).
		Where("upload_video_uuid = ? AND is_deleted = 0", uploadVideoUUID).
		First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByUploadVideoUUID failed: %v", err)
		return nil, err
	}
	return &video, nil
}

func (d *VideoDao) QueryByVideoUUID(ctx context.Context, videoUUID string) (*po.VideoPo, error) {
	var video po.VideoPo
	err := d.db.WithContext(ctx).
		Model(&po.VideoPo{}).
		Where("video_uuid = ? AND is_deleted = 0", videoUUID).
		First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByVideoUUID failed: %v", err)
		return nil, err
	}
	return &video, nil
}
