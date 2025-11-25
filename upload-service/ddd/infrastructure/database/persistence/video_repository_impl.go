package persistence

import (
	"context"
	"time"

	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
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

func (r *videoRepositoryImpl) UpdateVideoTranscodeInfo(ctx context.Context, videoUUID string, status vo.VideoStatus, videoURL string, transcodeTaskUUID string, errorMessage string, publishedAt *time.Time) error {
	values := map[string]interface{}{
		"status":              status.Value(),
		"video_url":           videoURL,
		"transcode_task_uuid": transcodeTaskUUID,
		"error_message":       errorMessage,
	}
	if publishedAt != nil {
		values["published_at"] = *publishedAt
	} else if status.IsPublished() {
		values["published_at"] = time.Now()
	}
	return r.videoDao.UpdateTranscodeInfo(ctx, videoUUID, values)
}

func (r *videoRepositoryImpl) ListByUser(ctx context.Context, userUUID string, status string, offset, limit int) ([]*entity.VideoEntity, int64, error) {
	poList, total, err := r.videoDao.QueryByUser(ctx, userUUID, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return convertor.ToVideoEntities(poList), total, nil
}

func (r *videoRepositoryImpl) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.VideoEntity, int64, error) {
	poList, total, err := r.videoDao.QueryByStatus(ctx, status, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return convertor.ToVideoEntities(poList), total, nil
}

func (r *videoRepositoryImpl) ListByUserQ(ctx context.Context, q *repo.VideoByUserQuery) ([]*entity.VideoEntity, int64, error) {
	page := q.Page
	size := q.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 12
	} else if size > 50 {
		size = 50
	}
	offset := (page - 1) * size
	poList, total, err := r.videoDao.QueryByUser(ctx, q.UserUUID, q.Status, offset, size)
	if err != nil {
		return nil, 0, err
	}
	return convertor.ToVideoEntities(poList), total, nil
}

func (r *videoRepositoryImpl) ListByStatusQ(ctx context.Context, q *repo.VideoByStatusQuery) ([]*entity.VideoEntity, int64, error) {
	page := q.Page
	size := q.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 12
	} else if size > 50 {
		size = 50
	}
	offset := (page - 1) * size
	poList, total, err := r.videoDao.QueryByStatus(ctx, q.Status, offset, size)
	if err != nil {
		return nil, 0, err
	}
	return convertor.ToVideoEntities(poList), total, nil
}
