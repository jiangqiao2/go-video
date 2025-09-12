package po

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// TranscodeTaskPO 转码任务持久化对象
type TranscodeTaskPO struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID          string    `gorm:"uniqueIndex;size:36;not null" json:"task_id"`
	UserID          string    `gorm:"index;size:36;not null" json:"user_id"`
	SourceVideoPath string    `gorm:"size:500;not null" json:"source_video_path"`
	OutputPath      string    `gorm:"size:500;not null" json:"output_path"`
	Config          JSONMap   `gorm:"type:json" json:"config"`
	Status          string    `gorm:"index;size:20;not null" json:"status"`
	WorkerID        string    `gorm:"index;size:36" json:"worker_id"`
	Priority        int       `gorm:"index;default:5" json:"priority"`
	RetryCount      int       `gorm:"default:0" json:"retry_count"`
	MaxRetryCount   int       `gorm:"default:3" json:"max_retry_count"`
	ErrorMessage    string    `gorm:"type:text" json:"error_message"`
	Progress        float64   `gorm:"default:0" json:"progress"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	EstimatedTime   *int64    `json:"estimated_time"` // 纳秒
	ActualTime      *int64    `json:"actual_time"`    // 纳秒
	Metadata        JSONMap   `gorm:"type:json" json:"metadata"`
}

// TableName 指定表名
func (TranscodeTaskPO) TableName() string {
	return "transcode_tasks"
}

// WorkerPO Worker持久化对象
type WorkerPO struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkerID        string    `gorm:"uniqueIndex;size:36;not null" json:"worker_id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	Status          string    `gorm:"index;size:20;not null" json:"status"`
	MaxTasks        int       `gorm:"not null" json:"max_tasks"`
	CurrentTasks    int       `gorm:"default:0" json:"current_tasks"`
	CPUUsage        float64   `gorm:"default:0" json:"cpu_usage"`
	MemoryUsage     float64   `gorm:"default:0" json:"memory_usage"`
	LastHeartbeatAt time.Time `gorm:"index" json:"last_heartbeat_at"`
	RegisteredAt    time.Time `json:"registered_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	SystemInfo      JSONMap   `gorm:"type:json" json:"system_info"`
	Metadata        JSONMap   `gorm:"type:json" json:"metadata"`
}

// TableName 指定表名
func (WorkerPO) TableName() string {
	return "workers"
}

// JSONMap 自定义JSON类型
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
}