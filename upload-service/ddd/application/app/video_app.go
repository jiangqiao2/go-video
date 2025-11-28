package app

import (
	"context"
	"time"
	"upload-service/pkg/logger"

	transcodepb "transcode-service/proto/transcode"
	videopb "video-service/proto/video"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/repo"
	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	"upload-service/ddd/infrastructure/database/persistence"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/errno"
)

// VideoApp exposes video publishing application services.
type VideoApp interface {
	PublishVideo(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDetailDto, error)
	ListUserVideos(ctx context.Context, req *cqe.ListVideosReq) (*dto.VideoListDto, error)
	ListOpenVideos(ctx context.Context, req *cqe.ListOpenVideosReq) (*dto.VideoListDto, error)
}

type videoAppImpl struct {
	videoService           service.VideoPublishService
	videoRepo              repo.VideoRepository
	userServiceClient      *grpcClient.UserServiceClient
	transcodeServiceClient *grpcClient.TranscodeServiceClient
	videoServiceClient     *grpcClient.VideoServiceClient
	pollInterval           time.Duration
	userQueryService       service.UserQueryService
}

// DefaultVideoApp constructs a VideoApp with default infrastructure dependencies.
func DefaultVideoApp() VideoApp {
	return &videoAppImpl{
		videoService:           service.NewVideoPublishService(),
		videoRepo:              persistence.NewVideoRepository(),
		userServiceClient:      grpcClient.DefaultUserServiceClient(),
		transcodeServiceClient: grpcClient.DefaultTranscodeServiceClient(),
		videoServiceClient:     grpcClient.DefaultVideoServiceClient(),
		pollInterval:           5 * time.Second,
		userQueryService:       service.NewUserQueryService(),
	}
}

func (a *videoAppImpl) PublishVideo(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDetailDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}

	userExists, err := a.userServiceClient.ValidateUser(ctx, req.UserUUID)
	if err != nil {
		logger.Errorf("PublishVideo ValidateUser failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	if !userExists {
		logger.Warnf("PublishVideo user not found: %s", req.UserUUID)
		return nil, errno.ErrNotFound
	}

	videoEntity, uploadVideoEntity, err := a.videoService.PublishVideo(ctx, req)
	if err != nil {
		return nil, err
	}

	// 预占位：写入 video-service，状态置 processing
	if a.videoServiceClient != nil {
		preReq := &videopb.PrecreateRequest{
			VideoUuid:       videoEntity.VideoUUID(),
			UploadVideoUuid: uploadVideoEntity.UploadVideoUUID(),
			UserUuid:        req.UserUUID,
			Title:           req.Title,
			Description:     req.Description,
			CoverUrl:        req.CoverURL,
		}
		if preResp, preErr := a.videoServiceClient.Precreate(ctx, preReq); preErr != nil {
			logger.Errorf("Precreate video-service failed: %v", preErr)
			return nil, errno.ErrInternalServer
		} else if preResp != nil && !preResp.Success {
			logger.Warnf("Precreate rejected: %s", preResp.Message)
			return nil, errno.ErrInternalServer
		}
	}

	createResp, err := a.transcodeServiceClient.CreateTranscodeTask(ctx, &transcodepb.CreateTranscodeTaskRequest{
		UserUuid:         req.UserUUID,
		VideoUuid:        videoEntity.VideoUUID(),
		InputPath:        uploadVideoEntity.StoragePath(),
		TargetResolution: req.TargetResolution,
		TargetBitrate:    req.TargetBitrate,
	})
	if err != nil {
		logger.Errorf("CreateTranscodeTask failed: %v", err)
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

	videos, total, err := a.videoRepo.ListByUserQ(ctx, &repo.VideoByUserQuery{UserUUID: req.UserUUID, Status: req.Status, Page: req.Page, Size: req.Size})
	if err != nil {
		logger.Errorf("ListUserVideos failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	userMap, _ := a.userQueryService.GetUsersForVideos(ctx, videos)

	items := make([]dto.VideoDetailDto, 0, len(videos))
	for _, v := range videos {
		d := dto.NewVideoDetailDto(v)
		if d == nil {
			continue
		}
		if info, ok := userMap[v.UserUUID()]; ok && info != nil {
			d.UploaderAccount = info.Account
			d.UploaderAvatarURL = info.AvatarUrl
		}
		items = append(items, *d)
	}

	return dto.NewVideoListFromItems(items, total, req.Page, req.Size), nil
}

func (a *videoAppImpl) ListOpenVideos(ctx context.Context, req *cqe.ListOpenVideosReq) (*dto.VideoListDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	videos, total, err := a.videoRepo.ListByStatusQ(ctx, &repo.VideoByStatusQuery{Status: req.Status, Page: req.Page, Size: req.Size})
	if err != nil {
		logger.Errorf("ListOpenVideos failed: %v", err)
		return nil, errno.ErrInternalServer
	}
	userMap, _ := a.userQueryService.GetUsersForVideos(ctx, videos)

	items := make([]dto.VideoDetailDto, 0, len(videos))
	for _, v := range videos {
		d := dto.NewVideoDetailDto(v)
		if d == nil {
			continue
		}
		if info, ok := userMap[v.UserUUID()]; ok && info != nil {
			d.UploaderAccount = info.Account
			d.UploaderAvatarURL = info.AvatarUrl
		}
		items = append(items, *d)
	}

	return dto.NewVideoListFromItems(items, total, req.Page, req.Size), nil
}
