package persistence

import (
	"context"
	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/repo"
	"go-video/ddd/video/domain/vo"
	"go-video/ddd/video/infrastructure/database/convertor"
	"go-video/ddd/video/infrastructure/database/dao"
)

// videoRepositoryImpl 视频仓储实现
type videoRepositoryImpl struct {
	videoDao       *dao.VideoDao
	videoConvertor *convertor.VideoConvertor
}

// NewVideoRepository 创建视频仓储实例（支持依赖注入）
func NewVideoRepository() repo.VideoRepository {
	return &videoRepositoryImpl{
		videoDao:       dao.NewVideoDao(),
		videoConvertor: convertor.NewVideoConvertor(),
	}
}

// Save 保存视频
func (r *videoRepositoryImpl) Save(ctx context.Context, video *entity.Video) error {
	videoPO := r.videoConvertor.EntityToPO(video)
	return r.videoDao.Create(ctx, videoPO)
}

func (r *videoRepositoryImpl) CreateVideo(ctx context.Context, video *entity.Video, videoUploadTask *entity.VideoUploadTaskEntity) error {
	videoPO := r.videoConvertor.EntityToPO(video)
	videoUploadTaskPo := r.videoConvertor.VideoUploadTaskEntityToPO(videoUploadTask)
	return r.videoDao.CreateVideoAndTask(ctx, videoPO, videoUploadTaskPo)
}

func (r *videoRepositoryImpl) UpdateVideoStatus(ctx context.Context, videoUUID string, videoStatus vo.VideoStatus, videoUploadTaskUUID string, videotaskStatus vo.VideoUploadTaskStatus) error {
	return r.videoDao.UpdateVideoStatus(ctx, videoUUID, videoStatus.Value(), videoUploadTaskUUID, videotaskStatus.Value())
}
