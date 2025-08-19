package app

import (
	"context"
	"go-video/ddd/video/application/cqe"
	"go-video/ddd/video/domain/service"
	"go-video/pkg/assert"
	"sync"
)

var (
	onceVideoApp      sync.Once
	singletonVideoApp VideoApp
)

type VideoApp interface {
	Create(ctx context.Context, cmd *cqe.UploadVideoCommand) error
}

type videoApp struct {
	videoService *service.VideoService
}

func DefaultVideoApp() VideoApp {
	assert.NotCircular()
	onceVideoApp.Do(func() {
		singletonVideoApp = &videoApp{
			videoService: service.DefaultVideoService(),
		}
	})
	assert.NotNil(singletonVideoApp)
	return singletonVideoApp
}

func (v *videoApp) Create(ctx context.Context, cmd *cqe.UploadVideoCommand) error {

}
