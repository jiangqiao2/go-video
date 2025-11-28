package dao

import (
	"context"
	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LikeDao struct {
	db *gorm.DB
}

func NewLikeDao() *LikeDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &LikeDao{db: deps.DB}
}

func (d *LikeDao) Add(ctx context.Context, videoUUID, userUUID string) (bool, error) {
	like := &po.VideoLike{UserUUID: userUUID, VideoUUID: videoUUID, Status: "Liked"}
	tx := d.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(like)
	return tx.RowsAffected > 0, tx.Error
}

func (d *LikeDao) Remove(ctx context.Context, videoUUID, userUUID string) error {
	return d.db.WithContext(ctx).Where("video_uuid = ? AND user_uuid = ?", videoUUID, userUUID).Delete(&po.VideoLike{}).Error
}

func (d *LikeDao) CountByVideo(ctx context.Context, videoUUID string) (int64, error) {
	var cnt int64
	if err := d.db.WithContext(ctx).Model(&po.VideoLike{}).Where("video_uuid = ?", videoUUID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}
