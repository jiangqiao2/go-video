package persistence

import (
	"context"

	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/infrastructure/database/convertor"
	"upload-service/ddd/infrastructure/database/dao"
)

type videoRepositoryImpl struct {
	videoDao *dao.VideoDao
}

// NewVideoRepository builds a repository backed by MySQL dao.
func NewVideoRepository() repo.VideoRepository {
	return &videoRepositoryImpl{
		videoDao: dao.NewVideoDao(),
	}
}

func (r *videoRepositoryImpl) Create(ctx context.Context, video *entity.VideoEntity) error {
	return r.videoDao.Create(ctx, convertor.ToVideoPo(video))
}

func (r *videoRepositoryImpl) FindByUploadVideoUUID(ctx context.Context, uploadVideoUUID string) (*entity.VideoEntity, error) {
	po, err := r.videoDao.QueryByUploadVideoUUID(ctx, uploadVideoUUID)
	if err != nil {
		return nil, err
	}
	return convertor.ToVideoEntity(po), nil
}

func (r *videoRepositoryImpl) FindByVideoUUID(ctx context.Context, videoUUID string) (*entity.VideoEntity, error) {
	po, err := r.videoDao.QueryByVideoUUID(ctx, videoUUID)
	if err != nil {
		return nil, err
	}
	return convertor.ToVideoEntity(po), nil
}
