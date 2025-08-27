package dao

import (
	"context"
	"go-video/ddd/internal/resource"
	"go-video/ddd/upload/infrastructure/database/po"
	"gorm.io/gorm"
)

type UploadVideoDao struct {
	db *gorm.DB
}

func NewUploadVideoDao() *UploadVideoDao {
	return &UploadVideoDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *UploadVideoDao) Create(ctx context.Context, uploadVideoPo *po.UploadVideoPo) error {
	return d.db.Model(&po.UploadVideoPo{}).Create(uploadVideoPo).Error
}

func (d *UploadVideoDao) UpdateStatusByUUID(ctx context.Context, uploadVideoUUID string, status string) error {
	return d.db.Model(&po.UploadVideoPo{}).Where("upload_video_uuid = ?", uploadVideoUUID).Update("status", status).Error
}
