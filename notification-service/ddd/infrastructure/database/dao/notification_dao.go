package dao

import (
	"context"
	"time"

	"notification-service/ddd/infrastructure/database/po"
	"notification-service/internal/resource"

	"gorm.io/gorm"
)

type NotificationDao struct {
	db *gorm.DB
}

func NewNotificationDao() *NotificationDao {
	return &NotificationDao{db: resource.MainDB()}
}

func (d *NotificationDao) Create(ctx context.Context, p *po.Notification) error {
	return d.db.WithContext(ctx).Create(p).Error
}

func (d *NotificationDao) ListByUser(ctx context.Context, userUUID string, offset, limit int) ([]po.Notification, error) {
	var pos []po.Notification
	err := d.db.WithContext(ctx).
		Where("user_uuid = ?", userUUID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return pos, nil
}

func (d *NotificationDao) CountUnread(ctx context.Context, userUUID string) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&po.Notification{}).
		Where("user_uuid = ? AND is_read = 0", userUUID).
		Count(&count).Error
	return count, err
}

func (d *NotificationDao) MarkRead(ctx context.Context, userUUID string, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return d.db.WithContext(ctx).
		Model(&po.Notification{}).
		Where("user_uuid = ? AND id IN ?", userUUID, ids).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}
