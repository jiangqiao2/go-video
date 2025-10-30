package cqe

import (
	"strings"
	"unicode/utf8"

	"upload-service/pkg/errno"
)

const (
	maxVideoTitleLen       = 120
	maxVideoDescriptionLen = 2000
	maxVideoTags           = 10
	maxVideoTagLen         = 32
)

// PublishVideoReq carries information required to publish a video.
type PublishVideoReq struct {
	UploadVideoUUID string   `json:"upload_video_uuid"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
	CoverURL        string   `json:"cover_url"`
	UserUUID        string   `json:"user_uuid"`
}

// Normalize trims strings and deduplicates tags.
func (r *PublishVideoReq) Normalize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
	r.Tags = sanitizeTags(r.Tags)
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
