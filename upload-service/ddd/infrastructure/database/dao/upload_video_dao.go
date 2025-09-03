package dao

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"upload-service/ddd/infrastructure/database/po"
	"upload-service/internal/resource"
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

func (d *UploadVideoDao) BatchCreate(ctx context.Context, uploadVideoPo *po.UploadVideoPo, uploadChunks []*po.UploadChunkPo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.UploadVideoPo{}).Create(uploadVideoPo).Error; err != nil {
			log.Errorf("BatchCreate failed to insert upload_video_po: %v", err)
			return err
		}
		if err := tx.Model(&po.UploadChunkPo{}).Create(uploadChunks).Error; err != nil {
			log.Errorf("BatchCreate failed to insert upload_chunk_po: %v", err)
			return err
		}
		return nil
	})
}

func (d *UploadVideoDao) QueryByFileNameAndHash(ctx context.Context, userUUID string, fileName string, fileSize int, fileHash string) (*po.UploadVideoPo, error) {
	var uploadVideoPo po.UploadVideoPo
	err := d.db.Model(&po.UploadVideoPo{}).Where("user_uuid = ? AND file_name = ? AND file_hash = ? AND file_size = ? AND is_deleted = 0", userUUID, fileName, fileHash, fileSize).First(&uploadVideoPo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByFileNameAndHash failed to query upload_video_po: %v", err)
		return nil, err
	}
	return &uploadVideoPo, nil
}

func (d *UploadVideoDao) QueryByUUID(ctx context.Context, uploadVideoUUID string) (*po.UploadVideoPo, error) {
	var uploadVideoPo po.UploadVideoPo
	err := d.db.Model(&po.UploadVideoPo{}).Where("upload_video_uuid = ? AND is_deleted = 0", uploadVideoUUID).First(&uploadVideoPo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByUUID failed to query upload_video_po: %v uuid:%v", err, uploadVideoUUID)
		return nil, err
	}
	return &uploadVideoPo, nil
}

func (d *UploadVideoDao) QueryByUserUUIDAndUUID(ctx context.Context, uploadVideoUUID string, userUUID string) (*po.UploadVideoPo, error) {
	var uploadVideoPo po.UploadVideoPo
	err := d.db.Model(&po.UploadVideoPo{}).Where("upload_video_uuid = ? AND user_uuid = ? AND is_deleted = 0", uploadVideoUUID, userUUID).First(&uploadVideoPo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &uploadVideoPo, nil
}
