package vo

import "time"

// WorkerStatus Worker状态
type WorkerStatus string

const (
	// WorkerStatusOnline 在线
	WorkerStatusOnline WorkerStatus = "online"
	// WorkerStatusOffline 离线
	WorkerStatusOffline WorkerStatus = "offline"
	// WorkerStatusBusy 忙碌
	WorkerStatusBusy WorkerStatus = "busy"
	// WorkerStatusIdle 空闲
	WorkerStatusIdle WorkerStatus = "idle"
	// WorkerStatusMaintenance 维护中
	WorkerStatusMaintenance WorkerStatus = "maintenance"
)

// IsValid 检查状态是否有效
func (s WorkerStatus) IsValid() bool {
	switch s {
	case WorkerStatusOnline, WorkerStatusOffline, WorkerStatusBusy,
		WorkerStatusIdle, WorkerStatusMaintenance:
		return true
	default:
		return false
	}
}

// String 返回状态字符串
func (s WorkerStatus) String() string {
	return string(s)
}

// CanAcceptTask 检查是否可以接受新任务
func (s WorkerStatus) CanAcceptTask() bool {
	return s == WorkerStatusOnline || s == WorkerStatusIdle
}

// TranscodeConfig 转码配置值对象
type TranscodeConfig struct {
	Resolution string `json:"resolution"` // 分辨率 如: 1280x720
	Bitrate    string `json:"bitrate"`    // 码率 如: 2000k
	Codec      string `json:"codec"`      // 编码器 如: libx264
	Preset     string `json:"preset"`     // 预设 如: medium
	Format     string `json:"format"`     // 格式 如: mp4
}

// IsValid 检查转码配置是否有效
func (c *TranscodeConfig) IsValid() bool {
	return c.Resolution != "" && c.Bitrate != "" && c.Codec != "" && c.Format != ""
}

// WorkerHeartbeat Worker心跳信息
type WorkerHeartbeat struct {
	WorkerID         string            `json:"worker_id"`
	Status           WorkerStatus      `json:"status"`
	CurrentTasks     int               `json:"current_tasks"`
	MaxTasks         int               `json:"max_tasks"`
	CPUUsage         float64           `json:"cpu_usage"`
	MemoryUsage      float64           `json:"memory_usage"`
	LastHeartbeatAt  time.Time         `json:"last_heartbeat_at"`
	SystemInfo       map[string]string `json:"system_info"`
}

// IsHealthy 检查Worker是否健康
func (h *WorkerHeartbeat) IsHealthy(timeout time.Duration) bool {
	return time.Since(h.LastHeartbeatAt) <= timeout
}

// CanAcceptTask 检查是否可以接受新任务
func (h *WorkerHeartbeat) CanAcceptTask() bool {
	return h.Status.CanAcceptTask() && h.CurrentTasks < h.MaxTasks
}