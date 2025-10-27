package po

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
}

// TableName 指定表名
func (TranscodeTask) TableName() string {
	return "transcode_tasks"
}
