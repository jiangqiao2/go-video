package repo

import (
	"context"
	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/vo"
)

// VideoRepository 视频仓储接口
type VideoRepository interface {
	// Save 保存视频
	Save(ctx context.Context, video *entity.Video) error
	CreateVideo(ctx context.Context, video *entity.Video, videoUploadTask *entity.VideoUploadTaskEntity) error
	UpdateVideoStatus(ctx context.Context, videoUUID string, videoStatus vo.VideoStatus, videoUploadTaskUUID string, videotaskStatus vo.VideoUploadTaskStatus) error
}
