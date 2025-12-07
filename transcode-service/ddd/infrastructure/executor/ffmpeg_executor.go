package executor

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

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/port"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

// FFmpegExecutor implements port.TranscodeExecutor using local ffmpeg and StorageGateway.
type FFmpegExecutor struct {
	cfg     *config.Config
	storage gateway.StorageGateway
}

func NewFFmpegExecutor(cfg *config.Config, storage gateway.StorageGateway) *FFmpegExecutor {
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	return &FFmpegExecutor{cfg: cfg, storage: storage}
}

// Execute runs ffmpeg, uploads result unless SkipUpload, and cleans temporary files.
func (e *FFmpegExecutor) Execute(ctx context.Context, task *entity.TranscodeTaskEntity, opts port.TranscodeOptions) (string, string, error) {
	if task == nil {
		return "", "", errors.New("nil task")
	}
	cfg := e.cfg
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	tempDir := os.TempDir()
	if cfg != nil && strings.TrimSpace(cfg.Transcode.FFmpeg.TempDir) != "" {
		tempDir = cfg.Transcode.FFmpeg.TempDir
	}

	// Prepare paths
	localInputPath := filepath.Join(tempDir, "inputs", fmt.Sprintf("input_%s_%s", task.TaskUUID(), filepath.Base(task.OriginalPath())))
	if err := os.MkdirAll(filepath.Dir(localInputPath), 0o755); err != nil {
		return "", "", fmt.Errorf("create temp dir: %w", err)
	}
	localOutputPath := filepath.Join(tempDir, strings.TrimPrefix(task.OutputPath(), "/"))
	if err := os.MkdirAll(filepath.Dir(localOutputPath), 0o755); err != nil {
		return "", "", fmt.Errorf("create output dir: %w", err)
	}

	// Download input
	if e.storage != nil {
		if err := e.storage.DownloadFile(ctx, task.OriginalPath(), localInputPath); err != nil {
			return "", "", fmt.Errorf("download input: %w", err)
		}
	}
	defer func() {
		_ = os.Remove(localInputPath)
	}()

	durationSec := e.probeDurationSeconds(localInputPath)
	cmd := e.buildFFmpegCommand(ctx, task, localInputPath, localOutputPath)
	logger.Infof("ffmpeg command task_uuid=%s command=%s", task.TaskUUID(), strings.Join(cmd.Args, " "))
	if err := e.executeFFmpegCommand(ctx, cmd, durationSec, opts.ProgressCb); err != nil {
		return "", "", err
	}

	var objectKey, publicURL string
	if opts.SkipUpload {
		// 不上传完整视频，直接清理本地产物
		_ = os.Remove(localOutputPath)
		return "", "", nil
	}

	if e.storage == nil {
		return "", "", errors.New("storage gateway not configured")
	}
	objectKey = strings.TrimPrefix(task.OutputPath(), "/")
	if objectKey == "" {
		objectKey = filepath.Base(localOutputPath)
	}

	uploadedKey, err := e.storage.UploadTranscodedFile(ctx, localOutputPath, objectKey, "video/mp4")
	if err != nil {
		return "", "", fmt.Errorf("upload output: %w", err)
	}
	_ = os.Remove(localOutputPath)
	objectKey = uploadedKey
	publicURL = e.buildFileURL(uploadedKey)
	return objectKey, publicURL, nil
}

// --- internal helpers (mostly migrated from old domain service) ---

func (e *FFmpegExecutor) executeFFmpegCommand(ctx context.Context, cmd *exec.Cmd, durationSec float64, progressCb port.ProgressCallback) error {
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
		e.scanFFmpegProgress(ctx, stderr, durationSec, &buf, progressCb)
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
				logger.Errorf("ffmpeg failed tail_stderr=%s", strings.Join(tail, "\n"))
			}
		}
		return err
	}
}

func (e *FFmpegExecutor) scanFFmpegProgress(ctx context.Context, stderr io.ReadCloser, durationSec float64, capture *[]string, progressCb port.ProgressCallback) {
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
				e.emitProgress(sec, durationSec, progressCb)
			}
			continue
		}

		if m := reTime.FindStringSubmatch(line); len(m) == 4 && durationSec > 0 {
			hh, _ := strconv.ParseFloat(m[1], 64)
			mm, _ := strconv.ParseFloat(m[2], 64)
			ss, _ := strconv.ParseFloat(m[3], 64)
			sec := hh*3600 + mm*60 + ss
			e.emitProgress(sec, durationSec, progressCb)
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

func (e *FFmpegExecutor) emitProgress(currentSec, totalSec float64, cb port.ProgressCallback) {
	if cb == nil || totalSec <= 0 {
		return
	}
	pct := int((currentSec / totalSec) * 100)
	if pct > 99 {
		pct = 99
	}
	if pct < 0 {
		pct = 0
	}
	cb(pct)
}

// probeDurationSeconds 调用 ffprobe 获取输入时长（秒），失败则返回 0。
func (e *FFmpegExecutor) probeDurationSeconds(inputPath string) float64 {
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

func (e *FFmpegExecutor) probeVideoCodec(inputPath string) (codec string, pixFmt string) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-probesize", "5M",
		"-analyzeduration", "5M",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,pix_fmt",
		"-of", "csv=p=0",
		inputPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return "", ""
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) > 0 {
		codec = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		pixFmt = strings.TrimSpace(parts[1])
	}
	return
}

func (e *FFmpegExecutor) buildFFmpegCommand(ctx context.Context, task *entity.TranscodeTaskEntity, inputPath, outputPath string) *exec.Cmd {
	params := task.GetParams()
	cfg := e.cfg

	videoCodec := "libx264"
	videoPreset := "medium"
	hardwareAccel := ""
	threads := 0
	useHwDecode := false
	decThreads := 0
	decSurfaces := 0
	inputCodec, _ := e.probeVideoCodec(inputPath)
	if cfg != nil {
		if strings.TrimSpace(cfg.Transcode.FFmpeg.VideoCodec) != "" {
			videoCodec = cfg.Transcode.FFmpeg.VideoCodec
		}
		if strings.TrimSpace(cfg.Transcode.FFmpeg.VideoPreset) != "" {
			videoPreset = cfg.Transcode.FFmpeg.VideoPreset
		}
		if strings.TrimSpace(cfg.Transcode.FFmpeg.HardwareAccel) != "" {
			hardwareAccel = cfg.Transcode.FFmpeg.HardwareAccel
		}
		if cfg.Transcode.FFmpeg.Threads > 0 {
			threads = cfg.Transcode.FFmpeg.Threads
		}
		useHwDecode = cfg.Transcode.FFmpeg.UseHardwareDecode && strings.EqualFold(hardwareAccel, "cuda")
		if cfg.Transcode.FFmpeg.DecoderThreads > 0 {
			decThreads = cfg.Transcode.FFmpeg.DecoderThreads
		}
		if cfg.Transcode.FFmpeg.CuvidSurfaces > 0 {
			decSurfaces = cfg.Transcode.FFmpeg.CuvidSurfaces
		}
	}

	args := make([]string, 0, 16)
	if useHwDecode {
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
		switch strings.ToLower(inputCodec) {
		case "h264", "avc1":
			args = append(args, "-c:v", "h264_cuvid")
		case "hevc", "hvc1", "hev1":
			args = append(args, "-c:v", "hevc_cuvid")
		default:
		}
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
		"-probesize", "5M",
		"-analyzeduration", "5M",
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

	binary := "ffmpeg"
	if cfg != nil && cfg.Transcode.FFmpeg.BinaryPath != "" {
		binary = cfg.Transcode.FFmpeg.BinaryPath
	}
	return exec.CommandContext(ctx, binary, args...)
}

func (e *FFmpegExecutor) buildFileURL(objectKey string) string {
	if strings.TrimSpace(objectKey) == "" {
		return ""
	}
	cfg := e.cfg
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
