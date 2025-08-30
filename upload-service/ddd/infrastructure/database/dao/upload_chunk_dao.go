package dao

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"upload-service/ddd/domain/repo"
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

func (d *UploadChunkDao) QueryUploadVideoByUUID(ctx context.Context, query *repo.UploadChunkCheckQuery) (*po.UploadChunkPo, error) {
	var result po.UploadChunkPo
	err := d.db.Model(&po.UploadVideoPo{}).Where("user_uuid = ? AND chunk_uuid AND upload_video_uuid AND chunk_index = ?",
		query.UserUUID, query.ChunkUUID, query.UploadVideoUUID, query.ChunkIndex).Find(result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryUploadVideoByUUID error: %v, uuid : %v", err, query.ChunkUUID)
		return nil, err
	}
	return &result, nil
}

func (d *UploadChunkDao) UpdateStatusByUUID(ctx context.Context, uuid string, status string) error {
	err := d.db.Model(&po.UploadChunkPo{}).Where("chunk_uuid = ? AND is_deleted = 0", uuid).Update("status", status).Error
	if err != nil {
		log.Errorf("UpdateStatusByUUID error: %v, uuid : %v", err, uuid)
		return err
	}
	return nil
}
