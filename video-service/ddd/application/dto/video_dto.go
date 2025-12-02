package dto

import (
	"video-service/ddd/domain/entity"
)

type VideoDto struct {
	VideoUUID   string `json:"video_uuid"`
	UploadVideo string `json:"upload_video_uuid,omitempty"`
	UserUUID    string `json:"user_uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	VideoURL    string `json:"video_url"`
	Status      string `json:"status"`
	LikeCount   int    `json:"like_count"`
	Liked       bool   `json:"liked,omitempty"`
	PlayCount   int    `json:"play_count"`
	CreatedAt   int64  `json:"created_at"`
	PublishedAt int64  `json:"published_at,omitempty"`
}

type CommentDto struct {
	CommentUUID string `json:"comment_uuid"`
	VideoUUID   string `json:"video_uuid"`
	UserUUID    string `json:"user_uuid"`
	Content     string `json:"content"`
	ParentUUID  string `json:"parent_uuid,omitempty"`
	CreatedAt   int64  `json:"created_at"` // 毫秒
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

type LikeDto struct {
	VideoUUID string `json:"video_uuid"`
	UserUUID  string `json:"user_uuid"`
	Liked     bool   `json:"liked"`
	LikeCount int64  `json:"like_count,omitempty"`
}

type PlayDto struct {
	VideoUUID string `json:"video_uuid"`
	Autoplay  bool   `json:"autoplay"`
	Muted     bool   `json:"muted"`
}

type FullscreenDto struct {
	Fullscreen bool `json:"fullscreen"`
}

func NewVideoDto(v *entity.Video, liked bool) *VideoDto {
	if v == nil {
		return nil
	}
	// 前端使用 JS Date/ms 时间戳，返回毫秒级可避免二次换算
	createdAt := v.CreatedAt.UnixMilli()
	var publishedTs int64
	if v.PublishedAt != nil {
		publishedTs = v.PublishedAt.UnixMilli()
	}
	uploadVideo := v.UploadVideoUUID
	return &VideoDto{
		VideoUUID:   v.VideoUUID,
		UploadVideo: uploadVideo,
		UserUUID:    v.UserUUID,
		Title:       v.Title,
		Description: v.Description,
		CoverURL:    v.CoverURL,
		VideoURL:    v.VideoURL,
		Status:      v.Status,
		LikeCount:   int(v.LikeCount),
		Liked:       liked,
		PlayCount:   int(v.PlayCount),
		CreatedAt:   createdAt,
		PublishedAt: publishedTs,
	}
}

func NewCommentDto(c *entity.Comment) *CommentDto {
	if c == nil {
		return nil
	}
	return &CommentDto{
		CommentUUID: c.CommentUUID,
		VideoUUID:   c.VideoUUID,
		UserUUID:    c.UserUUID,
		Content:     c.Content,
		ParentUUID:  c.ParentUUID,
		CreatedAt:   c.CreatedAt.UnixMilli(),
	}
}

func NewVideoListDto(videos []*entity.Video, total int64, page, size int) *VideoListDto {
	items := make([]*VideoDto, 0, len(videos))
	for _, v := range videos {
		items = append(items, NewVideoDto(v, false))
	}
	return &VideoListDto{List: items, Page: page, Size: size, Total: total}
}

func NewCommentListDto(comments []*entity.Comment, total int64, page, size int) *CommentListDto {
	items := make([]*CommentDto, 0, len(comments))
	for _, c := range comments {
		items = append(items, NewCommentDto(c))
	}
	return &CommentListDto{List: items, Page: page, Size: size, Total: total}
}
