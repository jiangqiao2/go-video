package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
	"upload-service/pkg/config"
	"upload-service/pkg/errno"
	"upload-service/pkg/logger"

	log "github.com/sirupsen/logrus"
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
	QueryUploadStatus(ctx context.Context, req *cqe.UploadVideoStatusReq) (*dto.UploadVideoStatusDto, error)
	PresignImage(ctx context.Context, req *cqe.PresignImageReq) (*dto.PresignImageDto, error)
	UploadImage(ctx context.Context, userUUID, fileName, category, contentType string, reader io.Reader, size int64) (*dto.UploadImageDto, error)
}

type uploadVideoAppImpl struct {
	minioService      gateway.MinioService
	uploadVideoRepo   repo.UploadVideoRepository
	uploadVideoSrv    service.UploadVideoService
	userServiceClient *grpcClient.UserServiceClient
}

func DefaultUploadVideoApp() UploadVideoApp {
	return &uploadVideoAppImpl{
		minioService:      rustfsInfra.DefaultRustFSService(),
		uploadVideoRepo:   persistence.NewUploadVideoRepository(),
		uploadVideoSrv:    service.NewUploadVideoService(),
		userServiceClient: grpcClient.DefaultUserServiceClient(),
	}
}

func (u *uploadVideoAppImpl) UploadVideoInit(ctx context.Context, req *cqe.UploadVideoInitReq) (*dto.UploadVideoDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用user服务检查用户ID是否存在
	userExists, err := u.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		log.Errorf("UploadVideoInit ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		log.Warnf("UploadVideoInit user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

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
	return dto.NewUpadVideoDto(uploadVideoEntity, uploadChunkEntityArr), nil
}

func (u *uploadVideoAppImpl) UploadVideoChunk(ctx context.Context, req *cqe.UploadChunkReq) (*dto.UploadVideoChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用user服务检查用户ID是否存在
	userExists, err := u.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		log.Errorf("UploadVideoChunk ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		log.Warnf("UploadVideoChunk user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

	return u.uploadVideoSrv.UploadChunk(ctx, req)
}

func (u *uploadVideoAppImpl) MergeChunks(ctx context.Context, req *cqe.MergeChunkReq) (*dto.MergeChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 调用user服务检查用户ID是否存在
	userExists, err := u.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		log.Errorf("MergeChunks ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		log.Warnf("MergeChunks user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

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

func (u *uploadVideoAppImpl) QueryUploadStatus(ctx context.Context, req *cqe.UploadVideoStatusReq) (*dto.UploadVideoStatusDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	uploadVideoEntity, err := u.uploadVideoRepo.QueryByUserAndUUID(ctx, req.UploadVideoUUID, req.UserUUID)
	if err != nil {
		return nil, err
	}
	if uploadVideoEntity == nil {
		return nil, errno.ErrNotFound
	}
	return &dto.UploadVideoStatusDto{
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		Status:          uploadVideoEntity.Status().Value(),
	}, nil
}

func (u *uploadVideoAppImpl) PresignImage(ctx context.Context, req *cqe.PresignImageReq) (*dto.PresignImageDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	key := u.minioService.GenerateImagePath(ctx, vo.NewGenerateImagePathVO(strings.TrimSpace(req.UserUUID), req.FileName, req.Category))
	putURL, err := u.minioService.PresignPutURL(ctx, "image", key, time.Duration(req.ExpiresSeconds)*time.Second)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrInternalServer, err)
	}
	return &dto.PresignImageDto{Bucket: "image", Key: key, PutURL: putURL}, nil
}

func (u *uploadVideoAppImpl) UploadImage(ctx context.Context, userUUID, fileName, category, contentType string, reader io.Reader, size int64) (*dto.UploadImageDto, error) {
	if fileName == "" || size <= 0 {
		return nil, errno.ErrFileNameIllegal
	}
	key := u.minioService.GenerateImagePath(ctx, vo.NewGenerateImagePathVO(userUUID, fileName, category))
	bucket := "image"
	err := u.minioService.UploadChunk(ctx, vo.NewMinIoUploadChunkVo(key, bucket, reader, size, contentType))
	if err != nil {
		logger.Errorf("UploadImage UploadChunk error :%v", err)
		return nil, err
	}
	cfg := config.GetGlobalConfig()
	base := strings.TrimSpace(cfg.Public.StorageBase)
	path := "/storage/" + bucket + "/" + strings.TrimLeft(key, "/")
	if base != "" {
		if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
			base = "http://" + strings.TrimRight(base, "/")
		}
		base = strings.TrimRight(base, "/")
		path = base + path
	}
	url := path
	return &dto.UploadImageDto{Bucket: bucket, Key: key, URL: url}, nil
}
