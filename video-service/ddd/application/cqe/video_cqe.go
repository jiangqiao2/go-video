package cqe

type PublishVideoReq struct {
	VideoUUID       string   `json:"video_uuid"`        // new video uuid (optional)
	UploadVideoUUID string   `json:"upload_video_uuid"` // source upload video uuid
	UserUUID        string   `json:"user_uuid"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	CoverURL        string   `json:"cover_url"`
	VideoURL        string   `json:"video_url"`
	Tags            []string `json:"tags"`
}

type LikeReq struct {
	VideoUUID string `json:"video_uuid"`
}

type PlayReq struct {
	VideoUUID string `json:"video_uuid"`
}

type CommentCreateReq struct {
	VideoUUID  string `json:"video_uuid"`
	UserUUID   string `json:"user_uuid"`
	Content    string `json:"content"`
	ParentUUID string `json:"parent_uuid,omitempty"`
}
