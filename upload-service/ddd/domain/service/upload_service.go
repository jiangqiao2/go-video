package service

import (
	"bytes"
	"context"
	"upload-service/ddd/adapter/task"

	log "github.com/sirupsen/logrus"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/minio"
	"upload-service/pkg/errno"
	"upload-service/pkg/logger"
)

type UploadVideoService interface {
	UploadChunk(ctx context.Context, cmd *cqe.UploadChunkReq) (*dto.UploadVideoChunkDto, error)
	MergeChunk(ctx context.Context, cmd *cqe.MergeChunkReq) (*dto.MergeChunkDto, error)
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
		log.Errorf("query upload chunk by uuid failed, err:%v", err)
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

func (s *uploadServiceImpl) UploadChunk(ctx context.Context, cmd *cqe.UploadChunkReq) (*dto.UploadVideoChunkDto, error) {
	uploadVideoEntity, uploadChunkEntity, err := s.checkUploadChunk(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	// 如果是第一个分片且上传视频状态为初始化，则更新为上传中
	if cmd.ChunkIndex == 0 && uploadVideoEntity.Status() == vo.UploadVideoStatusInit {
		err = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusUploading)
		if err != nil {
			log.Errorf("UploadChunk update video status to uploading failed: %v", err)
			return nil, err
		}
	}
	
	// 更新状态为上传中
	err = s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, uploadChunkEntity.ChunkUUID(), vo.UploadChunkStatusUploading)
	if err != nil {
		return nil, err
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
			logger.Errorf("failed to update upload chunk status failed, err:%v", updateErr)
		}
		return &dto.UploadVideoChunkDto{
			Status: "failed",
		}, err
	}

	// 上传成功，更新状态为完成
	err = s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, uploadChunkEntity.ChunkUUID(), vo.UploadChunkStatusCompleted)
	if err != nil {
		logger.Errorf("failed to update upload chunk status completed, err:%v", err)
		return nil, err
	}

	return &dto.UploadVideoChunkDto{
		Status: "success",
	}, nil
}

func (s *uploadServiceImpl) checkMergeChunk(ctx context.Context, uploadVideoUUID, userUUID string) (*entity.UploadVideoEntity, error) {
	// 合并次数
	uploadVideoEntity, err := s.uploadVideoRepo.QueryByUserAndUUID(ctx, uploadVideoUUID, userUUID)
	if err != nil {
		return nil, err
	}
	if uploadVideoEntity == nil {
		return nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload video is not exist")
	}
	chunksCount, err := s.uploadVideoRepo.CountChunkByUploadVideoUUID(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadChunkStatusCompleted.Value())
	if err != nil {
		return nil, err
	}
	if chunksCount != (int64(uploadVideoEntity.TotalChunks())) {
		return nil, errno.NewSimpleBizError(errno.ErrChunkIncomplete, nil, "upload chunks is not complete")
	}
	return uploadVideoEntity, nil
}

func (s *uploadServiceImpl) MergeChunk(ctx context.Context, cmd *cqe.MergeChunkReq) (*dto.MergeChunkDto, error) {
	// 查询user_uuid upload_video_uuid 是否存在
	uploadVideoEntity, err := s.checkMergeChunk(ctx, cmd.UploadVideoUUID, cmd.UserUUID)
	if err != nil {
		return nil, err
	}
	
	// 更新上传视频状态为合并中
	err = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusMerging)
	if err != nil {
		log.Errorf("MergeChunk update status to merging failed: %v", err)
		return nil, err
	}
	
	// 合并操作
	err = s.minioSrv.MergeChunk(ctx, vo.NewMergeChunkVo(uploadVideoEntity.StoragePath(), uploadVideoEntity.ChunkStoragePath(), int64(uploadVideoEntity.TotalChunks())))
	if err != nil {
		// 合并失败，更新状态为失败
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusFailed)
		return nil, err
	}

	// 合并成功，更新状态为成功
	err = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusSuccess)
	if err != nil {
		log.Errorf("MergeChunk update status to success failed: %v", err)
		return nil, err
	}

	task.EnqueueChunkCleanup(uploadVideoEntity.ChunkStoragePath(), int64(uploadVideoEntity.TotalChunks()))

	return &dto.MergeChunkDto{
		Status:          "success",
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
	}, nil
}
