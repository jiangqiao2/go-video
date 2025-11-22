package dao

import (
    "context"
    "gorm.io/gorm"
    "transcode-service/ddd/infrastructure/database/po"
    "transcode-service/internal/resource"
)

type HLSJobDAO struct { db *gorm.DB }

func NewHLSJobDAO() *HLSJobDAO { return &HLSJobDAO{db: resource.DefaultMysqlResource().MainDB()} }

func (d *HLSJobDAO) Create(ctx context.Context, job *po.HLSJob) error {
    return d.db.WithContext(ctx).Model(&po.HLSJob{}).Create(job).Error
}

func (d *HLSJobDAO) UpdateProgress(ctx context.Context, jobUUID string, progress int) error {
    return d.db.WithContext(ctx).Model(&po.HLSJob{}).Where("job_uuid = ?", jobUUID).Update("progress", progress).Error
}

func (d *HLSJobDAO) UpdateStatus(ctx context.Context, jobUUID, status string) error {
    return d.db.WithContext(ctx).Model(&po.HLSJob{}).Where("job_uuid = ?", jobUUID).Update("status", status).Error
}

func (d *HLSJobDAO) UpdateOutput(ctx context.Context, jobUUID, master string) error {
    return d.db.WithContext(ctx).Model(&po.HLSJob{}).Where("job_uuid = ?", jobUUID).Update("master_playlist", master).Error
}

func (d *HLSJobDAO) UpdateError(ctx context.Context, jobUUID, msg string) error {
    return d.db.WithContext(ctx).Model(&po.HLSJob{}).Where("job_uuid = ?", jobUUID).Update("error_message", msg).Error
}

func (d *HLSJobDAO) FindByJobUUID(ctx context.Context, jobUUID string) (*po.HLSJob, error) {
    var job po.HLSJob
    if err := d.db.WithContext(ctx).Where("job_uuid = ?", jobUUID).First(&job).Error; err != nil { return nil, err }
    return &job, nil
}

func (d *HLSJobDAO) QueryByStatus(ctx context.Context, status string, limit int) ([]*po.HLSJob, error) {
    var jobs []*po.HLSJob
    q := d.db.WithContext(ctx).Where("status = ?", status).Order("updated_at ASC")
    if limit > 0 { q = q.Limit(limit) }
    if err := q.Find(&jobs).Error; err != nil { return nil, err }
    return jobs, nil
}