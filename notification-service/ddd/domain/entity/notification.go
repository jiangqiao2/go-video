package entity

import "time"

// Notification 聚合根，表示一条站内通知。
type Notification struct {
	ID        uint64
	UserUUID  string
	Type      string
	Title     string
	Content   string
	ExtraJSON string
	IsRead    bool
	CreatedAt time.Time
	ReadAt    *time.Time
}

// NewNotification 创建一条新的未读通知。
func NewNotification(userUUID, typ, title, content, extraJSON string) *Notification {
	return &Notification{
		UserUUID:  userUUID,
		Type:      typ,
		Title:     title,
		Content:   content,
		ExtraJSON: extraJSON,
		IsRead:    false,
	}
}
