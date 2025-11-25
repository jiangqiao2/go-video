package persistence

import (
	"context"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/convertor"
	"transcode-service/ddd/infrastructure/database/dao"
)

type transcodeRepositoryImpl struct {
	jobDao    *dao.TranscodeJobDAO
	convertor *convertor.TranscodeTaskConvertor
}

func NewTranscodeRepository() repo.TranscodeJobRepository {
	return &transcodeRepositoryImpl{jobDao: dao.NewTranscodeJobDAO(), convertor: convertor.NewTranscodeTaskConvertor()}
}

func (t *transcodeRepositoryImpl) CreateTranscodeJob(ctx context.Context, job *entity.TranscodeTaskEntity) error {
	return t.jobDao.Create(ctx, t.convertor.ToPO(job))
}

func (t *transcodeRepositoryImpl) UpdateTranscodeJobProgress(ctx context.Context, jobUUID string, progress int) error {
	return t.jobDao.UpdateProgress(ctx, jobUUID, progress)
}

func (t *transcodeRepositoryImpl) UpdateTranscodeJob(ctx context.Context, job *entity.TranscodeTaskEntity) error {
	return t.jobDao.UpdateJob(ctx, t.convertor.ToPO(job))
}

func (t *transcodeRepositoryImpl) GetTranscodeJob(ctx context.Context, jobUUID string) (*entity.TranscodeTaskEntity, error) {
	jobPo, err := t.jobDao.FindByJobUUID(ctx, jobUUID)
	if err != nil {
		return nil, err
	}
	return t.convertor.ToEntity(jobPo), nil
}

func (t *transcodeRepositoryImpl) UpdateTranscodeJobStatus(ctx context.Context, jobUUID string, status vo.TaskStatus, message, outputPath string, progress int) error {
	return t.jobDao.UpdateStatus(ctx, jobUUID, status.String(), message, outputPath, progress)
}

func (t *transcodeRepositoryImpl) QueryTranscodeJobsByStatus(ctx context.Context, status vo.TaskStatus, limit int) ([]*entity.TranscodeTaskEntity, error) {
	jobs, err := t.jobDao.QueryByStatus(ctx, status.String(), limit)
	if err != nil {
		return nil, err
	}
	return t.convertor.ToEntities(jobs), nil
}
