package grpc

import (
	"context"
	"errors"
	"fmt"

	transcodepb "go-vedio-1/proto/transcode"

	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/cqe"
	"transcode-service/pkg/errno"
	"transcode-service/pkg/logger"
)

// TranscodeGrpcServer implements the gRPC TranscodeService.
type TranscodeGrpcServer struct {
	transcodepb.UnimplementedTranscodeServiceServer
	app app.TranscodeApp
}

// NewTranscodeGrpcServer creates a new gRPC server implementation.
func NewTranscodeGrpcServer(transcodeApp app.TranscodeApp) *TranscodeGrpcServer {
	return &TranscodeGrpcServer{
		app: transcodeApp,
	}
}

// CreateTranscodeTask 创建转码任务
func (s *TranscodeGrpcServer) CreateTranscodeTask(ctx context.Context, req *transcodepb.CreateTranscodeTaskRequest) (*transcodepb.CreateTranscodeTaskResponse, error) {
	// 检查应用层是否初始化
	if s.app == nil {
		logger.Error("transcode app not initialised for gRPC server", nil)
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "service unavailable",
		}, nil
	}

	// 参数验证
	userUUID := req.GetUserUuid()
	if userUUID == "" {
		logger.Warn("CreateTranscodeTask called with empty user_uuid", nil)
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "user_uuid is required",
		}, nil
	}

	videoUUID := req.GetVideoUuid()
	if videoUUID == "" {
		logger.Warn("CreateTranscodeTask called with empty video_uuid", map[string]interface{}{
			"user_uuid": userUUID,
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "video_uuid is required",
		}, nil
	}

	inputPath := req.GetInputPath()
	if inputPath == "" {
		logger.Warn("CreateTranscodeTask called with empty input_path", map[string]interface{}{
			"user_uuid":  userUUID,
			"video_uuid": videoUUID,
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "input_path is required",
		}, nil
	}

	logger.Info("CreateTranscodeTask called", map[string]interface{}{
		"user_uuid":         userUUID,
		"video_uuid":        videoUUID,
		"input_path":        inputPath,
		"target_resolution": req.GetTargetResolution(),
		"target_bitrate":    req.GetTargetBitrate(),
	})

	// 构建应用层请求
	createReq := &cqe.CreateTranscodeTaskReq{
		UserUUID:     userUUID,
		VideoUUID:    videoUUID,
		OriginalPath: inputPath,
		Resolution:   req.GetTargetResolution(),
		Bitrate:      req.GetTargetBitrate(),
	}

	// 调用应用层服务
	taskDto, err := s.app.CreateTranscodeTask(ctx, createReq)
	if err != nil {
		logger.Error("failed to create transcode task", map[string]interface{}{
			"user_uuid":  userUUID,
			"video_uuid": videoUUID,
			"error":      err.Error(),
		})

		// 检查特定错误类型
		if errors.Is(err, errno.ErrTranscodeTaskExists) {
			return &transcodepb.CreateTranscodeTaskResponse{
				Success: false,
				Message: "transcode task already exists for this video",
			}, nil
		}

		if errors.Is(err, errno.ErrWorkerNotAvailable) || errors.Is(err, errno.ErrQueueFull) {
			return &transcodepb.CreateTranscodeTaskResponse{
				Success: false,
				Message: "service temporarily busy, please try again later",
			}, nil
		}

		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "failed to create transcode task",
		}, nil
	}

	logger.Info("transcode task created successfully", map[string]interface{}{
		"task_uuid":  taskDto.TaskUUID,
		"user_uuid":  userUUID,
		"video_uuid": videoUUID,
	})

	return &transcodepb.CreateTranscodeTaskResponse{
		Success:  true,
		TaskUuid: taskDto.TaskUUID,
		Message:  "transcode task created successfully",
	}, nil
}

// GetTranscodeTask 获取转码任务信息
func (s *TranscodeGrpcServer) GetTranscodeTask(ctx context.Context, req *transcodepb.GetTranscodeTaskRequest) (*transcodepb.GetTranscodeTaskResponse, error) {
	// 检查应用层是否初始化
	if s.app == nil {
		logger.Error("transcode app not initialised for gRPC server", nil)
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: "service unavailable",
		}, nil
	}

	// 参数验证
	taskUUID := req.GetTaskUuid()
	if taskUUID == "" {
		logger.Warn("GetTranscodeTask called with empty task_uuid", nil)
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: "task_uuid is required",
		}, nil
	}

	logger.Info("GetTranscodeTask called", map[string]interface{}{
		"task_uuid": taskUUID,
	})

	// 调用应用层服务
	taskDto, err := s.app.GetTranscodeTask(ctx, taskUUID)
	if err != nil {
		if err == errno.ErrTranscodeTaskNotFound {
			logger.Warn("transcode task not found", map[string]interface{}{
				"task_uuid": taskUUID,
			})
			return &transcodepb.GetTranscodeTaskResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("transcode task not found: %s", taskUUID),
			}, nil
		}

		logger.Error("failed to get transcode task", map[string]interface{}{
			"task_uuid": taskUUID,
			"error":     err.Error(),
		})
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get transcode task: %v", err),
		}, nil
	}

	// 构建响应数据
	progress := int32(taskDto.Progress)
	outputPath := taskDto.OutputPath
	errorMessage := taskDto.ErrorMessage

	// 确保完成状态有输出路径
	if taskDto.Status == "completed" && outputPath == "" {
		outputPath = taskDto.OutputPath
	}

	// 确保失败状态有错误信息
	if taskDto.Status == "failed" && errorMessage == "" {
		errorMessage = "transcode task failed"
	}

	logger.Info("transcode task retrieved successfully", map[string]interface{}{
		"task_uuid":  taskDto.TaskUUID,
		"video_uuid": taskDto.VideoUUID,
		"status":     taskDto.Status,
		"progress":   progress,
	})

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
