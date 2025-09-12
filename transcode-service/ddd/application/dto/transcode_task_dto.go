package dto

import (
	"time"
	"transcode-service/ddd/domain/vo"
)

// CreateTranscodeTaskRequest 创建转码任务请求
type CreateTranscodeTaskRequest struct {
	UserID          string                 `json:"user_id" binding:"required"`
	SourceVideoPath string                 `json:"source_video_path" binding:"required"`
	OutputPath      string                 `json:"output_path" binding:"required"`
	Config          *vo.TranscodeConfig    `json:"config" binding:"required"`
	Priority        int                    `json:"priority,omitempty"` // 1-10，默认5
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CreateTranscodeTaskResponse 创建转码任务响应
type CreateTranscodeTaskResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// TranscodeTaskDTO 转码任务DTO
type TranscodeTaskDTO struct {
	TaskID          string                 `json:"task_id"`
	UserID          string                 `json:"user_id"`
	SourceVideoPath string                 `json:"source_video_path"`
	OutputPath      string                 `json:"output_path"`
	Config          *vo.TranscodeConfig    `json:"config"`
	Status          string                 `json:"status"`
	WorkerID        string                 `json:"worker_id,omitempty"`
	Priority        int                    `json:"priority"`
	RetryCount      int                    `json:"retry_count"`
	MaxRetryCount   int                    `json:"max_retry_count"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Progress        float64                `json:"progress"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	StartedAt       *string                `json:"started_at,omitempty"`
	CompletedAt     *string                `json:"completed_at,omitempty"`
	EstimatedTime   *string                `json:"estimated_time,omitempty"`
	ActualTime      *string                `json:"actual_time,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GetTaskRequest 获取任务请求
type GetTaskRequest struct {
	TaskID string `uri:"task_id" binding:"required"`
	UserID string `json:"user_id,omitempty"` // 可选，用于权限验证
}

// GetTaskResponse 获取任务响应
type GetTaskResponse struct {
	Task *TranscodeTaskDTO `json:"task"`
}

// ListTasksRequest 获取任务列表请求
type ListTasksRequest struct {
	UserID   string   `json:"user_id,omitempty"`
	Status   []string `json:"status,omitempty"`
	WorkerID string   `json:"worker_id,omitempty"`
	Priority *int     `json:"priority,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
	OrderBy  string   `json:"order_by,omitempty"`
	OrderDesc bool    `json:"order_desc,omitempty"`
}

// ListTasksResponse 获取任务列表响应
type ListTasksResponse struct {
	Tasks      []*TranscodeTaskDTO `json:"tasks"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	HasMore    bool                `json:"has_more"`
}

// UpdateTaskStatusRequest 更新任务状态请求
type UpdateTaskStatusRequest struct {
	TaskID string `uri:"task_id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

// UpdateTaskProgressRequest 更新任务进度请求
type UpdateTaskProgressRequest struct {
	TaskID   string  `uri:"task_id" binding:"required"`
	Progress float64 `json:"progress" binding:"required,min=0,max=100"`
}

// CancelTaskRequest 取消任务请求
type CancelTaskRequest struct {
	TaskID string `uri:"task_id" binding:"required"`
	UserID string `json:"user_id,omitempty"` // 可选，用于权限验证
}

// RetryTaskRequest 重试任务请求
type RetryTaskRequest struct {
	TaskID string `uri:"task_id" binding:"required"`
	UserID string `json:"user_id,omitempty"` // 可选，用于权限验证
}

// TaskStatisticsResponse 任务统计响应
type TaskStatisticsResponse struct {
	TotalTasks      int64 `json:"total_tasks"`
	PendingTasks    int64 `json:"pending_tasks"`
	ProcessingTasks int64 `json:"processing_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	FailedTasks     int64 `json:"failed_tasks"`
	CancelledTasks  int64 `json:"cancelled_tasks"`
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	TaskIDs   []string `json:"task_ids" binding:"required"`
	Operation string   `json:"operation" binding:"required"` // cancel, retry, delete
	UserID    string   `json:"user_id,omitempty"`           // 可选，用于权限验证
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedTasks  []string `json:"failed_tasks,omitempty"`
	Message      string   `json:"message"`
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}

// FormatTimePtr 格式化时间指针
func FormatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := FormatTime(*t)
	return &formatted
}

// FormatDuration 格式化时长
func FormatDuration(d time.Duration) string {
	return d.String()
}

// FormatDurationPtr 格式化时长指针
func FormatDurationPtr(d *time.Duration) *string {
	if d == nil {
		return nil
	}
	formatted := FormatDuration(*d)
	return &formatted
}