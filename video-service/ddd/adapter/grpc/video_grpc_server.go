package grpc

import (
	"context"
	apppkg "video-service/ddd/application/app"
	cqe "video-service/ddd/application/cqe"
	"video-service/pkg/logger"
	videopb "video-service/proto/video"
)

type VideoGRPCServer struct {
	videopb.UnimplementedVideoServiceServer
}

func (s *VideoGRPCServer) Precreate(ctx context.Context, in *videopb.PrecreateRequest) (*videopb.PrecreateResponse, error) {
	logger.WithContext(ctx).Infof("gRPC Precreate received video_uuid=%s upload_video_uuid=%s user_uuid=%s", in.GetVideoUuid(), in.GetUploadVideoUuid(), in.GetUserUuid())
	req := &cqe.PrecreateReq{
		VideoUUID:         in.GetVideoUuid(),
		UploadVideoUUID:   in.GetUploadVideoUuid(),
		UserUUID:          in.GetUserUuid(),
		Title:             in.GetTitle(),
		Description:       in.GetDescription(),
		CoverURL:          in.GetCoverUrl(),
		TranscodeTaskUUID: in.GetTaskUuid(),
	}
	res, err := apppkg.DefaultVideoApp().Precreate(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Warnf("Precreate failed video_uuid=%s error=%v", in.GetVideoUuid(), err)
		return &videopb.PrecreateResponse{Success: false, Message: err.Error()}, nil
	}
	logger.WithContext(ctx).Infof("Precreate success video_uuid=%s", res.VideoUUID)
	return &videopb.PrecreateResponse{Success: true, Message: "", VideoUuid: res.VideoUUID}, nil
}

func (s *VideoGRPCServer) UpdateTranscodeResult(ctx context.Context, in *videopb.UpdateTranscodeResultRequest) (*videopb.UpdateTranscodeResultResponse, error) {
	logger.WithContext(ctx).Infof("gRPC UpdateTranscodeResult received video_uuid=%s task_uuid=%s status=%s url=%s", in.GetVideoUuid(), in.GetTaskUuid(), in.GetStatus(), in.GetVideoUrl())
	dur := int(in.GetDurationSec())
	size := in.GetSizeBytes()
	req := &cqe.UpdateTranscodeResultReq{
		VideoUUID:   in.GetVideoUuid(),
		TaskUUID:    in.GetTaskUuid(),
		Status:      in.GetStatus(),
		VideoURL:    in.GetVideoUrl(),
		ErrorMsg:    in.GetErrorMsg(),
		DurationSec: &dur,
		SizeBytes:   &size,
	}
	_, err := apppkg.DefaultVideoApp().UpdateTranscodeResult(ctx, req)
	if err != nil {
		logger.WithContext(ctx).Warnf("UpdateTranscodeResult failed video_uuid=%s task_uuid=%s error=%v", in.GetVideoUuid(), in.GetTaskUuid(), err)
		return &videopb.UpdateTranscodeResultResponse{Success: false, Message: err.Error()}, nil
	}
	logger.WithContext(ctx).Infof("UpdateTranscodeResult success video_uuid=%s task_uuid=%s", in.GetVideoUuid(), in.GetTaskUuid())
	return &videopb.UpdateTranscodeResultResponse{Success: true, Message: ""}, nil
}
