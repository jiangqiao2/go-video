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
}

// NewHLSService 创建HLS切片服务
func NewHLSService(log *logger.Logger, hlsRepo repo.HLSJobRepository) HLSService {
	return &hlsServiceImpl{
		logger:  log,
		hlsRepo: hlsRepo,
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

	// 构建输出文件名
	resolutionName := resolution.Resolution
	playlistName := fmt.Sprintf("playlist_%s.m3u8", resolutionName)
	segmentPattern := fmt.Sprintf("segment_%s_%%03d.ts", resolutionName)

	playlistPath := filepath.Join(outputDir, playlistName)
	segmentPath := filepath.Join(outputDir, segmentPattern)

	// 构建FFmpeg命令
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-vf", fmt.Sprintf("scale=-2:%s", strings.TrimSuffix(resolution.Resolution, "p")),
		"-b:v", resolution.Bitrate,
		"-b:a", "128k",
		"-hls_time", strconv.Itoa(hlsConfig.SegmentDuration),
		"-hls_list_size", strconv.Itoa(hlsConfig.ListSize),
		"-hls_segment_filename", segmentPath,
		"-f", "hls",
		playlistPath,
	}

	h.logger.Debug(fmt.Sprintf("执行FFmpeg命令 job_uuid=%s command=%s", job.JobUUID(), "ffmpeg "+strings.Join(args, " ")))

	// 执行FFmpeg命令
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
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

// createMasterPlaylistEntry 创建master playlist条目
func (h *hlsServiceImpl) createMasterPlaylistEntry(resolution vo.ResolutionConfig, playlistPath string) string {
	// 解析比特率数值（去掉单位）
	bitrateStr := strings.TrimSuffix(resolution.Bitrate, "k")
	bitrateStr = strings.TrimSuffix(bitrateStr, "K")
	bitrate, _ := strconv.Atoi(bitrateStr)
	bitrate *= 1000 // 转换为bps

	// 从分辨率字符串解析高度（如 "720p" -> 720）
	heightStr := strings.TrimSuffix(resolution.Resolution, "p")
	height, _ := strconv.Atoi(heightStr)
	width := height * 16 / 9 // 假设16:9比例

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
