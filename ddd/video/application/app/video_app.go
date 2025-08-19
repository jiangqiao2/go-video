package app

import (
	"context"
	"go-video/ddd/video/application/cqe"
	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/gateway"
	"go-video/ddd/video/domain/repo"
	"go-video/ddd/video/infrastructure/database/persistence"
	"go-video/ddd/video/infrastructure/minio"
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
	minioService gateway.MinioService
	videoRepo    repo.VideoRepository
}

func DefaultVideoApp() VideoApp {
	assert.NotCircular()
	onceVideoApp.Do(func() {
		singletonVideoApp = &videoApp{
			minioService: minio.DefaultMinioService(),
			videoRepo:    persistence.NewVideoRepository(),
		}
	})
	assert.NotNil(singletonVideoApp)
	return singletonVideoApp
}

func (v *videoApp) Create(ctx context.Context, cmd *cqe.UploadVideoCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}
	fileName, err := v.minioService.UploadVideo(ctx, "", cmd.File)
	if err != nil {
		return err
	}
	videoEntity := entity.NewVideo(
		"", cmd.Title, cmd.Description, fileName, cmd.FileSize, cmd.Format,
	)

	return v.videoRepo.Save(ctx, videoEntity)
}
