package service

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/domain/entity"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/event"
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
}

// NewVideoPublishService builds a VideoPublishService with default repositories.
func NewVideoPublishService() VideoPublishService {
	return &videoPublishServiceImpl{
		videoRepo:       persistence.NewVideoRepository(),
		uploadVideoRepo: persistence.NewUploadVideoRepository(),
		eventPublisher:  event.DefaultVideoEventPublisher(),
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
		return nil, nil, errno.NewSimpleBizError(errno.ErrUploadVideoNotReady, nil)
	}

	existingVideo, err := s.videoRepo.FindByUploadVideoUUID(ctx, cmd.UploadVideoUUID)
	if err != nil {
		return nil, nil, err
	}
	if existingVideo != nil {
		return nil, nil, errno.NewSimpleBizError(errno.ErrVideoAlreadyPublished, nil)
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

	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishStatusChanged(ctx, videoEntity); err != nil {
			logger.Warn("发布视频状态事件失败", map[string]interface{}{
				"video_uuid": videoUUID,
				"status":     status.Value(),
				"error":      err.Error(),
			})
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
