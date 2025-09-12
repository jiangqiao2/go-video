package dto

import "time"

// RegisterWorkerRequest 注册Worker请求
type RegisterWorkerRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	MaxTasks int    `json:"max_tasks" binding:"required,min=1"`
}

// RegisterWorkerResponse 注册Worker响应
type RegisterWorkerResponse struct {
	WorkerID     string `json:"worker_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	RegisteredAt string `json:"registered_at"`
}

// WorkerDTO Worker DTO
type WorkerDTO struct {
	WorkerID        string                 `json:"worker_id"`
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	MaxTasks        int                    `json:"max_tasks"`
	CurrentTasks    int                    `json:"current_tasks"`
	CPUUsage        float64                `json:"cpu_usage"`
	MemoryUsage     float64                `json:"memory_usage"`
	LoadFactor      float64                `json:"load_factor"`
	LastHeartbeatAt string                 `json:"last_heartbeat_at"`
	RegisteredAt    string                 `json:"registered_at"`
	UpdatedAt       string                 `json:"updated_at"`
	SystemInfo      map[string]string      `json:"system_info,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	IsHealthy       bool                   `json:"is_healthy"`
}

// GetWorkerRequest 获取Worker请求
type GetWorkerRequest struct {
	WorkerID string `uri:"worker_id" binding:"required"`
}

// GetWorkerResponse 获取Worker响应
type GetWorkerResponse struct {
	Worker *WorkerDTO `json:"worker"`
}

// ListWorkersRequest 获取Worker列表请求
type ListWorkersRequest struct {
	Status    []string `json:"status,omitempty"`
	Healthy   *bool    `json:"healthy,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	OrderBy   string   `json:"order_by,omitempty"`
	OrderDesc bool     `json:"order_desc,omitempty"`
}

// ListWorkersResponse 获取Worker列表响应
type ListWorkersResponse struct {
	Workers []*WorkerDTO `json:"workers"`
	Total   int64        `json:"total"`
	Limit   int          `json:"limit"`
	Offset  int          `json:"offset"`
	HasMore bool         `json:"has_more"`
}

// UpdateWorkerStatusRequest 更新Worker状态请求
type UpdateWorkerStatusRequest struct {
	WorkerID string `uri:"worker_id" binding:"required"`
	Status   string `json:"status" binding:"required"`
}

// WorkerHeartbeatRequest Worker心跳请求
type WorkerHeartbeatRequest struct {
	WorkerID     string            `json:"worker_id" binding:"required"`
	Status       string            `json:"status" binding:"required"`
	CurrentTasks int               `json:"current_tasks" binding:"min=0"`
	MaxTasks     int               `json:"max_tasks" binding:"required,min=1"`
	CPUUsage     float64           `json:"cpu_usage" binding:"min=0,max=100"`
	MemoryUsage  float64           `json:"memory_usage" binding:"min=0,max=100"`
	SystemInfo   map[string]string `json:"system_info,omitempty"`
}

// WorkerHeartbeatResponse Worker心跳响应
type WorkerHeartbeatResponse struct {
	WorkerID  string `json:"worker_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// WorkerStatisticsResponse Worker统计响应
type WorkerStatisticsResponse struct {
	TotalWorkers       int64   `json:"total_workers"`
	OnlineWorkers      int64   `json:"online_workers"`
	OfflineWorkers     int64   `json:"offline_workers"`
	BusyWorkers        int64   `json:"busy_workers"`
	IdleWorkers        int64   `json:"idle_workers"`
	MaintenanceWorkers int64   `json:"maintenance_workers"`
	TotalTasks         int64   `json:"total_tasks"`
	AverageCPUUsage    float64 `json:"average_cpu_usage"`
	AverageMemoryUsage float64 `json:"average_memory_usage"`
	AverageLoadFactor  float64 `json:"average_load_factor"`
}

// DeleteWorkerRequest 删除Worker请求
type DeleteWorkerRequest struct {
	WorkerID string `uri:"worker_id" binding:"required"`
	Force    bool   `json:"force,omitempty"` // 强制删除，即使有正在处理的任务
}

// WorkerTasksRequest 获取Worker任务请求
type WorkerTasksRequest struct {
	WorkerID string   `uri:"worker_id" binding:"required"`
	Status   []string `json:"status,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// WorkerTasksResponse 获取Worker任务响应
type WorkerTasksResponse struct {
	WorkerID string             `json:"worker_id"`
	Tasks    []*TranscodeTaskDTO `json:"tasks"`
	Total    int64              `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	HasMore  bool               `json:"has_more"`
}

// BatchWorkerOperationRequest 批量Worker操作请求
type BatchWorkerOperationRequest struct {
	WorkerIDs []string `json:"worker_ids" binding:"required"`
	Operation string   `json:"operation" binding:"required"` // online, offline, maintenance, delete
	Force     bool     `json:"force,omitempty"`             // 强制操作
}

// BatchWorkerOperationResponse 批量Worker操作响应
type BatchWorkerOperationResponse struct {
	SuccessCount   int      `json:"success_count"`
	FailedCount    int      `json:"failed_count"`
	FailedWorkers  []string `json:"failed_workers,omitempty"`
	Message        string   `json:"message"`
}

// WorkerHealthCheckResponse Worker健康检查响应
type WorkerHealthCheckResponse struct {
	WorkerID      string `json:"worker_id"`
	IsHealthy     bool   `json:"is_healthy"`
	LastHeartbeat string `json:"last_heartbeat"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

// CalculateLoadFactor 计算负载因子
func (w *WorkerDTO) CalculateLoadFactor() {
	if w.MaxTasks > 0 {
		w.LoadFactor = float64(w.CurrentTasks) / float64(w.MaxTasks)
	} else {
		w.LoadFactor = 0.0
	}
}

// CheckHealthy 检查是否健康
func (w *WorkerDTO) CheckHealthy(timeout time.Duration) {
	lastHeartbeat, err := time.Parse("2006-01-02T15:04:05Z07:00", w.LastHeartbeatAt)
	if err != nil {
		w.IsHealthy = false
		return
	}
	
	w.IsHealthy = time.Since(lastHeartbeat) <= timeout && 
		(w.Status == "online" || w.Status == "idle" || w.Status == "busy")
}