package entity

import "time"

// Comment represents a comment on a video.
type Comment struct {
	ID          uint64
	CommentUUID string
	VideoUUID   string
	UserUUID    string
	Content     string
	ParentUUID  string
	LikeCount   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
