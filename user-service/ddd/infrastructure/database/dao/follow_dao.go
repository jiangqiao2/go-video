package dao

import (
	"context"
	"errors"
	"time"
	"user-service/ddd/infrastructure/database/po"
	"user-service/internal/resource"
	"user-service/pkg/logger"

	"gorm.io/gorm"
)

type FollowDao struct {
	db *gorm.DB
}

func NewFollowDao() *FollowDao {
	return &FollowDao{db: resource.DefaultMysqlResource().MainDB()}
}

func (d *FollowDao) Create(ctx context.Context, follow *po.FollowPo) error {
	return d.db.WithContext(ctx).Create(follow).Error
}

func (d *FollowDao) SoftDelete(ctx context.Context, userUUID, targetUUID string) error {
	now := time.Now()
	return d.db.WithContext(ctx).Model(&po.FollowPo{}).
		Where("user_uuid = ? AND target_uuid = ? AND deleted_at IS NULL", userUUID, targetUUID).
		Updates(map[string]interface{}{"deleted_at": now, "status": "Unfollowed"}).Error
}

func (d *FollowDao) Exists(ctx context.Context, userUUID, targetUUID string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&po.FollowPo{}).
		Where("user_uuid = ? AND target_uuid = ? AND deleted_at IS NULL", userUUID, targetUUID).
		Count(&count).Error
	return count > 0, err
}

func (d *FollowDao) QueryFollowers(ctx context.Context, targetUUID string, offset, limit int) ([]*po.FollowPo, int64, error) {
	var list []*po.FollowPo
	q := d.db.WithContext(ctx).Model(&po.FollowPo{}).Where("target_uuid = ? AND deleted_at IS NULL", targetUUID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	if err := q.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *FollowDao) QueryFollowings(ctx context.Context, userUUID string, offset, limit int) ([]*po.FollowPo, int64, error) {
	var list []*po.FollowPo
	q := d.db.WithContext(ctx).Model(&po.FollowPo{}).Where("user_uuid = ? AND deleted_at IS NULL", userUUID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	if err := q.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *FollowDao) Upsert(ctx context.Context, follow *po.FollowPo) error {
	err := d.db.WithContext(ctx).Model(&po.FollowPo{}).Where("user_uuid = ? AND target_uuid = ?", follow.UserUUID, follow.TargetUUID).First(&po.FollowPo{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d.Create(ctx, follow)
		}
		logger.Errorf("follow upsert user_uuid: %v, target_uuid: %v error: %v", follow.UserUUID, follow.TargetUUID, err)
		return err
	}
	// 已存在且可能被软删，重置删除时间
	return d.db.WithContext(ctx).Model(&po.FollowPo{}).Where("user_uuid = ? AND target_uuid = ?", follow.UserUUID, follow.TargetUUID).
		Updates(map[string]interface{}{"deleted_at": nil, "status": follow.Status, "updated_at": time.Now()}).Error
}
