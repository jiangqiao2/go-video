package po

import "time"

// VideoComment represents a comment on a video.
type VideoCommentRoot struct {
	BaseModel
	RootUUID   string     `gorm:"column:root_uuid"`
	VideoUUID  string     `gorm:"column:video_uuid"`
	UserUUID   string     `gorm:"column:user_uuid"`
	Content    string     `gorm:"column:content"`
	LikeCount  int64      `gorm:"column:like_count"`
	ReplyCount int64      `gorm:"column:reply_count"`
	IsDeleted  uint64     `gorm:"column:is_deleted"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

type VideoCommentReply struct {
	BaseModel
	CommentUUID string     `gorm:"column:comment_uuid"`
	RootUUID    string     `gorm:"column:root_uuid"`
	ParentUUID  string     `gorm:"column:parent_uuid"`
	ParentType  string     `gorm:"column:parent_type"`
	Depth       int        `gorm:"column:depth"`
	Path        string     `gorm:"column:path"`
	VideoUUID   string     `gorm:"column:video_uuid"`
	UserUUID    string     `gorm:"column:user_uuid"`
	Content     string     `gorm:"column:content"`
	LikeCount   int64      `gorm:"column:like_count"`
	ReplyCount  int64      `gorm:"column:reply_count"`
	IsDeleted   uint64     `gorm:"column:is_deleted"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (VideoCommentRoot) TableName() string  { return "video_comment_root" }
func (VideoCommentReply) TableName() string { return "video_comment_reply" }
