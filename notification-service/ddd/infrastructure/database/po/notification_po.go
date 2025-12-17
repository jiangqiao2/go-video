package po

import "time"

// Notification 持久化对象，对应 notifications 表。
type Notification struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserUUID  string     `gorm:"column:user_uuid"`
	Type      string     `gorm:"column:type"`
	Title     string     `gorm:"column:title"`
	Content   string     `gorm:"column:content"`
	ExtraJSON string     `gorm:"column:extra_json"`
	IsRead    bool       `gorm:"column:is_read"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	ReadAt    *time.Time `gorm:"column:read_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
