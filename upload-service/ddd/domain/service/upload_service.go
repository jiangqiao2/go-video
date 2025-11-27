package service

import (
	"context"
	"fmt"
	"time"
	"upload-service/ddd/adapter/task"

	log "github.com/sirupsen/logrus"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
	"upload-service/pkg/errno"
)

type UploadVideoService interface {
	UploadVideoInit(ctx context.Context, cmd *cqe.UploadVideoInitReq) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error)
	MergeChunk(ctx context.Context, cmd *cqe.MergeChunkReq) (*dto.MergeChunkDto, error)
	PresignChunk(ctx context.Context, cmd *cqe.PresignChunkReq) (*vo.PresignChunkResult, error)
	CompleteChunk(ctx context.Context, cmd *cqe.CompleteChunkReq) (*vo.CompleteChunkResult, error)
}

type uploadServiceImpl struct {
	uploadVideoRepo repo.UploadVideoRepository
	minioSrv        gateway.MinioService
}

func NewUploadVideoService() UploadVideoService {
	return &uploadServiceImpl{
		uploadVideoRepo: persistence.NewUploadVideoRepository(),
		minioSrv:        rustfsInfra.DefaultRustFSService(),
	}
}

func (s *uploadServiceImpl) UploadVideoInit(ctx context.Context, cmd *cqe.UploadVideoInitReq) (*entity.UploadVideoEntity, []*entity.UploadChunkEntity, error) {
	if cmd == nil {
		return nil, nil, errno.ErrParameterInvalid
	}
	if err := cmd.Validate(); err != nil {
		return nil, nil, err
	}

	uploadEntity, uploadChunkEntity, err := s.uploadVideoRepo.QueryByUploadVideoFileHash(ctx, &repo.UploadVideoHashQuery{
		UserUUID: cmd.UserUUID,
		FileName: cmd.FileName,
		FileHash: cmd.FileHash,
		FileSize: cmd.FileSize,
	})
	if err != nil {
		return nil, nil, err
	}

	// 命中历史记录，走断点续传
	if uploadEntity != nil {
		s.refreshChunkPresign(ctx, uploadEntity, uploadChunkEntity)
		return uploadEntity, uploadChunkEntity, nil
	}

	uploadEntity = entity.DefaultUploadVideoEntity(
		cmd.UserUUID,
		cmd.FileName,
		cmd.FileSize,
		cmd.FileHash,
		cmd.TotalChunks,
		0,
		vo.UploadVideoStatusInit,
		"",
		nil,
	)

	// 生成合并后存储路径以及分片前缀
	storagePath := s.minioSrv.GenerateStoragePath(ctx, vo.NewGenerateStoragePathVO(
		cmd.UserUUID,
		uploadEntity.UploadVideoUUID(),
		cmd.FileName,
	))
	chunkStoragePath := s.minioSrv.GenerateChunkStoragePath(ctx, uploadEntity.UploadVideoUUID())
	uploadEntity = uploadEntity.SetStoragePath(storagePath).SetChunkStoragePath(chunkStoragePath)

	uploadChunkEntityArr := make([]*entity.UploadChunkEntity, 0, uploadEntity.TotalChunks())
	for i := 0; i < uploadEntity.TotalChunks(); i++ {
		curChunkPath := fmt.Sprintf("%s%d", chunkStoragePath, i)
		uploadChunkEntityArr = append(uploadChunkEntityArr, entity.DefaultUploadChunkEntity(
			uploadEntity.UploadVideoUUID(), i, "", 0, curChunkPath, nil, vo.UploadChunkStatusInitialized,
		))
	}
	if err = s.uploadVideoRepo.CreateUploadVideoAndChunks(ctx, uploadEntity, uploadChunkEntityArr); err != nil {
		log.Errorf("UploadVideoInit CreateUploadVideoAndChunks error: %v", err)
		return nil, nil, err
	}

	s.refreshChunkPresign(ctx, uploadEntity, uploadChunkEntityArr)
	return uploadEntity, uploadChunkEntityArr, nil
}

func (s *uploadServiceImpl) refreshChunkPresign(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity, chunks []*entity.UploadChunkEntity) {
	if uploadVideoEntity == nil || len(chunks) == 0 {
		return
	}
	const bucket = "uploads"
	for _, ch := range chunks {
		if ch == nil {
			continue
		}
		if ch.Status().IsCompleted() {
			continue
		}
		key := ch.StoragePath()
		if key == "" {
			key = fmt.Sprintf("%s%d", uploadVideoEntity.ChunkStoragePath(), ch.ChunkIndex())
		}
		if key == "" {
			continue
		}
		expiredAt := ch.PresignExpiredAt()
		if ch.PutURL() != "" && expiredAt != nil && expiredAt.After(time.Now().Add(1*time.Minute)) {
			// 现有链接仍有效，跳过重新生成
			continue
		}
		putURL, err := s.minioSrv.PresignPutURL(ctx, bucket, key, 15*time.Minute)
		if err != nil {
			log.Errorf("refresh presign failed: %v", err)
			continue
		}
		newExpiredAt := time.Now().Add(15 * time.Minute)
		if err := s.uploadVideoRepo.UpdateUploadChunkPresign(ctx, ch.ChunkUUID(), putURL, newExpiredAt); err != nil {
			log.Errorf("update chunk presign failed: %v", err)
			continue
		}
		ch.SetPresign(putURL, newExpiredAt)
	}
}

// AttachPresignForChunks 将实体中的直传信息映射到 DTO，供上层返回给前端
func AttachPresignForChunks(uploadVideoEntity *entity.UploadVideoEntity, res *dto.UploadVideoDto, chunks []*entity.UploadChunkEntity) {
	if uploadVideoEntity == nil || res == nil || len(res.UploadChunks) == 0 {
		return
	}
	entities := make(map[string]*entity.UploadChunkEntity, len(chunks))
	for _, c := range chunks {
		if c == nil {
			continue
		}
		entities[c.ChunkUUID()] = c
	}
	for i := range res.UploadChunks {
		ch := &res.UploadChunks[i]
		ent := entities[ch.ChunkUUID]
		key := ""
		if ent != nil {
			key = ent.StoragePath()
		}
		if key == "" {
			key = fmt.Sprintf("%s%d", uploadVideoEntity.ChunkStoragePath(), ch.ChunkIndex)
		}
		ch.StoragePath = key
		if ch.Status == vo.UploadChunkStatusCompleted.Value() || key == "" {
			continue
		}
		if ent != nil {
			ch.PutURL = ent.PutURL()
			if exp := ent.PresignExpiredAt(); exp != nil {
				remaining := int(exp.Sub(time.Now()).Seconds())
				if remaining < 0 {
					remaining = 0
				}
				ch.ExpiresInSec = remaining
			}
		}
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

	if uploadVideoEntity.Status().IsSuccess() {
		return &dto.MergeChunkDto{
			Status:          "success",
			UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		}, nil
	}
	if uploadVideoEntity.Status().IsMerging() {
		return &dto.MergeChunkDto{
			Status:          "processing",
			UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		}, nil
	}

	err = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusMerging)
	if err != nil {
		log.Errorf("MergeChunk update status to merging failed: %v", err)
		return nil, err
	}

	task.EnqueueMergeTask(uploadVideoEntity.UploadVideoUUID())

	return &dto.MergeChunkDto{
		Status:          "processing",
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
	}, nil
}

func (s *uploadServiceImpl) PresignChunk(ctx context.Context, cmd *cqe.PresignChunkReq) (*vo.PresignChunkResult, error) {
	uploadVideoEntity, err := s.uploadVideoRepo.QueryByUserAndUUID(ctx, cmd.UploadVideoUUID, cmd.UserUUID)
	if err != nil {
		return nil, err
	}
	if uploadVideoEntity == nil {
		return nil, errno.ErrUploadIllegal
	}
	chunkEntity, err := s.uploadVideoRepo.QueryUploadVideoByChunkUUID(ctx, &repo.UploadChunkCheckQuery{
		UserUUID:        cmd.UserUUID,
		ChunkUUID:       cmd.ChunkUUID,
		UploadVideoUUID: cmd.UploadVideoUUID,
		ChunkIndex:      cmd.ChunkIndex,
	})
	if err != nil {
		return nil, err
	}
	if chunkEntity == nil {
		return nil, errno.ErrUploadIllegal
	}
	key := chunkEntity.StoragePath()
	if key == "" {
		key = fmt.Sprintf("%s%d", uploadVideoEntity.ChunkStoragePath(), chunkEntity.ChunkIndex())
	}
	if chunkEntity.ChunkIndex() == 0 && uploadVideoEntity.Status() == vo.UploadVideoStatusInit {
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusUploading)
	}
	if err := s.uploadVideoRepo.UpdateUploadChunkStatus(ctx, chunkEntity.ChunkUUID(), vo.UploadChunkStatusUploading); err != nil {
		return nil, err
	}
	var putURL string
	if chunkEntity.Status().IsCompleted() {
		putURL = ""
	} else {
		putURL, err = s.minioSrv.PresignPutURL(ctx, "uploads", key, 15*time.Minute)
		if err != nil {
			return nil, errno.NewBizError(errno.ErrInternalServer, err)
		}
	}
	return &vo.PresignChunkResult{
		UploadVideoUUID: uploadVideoEntity.UploadVideoUUID(),
		ChunkUUID:       chunkEntity.ChunkUUID(),
		ChunkIndex:      chunkEntity.ChunkIndex(),
		Bucket:          "uploads",
		Key:             key,
		PutURL:          putURL,
		ExpiresSeconds:  900,
	}, nil
}

func (s *uploadServiceImpl) CompleteChunk(ctx context.Context, cmd *cqe.CompleteChunkReq) (*vo.CompleteChunkResult, error) {
	uploadVideoEntity, err := s.uploadVideoRepo.QueryByUserAndUUID(ctx, cmd.UploadVideoUUID, cmd.UserUUID)
	if err != nil {
		return nil, err
	}
	if uploadVideoEntity == nil {
		return nil, errno.ErrUploadIllegal
	}
	chunkEntity, err := s.uploadVideoRepo.QueryUploadVideoByChunkUUID(ctx, &repo.UploadChunkCheckQuery{
		UserUUID:        cmd.UserUUID,
		ChunkUUID:       cmd.ChunkUUID,
		UploadVideoUUID: cmd.UploadVideoUUID,
		ChunkIndex:      cmd.ChunkIndex,
	})
	if err != nil {
		return nil, err
	}
	if chunkEntity == nil {
		return nil, errno.ErrUploadIllegal
	}
	key := chunkEntity.StoragePath()
	if key == "" {
		key = fmt.Sprintf("%s%d", uploadVideoEntity.ChunkStoragePath(), chunkEntity.ChunkIndex())
	}
	size, err := s.minioSrv.HeadObject(ctx, "uploads", key)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrUploadIllegal, err)
	}
	if size != int64(cmd.ChunkSize) {
		return nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "chunk size mismatch")
	}
	if !chunkEntity.Status().IsCompleted() {
		if err := s.uploadVideoRepo.MarkChunkCompleted(ctx, chunkEntity.ChunkUUID(), cmd.ChunkHash, int(size)); err != nil {
			return nil, err
		}
	}

	// 所有分片完成后自动进入合并
	count, err := s.uploadVideoRepo.CountChunkByUploadVideoUUID(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadChunkStatusCompleted.Value())
	if err != nil {
		return nil, err
	}
	if int(count) == uploadVideoEntity.TotalChunks() {
		if uploadVideoEntity.Status().IsSuccess() {
			return &vo.CompleteChunkResult{Status: vo.UploadChunkStatusCompleted.Value()}, nil
		}
		if !uploadVideoEntity.Status().IsMerging() {
			_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusMerging)
		}
		return &vo.CompleteChunkResult{Status: vo.UploadChunkStatusCompleted.Value()}, nil
	}

	return &vo.CompleteChunkResult{Status: vo.UploadChunkStatusCompleted.Value()}, nil
}
