package dto

import "time"

// NotificationDto 向上层暴露的通知视图模型。
type NotificationDto struct {
	ID        uint64     `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	ExtraJSON string     `json:"extra_json,omitempty"`
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// ListNotificationsResponse 列表响应结构，包含未读数。
type ListNotificationsResponse struct {
	Notifications []NotificationDto `json:"notifications"`
	UnreadCount   int64             `json:"unread_count"`
}
