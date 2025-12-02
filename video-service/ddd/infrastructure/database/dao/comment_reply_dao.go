package dao

import (
	"context"
	"errors"
	"fmt"

	"video-service/ddd/infrastructure/database/po"
	"video-service/pkg/manager"

	"gorm.io/gorm"
)

type CommentReplyDao struct {
	db *gorm.DB
}

func NewCommentReplyDao() *CommentReplyDao {
	deps := manager.GetDependencies()
	if deps == nil || deps.DB == nil {
		panic("video-service dependencies not initialized")
	}
	return &CommentReplyDao{db: deps.DB}
}

func (d *CommentReplyDao) Create(ctx context.Context, c *po.VideoCommentReply, parentPath string) error {
	// insert to get ID then set path
	if err := d.db.WithContext(ctx).Model(&po.VideoCommentReply{}).Create(c).Error; err != nil {
		return err
	}
	path := parentPath
	segment := fmt.Sprintf("%06d", c.Id)
	if parentPath == "" {
		path = segment
	} else {
		path = parentPath + "/" + segment
	}
	return d.db.WithContext(ctx).Model(&po.VideoCommentReply{}).
		Where("id = ?", c.Id).
		Update("path", path).Error
}

func (d *CommentReplyDao) ListReplies(ctx context.Context, rootUUID string, parentUUID string, sortBy string, page, size int) ([]po.VideoCommentReply, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	offset := (page - 1) * size
	query := d.db.WithContext(ctx).Model(&po.VideoCommentReply{}).Where("root_uuid = ? AND is_deleted = 0", rootUUID)
	if parentUUID != "" {
		query = query.Where("parent_uuid = ?", parentUUID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 子评论按创建时间正序展示，新回复排在列表末尾；增加 id 次序兜底
	query = query.Order("created_at ASC").Order("id ASC")
	var list []po.VideoCommentReply
	if err := query.Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *CommentReplyDao) FindByUUID(ctx context.Context, commentUUID string) (*po.VideoCommentReply, error) {
	var c po.VideoCommentReply
	if err := d.db.WithContext(ctx).Where("comment_uuid = ? AND is_deleted = 0", commentUUID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (d *CommentReplyDao) UpdateLikeCount(ctx context.Context, commentUUID string, likeCount int64) error {
	return d.db.WithContext(ctx).Model(&po.VideoCommentReply{}).
		Where("comment_uuid = ? AND is_deleted = 0", commentUUID).
		Update("like_count", likeCount).Error
}

func (d *CommentReplyDao) IncrementReplyCount(ctx context.Context, commentUUID string, delta int64) error {
	return d.db.WithContext(ctx).Model(&po.VideoCommentReply{}).
		Where("comment_uuid = ? AND is_deleted = 0", commentUUID).
		UpdateColumn("reply_count", gorm.Expr("GREATEST(reply_count + ?, 0)", delta)).Error
}
