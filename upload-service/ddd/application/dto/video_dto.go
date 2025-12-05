package dto

import (
	"time"

	"upload-service/ddd/domain/entity"
)

// VideoDetailDto describes published video metadata returned to clients.
type VideoDetailDto struct {
	VideoUUID         string     `json:"video_uuid"`
	UploadVideoUUID   string     `json:"upload_video_uuid"`
	UserUUID          string     `json:"user_uuid"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Tags              []string   `json:"tags"`
	CoverURL          string     `json:"cover_url"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	UploaderAccount   string     `json:"uploader_account,omitempty"`
	UploaderAvatarURL string     `json:"uploader_avatar_url,omitempty"`
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
		CreatedAt:       video.CreatedAt(),
		PublishedAt:     publishedAt,
	}
}

// VideoListDto describes a paginated list of videos.
type VideoListDto struct {
	Videos     []VideoDetailDto `json:"videos"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Size       int              `json:"size"`
	TotalPages int              `json:"total_pages"`
}

// NewVideoListDto builds a VideoListDto from entities.
func NewVideoListDto(videos []*entity.VideoEntity, total int64, page, size int) *VideoListDto {
	items := make([]VideoDetailDto, 0, len(videos))
	for _, video := range videos {
		if dto := NewVideoDetailDto(video); dto != nil {
			items = append(items, *dto)
		}
	}

	totalPages := 0
	if size > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}

	return &VideoListDto{
		Videos:     items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}

func NewVideoListFromItems(items []VideoDetailDto, total int64, page, size int) *VideoListDto {
	totalPages := 0
	if size > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return &VideoListDto{
		Videos:     items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}
