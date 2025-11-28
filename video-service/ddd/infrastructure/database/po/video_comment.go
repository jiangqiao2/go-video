package po

import "time"

// VideoComment represents a comment on a video.
type VideoComment struct {
	ID          uint64     `gorm:"primaryKey;column:id"`
	CommentUUID string     `gorm:"column:comment_uuid"`
	VideoUUID   string     `gorm:"column:video_uuid"`
	UserUUID    string     `gorm:"column:user_uuid"`
	Content     string     `gorm:"column:content"`
	ParentUUID  *string    `gorm:"column:parent_uuid"`
	LikeCount   int64      `gorm:"column:like_count"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (VideoComment) TableName() string {
	return "video_comment"
}
