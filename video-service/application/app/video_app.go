package app

import (
	"context"
	"fmt"
	"go-video/ddd/video/application/cqe"
	"go-video/ddd/video/application/dto"
	"go-video/ddd/video/domain/entity"
	"go-video/ddd/video/domain/gateway"
	"go-video/ddd/video/domain/repo"
	"go-video/ddd/video/domain/vo"
	"go-video/ddd/video/infrastructure/database/persistence"
	"go-video/ddd/video/infrastructure/minio"
	"go-video/pkg/assert"
	"go-video/pkg/logger"
	"sync"
)

var (
	onceVideoApp      sync.Once
	singletonVideoApp VideoApp
)

type VideoApp interface {
	Create(ctx context.Context, cmd *cqe.UploadVideoCommand) (*dto.UploadVideoDto, error)
	SyncUploadVideo(ctx context.Context, cmd *cqe.UploadVideoCommand) (*dto.VideoSyncVideoDto, error)
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
	videoEntity := entity.DefaultVideo(
		cmd.UserUUID, cmd.Title, cmd.Description, cmd.File.Filename, cmd.FileSize, cmd.Format, fileName,
		vo.VideoStatusInit,
	)

	err = v.videoRepo.Save(ctx, videoEntity)
	if err != nil {
		return nil, err
	}
	return &dto.UploadVideoDto{
		VideoUUID: videoEntity.UUID(),
	}, nil
}

// SyncUploadVideo 异步上传视频
func (v *videoApp) SyncUploadVideo(ctx context.Context, cmd *cqe.UploadVideoCommand) (*dto.VideoSyncVideoDto, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}
	storagePath := v.minioService.GenerateObjectName(cmd.UserUUID, cmd.File.Filename)
	logger.Info(fmt.Sprintf("upload video %s to %s", cmd.UserUUID, storagePath))
	videoEntity := entity.DefaultVideo(cmd.UserUUID, cmd.Title, cmd.Description, cmd.File.Filename, cmd.FileSize, cmd.Format, storagePath, vo.VideoStatusInit)
	videoTaskEntity := entity.DefaultVideoUploadTaskEntity(
		cmd.UserUUID, videoEntity.UUID(), vo.VideoUploadTaskStatusInit, "", nil, storagePath)
	err := v.videoRepo.CreateVideo(ctx, videoEntity, videoTaskEntity)
	if err != nil {
		logger.Info(fmt.Sprintf("SyncUploadVideo CreateVideo Failed to sync upload user_uuid: %v video: %s", cmd.UserUUID, err))
		return nil, err
	}
	videoUploadVo := vo.NewUploadVideo(cmd.UserUUID, videoEntity.UUID(), videoTaskEntity.UUID(), storagePath, cmd.File)

	go v.minioService.SyncUploadVideo(ctx, videoUploadVo)

	return &dto.VideoSyncVideoDto{
		VideoUUID: videoEntity.UUID(),
		TaskUUID:  videoTaskEntity.UUID(),
	}, nil
}
