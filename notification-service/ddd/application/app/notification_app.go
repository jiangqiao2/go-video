package app

import (
	"context"

	"notification-service/ddd/application/cqe"
	"notification-service/ddd/application/dto"
	"notification-service/ddd/domain/entity"
	drepo "notification-service/ddd/domain/repo"
	"notification-service/ddd/infrastructure/database/persistence"
	"notification-service/pkg/errno"
)

// NotificationApp 应用服务接口，编排通知相关用例。
type NotificationApp interface {
	ListNotifications(ctx context.Context, userUUID string, req *cqe.ListNotificationsReq) (*dto.ListNotificationsResponse, error)
	MarkRead(ctx context.Context, userUUID string, req *cqe.MarkReadReq) error
	Create(ctx context.Context, req *cqe.CreateNotificationReq) error
}

type notificationAppImpl struct {
	repo drepo.NotificationRepository
}

// DefaultNotificationApp 返回默认的应用服务实现。
func DefaultNotificationApp() NotificationApp {
	return &notificationAppImpl{
		repo: persistence.NewNotificationRepository(),
	}
}

func (a *notificationAppImpl) ListNotifications(ctx context.Context, userUUID string, req *cqe.ListNotificationsReq) (*dto.ListNotificationsResponse, error) {
	if userUUID == "" {
		return nil, errno.ErrUnauthorized
	}
	req.Normalize()
	offset := (req.Page - 1) * req.PageSize

	list, err := a.repo.ListByUser(ctx, userUUID, offset, req.PageSize)
	if err != nil {
		return nil, err
	}
	unread, err := a.repo.CountUnread(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotificationDto, 0, len(list))
	for _, n := range list {
		items = append(items, dto.NotificationDto{
			ID:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Content:   n.Content,
			ExtraJSON: n.ExtraJSON,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
			ReadAt:    n.ReadAt,
		})
	}

	return &dto.ListNotificationsResponse{
		Notifications: items,
		UnreadCount:   unread,
	}, nil
}

func (a *notificationAppImpl) MarkRead(ctx context.Context, userUUID string, req *cqe.MarkReadReq) error {
	if userUUID == "" {
		return errno.ErrUnauthorized
	}
	if !req.Validate() {
		return errno.ErrParameterInvalid
	}
	return a.repo.MarkRead(ctx, userUUID, req.IDs)
}

// Create 创建一条新的通知记录（内部调用）。
func (a *notificationAppImpl) Create(ctx context.Context, req *cqe.CreateNotificationReq) error {
	if req == nil || !req.Validate() {
		return errno.ErrParameterInvalid
	}
	n := entity.NewNotification(
		req.UserUUID,
		req.Type,
		req.Title,
		req.Content,
		req.ExtraJSON,
	)
	return a.repo.Create(ctx, n)
}
