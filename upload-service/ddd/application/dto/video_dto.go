package dto

import (
	"time"

	"upload-service/ddd/domain/entity"
)

// VideoDetailDto describes published video metadata returned to clients.
type VideoDetailDto struct {
	VideoUUID       string     `json:"video_uuid"`
	UploadVideoUUID string     `json:"upload_video_uuid"`
	UserUUID        string     `json:"user_uuid"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Tags            []string   `json:"tags"`
	CoverURL        string     `json:"cover_url"`
	Status          string     `json:"status"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
}

// NewVideoDetailDto maps a domain entity to dto.
func NewVideoDetailDto(video *entity.VideoEntity) *VideoDetailDto {
	if video == nil {
		return nil
	}
	var publishedAt *time.Time
	if ts := video.PublishedAt(); ts != nil {
		t := *ts
		publishedAt = &t
	}
	return &VideoDetailDto{
		VideoUUID:       video.VideoUUID(),
		UploadVideoUUID: video.UploadVideoUUID(),
		UserUUID:        video.UserUUID(),
		Title:           video.Title(),
		Description:     video.Description(),
		Tags:            video.Tags(),
		CoverURL:        video.CoverURL(),
		Status:          video.Status().Value(),
		PublishedAt:     publishedAt,
	}
}
