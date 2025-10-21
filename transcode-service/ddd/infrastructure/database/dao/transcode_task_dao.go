package dao

import (
	"context"
	"gorm.io/gorm"
	"log"
	"transcode-service/ddd/infrastructure/database/po"
	"transcode-service/internal/resource"
)

// TranscodeTaskDAO 转码任务数据访问对象
type TranscodeTaskDAO struct {
	db *gorm.DB
}

// NewTranscodeTaskDAO 创建转码任务DAO实例
func NewTranscodeTaskDAO() *TranscodeTaskDAO {
	return &TranscodeTaskDAO{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

// Create 创建转码任务
func (d *TranscodeTaskDAO) Create(ctx context.Context, taskPo *po.TranscodeTask) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).Create(taskPo).Error
	if err != nil {
		log.Printf("Error creating transcode task %v", err)
		return err
	}
	return nil
}
