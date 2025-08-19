package service

import (
	"context"
	"sync"

	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/repo"
	"go-video/ddd/video/infrastructure/database/persistence"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
)

var (
	videoDomainServiceOnce      sync.Once
	singletonVideoDomainService *VideoService
)

// VideoService 视频领域服务
type VideoService struct {
	videoRepo repo.VideoRepository
}

// DefaultVideoService 获取默认视频服务实例
func DefaultVideoService() *VideoService {
	assert.NotCircular()
	videoDomainServiceOnce.Do(func() {
		singletonVideoDomainService = &VideoService{
			videoRepo: persistence.NewVideoRepository(),
		}
	})
	assert.NotNil(singletonVideoDomainService)
	return singletonVideoDomainService
}

// CreateVideo 创建视频
func (s *VideoService) CreateVideo(ctx context.Context, userUUID, title, description, filename string, fileSize int64, format string) (*entity.Video, error) {
	// 创建视频实体
	video, err := entity.NewVideo(userUUID, title, description, filename, fileSize, format) // userID暂时设为0，后续通过userUUID获取
	if err != nil {
		return nil, errno.NewBizErrorWithCause(errno.CodeInvalidParam, "视频信息无效", err)
	}

	// 保存视频
	if err := s.videoRepo.Save(ctx, video); err != nil {
		return nil, errno.NewBizErrorWithCause(errno.CodeDatabaseError, "保存视频失败", err)
	}

	return video, nil
}
