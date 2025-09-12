package entity

import (
	"time"
	"transcode-service/ddd/domain/vo"
)

// WorkerEntity Worker实体
type WorkerEntity struct {
	workerID        string                 // Worker ID
	name            string                 // Worker名称
	status          vo.WorkerStatus        // Worker状态
	maxTasks        int                    // 最大并发任务数
	currentTasks    int                    // 当前任务数
	cpuUsage        float64                // CPU使用率
	memoryUsage     float64                // 内存使用率
	lastHeartbeatAt time.Time              // 最后心跳时间
	registeredAt    time.Time              // 注册时间
	updatedAt       time.Time              // 更新时间
	systemInfo      map[string]string      // 系统信息
	metadata        map[string]interface{} // 元数据
}

// NewWorkerEntity 创建新的Worker实体
func NewWorkerEntity(workerID, name string, maxTasks int) *WorkerEntity {
	now := time.Now()
	return &WorkerEntity{
		workerID:        workerID,
		name:            name,
		status:          vo.WorkerStatusOffline,
		maxTasks:        maxTasks,
		currentTasks:    0,
		cpuUsage:        0.0,
		memoryUsage:     0.0,
		lastHeartbeatAt: now,
		registeredAt:    now,
		updatedAt:       now,
		systemInfo:      make(map[string]string),
		metadata:        make(map[string]interface{}),
	}
}

// Getters
func (w *WorkerEntity) WorkerID() string                 { return w.workerID }
func (w *WorkerEntity) Name() string                     { return w.name }
func (w *WorkerEntity) Status() vo.WorkerStatus          { return w.status }
func (w *WorkerEntity) MaxTasks() int                    { return w.maxTasks }
func (w *WorkerEntity) CurrentTasks() int                { return w.currentTasks }
func (w *WorkerEntity) CPUUsage() float64                { return w.cpuUsage }
func (w *WorkerEntity) MemoryUsage() float64             { return w.memoryUsage }
func (w *WorkerEntity) LastHeartbeatAt() time.Time       { return w.lastHeartbeatAt }
func (w *WorkerEntity) RegisteredAt() time.Time          { return w.registeredAt }
func (w *WorkerEntity) UpdatedAt() time.Time             { return w.updatedAt }
func (w *WorkerEntity) SystemInfo() map[string]string    { return w.systemInfo }
func (w *WorkerEntity) Metadata() map[string]interface{} { return w.metadata }

// Online Worker上线
func (w *WorkerEntity) Online() error {
	if w.status == vo.WorkerStatusMaintenance {
		return NewDomainError("worker is in maintenance mode")
	}
	
	w.status = vo.WorkerStatusOnline
	w.updatedAt = time.Now()
	return nil
}

// Offline Worker下线
func (w *WorkerEntity) Offline() error {
	if w.currentTasks > 0 {
		return NewDomainError("cannot go offline while processing tasks")
	}
	
	w.status = vo.WorkerStatusOffline
	w.updatedAt = time.Now()
	return nil
}

// SetMaintenance 设置维护模式
func (w *WorkerEntity) SetMaintenance() error {
	if w.currentTasks > 0 {
		return NewDomainError("cannot enter maintenance mode while processing tasks")
	}
	
	w.status = vo.WorkerStatusMaintenance
	w.updatedAt = time.Now()
	return nil
}

// UpdateHeartbeat 更新心跳
func (w *WorkerEntity) UpdateHeartbeat(heartbeat *vo.WorkerHeartbeat) error {
	if heartbeat.WorkerID != w.workerID {
		return NewDomainError("worker ID mismatch")
	}
	
	w.status = heartbeat.Status
	w.currentTasks = heartbeat.CurrentTasks
	w.cpuUsage = heartbeat.CPUUsage
	w.memoryUsage = heartbeat.MemoryUsage
	w.lastHeartbeatAt = heartbeat.LastHeartbeatAt
	w.updatedAt = time.Now()
	
	// 更新系统信息
	for k, v := range heartbeat.SystemInfo {
		w.systemInfo[k] = v
	}
	
	// 根据当前任务数更新状态
	if w.status == vo.WorkerStatusOnline {
		if w.currentTasks >= w.maxTasks {
			w.status = vo.WorkerStatusBusy
		} else {
			w.status = vo.WorkerStatusIdle
		}
	}
	
	return nil
}

// AssignTask 分配任务
func (w *WorkerEntity) AssignTask() error {
	if !w.CanAcceptTask() {
		return NewDomainError("worker cannot accept new tasks")
	}
	
	w.currentTasks++
	w.updatedAt = time.Now()
	
	// 更新状态
	if w.currentTasks >= w.maxTasks {
		w.status = vo.WorkerStatusBusy
	}
	
	return nil
}

// CompleteTask 完成任务
func (w *WorkerEntity) CompleteTask() error {
	if w.currentTasks <= 0 {
		return NewDomainError("no tasks to complete")
	}
	
	w.currentTasks--
	w.updatedAt = time.Now()
	
	// 更新状态
	if w.status == vo.WorkerStatusBusy && w.currentTasks < w.maxTasks {
		w.status = vo.WorkerStatusIdle
	}
	
	return nil
}

// CanAcceptTask 检查是否可以接受新任务
func (w *WorkerEntity) CanAcceptTask() bool {
	return w.status.CanAcceptTask() && w.currentTasks < w.maxTasks
}

// IsHealthy 检查Worker是否健康
func (w *WorkerEntity) IsHealthy(timeout time.Duration) bool {
	if w.status == vo.WorkerStatusOffline || w.status == vo.WorkerStatusMaintenance {
		return false
	}
	return time.Since(w.lastHeartbeatAt) <= timeout
}

// GetLoadFactor 获取负载因子 (0.0 - 1.0)
func (w *WorkerEntity) GetLoadFactor() float64 {
	if w.maxTasks == 0 {
		return 1.0
	}
	return float64(w.currentTasks) / float64(w.maxTasks)
}

// SetSystemInfo 设置系统信息
func (w *WorkerEntity) SetSystemInfo(key, value string) {
	w.systemInfo[key] = value
	w.updatedAt = time.Now()
}

// SetMetadata 设置元数据
func (w *WorkerEntity) SetMetadata(key string, value interface{}) {
	w.metadata[key] = value
	w.updatedAt = time.Now()
}

// GetMetadata 获取元数据
func (w *WorkerEntity) GetMetadata(key string) (interface{}, bool) {
	value, exists := w.metadata[key]
	return value, exists
}

// UpdateMaxTasks 更新最大任务数
func (w *WorkerEntity) UpdateMaxTasks(maxTasks int) error {
	if maxTasks <= 0 {
		return NewDomainError("max tasks must be greater than 0")
	}
	
	if maxTasks < w.currentTasks {
		return NewDomainError("max tasks cannot be less than current tasks")
	}
	
	w.maxTasks = maxTasks
	w.updatedAt = time.Now()
	
	// 更新状态
	if w.status == vo.WorkerStatusBusy && w.currentTasks < w.maxTasks {
		w.status = vo.WorkerStatusIdle
	} else if w.status == vo.WorkerStatusIdle && w.currentTasks >= w.maxTasks {
		w.status = vo.WorkerStatusBusy
	}
	
	return nil
}