package vo

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
