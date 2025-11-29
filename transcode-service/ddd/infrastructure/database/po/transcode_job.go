package po

import "time"

// TranscodeJob 完整视频转码作业持久化对象
type TranscodeJob struct {
	BaseModel
	JobUUID       string     `gorm:"column:job_uuid;type:varchar(36);uniqueIndex" json:"job_uuid"`
	UserUUID      string     `gorm:"column:user_uuid;type:varchar(36);index" json:"user_uuid"`
	VideoUUID     string     `gorm:"column:video_uuid;type:varchar(36);index" json:"video_uuid"`
	VideoPushUUID string     `gorm:"column:video_push_uuid;type:varchar(36);index" json:"video_push_uuid"`
	InputPath     string     `gorm:"column:input_path;type:varchar(512)" json:"input_path"`
	OutputPath    string     `gorm:"column:output_path;type:varchar(512)" json:"output_path"`
	Resolution    string     `gorm:"column:resolution;type:varchar(50)" json:"resolution"`
	Bitrate       string     `gorm:"column:bitrate;type:varchar(50)" json:"bitrate"`
	Status        string     `gorm:"column:status;type:varchar(20);index" json:"status"`
	Progress      int        `gorm:"column:progress;type:int" json:"progress"`
	Message       string     `gorm:"column:message;type:varchar(255)" json:"message"`
	WorkerID      *string    `gorm:"column:worker_id;type:varchar(36);index" json:"worker_id,omitempty"`
	Priority      int        `gorm:"column:priority;type:int;default:5" json:"priority"`
	RetryCount    int        `gorm:"column:retry_count;type:int;default:0" json:"retry_count"`
	MaxRetryCount int        `gorm:"column:max_retry_count;type:int;default:3" json:"max_retry_count"`
	StartedAt     *time.Time `gorm:"column:started_at;type:timestamp" json:"started_at,omitempty"`
	CompletedAt   *time.Time `gorm:"column:completed_at;type:timestamp" json:"completed_at,omitempty"`
	EstimatedTime *int64     `gorm:"column:estimated_time;type:bigint" json:"estimated_time,omitempty"`
	ActualTime    *int64     `gorm:"column:actual_time;type:bigint" json:"actual_time,omitempty"`
	Metadata      *string    `gorm:"column:metadata;type:json" json:"metadata,omitempty"`
}

// TableName 指定表名
func (TranscodeJob) TableName() string {
	return "transcode_jobs"
}
