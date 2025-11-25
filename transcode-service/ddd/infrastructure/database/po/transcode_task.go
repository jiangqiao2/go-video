package po

import "time"

// TranscodeTask 转码任务持久化对象
type TranscodeTask struct {
	BaseModel
	TaskUUID   string `gorm:"column:task_uuid;type:varchar(36);uniqueIndex" json:"task_uuid"`
	UserUUID   string `gorm:"column:user_uuid;type:varchar(36);index" json:"user_uuid"`
	VideoUUID  string `gorm:"column:video_uuid;type:varchar(36);index" json:"video_uuid"`
	InputPath  string `gorm:"column:input_path;type:varchar(512)" json:"input_path"`
	OutputPath string `gorm:"column:output_path;type:varchar(512)" json:"output_path"`
	Resolution string `gorm:"column:resolution;type:varchar(50)" json:"resolution"` // 480p, 720p, 1080p
	Bitrate    string `gorm:"column:bitrate;type:varchar(50)" json:"bitrate"`
	Status     string `gorm:"column:status;type:varchar(20);index" json:"status"` // pending, processing, completed, failed
	Progress   int    `gorm:"column:progress;type:int" json:"progress"`
	Message    string `gorm:"column:message;type:varchar(255)" json:"message"` // 错误信息

	// HLS相关字段
	HLSEnabled         bool       `gorm:"column:hls_enabled;type:tinyint;default:0;index" json:"hls_enabled"`
	HLSStatus          *string    `gorm:"column:hls_status;type:varchar(20);index" json:"hls_status"`
	HLSProgress        int        `gorm:"column:hls_progress;type:int;default:0" json:"hls_progress"`
	HLSOutputPath      *string    `gorm:"column:hls_output_path;type:varchar(512)" json:"hls_output_path"`
	HLSSegmentDuration int        `gorm:"column:hls_segment_duration;type:int;default:10" json:"hls_segment_duration"`
	HLSListSize        int        `gorm:"column:hls_list_size;type:int;default:0" json:"hls_list_size"`
	HLSFormat          string     `gorm:"column:hls_format;type:varchar(20);default:'ts'" json:"hls_format"`
	HLSErrorMessage    *string    `gorm:"column:hls_error_message;type:varchar(500)" json:"hls_error_message"`
	HLSStartedAt       *time.Time `gorm:"column:hls_started_at;type:timestamp" json:"hls_started_at"`
	HLSCompletedAt     *time.Time `gorm:"column:hls_completed_at;type:timestamp" json:"hls_completed_at"`
}

// TableName 指定表名
func (TranscodeTask) TableName() string {
	return "transcode_tasks"
}
