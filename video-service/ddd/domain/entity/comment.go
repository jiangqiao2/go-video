package entity

import "time"

// Comment represents a comment on a video.
type Comment struct {
	ID          uint64
	CommentUUID string
	RootUUID    string
	VideoUUID   string
	UserUUID    string
	Content     string
	ParentUUID  string
	ParentType  string
	Depth       int
	Path        string
	LikeCount   int64
	ReplyCount  int64
	Liked       bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
