package vo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// HLSStatus HLS切片状态
type HLSStatus string

const (
	HLSStatusDisabled   HLSStatus = "disabled"   // 未启用
	HLSStatusPending    HLSStatus = "pending"    // 待处理
	HLSStatusProcessing HLSStatus = "processing" // 处理中
	HLSStatusCompleted  HLSStatus = "completed"  // 已完成
	HLSStatusFailed     HLSStatus = "failed"     // 失败
)

// String 返回状态字符串
func (s HLSStatus) String() string {
	return string(s)
}

// IsValid 检查状态是否有效
func (s HLSStatus) IsValid() bool {
	switch s {
	case HLSStatusDisabled, HLSStatusPending, HLSStatusProcessing, HLSStatusCompleted, HLSStatusFailed:
		return true
	default:
		return false
	}
}

// ResolutionConfig 分辨率配置
type ResolutionConfig struct {
	Resolution string `json:"resolution"` // 分辨率，如 "720p", "480p", "360p"
	Bitrate    string `json:"bitrate"`    // 码率，如 "2000k", "1000k", "500k"
}

// NewResolutionConfig 创建分辨率配置
func NewResolutionConfig(resolution, bitrate string) (*ResolutionConfig, error) {
	if err := validateResolution(resolution); err != nil {
		return nil, err
	}
	if err := validateBitrate(bitrate); err != nil {
		return nil, err
	}
	return &ResolutionConfig{
		Resolution: resolution,
		Bitrate:    bitrate,
	}, nil
}

// Validate 验证分辨率配置
func (rc *ResolutionConfig) Validate() error {
	if err := validateResolution(rc.Resolution); err != nil {
		return err
	}
	if err := validateBitrate(rc.Bitrate); err != nil {
		return err
	}
	return nil
}

// HLSConfig HLS配置值对象
type HLSConfig struct {
	EnableHLS         bool               `json:"enable_hls"`          // 是否启用HLS切片
	Resolutions       []ResolutionConfig `json:"resolutions"`         // 多分辨率配置
	SegmentDuration   int                `json:"segment_duration"`    // 切片时长(秒)
	ListSize          int                `json:"list_size"`           // 播放列表大小(0表示无限制)
	Format            string             `json:"format"`              // HLS格式(mpegts/fmp4)
	Status            HLSStatus          `json:"status"`              // HLS状态
	Progress          int                `json:"progress"`            // 进度(0-100)
	OutputPath        string             `json:"output_path"`         // 输出路径
	ErrorMessage      string             `json:"error_message"`       // 错误信息
}

// NewHLSConfig 创建HLS配置
func NewHLSConfig(enableHLS bool, resolutions []ResolutionConfig) (*HLSConfig, error) {
	if enableHLS && len(resolutions) == 0 {
		return nil, fmt.Errorf("启用HLS时必须指定至少一个分辨率配置")
	}

	// 验证所有分辨率配置
	for i, res := range resolutions {
		if err := res.Validate(); err != nil {
			return nil, fmt.Errorf("分辨率配置[%d]无效: %w", i, err)
		}
	}

	status := HLSStatusDisabled
	if enableHLS {
		status = HLSStatusPending
	}

	return &HLSConfig{
		EnableHLS:       enableHLS,
		Resolutions:     resolutions,
		SegmentDuration: 10,    // 默认10秒
		ListSize:        0,     // 默认无限制
		Format:          "mpegts", // 默认mpegts格式
		Status:          status,
		Progress:        0,
		OutputPath:      "",
		ErrorMessage:    "",
	}, nil
}

// DefaultHLSConfig 创建默认HLS配置（禁用状态）
func DefaultHLSConfig() *HLSConfig {
	return &HLSConfig{
		EnableHLS:       false,
		Resolutions:     []ResolutionConfig{},
		SegmentDuration: 10,
		ListSize:        0,
		Format:          "mpegts",
		Status:          HLSStatusDisabled,
		Progress:        0,
		OutputPath:      "",
		ErrorMessage:    "",
	}
}

// Validate 验证HLS配置
func (hc *HLSConfig) Validate() error {
	if hc.EnableHLS && len(hc.Resolutions) == 0 {
		return fmt.Errorf("启用HLS时必须指定至少一个分辨率配置")
	}

	// 验证分辨率配置
	for i, res := range hc.Resolutions {
		if err := res.Validate(); err != nil {
			return fmt.Errorf("分辨率配置[%d]无效: %w", i, err)
		}
	}

	// 验证切片时长
	if hc.SegmentDuration <= 0 || hc.SegmentDuration > 60 {
		return fmt.Errorf("切片时长必须在1-60秒之间")
	}

	// 验证播放列表大小
	if hc.ListSize < 0 {
		return fmt.Errorf("播放列表大小不能为负数")
	}

	// 验证格式
	if hc.Format != "mpegts" && hc.Format != "fmp4" {
		return fmt.Errorf("HLS格式必须是mpegts或fmp4")
	}

	// 验证状态
	if !hc.Status.IsValid() {
		return fmt.Errorf("无效的HLS状态: %s", hc.Status)
	}

	// 验证进度
	if hc.Progress < 0 || hc.Progress > 100 {
		return fmt.Errorf("进度必须在0-100之间")
	}

	return nil
}

// IsEnabled 检查是否启用HLS
func (hc *HLSConfig) IsEnabled() bool {
	return hc.EnableHLS
}

// IsCompleted 检查HLS是否完成
func (hc *HLSConfig) IsCompleted() bool {
	return hc.Status == HLSStatusCompleted
}

// IsFailed 检查HLS是否失败
func (hc *HLSConfig) IsFailed() bool {
	return hc.Status == HLSStatusFailed
}

// IsProcessing 检查HLS是否正在处理
func (hc *HLSConfig) IsProcessing() bool {
	return hc.Status == HLSStatusProcessing
}

// IsPending 检查HLS是否待处理
func (hc *HLSConfig) IsPending() bool {
	return hc.Status == HLSStatusPending
}

// SetStatus 设置HLS状态
func (hc *HLSConfig) SetStatus(status HLSStatus) {
	hc.Status = status
}

// SetProgress 设置进度
func (hc *HLSConfig) SetProgress(progress int) {
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	hc.Progress = progress
}

// SetOutputPath 设置输出路径
func (hc *HLSConfig) SetOutputPath(path string) {
	hc.OutputPath = path
}

// SetErrorMessage 设置错误信息
func (hc *HLSConfig) SetErrorMessage(message string) {
	hc.ErrorMessage = message
}

// ToJSON 转换为JSON字符串
func (hc *HLSConfig) ToJSON() (string, error) {
	data, err := json.Marshal(hc.Resolutions)
	if err != nil {
		return "", fmt.Errorf("序列化分辨率配置失败: %w", err)
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (hc *HLSConfig) FromJSON(jsonStr string) error {
	if jsonStr == "" {
		hc.Resolutions = []ResolutionConfig{}
		return nil
	}
	
	var resolutions []ResolutionConfig
	if err := json.Unmarshal([]byte(jsonStr), &resolutions); err != nil {
		return fmt.Errorf("解析分辨率配置失败: %w", err)
	}
	
	// 验证解析的配置
	for i, res := range resolutions {
		if err := res.Validate(); err != nil {
			return fmt.Errorf("分辨率配置[%d]无效: %w", i, err)
		}
	}
	
	hc.Resolutions = resolutions
	return nil
}

// GetResolutionCount 获取分辨率数量
func (hc *HLSConfig) GetResolutionCount() int {
	return len(hc.Resolutions)
}

// HasResolution 检查是否包含指定分辨率
func (hc *HLSConfig) HasResolution(resolution string) bool {
	for _, res := range hc.Resolutions {
		if res.Resolution == resolution {
			return true
		}
	}
	return false
}

// GetResolutionByIndex 根据索引获取分辨率配置
func (hc *HLSConfig) GetResolutionByIndex(index int) (*ResolutionConfig, error) {
	if index < 0 || index >= len(hc.Resolutions) {
		return nil, fmt.Errorf("分辨率配置索引超出范围: %d", index)
	}
	return &hc.Resolutions[index], nil
}

// 验证分辨率格式
func validateResolution(resolution string) error {
	if resolution == "" {
		return fmt.Errorf("分辨率不能为空")
	}

	// 支持的分辨率格式
	validResolutions := []string{
		"2160p", "1440p", "1080p", "720p", "480p", "360p", "240p",
		"4K", "2K", "1080", "720", "480", "360", "240",
	}

	resolution = strings.ToLower(resolution)
	for _, valid := range validResolutions {
		if strings.ToLower(valid) == resolution {
			return nil
		}
	}

	return fmt.Errorf("不支持的分辨率格式: %s", resolution)
}

// 验证码率格式
func validateBitrate(bitrate string) error {
	if bitrate == "" {
		return fmt.Errorf("码率不能为空")
	}

	// 移除单位后缀
	bitrateValue := strings.ToLower(bitrate)
	if strings.HasSuffix(bitrateValue, "k") {
		bitrateValue = strings.TrimSuffix(bitrateValue, "k")
	} else if strings.HasSuffix(bitrateValue, "m") {
		bitrateValue = strings.TrimSuffix(bitrateValue, "m")
	} else if strings.HasSuffix(bitrateValue, "kbps") {
		bitrateValue = strings.TrimSuffix(bitrateValue, "kbps")
	} else if strings.HasSuffix(bitrateValue, "mbps") {
		bitrateValue = strings.TrimSuffix(bitrateValue, "mbps")
	}

	// 验证是否为有效数字
	if _, err := strconv.ParseFloat(bitrateValue, 64); err != nil {
		return fmt.Errorf("无效的码率格式: %s", bitrate)
	}

	return nil
}