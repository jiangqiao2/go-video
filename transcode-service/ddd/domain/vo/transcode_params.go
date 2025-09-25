package vo

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TranscodeParams 转码参数值对象
type TranscodeParams struct {
	Resolution string
	Bitrate    string
}

// NewTranscodeParams 创建转码参数
func NewTranscodeParams(resolution, bitrate string) (*TranscodeParams, error) {
	if err := validateResolution(resolution); err != nil {
		return nil, err
	}
	if err := validateBitrate(bitrate); err != nil {
		return nil, err
	}
	return &TranscodeParams{
		Resolution: resolution,
		Bitrate:    bitrate,
	}, nil
}

// validateResolution 验证分辨率格式
func validateResolution(resolution string) error {
	validResolutions := []string{"480p", "720p", "1080p", "1440p", "2160p"}
	for _, valid := range validResolutions {
		if resolution == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid resolution: %s, supported: %v", resolution, validResolutions)
}

// validateBitrate 验证码率格式
func validateBitrate(bitrate string) error {
	if bitrate == "" {
		return errors.New("bitrate cannot be empty")
	}
	// 支持格式: 1000k, 2M, 5000
	if strings.HasSuffix(bitrate, "k") || strings.HasSuffix(bitrate, "K") {
		numStr := strings.TrimSuffix(strings.TrimSuffix(bitrate, "k"), "K")
		if _, err := strconv.Atoi(numStr); err != nil {
			return fmt.Errorf("invalid bitrate format: %s", bitrate)
		}
		return nil
	}
	if strings.HasSuffix(bitrate, "m") || strings.HasSuffix(bitrate, "M") {
		numStr := strings.TrimSuffix(strings.TrimSuffix(bitrate, "m"), "M")
		if _, err := strconv.Atoi(numStr); err != nil {
			return fmt.Errorf("invalid bitrate format: %s", bitrate)
		}
		return nil
	}
	// 纯数字
	if _, err := strconv.Atoi(bitrate); err != nil {
		return fmt.Errorf("invalid bitrate format: %s", bitrate)
	}
	return nil
}

// GetFFmpegArgs 获取FFmpeg参数
func (tp *TranscodeParams) GetFFmpegArgs() []string {
	args := []string{
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
	}
	
	// 设置分辨率
	switch tp.Resolution {
	case "480p":
		args = append(args, "-s", "854x480")
	case "720p":
		args = append(args, "-s", "1280x720")
	case "1080p":
		args = append(args, "-s", "1920x1080")
	case "1440p":
		args = append(args, "-s", "2560x1440")
	case "2160p":
		args = append(args, "-s", "3840x2160")
	}
	
	// 设置码率
	args = append(args, "-b:v", tp.Bitrate)
	
	return args
}