package grpc

import (
	"context"
	apppkg "video-service/ddd/application/app"
	cqe "video-service/ddd/application/cqe"
	videopb "video-service/proto/video"
)

type VideoGRPCServer struct {
	videopb.UnimplementedVideoServiceServer
}

func (s *VideoGRPCServer) Precreate(ctx context.Context, in *videopb.PrecreateRequest) (*videopb.PrecreateResponse, error) {
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
		return &videopb.PrecreateResponse{Success: false, Message: err.Error()}, nil
	}
	return &videopb.PrecreateResponse{Success: true, Message: "", VideoUuid: res.VideoUUID}, nil
}

func (s *VideoGRPCServer) UpdateTranscodeResult(ctx context.Context, in *videopb.UpdateTranscodeResultRequest) (*videopb.UpdateTranscodeResultResponse, error) {
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
		return &videopb.UpdateTranscodeResultResponse{Success: false, Message: err.Error()}, nil
	}
	return &videopb.UpdateTranscodeResultResponse{Success: true, Message: ""}, nil
}
