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
		logger.Error("video service not initialised for gRPC server", nil)
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "service unavailable",
		}, nil
	}

	videoUUID := strings.TrimSpace(req.GetVideoUuid())
	if videoUUID == "" {
		logger.Warn("UpdateTranscodeStatus called with empty video_uuid", nil)
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "video_uuid is required",
		}, nil
	}

	statusValue := strings.TrimSpace(req.GetStatus())
	if statusValue == "" {
		logger.Warnf("UpdateTranscodeStatus called with empty status %v", map[string]interface{}{
			"video_uuid": videoUUID,
		})
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "status is required",
		}, nil
	}

	status := vo.NewVideoStatus(statusValue)
	if status.Value() != statusValue {
		logger.Warnf("UpdateTranscodeStatus received invalid status %v", map[string]interface{}{
			"video_uuid": videoUUID,
			"status":     statusValue,
		})
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
		logger.Errorf("UpdateVideoTranscodeInfo failed %v", map[string]interface{}{
			"video_uuid": videoUUID,
			"task_uuid":  taskUUID,
			"status":     statusValue,
			"video_url":  videoURL,
			"error":      err.Error(),
			"error_msg":  errorMessage,
		})
		return &uploadpb.UpdateTranscodeStatusResponse{
			Success: false,
			Message: "failed to update video status",
		}, nil
	}

	logger.Infof("Video transcode status updated via gRPC %v", map[string]interface{}{
		"video_uuid": videoUUID,
		"task_uuid":  taskUUID,
		"status":     statusValue,
	})

	return &uploadpb.UpdateTranscodeStatusResponse{
		Success: true,
		Message: "video status updated",
	}, nil
}
