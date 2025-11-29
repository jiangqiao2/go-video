package repository

import (
	"context"
	"video-service/ddd/domain/entity"
	"video-service/ddd/domain/repo"
	"video-service/ddd/infrastructure/database/convertor"
	"video-service/ddd/infrastructure/database/dao"
	"video-service/pkg/logger"

	"gorm.io/gorm"
)

type VideoRepository struct {
	videoDao *dao.VideoDao
}

func NewVideoRepository(db *gorm.DB) *VideoRepository { // compatibility, ignored db
	return &VideoRepository{videoDao: dao.NewVideoDao()}
}

func DefaultVideoRepository() *VideoRepository {
	return &VideoRepository{videoDao: dao.NewVideoDao()}
}

var _ repo.VideoRepository = (*VideoRepository)(nil)

func (r *VideoRepository) Create(ctx context.Context, video *entity.Video) error {
	return r.videoDao.Create(ctx, convertor.ToVideoPo(video))
}

func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	po := convertor.ToVideoPo(video)
	logger.Infof("VideoRepository.Update video_uuid=%s id=%d status=%s url=%s", video.VideoUUID, video.ID, video.Status, video.VideoURL)
	return r.videoDao.Save(ctx, po)
}

func (r *VideoRepository) FindByUUID(ctx context.Context, videoUUID string) (*entity.Video, error) {
	po, err := r.videoDao.QueryByUUID(ctx, videoUUID)
	if err != nil || po == nil {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return convertor.ToVideoEntity(po), nil
}

func (r *VideoRepository) List(ctx context.Context, page, size int) ([]*entity.Video, int64, error) {
	list, total, err := r.videoDao.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Video, 0, len(list))
	for i := range list {
		res = append(res, convertor.ToVideoEntity(&list[i]))
	}
	return res, total, nil
}

func (r *VideoRepository) ListByUserStatus(ctx context.Context, userUUID string, status string, page, size int) ([]*entity.Video, int64, error) {
	list, total, err := r.videoDao.ListByUserStatus(ctx, userUUID, status, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Video, 0, len(list))
	for i := range list {
		res = append(res, convertor.ToVideoEntity(&list[i]))
	}
	return res, total, nil
}
