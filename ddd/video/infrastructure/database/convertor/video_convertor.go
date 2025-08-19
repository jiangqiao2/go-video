package convertor

import (
	"sync"
	"time"

	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/infrastructure/database/po"
	"go-video/pkg/assert"
)

var (
	videoConvertorOnce      sync.Once
	singletonVideoConvertor *VideoConvertor
)

// VideoConvertor 视频转换器
type VideoConvertor struct{}

// DefaultVideoConvertor 获取默认视频转换器实例
func DefaultVideoConvertor() *VideoConvertor {
	assert.NotCircular()
	videoConvertorOnce.Do(func() {
		singletonVideoConvertor = &VideoConvertor{}
	})
	assert.NotNil(singletonVideoConvertor)
	return singletonVideoConvertor
}

// NewVideoConvertor 创建视频转换器实例
func NewVideoConvertor() *VideoConvertor {
	return &VideoConvertor{}
}

// EntityToPO 实体转PO
func (c *VideoConvertor) EntityToPO(video *entity.Video) *po.VideoPo {
	if video == nil {
		return nil
	}

	videoPO := &po.VideoPo{
		UUID:        video.UUID(),
		UserUUID:    "", // 需要根据UserID获取UserUUID
		Title:       video.Title(),
		Description: video.Description(),
		Filename:    video.Filename(),
		FileSize:    video.FileSize(),
		Duration:    video.Duration(),
		Format:      video.Format(),
		StoragePath: video.StoragePath(),
		Status:      c.statusToString(video.Status()),
	}

	// 设置基础字段
	videoPO.Id = video.ID()
	if !video.CreatedAt().IsZero() {
		createdAt := video.CreatedAt()
		videoPO.CreatedAt = &createdAt
	}
	if !video.UpdatedAt().IsZero() {
		updatedAt := video.UpdatedAt()
		videoPO.UpdatedAt = &updatedAt
	}
	if video.IsDeleted() {
		videoPO.IsDeleted = 1
	}

	return videoPO
}

// POToEntity PO转实体
func (c *VideoConvertor) POToEntity(videoPO *po.VideoPo) *entity.Video {
	if videoPO == nil {
		return nil
	}

	video := entity.NewVideo(videoPO.UserUUID, videoPO.Title, videoPO.Description, videoPO.Filename, videoPO.FileSize, videoPO.Format)

	// 设置基础字段
	video.SetID(videoPO.Id)
	video.SetUUID(videoPO.UUID)
	video.SetDuration(videoPO.Duration)
	video.SetStoragePath(videoPO.StoragePath)
	video.SetStatus(c.stringToStatus(videoPO.Status))

	// 设置时间戳
	var createdAt, updatedAt time.Time
	var deletedAt *time.Time
	if videoPO.CreatedAt != nil {
		createdAt = *videoPO.CreatedAt
	}
	if videoPO.UpdatedAt != nil {
		updatedAt = *videoPO.UpdatedAt
	}
	if videoPO.IsDeleted == 1 {
		now := time.Now()
		deletedAt = &now
	}
	video.SetTimestamps(createdAt, updatedAt, deletedAt)

	return video
}

// statusToString 状态转字符串
func (c *VideoConvertor) statusToString(status entity.VideoStatus) string {
	switch status {
	case entity.VideoStatusUploading:
		return "uploading"
	case entity.VideoStatusProcessing:
		return "processing"
	case entity.VideoStatusReady:
		return "ready"
	case entity.VideoStatusFailed:
		return "failed"
	case entity.VideoStatusDeleted:
		return "deleted"
	default:
		return "uploading"
	}
}

// stringToStatus 字符串转状态
func (c *VideoConvertor) stringToStatus(status string) entity.VideoStatus {
	switch status {
	case "uploading":
		return entity.VideoStatusUploading
	case "processing":
		return entity.VideoStatusProcessing
	case "ready":
		return entity.VideoStatusReady
	case "failed":
		return entity.VideoStatusFailed
	case "deleted":
		return entity.VideoStatusDeleted
	default:
		return entity.VideoStatusUploading
	}
}
