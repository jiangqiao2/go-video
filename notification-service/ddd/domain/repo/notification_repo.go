package repo

import (
	"context"

	"notification-service/ddd/domain/entity"
)

// NotificationRepository 通知仓储接口，隐藏具体持久化实现。
type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	ListByUser(ctx context.Context, userUUID string, offset, limit int) ([]*entity.Notification, error)
	CountUnread(ctx context.Context, userUUID string) (int64, error)
	MarkRead(ctx context.Context, userUUID string, ids []uint64) error
}
