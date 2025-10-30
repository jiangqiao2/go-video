package app

import (
	"context"
	"time"

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
}

type videoAppImpl struct {
	videoService             service.VideoPublishService
	userServiceClient        *grpcClient.UserServiceClient
	transcodeServiceClient   *grpcClient.TranscodeServiceClient
	pollInterval             time.Duration
}

// DefaultVideoApp constructs a VideoApp with default infrastructure dependencies.
func DefaultVideoApp() VideoApp {
	return &videoAppImpl{
		videoService:             service.NewVideoPublishService(),
		userServiceClient:        grpcClient.DefaultUserServiceClient(),
		transcodeServiceClient:   grpcClient.DefaultTranscodeServiceClient(),
		pollInterval:             5 * time.Second,
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
		log.Errorf("UpdateVideoTranscodeInfo failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	videoEntity.SetTranscodeTaskUUID(taskUUID)
	videoEntity.SetStatus(vo.VideoStatusProcessing)

	go a.trackTranscodeTask(context.Background(), videoEntity.VideoUUID(), taskUUID)

	return dto.NewVideoDetailDto(videoEntity), nil
}

func (a *videoAppImpl) trackTranscodeTask(ctx context.Context, videoUUID, taskUUID string) {
	if a.transcodeServiceClient == nil {
		return
	}
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := a.transcodeServiceClient.GetTranscodeTask(ctx, &transcodepb.GetTranscodeTaskRequest{
				TaskUuid: taskUUID,
			})
			if err != nil {
				log.Errorf("GetTranscodeTask failed: %v", err)
				continue
			}
			if !resp.GetSuccess() {
				log.Warnf("GetTranscodeTask returned unsuccessful: %s", resp.GetErrorMessage())
				continue
			}
			status := resp.GetStatus()
			switch status {
			case "completed":
				publishedAt := time.Now()
				if err := a.videoService.UpdateVideoTranscodeInfo(context.Background(), videoUUID, vo.VideoStatusPublished, resp.GetOutputPath(), taskUUID, "", &publishedAt); err != nil {
					log.Errorf("UpdateVideoTranscodeInfo (completed) failed: %v", err)
				}
				return
			case "failed":
				if err := a.videoService.UpdateVideoTranscodeInfo(context.Background(), videoUUID, vo.VideoStatusFailed, "", taskUUID, resp.GetErrorMessage(), nil); err != nil {
					log.Errorf("UpdateVideoTranscodeInfo (failed) failed: %v", err)
				}
				return
			default:
				// continue polling for pending/processing
			}
		}
	}
}
