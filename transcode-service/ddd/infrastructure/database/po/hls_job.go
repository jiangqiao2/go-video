package po

import "time"

// HLSJob HLS切片作业持久化对象
type HLSJob struct {
    BaseModel
    JobUUID        string     `gorm:"column:job_uuid;type:varchar(36);uniqueIndex" json:"job_uuid"`
    UserUUID       string     `gorm:"column:user_uuid;type:varchar(36);index" json:"user_uuid"`
    VideoUUID      string     `gorm:"column:video_uuid;type:varchar(36);index" json:"video_uuid"`
    SourceJobUUID  *string    `gorm:"column:source_job_uuid;type:varchar(36);index" json:"source_job_uuid,omitempty"`
    SourceType     string     `gorm:"column:source_type;type:varchar(20)" json:"source_type"` // original|transcoded
    InputPath      string     `gorm:"column:input_path;type:varchar(512)" json:"input_path"`
    OutputDir      string     `gorm:"column:output_dir;type:varchar(512)" json:"output_dir"`
    MasterPlaylist *string    `gorm:"column:master_playlist;type:varchar(512)" json:"master_playlist,omitempty"`
    ProfilesJSON   *string    `gorm:"column:profiles_json;type:json" json:"profiles_json,omitempty"`
    Status         string     `gorm:"column:status;type:varchar(20);index" json:"status"`
    Progress       int        `gorm:"column:progress;type:int;default:0" json:"progress"`
    SegmentDuration int       `gorm:"column:segment_duration;type:int;default:10" json:"segment_duration"`
    ListSize       int        `gorm:"column:list_size;type:int;default:0" json:"list_size"`
    Format         string     `gorm:"column:format;type:varchar(20);default:'mpegts'" json:"format"`
    VariantCount   int        `gorm:"column:variant_count;type:int;default:0" json:"variant_count"`
    ErrorMessage   *string    `gorm:"column:error_message;type:varchar(500)" json:"error_message,omitempty"`
    StartedAt      *time.Time `gorm:"column:started_at;type:timestamp" json:"started_at,omitempty"`
    CompletedAt    *time.Time `gorm:"column:completed_at;type:timestamp" json:"completed_at,omitempty"`
}

// TableName 指定表名
func (HLSJob) TableName() string {
    return "hls_jobs"
}