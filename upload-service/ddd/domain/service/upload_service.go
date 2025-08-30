package service

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/minio"
	"upload-service/pkg/errno"
)

type UploadVideoService interface {
	UploadChunk(ctx context.Context, cmd *cqe.UploadChunkReq) error
}

type uploadServiceImpl struct {
	uploadVideoRepo repo.UploadVideoRepository
	minioSrv        gateway.MinioService
}

func NewUploadVideoService() UploadVideoService {
	return &uploadServiceImpl{
		uploadVideoRepo: persistence.NewUploadVideoRepository(),
		minioSrv:        minio.DefaultMinioService(),
	}
}

func (s *uploadServiceImpl) checkUploadChunk(ctx context.Context, cmd *cqe.UploadChunkReq) (*entity.UploadVideoEntity, *entity.UploadChunkEntity, error) {
	// 检查UploadVideo是否存在
	uploadVideoEntity, err := s.uploadVideoRepo.QueryUploadVideoByUUID(ctx, cmd.UploadVideoUUID)
	if err != nil {
		return nil, nil, err
	}
	if uploadVideoEntity == nil {
		return nil, nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload video is illegal")
	}
	// 检查UploadChunk是否合法
	uploadChunkEntity, err := s.uploadVideoRepo.QueryUploadVideoByChunkUUID(ctx, &repo.UploadChunkCheckQuery{
		UploadVideoUUID: cmd.UploadVideoUUID,
		UserUUID:        cmd.UserUUID,
		ChunkUUID:       cmd.ChunkUUID,
		ChunkIndex:      cmd.ChunkIndex,
	})
	if err != nil {
		return nil, nil, err
	}
	if uploadChunkEntity == nil {
		return nil, nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload chunk is illegal")
	}
	if uploadChunkEntity.Status().IsUploading() || uploadChunkEntity.Status().IsCompleted() {
		return nil, nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload chunk is loding")
	}
	return uploadVideoEntity, uploadChunkEntity, nil
}

func (s *uploadServiceImpl) UploadChunk(ctx context.Context, cmd *cqe.UploadChunkReq) error {
	_, uploadChunkEntity, err := s.checkUploadChunk(ctx, cmd)
	if err != nil {
		return err
	}

	// 更新状态为上传中
	err = s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, uploadChunkEntity.ChunkUUID(), vo.UploadChunkStatusUploading)
	if err != nil {
		return err
	}

	// 创建分片数据读取器
	reader := bytes.NewReader(cmd.ChunkData)

	// 上传分片到MinIO
	err = s.minioSrv.UploadChunk(ctx, vo.NewMinIoUploadChunkVo(
		uploadChunkEntity.StoragePath(),
		"uploads", // 默认bucket名称
		reader,
		int64(cmd.ChunkSize),
		"application/octet-stream",
	))

	if err != nil {
		// 上传失败，更新状态为失败
		updateErr := s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, uploadChunkEntity.ChunkUUID(), vo.UploadChunkStatusFailed)
		if updateErr != nil {
			log.Errorf("failed to update upload chunk status failed, err:%v", updateErr)
		}
		return err
	}

	// 上传成功，更新状态为完成
	err = s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, uploadChunkEntity.ChunkUUID(), vo.UploadChunkStatusCompleted)
	if err != nil {
		return err
	}

	return nil
}
