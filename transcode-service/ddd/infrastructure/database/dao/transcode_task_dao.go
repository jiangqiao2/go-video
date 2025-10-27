package dao

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"

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

func (d *TranscodeTaskDAO) UpdateTranscodeTaskProgress(ctx context.Context, taskUUID string, progress int) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).Where("task_uuid = ?", taskUUID).Update("progress", progress).Error
	if err != nil {
		log.Printf("Error updating transcode task %v", err)
		return err
	}
	return nil
}

func (d *TranscodeTaskDAO) UpdateTranscodeTask(ctx context.Context, task *po.TranscodeTask) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).Save(task).Error
	if err != nil {
		log.Printf("Error updating transcode task %v", err)
		return err
	}
	return nil
}

// FindByTaskUUID 根据任务UUID查询任务
func (d *TranscodeTaskDAO) FindByTaskUUID(ctx context.Context, taskUUID string) (*po.TranscodeTask, error) {
	var task po.TranscodeTask
	if err := d.db.WithContext(ctx).
		Where("task_uuid = ?", taskUUID).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		log.Printf("Error query transcode task by uuid %v", err)
		return nil, err
	}
	return &task, nil
}

// UpdateTranscodeTaskStatus 更新任务状态、输出路径、错误信息等
func (d *TranscodeTaskDAO) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, message, outputPath string, progress int) error {
	update := map[string]interface{}{
		"status":   status,
		"message":  message,
		"progress": progress,
	}
	if outputPath != "" {
		update["output_path"] = outputPath
	}

	if err := d.db.WithContext(ctx).
		Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Updates(update).Error; err != nil {
		log.Printf("Error updating transcode task status %v", err)
		return err
	}
	return nil
}

// QueryTranscodeTasksByStatus 根据状态查询任务
func (d *TranscodeTaskDAO) QueryTranscodeTasksByStatus(ctx context.Context, status string, limit int) ([]*po.TranscodeTask, error) {
	var tasks []*po.TranscodeTask
	query := d.db.WithContext(ctx).Where("status = ?", status).Order("updated_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&tasks).Error; err != nil {
		log.Printf("Error query transcode tasks by status %v", err)
		return nil, err
	}
	return tasks, nil
}
