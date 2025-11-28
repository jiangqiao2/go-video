package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

// TranscodeService 转码领域服务
type TranscodeService interface {
	// ExecuteTranscode 执行转码任务
	ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error
}

type transcodeServiceImpl struct {
	transcodeRepo  repo.TranscodeJobRepository
	hlsRepo        repo.HLSJobRepository
	storageGateway gateway.StorageGateway
	cfg            *config.Config
	resultReporter gateway.TranscodeResultReporter
}

// NewTranscodeService 创建转码领域服务
func NewTranscodeService(transcodeRepo repo.TranscodeJobRepository, hlsRepo repo.HLSJobRepository, storage gateway.StorageGateway, cfg *config.Config, reporter gateway.TranscodeResultReporter) TranscodeService {
	return &transcodeServiceImpl{
		transcodeRepo:  transcodeRepo,
		hlsRepo:        hlsRepo,
		storageGateway: storage,
		cfg:            cfg,
		resultReporter: reporter,
	}
}

// ExecuteTranscode 执行转码任务
func (s *transcodeServiceImpl) ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	logger.Infof("start transcode task task_uuid=%s video_uuid=%s resolution=%s bitrate=%s",
		task.TaskUUID(), task.VideoUUID(), task.GetParams().Resolution, task.GetParams().Bitrate)

	if s.cfg == nil {
		s.cfg = config.GetGlobalConfig()
	}

	// 更新任务状态为处理中
	task.SetStatus(vo.TaskStatusProcessing)
	task.SetProgress(0)
	task.SetErrorMessage("")
	if err := s.updateJobStatus(ctx, task, vo.TaskStatusProcessing, ""); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	localOutputPath := s.getLocalOutputPath(task)
	if err := os.MkdirAll(filepath.Dir(localOutputPath), 0o755); err != nil {
		return fmt.Errorf("创建本地输出目录失败: %w", err)
	}

	// 下载输入文件到本地
	localInputPath := s.getLocalInputPath(task)
	if err := s.storageGateway.DownloadFile(ctx, task.OriginalPath(), localInputPath); err != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(fmt.Sprintf("下载输入文件失败: %v", err))
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
		s.reportFailure(ctx, task)
		return fmt.Errorf("下载输入文件失败: %w", err)
	}

	// 确保在函数结束时清理本地输入文件
	defer func() {
		if err := os.Remove(localInputPath); err != nil {
			logger.Warnf("failed to clean local input file path=%s error=%s", localInputPath, err.Error())
		}
	}()

	binary := "ffmpeg"
	if s.cfg != nil && s.cfg.Transcode.FFmpeg.BinaryPath != "" {
		binary = s.cfg.Transcode.FFmpeg.BinaryPath
	}

	if err := s.runFFmpeg(ctx, task, binary, localInputPath, localOutputPath); err != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
		s.reportFailure(ctx, task)
		return fmt.Errorf("转码执行失败: %w", err)
	}

	uploadedKey, publicVideoURL, err := s.uploadTranscodedResult(ctx, task, localOutputPath)
	if err != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, err.Error())
		s.reportFailure(ctx, task)
		return fmt.Errorf("上传转码结果失败: %w", err)
	}

	// 已拆分：不在转码流程内直接处理 HLS 切片

	task.SetOutputPath(uploadedKey)
	task.SetStatus(vo.TaskStatusCompleted)
	task.SetProgress(100)
	task.SetErrorMessage("")

	if err := s.transcodeRepo.UpdateTranscodeJob(ctx, task); err != nil {
		errorMsg := fmt.Sprintf("更新任务完成状态失败: %v", err)
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(errorMsg)
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
		s.reportFailure(ctx, task)
		return fmt.Errorf("更新任务完成状态失败: %w", err)
	}

	if resCfg, err := vo.NewResolutionConfig(task.GetParams().Resolution, task.GetParams().Bitrate); err == nil {
		if hcfg, err2 := vo.NewHLSConfig(true, []vo.ResolutionConfig{*resCfg}); err2 == nil {
			hJobUUID := uuid.New().String()
			outputDir := filepath.ToSlash(filepath.Join("storage/hls", task.UserUUID(), task.VideoUUID(), hJobUUID))
			hJob := entity.NewHLSJobEntity(hJobUUID, task.UserUUID(), task.VideoUUID(), uploadedKey, outputDir, *hcfg)
			src := task.TaskUUID()
			hJob.SetSource(&src, "transcoded")
			_ = s.hlsRepo.CreateHLSJob(ctx, hJob)
			_ = queue.DefaultHLSJobQueue().Enqueue(ctx, hJob)
		}
	}

	logger.Infof("transcode task finished task_uuid=%s output_path=%s public_url=%s",
		task.TaskUUID(), uploadedKey, publicVideoURL)

	s.reportSuccess(ctx, task, publicVideoURL)

	return nil
}

func (s *transcodeServiceImpl) reportSuccess(ctx context.Context, task *entity.TranscodeTaskEntity, publicURL string) {
	result := vo.NewTranscodeResult(task.TaskUUID(), task.VideoUUID())
	if err := result.ReportSuccess(ctx, s.resultReporter, publicURL); err != nil {
		logger.Warnf("report transcode success to upload-service failed task_uuid=%s video_uuid=%s error=%s",
			task.TaskUUID(), task.VideoUUID(), err.Error())
	}
}

func (s *transcodeServiceImpl) reportFailure(ctx context.Context, task *entity.TranscodeTaskEntity) {
	result := vo.NewTranscodeResult(task.TaskUUID(), task.VideoUUID())
	if err := result.ReportFailure(ctx, s.resultReporter, task.ErrorMessage()); err != nil {
		logger.Warnf("report transcode failure to upload-service failed task_uuid=%s video_uuid=%s error=%s",
			task.TaskUUID(), task.VideoUUID(), err.Error())
	}
}

// updateJobStatus 封装状态更新，统一使用任务当前的输出路径与进度。
func (s *transcodeServiceImpl) updateJobStatus(ctx context.Context, task *entity.TranscodeTaskEntity, status vo.TaskStatus, message string) error {
	if s.transcodeRepo == nil {
		return errors.New("transcodeRepo is nil")
	}
	return s.transcodeRepo.UpdateTranscodeJobStatus(ctx, task.TaskUUID(), status, message, task.OutputPath(), task.Progress())
}

// runFFmpeg 执行转码，确保调用处只关心错误处理。
func (s *transcodeServiceImpl) runFFmpeg(ctx context.Context, task *entity.TranscodeTaskEntity, binary, inputPath, outputPath string) error {
	durationSec := s.probeDurationSeconds(inputPath)
	cmd := s.buildFFmpegCommand(ctx, task, binary, inputPath, outputPath)
	err := s.executeFFmpegCommand(ctx, cmd, task, durationSec)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Errorf("FFmpeg execution failed task_uuid=%s error=%s", task.TaskUUID(), err.Error())
	}
	return err
}

// scanFFmpegProgress 解析 FFmpeg 输出更新进度
func (s *transcodeServiceImpl) scanFFmpegProgress(ctx context.Context, task *entity.TranscodeTaskEntity, stderr io.ReadCloser, durationSec float64) {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	reTime := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.?\d*)`)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := scanner.Text()

		if strings.HasPrefix(line, "out_time_ms=") {
			if ms, err := strconv.ParseFloat(strings.TrimPrefix(line, "out_time_ms="), 64); err == nil && durationSec > 0 {
				sec := ms / 1e6
				s.setProgress(task, sec, durationSec)
			}
			continue
		}

		if m := reTime.FindStringSubmatch(line); len(m) == 4 && durationSec > 0 {
			hh, _ := strconv.ParseFloat(m[1], 64)
			mm, _ := strconv.ParseFloat(m[2], 64)
			ss, _ := strconv.ParseFloat(m[3], 64)
			sec := hh*3600 + mm*60 + ss
			s.setProgress(task, sec, durationSec)
		}
	}
}

func (s *transcodeServiceImpl) setProgress(task *entity.TranscodeTaskEntity, currentSec, totalSec float64) {
	if totalSec <= 0 {
		return
	}
	pct := int((currentSec / totalSec) * 100)
	if pct > 99 {
		pct = 99
	}
	if pct < 0 {
		pct = 0
	}
	task.SetProgress(pct)
	if err := s.transcodeRepo.UpdateTranscodeJobProgress(context.Background(), task.TaskUUID(), pct); err != nil {
		logger.Errorf("update transcode progress failed task_uuid=%s progress=%d error=%s", task.TaskUUID(), pct, err.Error())
	}
}

// probeDurationSeconds 调用 ffprobe 获取输入时长（秒），失败则返回 0。
func (s *transcodeServiceImpl) probeDurationSeconds(inputPath string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputPath)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return val
}

// uploadTranscodedResult 上传转码结果并返回存储Key和对外URL。
func (s *transcodeServiceImpl) uploadTranscodedResult(ctx context.Context, task *entity.TranscodeTaskEntity, localOutputPath string) (string, string, error) {
	if s.storageGateway == nil {
		return "", "", errors.New("storage gateway not initialized")
	}

	objectKey := strings.TrimPrefix(task.OutputPath(), "/")
	if objectKey == "" {
		objectKey = filepath.Base(localOutputPath)
	}

	uploadedKey, err := s.storageGateway.UploadTranscodedFile(ctx, localOutputPath, objectKey, "video/mp4")
	if err != nil {
		return "", "", err
	}

	publicVideoURL := s.buildFileURL(uploadedKey)
	if strings.TrimSpace(publicVideoURL) == "" {
		publicVideoURL = uploadedKey
	}
	return uploadedKey, publicVideoURL, nil
}

// buildFFmpegCommand 构建FFmpeg命令
func (s *transcodeServiceImpl) buildFFmpegCommand(ctx context.Context, task *entity.TranscodeTaskEntity, binaryPath, inputPath, outputPath string) *exec.Cmd {
	args := []string{
		"-i", inputPath,
		"-progress", "pipe:2",
		"-nostats",
	}
	params := task.GetParams()
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
func (s *transcodeServiceImpl) executeFFmpegCommand(ctx context.Context, cmd *exec.Cmd, task *entity.TranscodeTaskEntity, durationSec float64) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建FFmpeg stderr管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动FFmpeg命令失败: %w", err)
	}

	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		s.scanFFmpegProgress(ctx, task, stderr, durationSec)
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-progressDone
		return ctx.Err()
	case err := <-done:
		<-progressDone
		return err
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

func (s *transcodeServiceImpl) getLocalInputPath(task *entity.TranscodeTaskEntity) string {
	tempDir := os.TempDir()
	if s.cfg != nil && s.cfg.Transcode.FFmpeg.TempDir != "" {
		tempDir = s.cfg.Transcode.FFmpeg.TempDir
	}

	// 从原始路径中提取文件名
	originalPath := task.OriginalPath()
	fileName := filepath.Base(originalPath)

	// 为输入文件创建唯一的本地路径
	inputFileName := fmt.Sprintf("input_%s_%s", task.TaskUUID(), fileName)
	return filepath.Join(tempDir, "inputs", inputFileName)
}

func (s *transcodeServiceImpl) buildFileURL(objectKey string) string {
	if strings.TrimSpace(objectKey) == "" {
		return ""
	}
	cfg := s.cfg
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	if cfg == nil {
		return objectKey
	}

	key := strings.TrimLeft(objectKey, "/")
	if strings.HasPrefix(key, "transcode/") {
		key = strings.TrimPrefix(key, "transcode/")
	}

	path := fmt.Sprintf("/storage/transcode/%s", key)
	publicBase := strings.TrimSpace(cfg.Public.StorageBase)
	if publicBase != "" {
		if !strings.HasPrefix(publicBase, "http://") && !strings.HasPrefix(publicBase, "https://") {
			publicBase = "http://" + publicBase
		}
		return strings.TrimRight(publicBase, "/") + path
	}

	return path
}

func detectHLSContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
