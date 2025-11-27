package dto

type VideoDto struct {
	VideoUUID   string   `json:"video_uuid"`
	UploadVideo string   `json:"upload_video_uuid,omitempty"`
	UserUUID    string   `json:"user_uuid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	VideoURL    string   `json:"video_url"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
	LikeCount   int      `json:"like_count"`
	PlayCount   int      `json:"play_count"`
	CreatedAt   string   `json:"created_at"`
}

type CommentDto struct {
	CommentUUID string `json:"comment_uuid"`
	VideoUUID   string `json:"video_uuid"`
	UserUUID    string `json:"user_uuid"`
	Content     string `json:"content"`
	ParentUUID  string `json:"parent_uuid,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type VideoListDto struct {
	List  []*VideoDto `json:"list"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int64       `json:"total"`
}

type CommentListDto struct {
	List  []*CommentDto `json:"list"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
	Total int64         `json:"total"`
}
