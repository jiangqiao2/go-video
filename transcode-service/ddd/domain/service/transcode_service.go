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
	"sync"
	"time"

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
	progressMu     sync.Mutex
	lastPersist    map[string]time.Time
}

// NewTranscodeService 创建转码领域服务
func NewTranscodeService(transcodeRepo repo.TranscodeJobRepository, hlsRepo repo.HLSJobRepository, storage gateway.StorageGateway, cfg *config.Config, reporter gateway.TranscodeResultReporter) TranscodeService {
	return &transcodeServiceImpl{
		transcodeRepo:  transcodeRepo,
		hlsRepo:        hlsRepo,
		storageGateway: storage,
		cfg:            cfg,
		resultReporter: reporter,
		lastPersist:    make(map[string]time.Time),
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
	defer s.clearProgressThrottle(task.TaskUUID())

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
		return fmt.Errorf("转码执行失败: %w", err)
	}

	uploadedKey, _, err := s.uploadTranscodedResult(ctx, task, localOutputPath)
	if err != nil {
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, err.Error())
		return fmt.Errorf("上传转码结果失败: %w", err)
	}

	// 上传成功后清理本地输出文件，避免磁盘空间被持续占用
	if err := os.Remove(localOutputPath); err != nil {
		logger.Warnf("failed to clean local output file path=%s error=%s", localOutputPath, err.Error())
	}

	task.SetOutputPath(uploadedKey)
	task.SetStatus(vo.TaskStatusCompleted)
	task.SetProgress(100)
	task.SetErrorMessage("")

	if err := s.transcodeRepo.UpdateTranscodeJob(ctx, task); err != nil {
		errorMsg := fmt.Sprintf("更新任务完成状态失败: %v", err)
		task.SetStatus(vo.TaskStatusFailed)
		task.SetErrorMessage(errorMsg)
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
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

	logger.Infof("transcode task finished task_uuid=%s output_path=%s",
		task.TaskUUID(), uploadedKey)

	return nil
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
	logger.Infof("ffmpeg command task_uuid=%s command=%s", task.TaskUUID(), strings.Join(cmd.Args, " "))
	err := s.executeFFmpegCommand(ctx, cmd, task, durationSec)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Errorf("FFmpeg execution failed task_uuid=%s error=%s", task.TaskUUID(), err.Error())
	}
	return err
}

// scanFFmpegProgress 解析 FFmpeg 输出更新进度
func (s *transcodeServiceImpl) scanFFmpegProgress(ctx context.Context, task *entity.TranscodeTaskEntity, stderr io.ReadCloser, durationSec float64, capture *[]string) {
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
			continue
		}

		if capture != nil {
			b := *capture
			if len(b) >= 200 {
				b = b[1:]
			}
			b = append(b, line)
			*capture = b
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
	shouldPersist := false
	now := time.Now()
	s.progressMu.Lock()
	last := s.lastPersist[task.TaskUUID()]
	if last.IsZero() || now.Sub(last) >= time.Minute {
		s.lastPersist[task.TaskUUID()] = now
		shouldPersist = true
	}
	s.progressMu.Unlock()
	if shouldPersist {
		if err := s.transcodeRepo.UpdateTranscodeJobProgress(context.Background(), task.TaskUUID(), pct); err != nil {
			logger.Errorf("update transcode progress failed task_uuid=%s progress=%d error=%s", task.TaskUUID(), pct, err.Error())
		}
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
	params := task.GetParams()

	videoCodec := "libx264"
	videoPreset := "medium"
	hardwareAccel := ""
	threads := 0
	useHwDecode := false
	decThreads := 0
	decSurfaces := 0
	if s.cfg != nil {
		if strings.TrimSpace(s.cfg.Transcode.FFmpeg.VideoCodec) != "" {
			videoCodec = s.cfg.Transcode.FFmpeg.VideoCodec
		}
		if strings.TrimSpace(s.cfg.Transcode.FFmpeg.VideoPreset) != "" {
			videoPreset = s.cfg.Transcode.FFmpeg.VideoPreset
		}
		if strings.TrimSpace(s.cfg.Transcode.FFmpeg.HardwareAccel) != "" {
			hardwareAccel = s.cfg.Transcode.FFmpeg.HardwareAccel
		}
		if s.cfg.Transcode.FFmpeg.Threads > 0 {
			threads = s.cfg.Transcode.FFmpeg.Threads
		}
		useHwDecode = s.cfg.Transcode.FFmpeg.UseHardwareDecode && strings.EqualFold(hardwareAccel, "cuda")
		if s.cfg.Transcode.FFmpeg.DecoderThreads > 0 {
			decThreads = s.cfg.Transcode.FFmpeg.DecoderThreads
		}
		if s.cfg.Transcode.FFmpeg.CuvidSurfaces > 0 {
			decSurfaces = s.cfg.Transcode.FFmpeg.CuvidSurfaces
		}
	}

	args := make([]string, 0, 16)
	if useHwDecode {
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
		args = append(args, "-c:v", "h264_cuvid")
		if decSurfaces > 0 {
			args = append(args, "-surfaces", strconv.Itoa(decSurfaces))
		}
		if decThreads <= 0 {
			decThreads = 1
		}
		args = append(args, "-threads", strconv.Itoa(decThreads))
		threads = 0
	} else if hardwareAccel != "" && !strings.EqualFold(hardwareAccel, "cuda") {
		args = append(args, "-hwaccel", hardwareAccel)
	}
	args = append(args,
		"-i", inputPath,
		"-progress", "pipe:2",
		"-nostats",
	)
	baseArgs := (&params).GetFFmpegArgs(videoCodec, videoPreset)
	isNvenc := strings.Contains(strings.ToLower(videoCodec), "nvenc")
	if isNvenc {
		filtered := make([]string, 0, len(baseArgs))
		for i := 0; i < len(baseArgs); i++ {
			if baseArgs[i] == "-crf" && i+1 < len(baseArgs) {
				i++
				continue
			}
			filtered = append(filtered, baseArgs[i])
		}
		baseArgs = filtered
	}
	useCuda := strings.EqualFold(hardwareAccel, "cuda")
	if useCuda {
		filtered := make([]string, 0, len(baseArgs))
		for i := 0; i < len(baseArgs); i++ {
			if baseArgs[i] == "-s" && i+1 < len(baseArgs) {
				i++
				continue
			}
			filtered = append(filtered, baseArgs[i])
		}
		baseArgs = filtered
	}
	args = append(args, baseArgs...)

	w, h := 0, 0
	switch strings.TrimSpace(params.Resolution) {
	case "480p":
		w, h = 854, 480
	case "720p":
		w, h = 1280, 720
	case "1080p":
		w, h = 1920, 1080
	case "1440p":
		w, h = 2560, 1440
	case "2160p":
		w, h = 3840, 2160
	}
	if useCuda && w > 0 && h > 0 {
		if useHwDecode {
			args = append(args, "-vf", fmt.Sprintf("scale_npp=%d:%d:format=yuv420p", w, h))
		} else {
			args = append(args, "-vf", fmt.Sprintf("hwupload_cuda,scale_npp=%d:%d:format=yuv420p", w, h))
		}
	}
	if threads > 0 {
		args = append(args, "-threads", strconv.Itoa(threads))
	}
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
	buf := make([]string, 0, 200)
	go func() {
		defer close(progressDone)
		s.scanFFmpegProgress(ctx, task, stderr, durationSec, &buf)
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
		if err != nil {
			tail := buf
			if n := len(tail); n > 50 {
				tail = tail[n-50:]
			}
			if len(tail) > 0 {
				logger.Errorf("ffmpeg failed task_uuid=%s tail_stderr=%s", task.TaskUUID(), strings.Join(tail, "\n"))
			}
		}
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

func (s *transcodeServiceImpl) clearProgressThrottle(taskUUID string) {
	s.progressMu.Lock()
	delete(s.lastPersist, taskUUID)
	s.progressMu.Unlock()
}
