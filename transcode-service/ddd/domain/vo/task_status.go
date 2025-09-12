package vo

// TaskStatus 转码任务状态
type TaskStatus string

const (
	// TaskStatusPending 待处理
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusAssigned 已分配
	TaskStatusAssigned TaskStatus = "assigned"
	// TaskStatusProcessing 处理中
	TaskStatusProcessing TaskStatus = "processing"
	// TaskStatusCompleted 已完成
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed 失败
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusRetrying 重试中
	TaskStatusRetrying TaskStatus = "retrying"
	// TaskStatusCancelled 已取消
	TaskStatusCancelled TaskStatus = "cancelled"
)

// IsValid 检查状态是否有效
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusAssigned, TaskStatusProcessing,
		TaskStatusCompleted, TaskStatusFailed, TaskStatusRetrying, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// String 返回状态字符串
func (s TaskStatus) String() string {
	return string(s)
}

// IsFinalStatus 检查是否为最终状态
func (s TaskStatus) IsFinalStatus() bool {
	return s == TaskStatusCompleted || s == TaskStatusFailed || s == TaskStatusCancelled
}

// CanTransitionTo 检查是否可以转换到目标状态
func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	switch s {
	case TaskStatusPending:
		return target == TaskStatusAssigned || target == TaskStatusCancelled
	case TaskStatusAssigned:
		return target == TaskStatusProcessing || target == TaskStatusCancelled
	case TaskStatusProcessing:
		return target == TaskStatusCompleted || target == TaskStatusFailed
	case TaskStatusFailed:
		return target == TaskStatusRetrying || target == TaskStatusCancelled
	case TaskStatusRetrying:
		return target == TaskStatusAssigned || target == TaskStatusCancelled
	case TaskStatusCompleted, TaskStatusCancelled:
		return false // 最终状态不能转换
	default:
		return false
	}
}