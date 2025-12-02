package convertor

import (
	"video-service/ddd/domain/entity"
	"video-service/ddd/infrastructure/database/po"
)

func ToCommentRootPo(c *entity.Comment) *po.VideoCommentRoot {
	if c == nil {
		return nil
	}
	return &po.VideoCommentRoot{
		BaseModel:  po.BaseModel{Id: c.ID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt},
		RootUUID:   c.CommentUUID,
		VideoUUID:  c.VideoUUID,
		UserUUID:   c.UserUUID,
		Content:    c.Content,
		LikeCount:  c.LikeCount,
		ReplyCount: c.ReplyCount,
		IsDeleted:  0,
	}
}

func ToCommentReplyPo(c *entity.Comment) *po.VideoCommentReply {
	if c == nil {
		return nil
	}
	return &po.VideoCommentReply{
		BaseModel:   po.BaseModel{Id: c.ID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt},
		CommentUUID: c.CommentUUID,
		RootUUID:    c.RootUUID,
		ParentUUID:  c.ParentUUID,
		ParentType:  c.ParentType,
		Depth:       c.Depth,
		Path:        c.Path,
		VideoUUID:   c.VideoUUID,
		UserUUID:    c.UserUUID,
		Content:     c.Content,
		LikeCount:   c.LikeCount,
		ReplyCount:  c.ReplyCount,
		IsDeleted:   0,
	}
}

func ToCommentRootEntity(p *po.VideoCommentRoot) *entity.Comment {
	if p == nil {
		return nil
	}
	return &entity.Comment{
		ID:          p.Id,
		CommentUUID: p.RootUUID,
		RootUUID:    p.RootUUID,
		VideoUUID:   p.VideoUUID,
		UserUUID:    p.UserUUID,
		Content:     p.Content,
		LikeCount:   p.LikeCount,
		ReplyCount:  p.ReplyCount,
		Depth:       0,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToCommentReplyEntity(p *po.VideoCommentReply) *entity.Comment {
	if p == nil {
		return nil
	}
	return &entity.Comment{
		ID:          p.Id,
		CommentUUID: p.CommentUUID,
		RootUUID:    p.RootUUID,
		VideoUUID:   p.VideoUUID,
		UserUUID:    p.UserUUID,
		Content:     p.Content,
		ParentUUID:  p.ParentUUID,
		ParentType:  p.ParentType,
		Depth:       p.Depth,
		Path:        p.Path,
		LikeCount:   p.LikeCount,
		ReplyCount:  p.ReplyCount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
