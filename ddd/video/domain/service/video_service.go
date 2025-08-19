package service

import (
	"sync"

	"go-video/ddd/video/domain/repo"
	"go-video/ddd/video/infrastructure/database/persistence"
	"go-video/pkg/assert"
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
