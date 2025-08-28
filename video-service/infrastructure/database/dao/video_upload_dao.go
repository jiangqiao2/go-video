package dao

import (
	"context"
	"errors"
	"go-video/ddd/internal/resource"
	"go-video/ddd/video/infrastructure/database/po"
	"gorm.io/gorm"
)

type VideoUploadDao struct {
	db *gorm.DB
}

func NewVideoUploadDao() *VideoUploadDao {
	return &VideoUploadDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *VideoUploadDao) Create(ctx context.Context, videoUploadTaskPo *po.VideoUploadTaskPo) error {
	return d.db.WithContext(ctx).Create(videoUploadTaskPo).Error
}

func (d *VideoUploadDao) QueryByUUID(ctx context.Context, uuid string) (videoUploadTaskPo *po.VideoUploadTaskPo, err error) {
	videoUploadTaskPo = &po.VideoUploadTaskPo{}
	err = d.db.WithContext(ctx).First(videoUploadTaskPo, "uuid = ? AND is_deleted = 0", uuid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
	}
	return videoUploadTaskPo, err
}
