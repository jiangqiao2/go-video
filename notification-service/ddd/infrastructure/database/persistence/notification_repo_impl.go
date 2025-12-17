package persistence

import (
	"context"
	"notification-service/ddd/domain/entity"
	drepo "notification-service/ddd/domain/repo"
	"notification-service/ddd/infrastructure/database/dao"
	"notification-service/ddd/infrastructure/database/po"
)

type notificationRepositoryImpl struct {
	dao *dao.NotificationDao
}

func NewNotificationRepository() drepo.NotificationRepository {
	return &notificationRepositoryImpl{dao: dao.NewNotificationDao()}
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
	return r.dao.Create(ctx, p)
}

func (r *notificationRepositoryImpl) ListByUser(ctx context.Context, userUUID string, offset, limit int) ([]*entity.Notification, error) {
	pos, err := r.dao.ListByUser(ctx, userUUID, offset, limit)
	if err != nil {
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
	return r.dao.CountUnread(ctx, userUUID)
}

func (r *notificationRepositoryImpl) MarkRead(ctx context.Context, userUUID string, ids []uint64) error {
	return r.dao.MarkRead(ctx, userUUID, ids)
}
