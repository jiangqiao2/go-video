package app

import (
	"context"
	"go-video/ddd/video/application/cqe"
	"go-video/ddd/video/application/dto"
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
	Create(ctx context.Context, cmd *cqe.UploadVideoCommand) (*dto.UploadVideoDto, error)
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

func (v *videoApp) Create(ctx context.Context, cmd *cqe.UploadVideoCommand) (*dto.UploadVideoDto, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}
	fileName, err := v.minioService.UploadVideo(ctx, cmd.UserUUID, cmd.File)
	if err != nil {
		return nil, err
	}
	videoEntity := entity.NewVideo(
		cmd.UserUUID, cmd.Title, cmd.Description, fileName, cmd.FileSize, cmd.Format,
	)

	err = v.videoRepo.Save(ctx, videoEntity)
	if err != nil {
		return nil, err
	}
	return &dto.UploadVideoDto{
		VideoUUID: videoEntity.UUID(),
	}, nil
}
