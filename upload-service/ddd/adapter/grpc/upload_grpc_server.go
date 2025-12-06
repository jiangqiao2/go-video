package grpc

import (
	"context"
	"strings"

	uploadpb "upload-service/proto/upload"

	"upload-service/ddd/domain/service"
	"upload-service/ddd/domain/vo"
	"upload-service/pkg/logger"
)

// UploadGrpcServer implements uploadpb.UploadServiceServer.
type UploadGrpcServer struct {
	uploadpb.UnimplementedUploadServiceServer
	videoService service.VideoPublishService
}

// NewUploadGrpcServer builds a gRPC server using the provided domain service.
func NewUploadGrpcServer(videoService service.VideoPublishService) *UploadGrpcServer {
	return &UploadGrpcServer{
		videoService: videoService,
	}
}

// UpdateTranscodeStatus updates persisted video metadata based on the transcode result.
func (s *UploadGrpcServer) UpdateTranscodeStatus(ctx context.Context, req *uploadpb.UpdateTranscodeStatusRequest) (*uploadpb.UpdateTranscodeStatusResponse, error) {
	if s.videoService == nil {
		logger.WithContext(ctx).Errorf("video service not initialised for gRPC server")
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "service unavailable",
		}, nil
	}

	videoUUID := strings.TrimSpace(req.GetVideoUuid())
	if videoUUID == "" {
		logger.WithContext(ctx).Warnf("UpdateTranscodeStatus called with empty video_uuid")
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "video_uuid is required",
		}, nil
	}

	statusValue := strings.TrimSpace(req.GetStatus())
	if statusValue == "" {
		logger.WithContext(ctx).Warnf("UpdateTranscodeStatus called with empty status video_uuid=%s", videoUUID)
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "status is required",
		}, nil
	}

	status := vo.NewVideoStatus(statusValue)
	if status.Value() != statusValue {
		logger.WithContext(ctx).Warnf("UpdateTranscodeStatus received invalid status video_uuid=%s status=%s", videoUUID, statusValue)
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "invalid status value",
		}, nil
	}

	videoURL := strings.TrimSpace(req.GetVideoUrl())
	errorMessage := strings.TrimSpace(req.GetErrorMessage())
	taskUUID := strings.TrimSpace(req.GetTranscodeTaskUuid())

	err := s.videoService.UpdateVideoTranscodeInfo(ctx, videoUUID, status, videoURL, taskUUID, errorMessage, nil)
	if err != nil {
		logger.WithContext(ctx).Errorf("UpdateVideoTranscodeInfo failed video_uuid=%s task_uuid=%s status=%s video_url=%s error=%v error_msg=%s", videoUUID, taskUUID, statusValue, videoURL, err, errorMessage)
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "failed to update video status",
		}, nil
	}

	logger.WithContext(ctx).Infof("Video transcode status updated via gRPC video_uuid=%s task_uuid=%s status=%s", videoUUID, taskUUID, statusValue)

	return &uploadpb.UpdateTranscodeStatusResponse{
		Success: true,
		Message: "video status updated",
	}, nil
}
