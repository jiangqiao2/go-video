package dao

import (
	"context"
	"log"

	"gorm.io/gorm"

	"transcode-service/ddd/infrastructure/database/po"
	"transcode-service/internal/resource"
)

type TranscodeJobDAO struct {
	db *gorm.DB
}

func NewTranscodeJobDAO() *TranscodeJobDAO {
	return &TranscodeJobDAO{db: resource.DefaultMysqlResource().MainDB()}
}

func (d *TranscodeJobDAO) Create(ctx context.Context, job *po.TranscodeJob) error {
	return d.db.WithContext(ctx).Model(&po.TranscodeJob{}).Create(job).Error
}

func (d *TranscodeJobDAO) UpdateProgress(ctx context.Context, jobUUID string, progress int) error {
	return d.db.WithContext(ctx).Model(&po.TranscodeJob{}).Where("job_uuid = ?", jobUUID).Update("progress", progress).Error
}

func (d *TranscodeJobDAO) UpdateJob(ctx context.Context, job *po.TranscodeJob) error {
	return d.db.WithContext(ctx).Model(&po.TranscodeJob{}).Where("job_uuid = ?", job.JobUUID).Updates(job).Error
}

func (d *TranscodeJobDAO) FindByJobUUID(ctx context.Context, jobUUID string) (*po.TranscodeJob, error) {
	var job po.TranscodeJob
	if err := d.db.WithContext(ctx).Where("job_uuid = ?", jobUUID).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (d *TranscodeJobDAO) UpdateStatus(ctx context.Context, jobUUID, status, message, outputPath string, progress int) error {
	update := map[string]interface{}{"status": status, "message": message, "progress": progress}
	if outputPath != "" {
		update["output_path"] = outputPath
	}
	return d.db.WithContext(ctx).Model(&po.TranscodeJob{}).Where("job_uuid = ?", jobUUID).Updates(update).Error
}

func (d *TranscodeJobDAO) QueryByStatus(ctx context.Context, status string, limit int) ([]*po.TranscodeJob, error) {
	var jobs []*po.TranscodeJob
	q := d.db.WithContext(ctx).Where("status = ?", status).Order("updated_at ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		log.Printf("Error query transcode jobs %v", err)
		return nil, err
	}
	return jobs, nil
}
