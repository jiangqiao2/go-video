package service

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"upload-service/ddd/adapter/task"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/event"
	rustfsInfra "upload-service/ddd/infrastructure/rustfs"
	"upload-service/pkg/errno"
	"upload-service/pkg/logger"
)

// VideoPublishService defines domain logic for publishing videos.
type VideoPublishService interface {
	PublishVideo(ctx context.Context, cmd *cqe.PublishVideoReq) (*entity.VideoEntity, *entity.UploadVideoEntity, error)
	UpdateVideoTranscodeInfo(ctx context.Context, videoUUID string, status vo.VideoStatus, videoURL string, transcodeTaskUUID string, errorMessage string, publishedAt *time.Time) error
	ListVideos(ctx context.Context, userUUID string, status string, offset, limit int) ([]*entity.VideoEntity, int64, error)
	ListPublicVideos(ctx context.Context, status string, offset, limit int) ([]*entity.VideoEntity, int64, error)
}

type videoPublishServiceImpl struct {
	videoRepo       repo.VideoRepository
	uploadVideoRepo repo.UploadVideoRepository
	eventPublisher  event.VideoEventPublisher
	minioSrv        gateway.MinioService
}

// NewVideoPublishService builds a VideoPublishService with default repositories.
func NewVideoPublishService() VideoPublishService {
	return &videoPublishServiceImpl{
		videoRepo:       persistence.NewVideoRepository(),
		uploadVideoRepo: persistence.NewUploadVideoRepository(),
		eventPublisher:  event.DefaultVideoEventPublisher(),
		minioSrv:        rustfsInfra.DefaultRustFSService(),
	}
}

func (s *videoPublishServiceImpl) PublishVideo(ctx context.Context, cmd *cqe.PublishVideoReq) (*entity.VideoEntity, *entity.UploadVideoEntity, error) {
	uploadVideoEntity, err := s.uploadVideoRepo.QueryByUserAndUUID(ctx, cmd.UploadVideoUUID, cmd.UserUUID)
	if err != nil {
		return nil, nil, err
	}
	if uploadVideoEntity == nil {
		return nil, nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload video not found")
	}
	if !uploadVideoEntity.Status().IsSuccess() {
		uploadVideoEntity, err = s.mergeUploadVideo(ctx, uploadVideoEntity)
		if err != nil {
			return nil, nil, err
		}
		if !uploadVideoEntity.Status().IsSuccess() {
			return nil, nil, errno.NewSimpleBizError(errno.ErrUploadVideoNotReady, nil)
		}
	}

	videoEntity := entity.DefaultVideoEntity(
		cmd.UserUUID,
		cmd.UploadVideoUUID,
		cmd.Title,
		cmd.Description,
		cmd.CoverURL,
		cmd.Tags,
		vo.VideoStatusProcessing,
	)

	if err := s.videoRepo.Create(ctx, videoEntity); err != nil {
		log.Errorf("PublishVideo create video failed: %v", err)
		return nil, nil, err
	}

	return videoEntity, uploadVideoEntity, nil
}

func (s *videoPublishServiceImpl) mergeUploadVideo(ctx context.Context, uploadVideoEntity *entity.UploadVideoEntity) (*entity.UploadVideoEntity, error) {
	if uploadVideoEntity == nil {
		return nil, errno.NewSimpleBizError(errno.ErrUploadIllegal, nil, "upload video not found")
	}
	if uploadVideoEntity.Status().IsSuccess() {
		return uploadVideoEntity, nil
	}

	count, err := s.uploadVideoRepo.CountChunkByUploadVideoUUID(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadChunkStatusCompleted.Value())
	if err != nil {
		return nil, err
	}
	if int(count) != uploadVideoEntity.TotalChunks() {
		return nil, errno.NewSimpleBizError(errno.ErrUploadVideoNotReady, nil)
	}
	if !uploadVideoEntity.Status().IsMerging() {
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusMerging)
	}

	err = s.minioSrv.MergeChunk(ctx, vo.NewMergeChunkVo(uploadVideoEntity.StoragePath(), uploadVideoEntity.ChunkStoragePath(), int64(uploadVideoEntity.TotalChunks())))
	if err != nil {
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusFailed)
		return nil, errno.NewBizError(errno.ErrInternalServer, err)
	}
	if err := s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, uploadVideoEntity.UploadVideoUUID(), vo.UploadVideoStatusSuccess); err != nil {
		logger.Errorf("update upload video status to success failed: %v", err)
		return nil, err
	}
	task.EnqueueChunkCleanup(uploadVideoEntity.ChunkStoragePath(), int64(uploadVideoEntity.TotalChunks()))

	updated, err := s.uploadVideoRepo.QueryByUserAndUUID(ctx, uploadVideoEntity.UploadVideoUUID(), uploadVideoEntity.UserUUID())
	if err != nil || updated == nil {
		return uploadVideoEntity, err
	}
	return updated, nil
}

func (s *videoPublishServiceImpl) UpdateVideoTranscodeInfo(ctx context.Context, videoUUID string, status vo.VideoStatus, videoURL string, transcodeTaskUUID string, errorMessage string, publishedAt *time.Time) error {
	videoEntity, err := s.videoRepo.FindByVideoUUID(ctx, videoUUID)
	if err != nil {
		return err
	}
	if videoEntity == nil {
		return errno.ErrNotFound
	}

	videoEntity.SetStatus(status)
	videoEntity.SetVideoURL(videoURL)
	videoEntity.SetTranscodeTaskUUID(transcodeTaskUUID)
	videoEntity.SetErrorMessage(errorMessage)

	var effectivePublishedAt *time.Time
	if publishedAt != nil {
		effectivePublishedAt = publishedAt
	} else if status.IsPublished() {
		now := time.Now().UTC()
		effectivePublishedAt = &now
	}
	videoEntity.SetPublishedAt(effectivePublishedAt)

	if err := s.videoRepo.UpdateVideoTranscodeInfo(ctx, videoUUID, status, videoURL, transcodeTaskUUID, errorMessage, effectivePublishedAt); err != nil {
		return err
	}

	if status.IsPublished() {
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, videoEntity.UploadVideoUUID(), vo.UploadVideoStatusSuccess)
	} else if status.Value() == vo.VideoStatusFailed.Value() {
		_ = s.uploadVideoRepo.UpdateUploadVideoStatus(ctx, videoEntity.UploadVideoUUID(), vo.UploadVideoStatusFailed)
	}

	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishStatusChanged(ctx, videoEntity); err != nil {
			logger.Warnf("发布视频状态事件失败 video_uuid=%s status=%s error=%s", videoUUID, status.Value(), err.Error())
		}
	}

	return nil
}

func (s *videoPublishServiceImpl) ListVideos(ctx context.Context, userUUID string, status string, offset, limit int) ([]*entity.VideoEntity, int64, error) {
	return s.videoRepo.ListByUser(ctx, userUUID, status, offset, limit)
}

func (s *videoPublishServiceImpl) ListPublicVideos(ctx context.Context, status string, offset, limit int) ([]*entity.VideoEntity, int64, error) {
	return s.videoRepo.ListByStatus(ctx, status, offset, limit)
}
