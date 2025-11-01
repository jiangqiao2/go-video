package app

import (
	"context"
	"time"
	"upload-service/pkg/logger"

	log "github.com/sirupsen/logrus"

	transcodepb "go-vedio-1/proto/transcode"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/errno"
)

// VideoApp exposes video publishing application services.
type VideoApp interface {
	PublishVideo(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDetailDto, error)
	ListUserVideos(ctx context.Context, req *cqe.ListVideosReq) (*dto.VideoListDto, error)
}

type videoAppImpl struct {
	videoService           service.VideoPublishService
	userServiceClient      *grpcClient.UserServiceClient
	transcodeServiceClient *grpcClient.TranscodeServiceClient
	pollInterval           time.Duration
}

// DefaultVideoApp constructs a VideoApp with default infrastructure dependencies.
func DefaultVideoApp() VideoApp {
	return &videoAppImpl{
		videoService:           service.NewVideoPublishService(),
		userServiceClient:      grpcClient.DefaultUserServiceClient(),
		transcodeServiceClient: grpcClient.DefaultTranscodeServiceClient(),
		pollInterval:           5 * time.Second,
	}
}

func (a *videoAppImpl) PublishVideo(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDetailDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}

	userExists, err := a.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		log.Errorf("PublishVideo ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		log.Warnf("PublishVideo user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

	videoEntity, uploadVideoEntity, err := a.videoService.PublishVideo(ctx, req)
	if err != nil {
		return nil, err
	}

	createResp, err := a.transcodeServiceClient.CreateTranscodeTask(ctx, &transcodepb.CreateTranscodeTaskRequest{
		UserUuid:         req.UserUUID,
		VideoUuid:        videoEntity.VideoUUID(),
		InputPath:        uploadVideoEntity.StoragePath(),
		TargetResolution: req.TargetResolution,
		TargetBitrate:    req.TargetBitrate,
	})
	if err != nil {
		log.Errorf("CreateTranscodeTask failed: %v", err)
		_ = a.videoService.UpdateVideoTranscodeInfo(ctx, videoEntity.VideoUUID(), vo.VideoStatusFailed, "", "", err.Error(), nil)
		return nil, errno.ErrInternalServer
	}
	if !createResp.GetSuccess() {
		errMsg := createResp.GetMessage()
		if errMsg == "" {
			errMsg = "transcode service rejected task"
		}
		_ = a.videoService.UpdateVideoTranscodeInfo(ctx, videoEntity.VideoUUID(), vo.VideoStatusFailed, "", "", errMsg, nil)
		return nil, errno.ErrInternalServer
	}

	taskUUID := createResp.GetTaskUuid()
	if err := a.videoService.UpdateVideoTranscodeInfo(ctx, videoEntity.VideoUUID(), vo.VideoStatusProcessing, "", taskUUID, "", nil); err != nil {
		logger.Errorf("UpdateVideoTranscodeInfo failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	videoEntity.SetTranscodeTaskUUID(taskUUID)
	videoEntity.SetStatus(vo.VideoStatusProcessing)

	return dto.NewVideoDetailDto(videoEntity), nil
}

func (a *videoAppImpl) ListUserVideos(ctx context.Context, req *cqe.ListVideosReq) (*dto.VideoListDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.Size
	videos, total, err := a.videoService.ListVideos(ctx, req.UserUUID, req.Status, offset, req.Size)
	if err != nil {
		logger.Errorf("ListUserVideos failed: %v", err)
		return nil, errno.ErrInternalServer
	}

	return dto.NewVideoListDto(videos, total, req.Page, req.Size), nil
}
