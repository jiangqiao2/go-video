package convertor

import (
	"video-service/ddd/domain/entity"
	"video-service/ddd/infrastructure/database/po"
)

func ToCommentPo(c *entity.Comment) *po.VideoComment {
	if c == nil {
		return nil
	}
	var parentPtr *string
	if c.ParentUUID != "" {
		s := c.ParentUUID
		parentPtr = &s
	}
	return &po.VideoComment{
		ID:          c.ID,
		CommentUUID: c.CommentUUID,
		VideoUUID:   c.VideoUUID,
		UserUUID:    c.UserUUID,
		Content:     c.Content,
		ParentUUID:  parentPtr,
		LikeCount:   c.LikeCount,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func ToCommentEntity(p *po.VideoComment) *entity.Comment {
	if p == nil {
		return nil
	}
	var parent string
	if p.ParentUUID != nil {
		parent = *p.ParentUUID
	}
	return &entity.Comment{
		ID:          p.ID,
		CommentUUID: p.CommentUUID,
		VideoUUID:   p.VideoUUID,
		UserUUID:    p.UserUUID,
		Content:     p.Content,
		ParentUUID:  parent,
		LikeCount:   p.LikeCount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
