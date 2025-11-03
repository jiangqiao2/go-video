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
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).Where("task_uuid = ?", task.TaskUUID).Updates(task).Error
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

// UpdateHLSProgress 更新HLS进度
func (d *TranscodeTaskDAO) UpdateHLSProgress(ctx context.Context, taskUUID string, progress int) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Update("hls_progress", progress).Error
	if err != nil {
		log.Printf("Error updating HLS progress %v", err)
		return err
	}
	return nil
}

// UpdateHLSStatus 更新HLS状态
func (d *TranscodeTaskDAO) UpdateHLSStatus(ctx context.Context, taskUUID, status string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Update("hls_status", status).Error
	if err != nil {
		log.Printf("Error updating HLS status %v", err)
		return err
	}
	return nil
}

// UpdateHLSOutputPath 更新HLS输出路径
func (d *TranscodeTaskDAO) UpdateHLSOutputPath(ctx context.Context, taskUUID, outputPath string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Update("hls_output_path", outputPath).Error
	if err != nil {
		log.Printf("Error updating HLS output path %v", err)
		return err
	}
	return nil
}

// UpdateHLSError 更新HLS错误信息
func (d *TranscodeTaskDAO) UpdateHLSError(ctx context.Context, taskUUID, errorMessage string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Update("hls_error_message", errorMessage).Error
	if err != nil {
		log.Printf("Error updating HLS error message %v", err)
		return err
	}
	return nil
}

// UpdateHLSCompleted 标记HLS完成
func (d *TranscodeTaskDAO) UpdateHLSCompleted(ctx context.Context, taskUUID string) error {
	updates := map[string]interface{}{
		"hls_status":       "completed",
		"hls_progress":     100,
		"hls_completed_at": "NOW()",
	}
	
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Updates(updates).Error
	if err != nil {
		log.Printf("Error updating HLS completed %v", err)
		return err
	}
	return nil
}

// UpdateHLSFailed 标记HLS失败
func (d *TranscodeTaskDAO) UpdateHLSFailed(ctx context.Context, taskUUID, errorMessage string) error {
	updates := map[string]interface{}{
		"hls_status":        "failed",
		"hls_error_message": errorMessage,
		"hls_completed_at":  "NOW()",
	}
	
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ?", taskUUID).
		Updates(updates).Error
	if err != nil {
		log.Printf("Error updating HLS failed %v", err)
		return err
	}
	return nil
}

// QueryHLSEnabledTasks 查询启用HLS的任务
func (d *TranscodeTaskDAO) QueryHLSEnabledTasks(ctx context.Context, status string, limit int) ([]*po.TranscodeTask, error) {
	var tasks []*po.TranscodeTask
	query := d.db.WithContext(ctx).
		Where("hls_enabled = ? AND hls_status = ?", true, status).
		Order("updated_at ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&tasks).Error; err != nil {
		log.Printf("Error query HLS enabled tasks %v", err)
		return nil, err
	}
	return tasks, nil
}
