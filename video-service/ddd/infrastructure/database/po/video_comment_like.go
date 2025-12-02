package po

// VideoCommentLike represents likes on a comment.
type VideoCommentLike struct {
	BaseModel
	UserUUID    string `gorm:"column:user_uuid"`
	CommentUUID string `gorm:"column:comment_uuid"`
	Status      string `gorm:"column:status"`
}

func (VideoCommentLike) TableName() string {
	return "video_comment_like"
}
