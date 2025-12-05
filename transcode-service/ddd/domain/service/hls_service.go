package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

// HLSService HLS切片服务接口
type HLSService interface {
	GenerateHLSSlices(ctx context.Context, job *entity.HLSJobEntity, inputPath string) error
}

// hlsServiceImpl HLS切片服务实现
type hlsServiceImpl struct {
	logger  *logger.Logger
	hlsRepo repo.HLSJobRepository
	cfg     *config.Config
}

// NewHLSService 创建HLS切片服务
func NewHLSService(log *logger.Logger, hlsRepo repo.HLSJobRepository, cfg *config.Config) HLSService {
	return &hlsServiceImpl{
		logger:  log,
		hlsRepo: hlsRepo,
		cfg:     cfg,
	}
}

// GenerateHLSSlices 生成HLS切片
func (h *hlsServiceImpl) GenerateHLSSlices(ctx context.Context, job *entity.HLSJobEntity, inputPath string) error {
	hlsConfig := job.GetConfig()
	if hlsConfig == nil || !hlsConfig.IsEnabled() {
		return fmt.Errorf("HLS is not enabled for job %s", job.JobUUID())
	}
	h.logger.Infof("开始生成HLS切片 job_uuid=%s input_path=%s resolutions=%d", job.JobUUID(), inputPath, len(hlsConfig.Resolutions))

	// 创建输出目录
	outputDir := h.generateOutputDir(job)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 设置HLS状态为处理中
	hlsConfig.SetStatus(vo.HLSStatusProcessing)
	h.updateProgress(ctx, job, 0)

	// 生成各分辨率的HLS切片
	var masterPlaylistEntries []string
	resolutions := hlsConfig.Resolutions

	for i, resolution := range resolutions {
		h.logger.Infof("生成分辨率切片 job_uuid=%s resolution=%s bitrate=%s", job.JobUUID(), resolution.Resolution, resolution.Bitrate)

		// 生成单个分辨率的HLS切片
		playlistPath, err := h.generateResolutionHLS(ctx, job, inputPath, outputDir, resolution, i)
		if err != nil {
			job.SetError(fmt.Sprintf("生成%s分辨率切片失败: %v", resolution.Resolution, err))
			return err
		}

		// 添加到master playlist
		masterPlaylistEntries = append(masterPlaylistEntries, h.createMasterPlaylistEntry(resolution, playlistPath))

		progress := (i + 1) * 100 / len(resolutions) // 以分辨率维度粗粒度进度
		h.updateProgress(ctx, job, progress)
	}

	// 生成master playlist
	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	if err := h.generateMasterPlaylist(masterPlaylistPath, masterPlaylistEntries); err != nil {
		job.SetError(fmt.Sprintf("生成master playlist失败: %v", err))
		return err
	}

	// 设置HLS完成
	job.SetMasterPlaylist(masterPlaylistPath)
	job.SetOutputDir(outputDir)
	job.SetStatus(vo.HLSStatusCompleted)

	h.logger.Infof("HLS切片生成完成 job_uuid=%s output_dir=%s master_path=%s", job.JobUUID(), outputDir, masterPlaylistPath)

	return nil
}

// generateResolutionHLS 生成单个分辨率的HLS切片
func (h *hlsServiceImpl) generateResolutionHLS(ctx context.Context, job *entity.HLSJobEntity, inputPath, outputDir string, resolution vo.ResolutionConfig, index int) (string, error) {
	hlsConfig := job.GetConfig()
	ffcfg := h.cfg.Transcode.FFmpeg

	videoCodec := "libx264"
	if strings.TrimSpace(ffcfg.VideoCodec) != "" {
		videoCodec = ffcfg.VideoCodec
	}
	lowerCodec := strings.ToLower(videoCodec)
	hardwareAccel := strings.TrimSpace(ffcfg.HardwareAccel)
	useHwDecode := ffcfg.UseHardwareDecode && strings.EqualFold(hardwareAccel, "cuda")
	threads := ffcfg.Threads
	if threads < 0 {
		threads = 0
	}

	// 构建输出文件名
	resolutionName := resolution.Resolution
	playlistName := fmt.Sprintf("playlist_%s.m3u8", resolutionName)
	segmentPattern := fmt.Sprintf("segment_%s_%%03d.ts", resolutionName)

	playlistPath := filepath.Join(outputDir, playlistName)
	segmentPath := filepath.Join(outputDir, segmentPattern)

	// 根据配置解析目标高度，无法解析时回退为源尺寸
	height, err := parseResolutionHeight(resolution.Resolution)
	scaleFilter := ""
	if err != nil {
		h.logger.Warnf("invalid HLS resolution; use source height job_uuid=%s resolution=%s err=%v",
			job.JobUUID(), resolution.Resolution, err)
	} else {
		scaleFilter = fmt.Sprintf("scale=-2:%d", height)
	}

	// 构建FFmpeg命令
	args := make([]string, 0, 32)
	if useHwDecode {
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda", "-c:v", "h264_cuvid")
	} else if hardwareAccel != "" && !strings.EqualFold(hardwareAccel, "cuda") {
		args = append(args, "-hwaccel", hardwareAccel)
	}
	args = append(args,
		"-probesize", "5M",
		"-analyzeduration", "5M",
		"-i", inputPath,
		"-c:v", videoCodec,
		"-c:a", "aac",
	)
	if strings.Contains(lowerCodec, "nvenc") {
		// NVENC: 用 scale_npp，目标格式使用 nv12 以避免 auto_scale 插入。
		if scaleFilter != "" {
			nppFilter := strings.ReplaceAll(scaleFilter, "scale=", "scale_npp=")
			if !strings.Contains(nppFilter, "format=") {
				nppFilter += ":format=nv12"
			}
			args = append(args, "-vf", nppFilter)
		}
		// 保持 GPU 链路，避免强制指定 nv12 触发 auto_scale 将帧拉回 CPU
	} else if scaleFilter != "" {
		// CPU 路径：加上 format=yuv420p，防止自动插入不兼容的 auto_scale
		cpuFilter := scaleFilter
		if !strings.Contains(cpuFilter, "format=") {
			cpuFilter = fmt.Sprintf("%s,format=yuv420p", cpuFilter)
		}
		args = append(args, "-vf", cpuFilter)
	} else if !strings.Contains(lowerCodec, "nvenc") {
		// 非 GPU 且无缩放时仍指定兼容像素格式
		args = append(args, "-pix_fmt", "yuv420p")
	}
	args = append(args,
		"-b:v", resolution.Bitrate,
		"-b:a", "128k",
		"-threads", strconv.Itoa(max(1, threads)),
		"-sc_threshold", "0",
		"-keyint_min", "48",
		"-g", "48",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n*%d)", hlsConfig.SegmentDuration),
		"-hls_flags", "independent_segments",
		"-hls_time", strconv.Itoa(hlsConfig.SegmentDuration),
		"-hls_list_size", strconv.Itoa(hlsConfig.ListSize),
		"-hls_segment_filename", segmentPath,
		"-f", "hls",
		playlistPath,
	)

	binary := "ffmpeg"
	if h.cfg != nil && strings.TrimSpace(h.cfg.Transcode.FFmpeg.BinaryPath) != "" {
		binary = h.cfg.Transcode.FFmpeg.BinaryPath
	}
	h.logger.Debug(fmt.Sprintf("执行FFmpeg命令 job_uuid=%s command=%s", job.JobUUID(), binary+" "+strings.Join(args, " ")))

	// 执行FFmpeg命令
	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Errorf("FFmpeg执行失败 job_uuid=%s error=%v output=%s", job.JobUUID(), err, string(output))
		return "", fmt.Errorf("FFmpeg执行失败: %w, output: %s", err, string(output))
	}

	return playlistName, nil
}

// generateMasterPlaylist 生成master playlist
func (h *hlsServiceImpl) generateMasterPlaylist(masterPath string, entries []string) error {
	content := "#EXTM3U\n#EXT-X-VERSION:3\n\n"
	for _, entry := range entries {
		content += entry + "\n"
	}

	return os.WriteFile(masterPath, []byte(content), 0644)
}

// parseBitrateToBps 将 "2000k"/"2M"/"2000kbps"/"2mbps" 等解析为 bps
func parseBitrateToBps(bitrate string) (int, error) {
	s := strings.TrimSpace(strings.ToLower(bitrate))
	if s == "" {
		return 0, fmt.Errorf("empty bitrate")
	}

	factor := 1.0
	switch {
	case strings.HasSuffix(s, "kbps"):
		factor = 1000
		s = strings.TrimSuffix(s, "kbps")
	case strings.HasSuffix(s, "mbps"):
		factor = 1000 * 1000
		s = strings.TrimSuffix(s, "mbps")
	case strings.HasSuffix(s, "k"):
		factor = 1000
		s = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "m"):
		factor = 1000 * 1000
		s = strings.TrimSuffix(s, "m")
	}

	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid bitrate: %s", bitrate)
	}
	return int(v * factor), nil
}

// parseResolutionHeight 将 "720p"/"1080"/"4K"/"2K" 转为高度值（720/1080/2160/1440）
func parseResolutionHeight(resolution string) (int, error) {
	s := strings.TrimSpace(strings.ToLower(resolution))
	if s == "" {
		return 0, fmt.Errorf("empty resolution")
	}
	switch s {
	case "4k":
		return 2160, nil
	case "2k":
		return 1440, nil
	}
	s = strings.TrimSuffix(s, "p")
	if s == "" {
		return 0, fmt.Errorf("invalid resolution: %s", resolution)
	}
	h, err := strconv.Atoi(s)
	if err != nil || h <= 0 {
		return 0, fmt.Errorf("invalid resolution: %s", resolution)
	}
	return h, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// createMasterPlaylistEntry 创建master playlist条目
func (h *hlsServiceImpl) createMasterPlaylistEntry(resolution vo.ResolutionConfig, playlistPath string) string {
	bitrate, err := parseBitrateToBps(resolution.Bitrate)
	if err != nil {
		h.logger.Warnf("invalid HLS bitrate bitrate=%s err=%v", resolution.Bitrate, err)
		bitrate = 0
	}

	height, err := parseResolutionHeight(resolution.Resolution)
	if err != nil {
		h.logger.Warnf("invalid HLS resolution resolution=%s err=%v", resolution.Resolution, err)
		height = 0
	}
	width := 0
	if height > 0 {
		width = height * 16 / 9 // 假设16:9比例
	}

	return fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n%s",
		bitrate, width, height, playlistPath)
}

// generateOutputDir 生成输出目录路径
func (h *hlsServiceImpl) generateOutputDir(job *entity.HLSJobEntity) string {
	baseDir := "storage/hls"
	return filepath.Join(baseDir, job.UserUUID(), job.VideoUUID(), job.JobUUID())
}

func (h *hlsServiceImpl) updateProgress(ctx context.Context, job *entity.HLSJobEntity, progress int) {
	job.SetProgress(progress)
	if h.hlsRepo == nil {
		return
	}
	if err := h.hlsRepo.UpdateHLSJobProgress(ctx, job.JobUUID(), progress); err != nil {
		h.logger.Warnf("update hls progress failed job_uuid=%s progress=%d error=%s", job.JobUUID(), progress, err.Error())
	}
}
