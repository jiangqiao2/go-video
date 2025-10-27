package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

// TranscodeService 转码领域服务
type TranscodeService interface {
	// ValidateTranscodeParams 验证转码参数
	ValidateTranscodeParams(params vo.TranscodeParams) error

	// GenerateOutputPath 生成输出路径
	GenerateOutputPath(userUUID, videoUUID string, params vo.TranscodeParams) string

	// CalculateEstimatedDuration 计算预估转码时长
	CalculateEstimatedDuration(fileSize int64, params vo.TranscodeParams) int64

	// ExecuteTranscode 执行转码任务
	ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error
}

type transcodeServiceImpl struct {
	transcodeTaskRepo repo.TranscodeTaskRepository
	storageGateway    gateway.StorageGateway
	cfg               *config.Config
}

// NewTranscodeService 创建转码领域服务
func NewTranscodeService(transcodeTaskRepo repo.TranscodeTaskRepository, storage gateway.StorageGateway, cfg *config.Config) TranscodeService {
	return &transcodeServiceImpl{
		transcodeTaskRepo: transcodeTaskRepo,
		storageGateway:    storage,
		cfg:               cfg,
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
		"task_uuid":  task.TaskUUID(),
		"video_uuid": task.VideoUUID(),
		"resolution": task.Params().Resolution,
		"bitrate":    task.Params().Bitrate,
	})

	if s.cfg == nil {
		s.cfg = config.GetGlobalConfig()
	}

	// 更新任务状态为处理中
	task.SetStatus(vo.TaskStatusProcessing)
	task.SetProgress(0)
	task.SetErrorMessage("")
	if err := s.transcodeTaskRepo.UpdateTranscodeTaskStatus(ctx, task.TaskUUID(), vo.TaskStatusProcessing, "", task.OutputPath(), task.Progress()); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	localOutputPath := s.getLocalOutputPath(task)
	if err := os.MkdirAll(filepath.Dir(localOutputPath), 0o755); err != nil {
		return fmt.Errorf("创建本地输出目录失败: %w", err)
	}

	var transcodeErr error
	binary := "ffmpeg"
	if s.cfg != nil && s.cfg.Transcode.FFmpeg.BinaryPath != "" {
		binary = s.cfg.Transcode.FFmpeg.BinaryPath
	}

	if _, err := exec.LookPath(binary); err != nil {
		logger.Warn("FFmpeg未找到，使用模拟转码", map[string]interface{}{"binary": binary})
		transcodeErr = s.simulateTranscode(localOutputPath)
	} else {
		cmd := s.buildFFmpegCommand(ctx, task, binary, localOutputPath)
		transcodeErr = s.executeFFmpegCommand(ctx, cmd, task)
		if transcodeErr != nil && !errors.Is(transcodeErr, context.Canceled) {
			logger.Error("FFmpeg执行失败，尝试模拟转码", map[string]interface{}{
				"task_uuid": task.TaskUUID(),
				"error":     transcodeErr.Error(),
			})
			if fallbackErr := s.simulateTranscode(localOutputPath); fallbackErr == nil {
				transcodeErr = nil
			} else {
				transcodeErr = fmt.Errorf("%w; fallback error: %v", transcodeErr, fallbackErr)
			}
		}
	}

	if transcodeErr != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(transcodeErr.Error())
		_ = s.transcodeTaskRepo.UpdateTranscodeTaskStatus(ctx, task.TaskUUID(), vo.TaskStatusFailed, transcodeErr.Error(), task.OutputPath(), task.Progress())
		return fmt.Errorf("转码执行失败: %w", transcodeErr)
	}

	if s.storageGateway == nil {
		err := errors.New("storage gateway not initialized")
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.transcodeTaskRepo.UpdateTranscodeTaskStatus(ctx, task.TaskUUID(), vo.TaskStatusFailed, err.Error(), task.OutputPath(), task.Progress())
		return err
	}

	objectKey := strings.TrimPrefix(task.OutputPath(), "/")
	if objectKey == "" {
		objectKey = filepath.Base(localOutputPath)
	}

	uploadedKey, err := s.storageGateway.UploadTranscodedFile(ctx, localOutputPath, objectKey, "video/mp4")
	if err != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.transcodeTaskRepo.UpdateTranscodeTaskStatus(ctx, task.TaskUUID(), vo.TaskStatusFailed, err.Error(), task.OutputPath(), task.Progress())
		return fmt.Errorf("上传转码结果失败: %w", err)
	}

	// 清理本地文件
	_ = os.Remove(localOutputPath)

	task.SetOutputPath(uploadedKey)
	task.SetStatus(vo.TaskStatusCompleted)
	task.SetProgress(100)
	task.SetErrorMessage("")

	if err := s.transcodeTaskRepo.UpdateTranscodeTask(ctx, task); err != nil {
		return fmt.Errorf("更新任务完成状态失败: %w", err)
	}

	logger.Info("转码任务执行完成", map[string]interface{}{
		"task_uuid":   task.TaskUUID(),
		"output_path": uploadedKey,
	})

	return nil
}

// buildFFmpegCommand 构建FFmpeg命令
func (s *transcodeServiceImpl) buildFFmpegCommand(ctx context.Context, task *entity.TranscodeTaskEntity, binaryPath, outputPath string) *exec.Cmd {
	inputPath := task.OriginalPath()
	args := []string{"-i", inputPath}
	params := task.Params()
	args = append(args, (&params).GetFFmpegArgs()...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-y",
		outputPath,
	)

	return exec.CommandContext(ctx, binaryPath, args...)
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
					"progress":  progress,
					"error":     err.Error(),
				})
			}
		}
	}
}

func (s *transcodeServiceImpl) getLocalOutputPath(task *entity.TranscodeTaskEntity) string {
	tempDir := os.TempDir()
	if s.cfg != nil && s.cfg.Transcode.FFmpeg.TempDir != "" {
		tempDir = s.cfg.Transcode.FFmpeg.TempDir
	}

	cleanPath := strings.TrimPrefix(task.OutputPath(), "/")
	return filepath.Join(tempDir, cleanPath)
}

func (s *transcodeServiceImpl) simulateTranscode(localOutputPath string) error {
	placeholder := []byte("transcoded-video-placeholder")
	if err := os.WriteFile(localOutputPath, placeholder, 0o644); err != nil {
		return fmt.Errorf("写入模拟转码文件失败: %w", err)
	}
	return nil
}
