package service

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/pkg/logger"
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
	
	// ExecuteTranscode 执行转码任务
	ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error
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

// ExecuteTranscode 执行转码任务
func (s *transcodeServiceImpl) ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	logger.Info("开始执行转码任务", map[string]interface{}{
		"task_uuid": task.TaskUUID(),
		"video_uuid": task.VideoUUID(),
		"resolution": task.Params().Resolution,
		"bitrate": task.Params().Bitrate,
	})

	// 更新任务状态为处理中
	task.SetStatus(vo.TaskStatusProcessing)
	task.SetProgress(0)
	if err := s.transcodeTaskRepo.UpdateTranscodeTaskProgress(ctx, task.TaskUUID(), 0); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 构建FFmpeg命令
	cmd := s.buildFFmpegCommand(task)

	// 执行转码
	err := s.executeFFmpegCommand(ctx, cmd, task)
	if err != nil {
		// 转码失败，更新任务状态
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		if updateErr := s.transcodeTaskRepo.UpdateTranscodeTask(ctx, task); updateErr != nil {
			logger.Error("更新失败任务状态失败", map[string]interface{}{
				"task_uuid": task.TaskUUID(),
				"error": updateErr.Error(),
			})
		}
		return fmt.Errorf("转码执行失败: %w", err)
	}

	// 转码成功，更新任务状态
	task.SetStatus(vo.TaskStatusCompleted)
	task.SetProgress(100)
	if err := s.transcodeTaskRepo.UpdateTranscodeTaskProgress(ctx, task.TaskUUID(), 100); err != nil {
		return fmt.Errorf("更新任务完成状态失败: %w", err)
	}

	logger.Info("转码任务执行完成", map[string]interface{}{
		"task_uuid": task.TaskUUID(),
		"output_path": task.OutputPath(),
	})

	return nil
}

// buildFFmpegCommand 构建FFmpeg命令
func (s *transcodeServiceImpl) buildFFmpegCommand(task *entity.TranscodeTaskEntity) *exec.Cmd {
	params := task.Params()
	inputPath := task.OriginalPath()
	outputPath := task.OutputPath()

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	exec.Command("mkdir", "-p", outputDir).Run()

	// 构建FFmpeg参数
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-y", // 覆盖输出文件
	}

	// 根据分辨率设置视频尺寸
	switch params.Resolution {
	case "480p":
		args = append(args, "-vf", "scale=854:480")
	case "720p":
		args = append(args, "-vf", "scale=1280:720")
	case "1080p":
		args = append(args, "-vf", "scale=1920:1080")
	case "1440p":
		args = append(args, "-vf", "scale=2560:1440")
	case "2160p":
		args = append(args, "-vf", "scale=3840:2160")
	}

	// 设置码率
	if params.Bitrate != "" {
		args = append(args, "-b:v", params.Bitrate)
	}

	args = append(args, outputPath)

	return exec.Command("ffmpeg", args...)
}

// executeFFmpegCommand 执行FFmpeg命令并监控进度
func (s *transcodeServiceImpl) executeFFmpegCommand(ctx context.Context, cmd *exec.Cmd, task *entity.TranscodeTaskEntity) error {
	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动FFmpeg命令失败: %w", err)
	}

	// 创建一个goroutine来监控进度
	progressDone := make(chan struct{})
	go s.monitorTranscodeProgress(ctx, task, progressDone)

	// 等待命令完成
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// 上下文取消，终止FFmpeg进程
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		close(progressDone)
		return ctx.Err()
	case err := <-done:
		close(progressDone)
		return err
	}
}

// monitorTranscodeProgress 监控转码进度
func (s *transcodeServiceImpl) monitorTranscodeProgress(ctx context.Context, task *entity.TranscodeTaskEntity, done <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	progress := float32(0)
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 模拟进度更新（实际应该解析FFmpeg输出）
			progress += 10
			if progress > 95 {
				progress = 95 // 不要设置为100，等转码完成后再设置
			}
			task.SetProgress(int(progress))
			if err := s.transcodeTaskRepo.UpdateTranscodeTaskProgress(ctx, task.TaskUUID(), int(progress)); err != nil {
				logger.Error("更新转码进度失败", map[string]interface{}{
					"task_uuid": task.TaskUUID(),
					"progress": progress,
					"error": err.Error(),
				})
			}
		}
	}
}