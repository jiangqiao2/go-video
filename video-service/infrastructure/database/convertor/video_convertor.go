package convertor

import (
	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/vo"
	"go-video/ddd/video/infrastructure/database/po"
	"go-video/pkg/assert"
	"sync"
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
		UserUUID:    video.UserUuid(), // 需要根据UserID获取UserUUID
		Title:       video.Title(),
		Description: video.Description(),
		Filename:    video.Filename(),
		FileSize:    video.FileSize(),
		Format:      video.Format(),
		StoragePath: video.StoragePath(),
		Status:      video.Status().Value(),
	}

	return videoPO
}

// POToEntity PO转实体
func (c *VideoConvertor) POToEntity(videoPO *po.VideoPo) *entity.Video {
	if videoPO == nil {
		return nil
	}

	video := entity.NewVideo(videoPO.UUID, videoPO.UserUUID, videoPO.Title, videoPO.Description, videoPO.Filename, videoPO.FileSize, videoPO.Format, vo.NewVideoStatus(videoPO.Status))

	return video
}

// VideoUploadTaskEntityToPO 视频上传任务实体转PO
func (c *VideoConvertor) VideoUploadTaskEntityToPO(task *entity.VideoUploadTaskEntity) *po.VideoUploadTaskPo {
	if task == nil {
		return nil
	}

	taskPO := &po.VideoUploadTaskPo{
		UUID:        task.UUID(),
		UserUUID:    task.UserUuid(),
		Status:      task.Status().String(),
		ErrorMsg:    task.ErrorMsg(),
		CompletedAt: task.CompletedAt(),
		StoragePath: task.ObjectName(),
	}

	return taskPO
}

// VideoUploadTaskPOToEntity 视频上传任务PO转实体
func (c *VideoConvertor) VideoUploadTaskPOToEntity(taskPO *po.VideoUploadTaskPo) *entity.VideoUploadTaskEntity {
	if taskPO == nil {
		return nil
	}
	return entity.NewVideoUploadTask(taskPO.UUID,
		taskPO.UserUUID,
		vo.NewVideoUploadTaskStatus(taskPO.Status),
		taskPO.ErrorMsg,
		taskPO.CompletedAt,
		taskPO.StoragePath)
}
