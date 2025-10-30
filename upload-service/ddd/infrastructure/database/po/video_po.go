package po

import "time"

// VideoPo stores video publishing metadata.
type VideoPo struct {
	BaseModel
	VideoUUID       string     `gorm:"column:video_uuid" json:"video_uuid"`
	UploadVideoUUID string     `gorm:"column:upload_video_uuid" json:"upload_video_uuid"`
	UserUUID        string     `gorm:"column:user_uuid" json:"user_uuid"`
	Title           string     `gorm:"column:title" json:"title"`
	Description     string     `gorm:"column:description" json:"description"`
	TagsJSON        string     `gorm:"column:tags_json" json:"tags_json"`
	CoverURL        string     `gorm:"column:cover_url" json:"cover_url"`
	Status          string     `gorm:"column:status" json:"status"`
	PublishedAt     *time.Time `gorm:"column:published_at" json:"published_at"`
	TranscodeTaskUUID string   `gorm:"column:transcode_task_uuid" json:"transcode_task_uuid"`
	VideoURL          string   `gorm:"column:video_url" json:"video_url"`
	ErrorMessage      string   `gorm:"column:error_message" json:"error_message"`
}

func (VideoPo) TableName() string {
	return "video_publish"
}
