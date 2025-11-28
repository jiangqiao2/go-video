package po

import "time"

// VideoLike represents a like relationship.
type VideoLike struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	UserUUID  string    `gorm:"column:user_uuid"`
	VideoUUID string    `gorm:"column:video_uuid"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (VideoLike) TableName() string {
	return "video_like"
}
