package gateway

import "context"

// NotificationService 抽象通知下游，用于在视频发布后提醒作者。
type NotificationService interface {
	// NotifyVideoPublished 在视频发布成功后发送一条站内通知给作者。
	NotifyVideoPublished(ctx context.Context, userUUID, videoUUID, title string) error
}
