package dao

import (
	"context"
	"gorm.io/gorm"
	"transcode-service/ddd/domain/repo"
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

// NewUploadChunkDao 兼容旧方法名
func NewUploadChunkDao() *TranscodeTaskDAO {
	return NewTranscodeTaskDAO()
}

// Create 创建转码任务
func (d *TranscodeTaskDAO) Create(ctx context.Context, taskPo *po.TranscodeTask) error {
	err := d.db.WithContext(ctx).Create(taskPo).Error
	if err != nil {
		return err
	}
	return nil
}

// GetByTaskUUID 根据任务UUID查询转码任务
func (d *TranscodeTaskDAO) GetByTaskUUID(ctx context.Context, taskUUID string) (*po.TranscodeTask, error) {
	var task po.TranscodeTask
	err := d.db.WithContext(ctx).Where("task_uuid = ? AND is_deleted = 0", taskUUID).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// GetByVideoUUID 根据视频UUID查询转码任务列表
func (d *TranscodeTaskDAO) GetByVideoUUID(ctx context.Context, videoUUID string) ([]*po.TranscodeTask, error) {
	var tasks []*po.TranscodeTask
	err := d.db.WithContext(ctx).Where("video_uuid = ? AND is_deleted = 0", videoUUID).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByUserUUID 根据用户UUID查询转码任务列表（支持分页）
func (d *TranscodeTaskDAO) GetByUserUUID(ctx context.Context, userUUID string, query *repo.TranscodeTaskQuery) ([]*po.TranscodeTask, int64, error) {
	var tasks []*po.TranscodeTask
	var total int64

	db := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).Where("user_uuid = ? AND is_deleted = 0", userUUID)

	// 添加状态过滤
	if query.Status != nil {
		db = db.Where("status = ?", query.Status.String())
	}

	// 统计总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}

	err = db.Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetByStatus 根据状态查询转码任务列表
func (d *TranscodeTaskDAO) GetByStatus(ctx context.Context, status string) ([]*po.TranscodeTask, error) {
	var tasks []*po.TranscodeTask
	err := d.db.WithContext(ctx).Where("status = ? AND is_deleted = 0", status).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateStatus 更新转码任务状态
func (d *TranscodeTaskDAO) UpdateStatus(ctx context.Context, taskUUID string, status string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ? AND is_deleted = 0", taskUUID).
		Update("status", status).Error
	return err
}

// UpdateProgress 更新转码任务进度
func (d *TranscodeTaskDAO) UpdateProgress(ctx context.Context, taskUUID string, progress float64) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ? AND is_deleted = 0", taskUUID).
		Update("progress", progress).Error
	return err
}

// UpdateOutputPath 更新转码任务输出路径
func (d *TranscodeTaskDAO) UpdateOutputPath(ctx context.Context, taskUUID string, outputPath string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ? AND is_deleted = 0", taskUUID).
		Update("output_path", outputPath).Error
	return err
}

// UpdateError 更新转码任务错误信息
func (d *TranscodeTaskDAO) UpdateError(ctx context.Context, taskUUID string, errorMsg string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ? AND is_deleted = 0", taskUUID).
		Updates(map[string]interface{}{
			"message": errorMsg,
			"status":  "failed",
		}).Error
	return err
}

// Delete 删除转码任务（软删除）
func (d *TranscodeTaskDAO) Delete(ctx context.Context, taskUUID string) error {
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("task_uuid = ? AND is_deleted = 0", taskUUID).
		Update("is_deleted", 1).Error
	return err
}

// CountByStatus 根据状态统计转码任务数量
func (d *TranscodeTaskDAO) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.TranscodeTask{}).
		Where("status = ? AND is_deleted = 0", status).
		Count(&count).Error
	return count, err
}
