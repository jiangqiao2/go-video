package persistence

import (
	"context"
	"time"

	"notification-service/ddd/domain/entity"
	drepo "notification-service/ddd/domain/repo"
	"notification-service/ddd/infrastructure/database/po"
	"notification-service/internal/resource"

	"gorm.io/gorm"
)

type notificationRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实现，使用全局主库连接。
func NewNotificationRepository() drepo.NotificationRepository {
	return &notificationRepositoryImpl{
		db: resource.MainDB(),
	}
}

func (r *notificationRepositoryImpl) Create(ctx context.Context, n *entity.Notification) error {
	p := &po.Notification{
		UserUUID:  n.UserUUID,
		Type:      n.Type,
		Title:     n.Title,
		Content:   n.Content,
		ExtraJSON: n.ExtraJSON,
		IsRead:    n.IsRead,
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *notificationRepositoryImpl) ListByUser(ctx context.Context, userUUID string, offset, limit int) ([]*entity.Notification, error) {
	var pos []po.Notification
	if err := r.db.WithContext(ctx).
		Where("user_uuid = ?", userUUID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&pos).Error; err != nil {
		return nil, err
	}

	res := make([]*entity.Notification, 0, len(pos))
	for _, p := range pos {
		n := &entity.Notification{
			ID:        p.ID,
			UserUUID:  p.UserUUID,
			Type:      p.Type,
			Title:     p.Title,
			Content:   p.Content,
			ExtraJSON: p.ExtraJSON,
			IsRead:    p.IsRead,
			CreatedAt: p.CreatedAt,
			ReadAt:    p.ReadAt,
		}
		res = append(res, n)
	}
	return res, nil
}

func (r *notificationRepositoryImpl) CountUnread(ctx context.Context, userUUID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&po.Notification{}).
		Where("user_uuid = ? AND is_read = 0", userUUID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepositoryImpl) MarkRead(ctx context.Context, userUUID string, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&po.Notification{}).
		Where("user_uuid = ? AND id IN ?", userUUID, ids).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}
