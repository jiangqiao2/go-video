package vo

import "fmt"

// TaskStatus 任务状态值对象（字符串枚举风格）
type TaskStatus struct {
	value string
}

var (
	TaskStatusPending    = TaskStatus{value: "pending"}
	TaskStatusProcessing = TaskStatus{value: "processing"}
	TaskStatusCompleted  = TaskStatus{value: "completed"}
	TaskStatusFailed     = TaskStatus{value: "failed"}
	TaskStatusCancelled  = TaskStatus{value: "cancelled"}
)

var taskStatusSet = []TaskStatus{
	TaskStatusPending,
	TaskStatusProcessing,
	TaskStatusCompleted,
	TaskStatusFailed,
	TaskStatusCancelled,
}

// NewTaskStatus 尝试从原始值构造，未知值回退为 pending。
func NewTaskStatus(value string) TaskStatus {
	for _, status := range taskStatusSet {
		if status.value == value {
			return status
		}
	}
	return TaskStatusPending
}

// NewTaskStatusFromString 从字符串创建任务状态，未知值报错。
func NewTaskStatusFromString(value string) (TaskStatus, error) {
	status := NewTaskStatus(value)
	if status == TaskStatusPending && value != TaskStatusPending.value {
		return status, fmt.Errorf("invalid task status string: %s", value)
	}
	return status, nil
}

// String 返回字符串值。
func (ts TaskStatus) String() string {
	return ts.value
}

// Value 返回字符串值（与 String 等价）。
func (ts TaskStatus) Value() string {
	return ts.value
}

// IsValid 是否属于定义的枚举集。
func (ts TaskStatus) IsValid() bool {
	for _, status := range taskStatusSet {
		if ts.value == status.value {
			return true
		}
	}
	return false
}

// CanTransitionTo 检查是否允许转换到目标状态。
func (ts TaskStatus) CanTransitionTo(target TaskStatus) bool {
	switch ts {
	case TaskStatusPending:
		return target == TaskStatusProcessing || target == TaskStatusFailed || target == TaskStatusCancelled
	case TaskStatusProcessing:
		return target == TaskStatusCompleted || target == TaskStatusFailed || target == TaskStatusCancelled
	case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return false // 终态不能再转换
	default:
		return false
	}
}
