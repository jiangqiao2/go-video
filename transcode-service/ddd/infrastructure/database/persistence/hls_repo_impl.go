package persistence

import (
	"context"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/infrastructure/database/convertor"
	"transcode-service/ddd/infrastructure/database/dao"
)

type hlsRepositoryImpl struct {
	dao *dao.HLSJobDAO
	cvt *convertor.HLSJobConvertor
}

func NewHLSRepository() repo.HLSJobRepository {
	return &hlsRepositoryImpl{dao: dao.NewHLSJobDAO(), cvt: convertor.NewHLSJobConvertor()}
}

func (r *hlsRepositoryImpl) CreateHLSJob(ctx context.Context, job *entity.HLSJobEntity) error {
	return r.dao.Create(ctx, r.cvt.ToPO(job))
}

func (r *hlsRepositoryImpl) UpdateHLSJobProgress(ctx context.Context, jobUUID string, progress int) error {
	return r.dao.UpdateProgress(ctx, jobUUID, progress)
}

func (r *hlsRepositoryImpl) UpdateHLSJobStatus(ctx context.Context, jobUUID string, status string) error {
	return r.dao.UpdateStatus(ctx, jobUUID, status)
}

func (r *hlsRepositoryImpl) UpdateHLSJobOutput(ctx context.Context, jobUUID string, masterPlaylist string) error {
	return r.dao.UpdateOutput(ctx, jobUUID, masterPlaylist)
}

func (r *hlsRepositoryImpl) UpdateHLSJobError(ctx context.Context, jobUUID string, errorMessage string) error {
	return r.dao.UpdateError(ctx, jobUUID, errorMessage)
}

func (r *hlsRepositoryImpl) GetHLSJob(ctx context.Context, jobUUID string) (*entity.HLSJobEntity, error) {
	jobPo, err := r.dao.FindByJobUUID(ctx, jobUUID)
	if err != nil {
		return nil, err
	}
	return r.cvt.ToEntity(jobPo), nil
}

func (r *hlsRepositoryImpl) QueryHLSJobsByStatus(ctx context.Context, status string, limit int) ([]*entity.HLSJobEntity, error) {
	pos, err := r.dao.QueryByStatus(ctx, status, limit)
	if err != nil {
		return nil, err
	}
	entities := make([]*entity.HLSJobEntity, 0, len(pos))
	for _, p := range pos {
		entities = append(entities, r.cvt.ToEntity(p))
	}
	return entities, nil
}
