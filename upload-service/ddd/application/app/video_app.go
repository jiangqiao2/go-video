package app

import (
	"context"

	log "github.com/sirupsen/logrus"

	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/service"
	grpcClient "upload-service/ddd/infrastructure/grpc"
	"upload-service/pkg/errno"
)

// VideoApp exposes video publishing application services.
type VideoApp interface {
	PublishVideo(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDetailDto, error)
}

type videoAppImpl struct {
	videoService      service.VideoPublishService
	userServiceClient *grpcClient.UserServiceClient
}

// DefaultVideoApp constructs a VideoApp with default infrastructure dependencies.
func DefaultVideoApp() VideoApp {
	return &videoAppImpl{
		videoService:      service.NewVideoPublishService(),
		userServiceClient: grpcClient.DefaultUserServiceClient(),
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

	videoEntity, err := a.videoService.PublishVideo(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.NewVideoDetailDto(videoEntity), nil
}
