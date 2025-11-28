package cqe

import (
	"strings"
	"video-service/pkg/errno"
)

type PublishVideoReq struct {
	VideoUUID       string `json:"video_uuid"`        // new video uuid (optional)
	UploadVideoUUID string `json:"upload_video_uuid"` // source upload video uuid
	UserUUID        string `json:"user_uuid"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	CoverURL        string `json:"cover_url"`
	VideoURL        string `json:"video_url"`
	DurationSec     *int   `json:"duration_sec,omitempty"`
	Resolution      string `json:"resolution,omitempty"`
	SizeBytes       *int64 `json:"size_bytes,omitempty"`
	Status          string `json:"status,omitempty"`
}

type LikeReq struct {
	VideoUUID string `json:"video_uuid"`
	UserUUID  string `json:"user_uuid"`
}

type PlayReq struct {
	VideoUUID string `json:"video_uuid"`
}

type GetVideoReq struct {
	VideoUUID string `json:"video_uuid"`
}

type ListVideosReq struct {
	Page     int    `json:"page" form:"page"`
	Size     int    `json:"size" form:"size"`
	UserUUID string `json:"user_uuid" form:"user_uuid"`
	Status   string `json:"status" form:"status"`
}

type ListCommentsReq struct {
	VideoUUID string `json:"video_uuid"`
	Page      int    `json:"page" form:"page"`
	Size      int    `json:"size" form:"size"`
}

type CommentCreateReq struct {
	VideoUUID  string `json:"video_uuid"`
	UserUUID   string `json:"user_uuid"`
	Content    string `json:"content"`
	ParentUUID string `json:"parent_uuid,omitempty"`
}

// 预创建占位，发布链路第一步，由 upload 调用。
type PrecreateReq struct {
	VideoUUID         string `json:"video_uuid"`
	UploadVideoUUID   string `json:"upload_video_uuid"`
	UserUUID          string `json:"user_uuid"`
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
	CoverURL          string `json:"cover_url,omitempty"`
	TranscodeTaskUUID string `json:"task_uuid,omitempty"`
}

// 转码结果回写，由 upload 调用。
type UpdateTranscodeResultReq struct {
	VideoUUID   string `json:"video_uuid"`
	TaskUUID    string `json:"task_uuid"`
	Status      string `json:"status"` // processing/published/failed
	VideoURL    string `json:"video_url,omitempty"`
	ErrorMsg    string `json:"error_message,omitempty"`
	DurationSec *int   `json:"duration_sec,omitempty"`
	SizeBytes   *int64 `json:"size_bytes,omitempty"`
}

// Normalize sets defaults.
func (r *PublishVideoReq) Normalize() {
	if r.Status == "" {
		r.Status = "processing"
	}
}

// Validate basic fields.
func (r *PublishVideoReq) Validate() error {
	if r.UserUUID == "" || r.Title == "" || r.VideoURL == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "user_uuid/title/video_url")
	}
	return nil
}

// Normalize sets defaults for comment create.
func (r *CommentCreateReq) Normalize() {}

func (r *CommentCreateReq) Validate() error {
	if r.VideoUUID == "" || r.UserUUID == "" || r.Content == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid/user_uuid/content")
	}
	return nil
}

func (r *PrecreateReq) Normalize() {}

func (r *PrecreateReq) Validate() error {
	if r.VideoUUID == "" || r.UserUUID == "" || r.UploadVideoUUID == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid/user_uuid/upload_video_uuid")
	}
	return nil
}

func (r *UpdateTranscodeResultReq) Normalize() {
	r.Status = strings.ToLower(r.Status)
}

func (r *UpdateTranscodeResultReq) Validate() error {
	if r.VideoUUID == "" || r.TaskUUID == "" || r.Status == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid/task_uuid/status")
	}
	switch r.Status {
	case "processing", "published", "failed":
	default:
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "status must be processing/published/failed")
	}
	if r.Status == "published" && r.VideoURL == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_url required when published")
	}
	return nil
}

func (r *LikeReq) Normalize() {}

func (r *LikeReq) Validate() error {
	if r.VideoUUID == "" || r.UserUUID == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid/user_uuid")
	}
	return nil
}

func (r *PlayReq) Normalize() {}

func (r *PlayReq) Validate() error {
	if r.VideoUUID == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid")
	}
	return nil
}

func (r *GetVideoReq) Normalize() {}

func (r *GetVideoReq) Validate() error {
	if r.VideoUUID == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid")
	}
	return nil
}

func (r *ListVideosReq) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = 20
	}
	if r.Size > 200 {
		r.Size = 200
	}
	if r.Status != "" {
		r.Status = strings.ToLower(r.Status)
	}
}

func (r *ListVideosReq) Validate() error { return nil }

func (r *ListCommentsReq) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = 20
	}
	if r.Size > 200 {
		r.Size = 200
	}
}

func (r *ListCommentsReq) Validate() error {
	if r.VideoUUID == "" {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "video_uuid")
	}
	return nil
}
