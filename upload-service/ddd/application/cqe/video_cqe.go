package cqe

import (
	"strings"
	"unicode/utf8"

	"upload-service/ddd/domain/vo"
	"upload-service/pkg/errno"
)

const (
	maxVideoTitleLen       = 120
	maxVideoDescriptionLen = 2000
	maxVideoTags           = 10
	maxVideoTagLen         = 32
	// 默认转码目标分辨率/码率：优先生成 1080p 主档，便于后续多码率 HLS 以此为上限
	defaultResolution = "1080p"
	defaultBitrate    = "4000k"
)

// PublishVideoReq carries information required to publish a video.
type PublishVideoReq struct {
	UploadVideoUUID  string   `json:"upload_video_uuid"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	CoverURL         string   `json:"cover_url"`
	UserUUID         string   `json:"user_uuid"`
	TargetResolution string   `json:"target_resolution"`
	TargetBitrate    string   `json:"target_bitrate"`
}

// Normalize trims strings and deduplicates tags.
func (r *PublishVideoReq) Normalize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
	r.Tags = sanitizeTags(r.Tags)
	if strings.TrimSpace(r.TargetResolution) == "" {
		r.TargetResolution = defaultResolution
	}
	if strings.TrimSpace(r.TargetBitrate) == "" {
		r.TargetBitrate = defaultBitrate
	}
}

// Validate ensures request values satisfy publishing constraints.
func (r *PublishVideoReq) Validate() error {
	if r.UploadVideoUUID == "" {
		return errno.NewSimpleBizError(errno.ErrMissingParam, nil, "upload_video_uuid")
	}
	if r.UserUUID == "" {
		return errno.NewSimpleBizError(errno.ErrMissingParam, nil, "user_uuid")
	}
	if r.Title == "" || utf8.RuneCountInString(r.Title) > maxVideoTitleLen {
		return errno.NewSimpleBizError(errno.ErrVideoTitleIllegal, nil)
	}
	if r.Description != "" && utf8.RuneCountInString(r.Description) > maxVideoDescriptionLen {
		return errno.NewSimpleBizError(errno.ErrVideoDescriptionIllegal, nil)
	}
	if len(r.Tags) > maxVideoTags {
		return errno.NewSimpleBizError(errno.ErrVideoTagsIllegal, nil)
	}
	for _, tag := range r.Tags {
		if tag == "" || utf8.RuneCountInString(tag) > maxVideoTagLen {
			return errno.NewSimpleBizError(errno.ErrVideoTagsIllegal, nil)
		}
	}
	if utf8.RuneCountInString(r.TargetResolution) == 0 || utf8.RuneCountInString(r.TargetResolution) > 20 {
		return errno.NewSimpleBizError(errno.ErrVideoTitleIllegal, nil)
	}
	if utf8.RuneCountInString(r.TargetBitrate) == 0 || utf8.RuneCountInString(r.TargetBitrate) > 20 {
		return errno.NewSimpleBizError(errno.ErrVideoTitleIllegal, nil)
	}
	return nil
}

func sanitizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// ListVideosReq carries query parameters for listing user's videos.
type ListVideosReq struct {
	Page     int    `form:"page"`
	Size     int    `form:"size"`
	Status   string `form:"status"`
	UserUUID string `form:"-"`
}

// Normalize applies default pagination values and trims filters.
func (r *ListVideosReq) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = 12
	} else if r.Size > 50 {
		r.Size = 50
	}
	r.Status = strings.TrimSpace(r.Status)
	if r.Status != "" {
		switch strings.ToLower(r.Status) {
		case strings.ToLower(vo.VideoStatusDraft.Value()):
			r.Status = vo.VideoStatusDraft.Value()
		case strings.ToLower(vo.VideoStatusProcessing.Value()):
			r.Status = vo.VideoStatusProcessing.Value()
		case strings.ToLower(vo.VideoStatusPublished.Value()):
			r.Status = vo.VideoStatusPublished.Value()
		case strings.ToLower(vo.VideoStatusFailed.Value()):
			r.Status = vo.VideoStatusFailed.Value()
		}
	}
}

// Validate ensures the list query parameters are acceptable.
func (r *ListVideosReq) Validate() error {
	if r.UserUUID == "" {
		return errno.ErrUnauthorized
	}
	if r.Status != "" {
		status := vo.NewVideoStatus(r.Status)
		if status.Value() != r.Status {
			return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "status")
		}
	}
	return nil
}

type ListOpenVideosReq struct {
	Page   int    `form:"page"`
	Size   int    `form:"size"`
	Status string `form:"status"`
}

func (r *ListOpenVideosReq) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = 12
	} else if r.Size > 50 {
		r.Size = 50
	}
	r.Status = strings.TrimSpace(r.Status)
	if r.Status == "" {
		r.Status = vo.VideoStatusPublished.Value()
	} else {
		switch strings.ToLower(r.Status) {
		case strings.ToLower(vo.VideoStatusPublished.Value()):
			r.Status = vo.VideoStatusPublished.Value()
		default:
			r.Status = vo.VideoStatusPublished.Value()
		}
	}
}

func (r *ListOpenVideosReq) Validate() error {
	if r.Status != "" {
		status := vo.NewVideoStatus(r.Status)
		if !status.IsPublished() {
			return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "status")
		}
	}
	return nil
}
