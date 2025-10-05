package vo

import (
	"errors"
	"fmt"
)

// 错误定义
var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// TaskStatus 任务状态值对象
type TaskStatus int

const (
	TaskStatusPending    TaskStatus = 0 // 待处理
	TaskStatusProcessing TaskStatus = 1 // 处理中
	TaskStatusCompleted  TaskStatus = 2 // 已完成
	TaskStatusFailed     TaskStatus = 3 // 失败
	TaskStatusCancelled  TaskStatus = 4 // 已取消
)

// String 返回状态的字符串表示
func (ts TaskStatus) String() string {
	switch ts {
	case TaskStatusPending:
		return "pending"
	case TaskStatusProcessing:
		return "processing"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	case TaskStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// IsValid 检查状态是否有效
func (ts TaskStatus) IsValid() bool {
	return ts >= TaskStatusPending && ts <= TaskStatusCancelled
}

// CanTransitionTo 检查是否可以转换到目标状态
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

// NewTaskStatus 创建任务状态
func NewTaskStatus(status int) (TaskStatus, error) {
	ts := TaskStatus(status)
	if !ts.IsValid() {
		return 0, fmt.Errorf("invalid task status: %d", status)
	}
	return ts, nil
}

// NewTaskStatusFromString 从字符串创建任务状态
func NewTaskStatusFromString(status string) (TaskStatus, error) {
	switch status {
	case "pending":
		return TaskStatusPending, nil
	case "processing":
		return TaskStatusProcessing, nil
	case "completed":
		return TaskStatusCompleted, nil
	case "failed":
		return TaskStatusFailed, nil
	case "cancelled":
		return TaskStatusCancelled, nil
	default:
		return 0, fmt.Errorf("invalid task status string: %s", status)
	}
}

// ToInt 转换为整数
func (ts TaskStatus) ToInt() int {
	return int(ts)
}