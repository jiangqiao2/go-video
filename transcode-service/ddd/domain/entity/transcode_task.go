package entity

import (
	"time"
	"transcode-service/ddd/domain/vo"
)

// TranscodeTaskEntity 转码任务实体
type TranscodeTaskEntity struct {
	taskUUID     string
	userUUID     string
	videoUUID    string
	originalPath string
	outputPath   string
	status       vo.TaskStatus
	progress     int
	errorMessage string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewTranscodeTaskEntity 创建转码任务实体
func NewTranscodeTaskEntity(
	taskUUID, userUUID, videoUUID, originalPath string,
) *TranscodeTaskEntity {
	now := time.Now()
	return &TranscodeTaskEntity{
		taskUUID:     taskUUID,
		userUUID:     userUUID,
		videoUUID:    videoUUID,
		originalPath: originalPath,
		status:       vo.TaskStatusPending,
		progress:     0,
		createdAt:    now,
		updatedAt:    now,
	}
}

// TaskUUID 获取任务UUID
func (t *TranscodeTaskEntity) TaskUUID() string {
	return t.taskUUID
}

// UserUUID 获取用户UUID
func (t *TranscodeTaskEntity) UserUUID() string {
	return t.userUUID
}

// VideoUUID 获取视频UUID
func (t *TranscodeTaskEntity) VideoUUID() string {
	return t.videoUUID
}

// OriginalPath 获取原始路径
func (t *TranscodeTaskEntity) OriginalPath() string {
	return t.originalPath
}

// OutputPath 获取输出路径
func (t *TranscodeTaskEntity) OutputPath() string {
	return t.outputPath
}

// Status 获取状态
func (t *TranscodeTaskEntity) Status() vo.TaskStatus {
	return t.status
}

// Progress 获取进度
func (t *TranscodeTaskEntity) Progress() int {
	return t.progress
}

// ErrorMessage 获取错误信息
func (t *TranscodeTaskEntity) ErrorMessage() string {
	return t.errorMessage
}

// CreatedAt 获取创建时间
func (t *TranscodeTaskEntity) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt 获取更新时间
func (t *TranscodeTaskEntity) UpdatedAt() time.Time {
	return t.updatedAt
}

// SetOutputPath 设置输出路径
func (t *TranscodeTaskEntity) SetOutputPath(outputPath string) {
	t.outputPath = outputPath
	t.updatedAt = time.Now()
}

// SetStatus 设置状态
func (t *TranscodeTaskEntity) SetStatus(status vo.TaskStatus) error {
	if !t.status.CanTransitionTo(status) {
		return vo.ErrInvalidStatusTransition
	}
	t.status = status
	t.updatedAt = time.Now()
	return nil
}

// SetProgress 设置进度
func (t *TranscodeTaskEntity) SetProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	t.progress = progress
	t.updatedAt = time.Now()
}

// SetErrorMessage 设置错误信息
func (t *TranscodeTaskEntity) SetErrorMessage(errorMessage string) {
	t.errorMessage = errorMessage
	t.updatedAt = time.Now()
}

// IsCompleted 检查是否已完成
func (t *TranscodeTaskEntity) IsCompleted() bool {
	return t.status == vo.TaskStatusCompleted
}

// IsFailed 检查是否失败
func (t *TranscodeTaskEntity) IsFailed() bool {
	return t.status == vo.TaskStatusFailed
}

// IsCancelled 检查是否已取消
func (t *TranscodeTaskEntity) IsCancelled() bool {
	return t.status == vo.TaskStatusCancelled
}

// IsProcessing 检查是否正在处理
func (t *TranscodeTaskEntity) IsProcessing() bool {
	return t.status == vo.TaskStatusProcessing
}

// IsPending 检查是否待处理
func (t *TranscodeTaskEntity) IsPending() bool {
	return t.status == vo.TaskStatusPending
}
