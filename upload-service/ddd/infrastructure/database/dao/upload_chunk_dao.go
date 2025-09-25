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
	err := d.db.Model(&po.UploadChunkPo{}).Where("upload_video_uuid = ? AND status = ? AND is_deleted = 0", uploadVideoUUID).Find(result).Error
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

func (d *UploadChunkDao) CountChunk(ctx context.Context, uploadVideoUUID, status string) (int64, error) {
	var count int64
	err := d.db.Model(&po.UploadChunkPo{}).Where("upload_video_uuid = ? AND status = ? AND is_deleted = 0", uploadVideoUUID, status).Count(&count).Error
	if err != nil {
		log.Errorf("CountChunk error: %v, uuid : %v", err, uploadVideoUUID)
		return 0, err
	}
	return count, nil
}

func (d *UploadChunkDao) QueryStoragePathByUUID(ctx context.Context, userUUID, chunkUUID, status string) (string, error) {
	var res string
	err := d.db.Model(&po.UploadChunkPo{}).Select("storage_path").Where("user_uuid = ? AND chunk_uuid = ? AND status = ? AND is_deleted = 0", userUUID, chunkUUID, status).Scan(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
	}
	return res, nil
}
