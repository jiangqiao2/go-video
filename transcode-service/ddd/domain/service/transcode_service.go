package service

import (
	"context"
	"fmt"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
)

// TranscodeService 转码领域服务
type TranscodeService interface {
	// ValidateTranscodeParams 验证转码参数
	ValidateTranscodeParams(params vo.TranscodeParams) error
	
	// GenerateOutputPath 生成输出路径
	GenerateOutputPath(userUUID, videoUUID string, params vo.TranscodeParams) string
	
	// CanCreateTask 检查是否可以创建转码任务
	CanCreateTask(ctx context.Context, userUUID, videoUUID string) error
	
	// CalculateEstimatedDuration 计算预估转码时长
	CalculateEstimatedDuration(fileSize int64, params vo.TranscodeParams) int64
}

type transcodeServiceImpl struct {
	transcodeTaskRepo repo.TranscodeTaskRepository
}

// NewTranscodeService 创建转码领域服务
func NewTranscodeService(transcodeTaskRepo repo.TranscodeTaskRepository) TranscodeService {
	return &transcodeServiceImpl{
		transcodeTaskRepo: transcodeTaskRepo,
	}
}

// ValidateTranscodeParams 验证转码参数
func (s *transcodeServiceImpl) ValidateTranscodeParams(params vo.TranscodeParams) error {
	if params.Resolution == "" {
		return fmt.Errorf("转码分辨率不能为空")
	}
	
	if params.Bitrate == "" {
		return fmt.Errorf("转码码率不能为空")
	}
	
	return nil
}

// GenerateOutputPath 生成输出路径
func (s *transcodeServiceImpl) GenerateOutputPath(userUUID, videoUUID string, params vo.TranscodeParams) string {
	return fmt.Sprintf("/transcoded/%s/%s_%s_%s.mp4", 
		userUUID, 
		videoUUID, 
		params.Resolution, 
		params.Bitrate)
}

// CanCreateTask 检查是否可以创建转码任务
func (s *transcodeServiceImpl) CanCreateTask(ctx context.Context, userUUID, videoUUID string) error {
	// 检查是否已有相同视频的处理中任务
	tasks, err := s.transcodeTaskRepo.QueryTranscodeTasksByVideoUUID(ctx, videoUUID)
	if err != nil {
		return fmt.Errorf("查询转码任务失败: %w", err)
	}
	
	for _, task := range tasks {
		if task.Status() == vo.TaskStatusPending || task.Status() == vo.TaskStatusProcessing {
			return fmt.Errorf("该视频已有转码任务正在处理中")
		}
	}
	
	return nil
}

// CalculateEstimatedDuration 计算预估转码时长
func (s *transcodeServiceImpl) CalculateEstimatedDuration(fileSize int64, params vo.TranscodeParams) int64 {
	// 简单的估算逻辑：文件大小(MB) * 分辨率系数 * 码率系数
	fileSizeMB := fileSize / (1024 * 1024)
	
	// 根据分辨率计算系数
	var resolutionFactor float64
	switch params.Resolution {
	case "480p":
		resolutionFactor = 1.0
	case "720p":
		resolutionFactor = 1.5
	case "1080p":
		resolutionFactor = 2.0
	case "1440p":
		resolutionFactor = 3.0
	case "2160p":
		resolutionFactor = 4.0
	default:
		resolutionFactor = 1.0
	}
	
	// 基础转码速度：每MB需要2秒
	estimatedSeconds := float64(fileSizeMB) * 2.0 * resolutionFactor
	return int64(estimatedSeconds)
}