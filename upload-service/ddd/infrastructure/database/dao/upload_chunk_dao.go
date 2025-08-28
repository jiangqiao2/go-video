package dao

import (
	"context"
	"gorm.io/gorm"
	"upload-service/ddd/infrastructure/database/po"
	"upload-service/internal/resource"
)

type UploadChunkDao struct {
	db *gorm.DB
}

func NewUploadChunkDao() *UploadChunkDao {
	return &UploadChunkDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *UploadChunkDao) QueryByUploadVideoUUIDAndStatus(ctx context.Context, uploadVideoUUID string, status string) ([]*po.UploadChunkPo, error) {
	var result []*po.UploadChunkPo
	err := d.db.Where("upload_video_uuid = ? AND status = ? AND is_deleted = 0", uploadVideoUUID).Find(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}
