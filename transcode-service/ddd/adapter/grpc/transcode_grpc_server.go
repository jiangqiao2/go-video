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
	app app.TranscodeApp
	transcodepb.UnimplementedTranscodeServiceServer
}

// NewTranscodeGrpcServer creates a new gRPC server implementation.
func NewTranscodeGrpcServer(transcodeApp app.TranscodeApp) *TranscodeGrpcServer {
	return &TranscodeGrpcServer{
		app: transcodeApp,
	}
}

// CreateTranscodeTask 创建转码任务
func (s *TranscodeGrpcServer) CreateTranscodeTask(ctx context.Context, req *transcodepb.CreateTranscodeTaskRequest) (*transcodepb.CreateTranscodeTaskResponse, error) {
	logger.Info("gRPC CreateTranscodeTask called", map[string]interface{}{
		"user_uuid":         req.GetUserUuid(),
		"video_uuid":        req.GetVideoUuid(),
		"input_path":        req.GetInputPath(),
		"target_resolution": req.GetTargetResolution(),
		"target_bitrate":    req.GetTargetBitrate(),
	})

	// 检查应用层是否初始化
	if s.app == nil {
		logger.Error("Transcode app not initialized", map[string]interface{}{
			"method": "CreateTranscodeTask",
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "Service temporarily unavailable",
		}, nil
	}

	// 参数验证
	if req.GetUserUuid() == "" {
		logger.Warn("Missing required parameter: user_uuid", map[string]interface{}{
			"method": "CreateTranscodeTask",
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "Invalid parameter: user_uuid is required",
		}, nil
	}

	if req.GetVideoUuid() == "" {
		logger.Warn("Missing required parameter: video_uuid", map[string]interface{}{
			"method": "CreateTranscodeTask",
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "Invalid parameter: video_uuid is required",
		}, nil
	}

	if req.GetInputPath() == "" {
		logger.Warn("Missing required parameter: input_path", map[string]interface{}{
			"method": "CreateTranscodeTask",
		})
		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "Invalid parameter: input_path is required",
		}, nil
	}

	// 构建应用层请求
	createReq := &cqe.CreateTranscodeTaskReq{
		UserUUID:     req.GetUserUuid(),
		VideoUUID:    req.GetVideoUuid(),
		OriginalPath: req.GetInputPath(),
		Resolution:   req.GetTargetResolution(),
		Bitrate:      req.GetTargetBitrate(),
	}

	// 调用应用层服务
	taskDto, err := s.app.CreateTranscodeTask(ctx, createReq)
	if err != nil {
		logger.Error("Failed to create transcode task", map[string]interface{}{
			"error":      err.Error(),
			"user_uuid":  req.GetUserUuid(),
			"video_uuid": req.GetVideoUuid(),
			"method":     "CreateTranscodeTask",
		})

		// 检查特定错误类型
		if errors.Is(err, errno.ErrTranscodeTaskExists) {
			return &transcodepb.CreateTranscodeTaskResponse{
				Success: false,
				Message: "Transcode task already exists for this video",
			}, nil
		}

		if errors.Is(err, errno.ErrWorkerNotAvailable) || errors.Is(err, errno.ErrQueueFull) {
			return &transcodepb.CreateTranscodeTaskResponse{
				Success: false,
				Message: "Service temporarily busy, please try again later",
			}, nil
		}

		return &transcodepb.CreateTranscodeTaskResponse{
			Success: false,
			Message: "Failed to create transcode task",
		}, nil
	}

	logger.Info("Transcode task created successfully", map[string]interface{}{
		"task_uuid":  taskDto.TaskUUID,
		"user_uuid":  req.GetUserUuid(),
		"video_uuid": req.GetVideoUuid(),
		"method":     "CreateTranscodeTask",
	})

	return &transcodepb.CreateTranscodeTaskResponse{
		Success:  true,
		TaskUuid: taskDto.TaskUUID,
		Message:  "Transcode task created successfully",
	}, nil
}

// GetTranscodeTask 获取转码任务信息
func (s *TranscodeGrpcServer) GetTranscodeTask(ctx context.Context, req *transcodepb.GetTranscodeTaskRequest) (*transcodepb.GetTranscodeTaskResponse, error) {
	logger.Info("gRPC GetTranscodeTask called", map[string]interface{}{
		"task_uuid": req.GetTaskUuid(),
	})

	// 检查应用层是否初始化
	if s.app == nil {
		logger.Error("Transcode app not initialized", map[string]interface{}{
			"method": "GetTranscodeTask",
		})
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: "transcode app not initialized",
		}, nil
	}

	// 参数验证
	if req.GetTaskUuid() == "" {
		logger.Warn("Missing required parameter: task_uuid", map[string]interface{}{
			"method": "GetTranscodeTask",
		})
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: errno.ErrTaskUUIDRequired.Error(),
		}, nil
	}

	// 调用应用层服务
	taskDto, err := s.app.GetTranscodeTask(ctx, req.GetTaskUuid())
	if err != nil {
		if err == errno.ErrTranscodeTaskNotFound {
			logger.Warn("Transcode task not found", map[string]interface{}{
				"task_uuid": req.GetTaskUuid(),
				"method":    "GetTranscodeTask",
			})
			return &transcodepb.GetTranscodeTaskResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Transcode task not found: %s", req.GetTaskUuid()),
			}, nil
		}

		logger.Error("Failed to get transcode task", map[string]interface{}{
			"error":     err.Error(),
			"task_uuid": req.GetTaskUuid(),
			"method":    "GetTranscodeTask",
		})
		return &transcodepb.GetTranscodeTaskResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to get transcode task: %v", err),
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
		errorMessage = "Transcode task failed"
	}

	logger.Info("Transcode task retrieved successfully", map[string]interface{}{
		"task_uuid":  taskDto.TaskUUID,
		"video_uuid": taskDto.VideoUUID,
		"status":     taskDto.Status,
		"progress":   progress,
		"method":     "GetTranscodeTask",
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
