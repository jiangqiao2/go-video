package app

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
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
	QueryStoragePath(ctx context.Context, req *cqe.UploadVideoStoragePathReq) (*dto.UploadVideoStoragePathDto, error)
	MergeChunks(ctx context.Context, req *cqe.MergeChunkReq) (*dto.MergeChunkDto, error)
	QueryUploadStatus(ctx context.Context, req *cqe.UploadVideoStatusReq) (*dto.UploadVideoStatusDto, error)
	PresignImage(ctx context.Context, req *cqe.PresignImageReq) (*dto.PresignImageDto, error)
	UploadImage(ctx context.Context, userUUID, fileName, category, contentType string, reader io.Reader, size int64) (*dto.UploadImageDto, error)
	PresignChunk(ctx context.Context, req *cqe.PresignChunkReq) (*dto.PresignChunkDto, error)
	CompleteChunk(ctx context.Context, req *cqe.CompleteChunkReq) (*dto.UploadVideoChunkDto, error)
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
		logger.Errorf("UploadVideoInit ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		logger.Warnf("UploadVideoInit user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

	uploadVideoEntity, chunkEntities, err := u.uploadVideoSrv.UploadVideoInit(ctx, req)
	if err != nil {
		logger.Errorf("UploadVideoInit domain service failed: %v", err)
		return nil, err
	}
	res := dto.NewUpadVideoDto(uploadVideoEntity, chunkEntities)
	service.AttachPresignForChunks(uploadVideoEntity, res, chunkEntities)
	return res, nil
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
	// 返回相对路径，前端或调用方自行拼接网关前缀
	url := "/storage/" + bucket + "/" + strings.TrimLeft(key, "/")
	return &dto.UploadImageDto{Bucket: bucket, Key: key, URL: url}, nil
}

func (u *uploadVideoAppImpl) PresignChunk(ctx context.Context, req *cqe.PresignChunkReq) (*dto.PresignChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	userExists, err := u.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		return nil, errno.ErrNotFound
	}
	res, err := u.uploadVideoSrv.PresignChunk(ctx, req)
	if err != nil {
		return nil, err
	}
	return &dto.PresignChunkDto{
		UploadVideoUUID: res.UploadVideoUUID,
		ChunkUUID:       res.ChunkUUID,
		ChunkIndex:      res.ChunkIndex,
		Bucket:          res.Bucket,
		Key:             res.Key,
		PutURL:          res.PutURL,
		ExpiresSeconds:  res.ExpiresSeconds,
	}, nil
}

func (u *uploadVideoAppImpl) CompleteChunk(ctx context.Context, req *cqe.CompleteChunkReq) (*dto.UploadVideoChunkDto, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	userExists, err := u.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		return nil, errno.ErrNotFound
	}
	res, err := u.uploadVideoSrv.CompleteChunk(ctx, req)
	if err != nil {
		return nil, err
	}
	return &dto.UploadVideoChunkDto{Status: res.Status}, nil
}
