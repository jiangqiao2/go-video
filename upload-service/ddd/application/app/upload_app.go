package app

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/minio"
	"upload-service/pkg/assert"
)

var (
	onceUploadVideoApp      sync.Once
	singletonUploadVideoApp UploadVideoApp
)

type UploadVideoApp interface {
	UploadVideoInit(ctx context.Context, req *cqe.UploadVideoInitReq) (*dto.UploadVideoDto, error)
	UploadVideoChunk(ctx context.Context, req *cqe.UploadChunkReq) (*dto.UploadVideoChunkDto, error)
	QueryStoragePath(ctx context.Context, req *cqe.UploadVideoStoragePathReq) (*dto.UploadVideoStoragePathDto, error)
	MergeChunks(ctx context.Context, req *cqe.MergeChunkReq) (*dto.MergeChunkDto, error)
}

type uploadVideoAppImpl struct {
	minioService    gateway.MinioService
	uploadVideoRepo repo.UploadVideoRepository
	uploadVideoSrv  service.UploadVideoService
}

func DefaultUploadVideoApp() UploadVideoApp {
	assert.NotCircular()
	onceUploadVideoApp.Do(func() {
		singletonUploadVideoApp = &uploadVideoAppImpl{
			minioService:    minio.DefaultMinioService(),
			uploadVideoRepo: persistence.NewUploadVideoRepository(),
			uploadVideoSrv:  service.NewUploadVideoService(),
		}
	})
	assert.NotNil(singletonUploadVideoApp)
	return singletonUploadVideoApp
}

func (u *uploadVideoAppImpl) UploadVideoInit(ctx context.Context, req *cqe.UploadVideoInitReq) (*dto.UploadVideoDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// TODO 调用user服务检查用户ID是否存在

	// 支持断点续传
	// (文件名+文件大小+文件Hash) 查询有没有
	uploadVideoEntity, uploadChunkEntity, err := u.uploadVideoRepo.QueryUploadVideoByName(ctx, req.UserUUID, req.FileName, req.FileSize, req.FileHash)
	if err != nil {
		log.Errorf("UploadVideoInit upload video QueryUploadVideoByName failed: %v", err)
		return nil, err
	}
	if uploadVideoEntity != nil {
		return dto.NewUpadVideoDto(uploadVideoEntity, uploadChunkEntity), nil
	}
	uploadVideoEntity = entity.DefaultUploadVideoEntity(req.UserUUID,
		req.FileName,
		req.FileSize,
		req.FileHash,
		req.TotalChunks,
		0, vo.UploadVideoStatusInit,
		"", nil)
	storagePath := u.minioService.GenerateStoragePath(ctx, vo.NewGenerateStoragePathVO(
		req.UserUUID,
		uploadVideoEntity.UploadVideoUUID(),
		req.FileName,
	))

	storageChunkPath := u.minioService.GenerateChunkStoragePath(ctx, uploadVideoEntity.UploadVideoUUID())
	uploadVideoEntity = uploadVideoEntity.SetStoragePath(storagePath).SetChunkStoragePath(storageChunkPath)
	uploadChunkEntityArr := make([]*entity.UploadChunkEntity, 0, uploadVideoEntity.TotalChunks())

	for i := 0; i < uploadVideoEntity.TotalChunks(); i++ {
		curChunkPath := fmt.Sprintf(storageChunkPath+"%d", i)
		uploadChunkEntityArr = append(uploadChunkEntityArr, entity.DefaultUploadChunkEntity(
			uploadVideoEntity.UploadVideoUUID(), i, "", 0, curChunkPath, nil, vo.UploadChunkStatusInitialized,
		))
	}
	if err = u.uploadVideoRepo.CreateUploadVideoAndChunks(ctx, uploadVideoEntity, uploadChunkEntityArr); err != nil {
		log.Errorf("UploadVideoInit CreateUploadVideoAndChunks error: %v", err)
		return nil, err
	}
	return dto.NewUpadVideoDto(uploadVideoEntity, nil), nil
}

func (u *uploadVideoAppImpl) UploadVideoChunk(ctx context.Context, req *cqe.UploadChunkReq) (*dto.UploadVideoChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// TODO 调用user服务检查用户ID是否存在 通过grpc吧
	return u.uploadVideoSrv.UploadChunk(ctx, req)
}

func (u *uploadVideoAppImpl) MergeChunks(ctx context.Context, req *cqe.MergeChunkReq) (*dto.MergeChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// TODO 调用user服务检查用户ID是否存在，grpc调用
	return u.uploadVideoSrv.MergeChunk(ctx, req)
}

func (u *uploadVideoAppImpl) QueryStoragePath(ctx context.Context, req *cqe.UploadVideoStoragePathReq) (*dto.UploadVideoStoragePathDto, error) {
	storagePath, err := u.uploadVideoRepo.QueryByStoragePath(ctx, req.UserUUID, req.ChunkUUID)
	if err != nil {
		return nil, err
	}
	return &dto.UploadVideoStoragePathDto{
		StoragePath: storagePath,
	}, nil
}
