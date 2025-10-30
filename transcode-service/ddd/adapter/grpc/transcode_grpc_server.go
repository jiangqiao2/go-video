package grpc

import (
	"context"

	transcodepb "go-vedio-1/proto/transcode"

	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/cqe"
	"transcode-service/pkg/errno"
)

// TranscodeGrpcServer implements the gRPC TranscodeService.
type TranscodeGrpcServer struct {
	app app.TranscodeApp
	transcodepb.UnimplementedTranscodeServiceServer
}

// NewTranscodeGrpcServer creates a new gRPC server implementation.
func NewTranscodeGrpcServer(transcodeApp app.TranscodeApp) *TranscodeGrpcServer {
	return &TranscodeGrpcServer{
		app: transcodeApp,
	}
}

func (s *TranscodeGrpcServer) CreateTranscodeTask(ctx context.Context, req *transcodepb.CreateTranscodeTaskRequest) (*transcodepb.CreateTranscodeTaskResponse, error) {
	if s.app == nil {
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			ErrorMessage: "transcode app not initialized",
		}, nil
	}

	createReq := &cqe.CreateTranscodeTaskReq{
		UserUUID:     req.GetUserUuid(),
		VideoUUID:    req.GetVideoUuid(),
		OriginalPath: req.GetInputPath(),
		Resolution:   req.GetTargetResolution(),
		Bitrate:      req.GetTargetBitrate(),
	}

	taskDto, err := s.app.CreateTranscodeTask(ctx, createReq)
	if err != nil {
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &transcodepb.CreateTranscodeTaskResponse{
		Success:  true,
		TaskUuid: taskDto.TaskUUID,
		Message:  "accepted",
	}, nil
}

func (s *TranscodeGrpcServer) GetTranscodeTask(ctx context.Context, req *transcodepb.GetTranscodeTaskRequest) (*transcodepb.GetTranscodeTaskResponse, error) {
	if s.app == nil {
		return &transcodepb.GetTranscodeTaskResponse{
			Success: false,
			ErrorMessage: "transcode app not initialized",
		}, nil
	}

	taskDto, err := s.app.GetTranscodeTask(ctx, req.GetTaskUuid())
	if err != nil {
		if err == errno.ErrTranscodeTaskNotFound {
			return &transcodepb.GetTranscodeTaskResponse{
				Success:      false,
				ErrorMessage: err.Error(),
			}, nil
		}
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	progress := int32(taskDto.Progress)
	outputPath := taskDto.OutputPath
	errorMessage := taskDto.ErrorMessage
	if taskDto.Status == "completed" && outputPath == "" {
		outputPath = taskDto.OutputPath
	}
	if taskDto.Status == "failed" && errorMessage == "" {
		errorMessage = "transcode failed"
	}

	return &transcodepb.GetTranscodeTaskResponse{
		Success:      true,
		TaskUuid:     taskDto.TaskUUID,
		VideoUuid:    taskDto.VideoUUID,
		Status:       taskDto.Status,
		Progress:     progress,
		OutputPath:   outputPath,
		ErrorMessage: errorMessage,
	}, nil
}
