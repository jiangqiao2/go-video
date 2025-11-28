package repository

import (
	"context"
	"video-service/ddd/domain/repo"
	"video-service/ddd/infrastructure/database/dao"

	"gorm.io/gorm"
)

type LikeRepository struct {
	likeDao *dao.LikeDao
}

func NewLikeRepository(db *gorm.DB) *LikeRepository { // compatibility
	return &LikeRepository{likeDao: dao.NewLikeDao()}
}

func DefaultLikeRepository() *LikeRepository {
	return &LikeRepository{likeDao: dao.NewLikeDao()}
}

var _ repo.LikeRepository = (*LikeRepository)(nil)

func (r *LikeRepository) Add(ctx context.Context, videoUUID, userUUID string) (bool, error) {
	return r.likeDao.Add(ctx, videoUUID, userUUID)
}

func (r *LikeRepository) Remove(ctx context.Context, videoUUID, userUUID string) error {
	return r.likeDao.Remove(ctx, videoUUID, userUUID)
}

func (r *LikeRepository) CountByVideo(ctx context.Context, videoUUID string) (int64, error) {
	return r.likeDao.CountByVideo(ctx, videoUUID)
}
