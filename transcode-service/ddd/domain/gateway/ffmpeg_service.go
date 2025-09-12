package gateway

import (
	"context"
	"transcode-service/ddd/domain/vo"
)

// FFmpegService FFmpeg服务接口
type FFmpegService interface {
	// TranscodeVideo 转码视频
	TranscodeVideo(ctx context.Context, request *TranscodeRequest) (*TranscodeResult, error)

	// GetVideoInfo 获取视频信息
	GetVideoInfo(ctx context.Context, videoPath string) (*VideoInfo, error)

	// ValidateVideo 验证视频文件
	ValidateVideo(ctx context.Context, videoPath string) error

	// CancelTranscode 取消转码
	CancelTranscode(ctx context.Context, taskID string) error

	// GetTranscodeProgress 获取转码进度
	GetTranscodeProgress(ctx context.Context, taskID string) (*TranscodeProgress, error)
}

// TranscodeRequest 转码请求
type TranscodeRequest struct {
	TaskID           string                 `json:"task_id"`
	InputPath        string                 `json:"input_path"`
	OutputPath       string                 `json:"output_path"`
	Config           *vo.TranscodeConfig    `json:"config"`
	ProgressCallback func(progress float64) // 进度回调
}

// TranscodeResult 转码结果
type TranscodeResult struct {
	TaskID       string                 `json:"task_id"`
	Success      bool                   `json:"success"`
	OutputPath   string                 `json:"output_path"`
	FileSize     int64                  `json:"file_size"`
	Duration     float64                `json:"duration"`
	Bitrate      string                 `json:"bitrate"`
	Resolution   string                 `json:"resolution"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// VideoInfo 视频信息
type VideoInfo struct {
	FilePath     string                 `json:"file_path"`
	FileSize     int64                  `json:"file_size"`
	Duration     float64                `json:"duration"`
	Bitrate      string                 `json:"bitrate"`
	Resolution   string                 `json:"resolution"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	FrameRate    float64                `json:"frame_rate"`
	Codec        string                 `json:"codec"`
	Format       string                 `json:"format"`
	AudioCodec   string                 `json:"audio_codec,omitempty"`
	AudioBitrate string                 `json:"audio_bitrate,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TranscodeProgress 转码进度
type TranscodeProgress struct {
	TaskID      string  `json:"task_id"`
	Progress    float64 `json:"progress"`     // 进度百分比 0-100
	CurrentTime float64 `json:"current_time"` // 当前处理时间（秒）
	TotalTime   float64 `json:"total_time"`   // 总时长（秒）
	Speed       string  `json:"speed"`        // 处理速度
	Bitrate     string  `json:"bitrate"`      // 当前码率
	FPS         float64 `json:"fps"`          // 当前帧率
	IsRunning   bool    `json:"is_running"`   // 是否正在运行
}
