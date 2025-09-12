package entity

import (
	"time"
	"transcode-service/ddd/domain/vo"
	"github.com/google/uuid"
)

// TranscodeTaskEntity 转码任务实体
type TranscodeTaskEntity struct {
	taskID          string                 // 任务ID
	userID          string                 // 用户ID
	sourceVideoPath string                 // 源视频路径
	outputPath      string                 // 输出路径
	config          *vo.TranscodeConfig    // 转码配置
	status          vo.TaskStatus          // 任务状态
	workerID        string                 // 分配的Worker ID
	priority        int                    // 任务优先级 (1-10, 数字越大优先级越高)
	retryCount      int                    // 重试次数
	maxRetryCount   int                    // 最大重试次数
	errorMessage    string                 // 错误信息
	progress        float64                // 进度百分比 (0-100)
	createdAt       time.Time              // 创建时间
	updatedAt       time.Time              // 更新时间
	startedAt       *time.Time             // 开始时间
	completedAt     *time.Time             // 完成时间
	estimatedTime   *time.Duration         // 预估耗时
	actualTime      *time.Duration         // 实际耗时
	metadata        map[string]interface{} // 元数据
}

// NewTranscodeTaskEntity 创建新的转码任务实体
func NewTranscodeTaskEntity(
	userID string,
	sourceVideoPath string,
	outputPath string,
	config *vo.TranscodeConfig,
	priority int,
	maxRetryCount int,
) *TranscodeTaskEntity {
	now := time.Now()
	return &TranscodeTaskEntity{
		taskID:          uuid.New().String(),
		userID:          userID,
		sourceVideoPath: sourceVideoPath,
		outputPath:      outputPath,
		config:          config,
		status:          vo.TaskStatusPending,
		priority:        priority,
		retryCount:      0,
		maxRetryCount:   maxRetryCount,
		progress:        0.0,
		createdAt:       now,
		updatedAt:       now,
		metadata:        make(map[string]interface{}),
	}
}

// Getters
func (t *TranscodeTaskEntity) TaskID() string                 { return t.taskID }
func (t *TranscodeTaskEntity) UserID() string                 { return t.userID }
func (t *TranscodeTaskEntity) SourceVideoPath() string        { return t.sourceVideoPath }
func (t *TranscodeTaskEntity) OutputPath() string             { return t.outputPath }
func (t *TranscodeTaskEntity) Config() *vo.TranscodeConfig     { return t.config }
func (t *TranscodeTaskEntity) Status() vo.TaskStatus           { return t.status }
func (t *TranscodeTaskEntity) WorkerID() string               { return t.workerID }
func (t *TranscodeTaskEntity) Priority() int                  { return t.priority }
func (t *TranscodeTaskEntity) RetryCount() int                { return t.retryCount }
func (t *TranscodeTaskEntity) MaxRetryCount() int             { return t.maxRetryCount }
func (t *TranscodeTaskEntity) ErrorMessage() string           { return t.errorMessage }
func (t *TranscodeTaskEntity) Progress() float64              { return t.progress }
func (t *TranscodeTaskEntity) CreatedAt() time.Time           { return t.createdAt }
func (t *TranscodeTaskEntity) UpdatedAt() time.Time           { return t.updatedAt }
func (t *TranscodeTaskEntity) StartedAt() *time.Time          { return t.startedAt }
func (t *TranscodeTaskEntity) CompletedAt() *time.Time        { return t.completedAt }
func (t *TranscodeTaskEntity) EstimatedTime() *time.Duration  { return t.estimatedTime }
func (t *TranscodeTaskEntity) ActualTime() *time.Duration     { return t.actualTime }
func (t *TranscodeTaskEntity) Metadata() map[string]interface{} { return t.metadata }

// AssignToWorker 分配给Worker
func (t *TranscodeTaskEntity) AssignToWorker(workerID string) error {
	if !t.status.CanTransitionTo(vo.TaskStatusAssigned) {
		return NewDomainError("cannot assign task in current status: " + t.status.String())
	}
	
	t.workerID = workerID
	t.status = vo.TaskStatusAssigned
	t.updatedAt = time.Now()
	return nil
}

// StartProcessing 开始处理
func (t *TranscodeTaskEntity) StartProcessing() error {
	if !t.status.CanTransitionTo(vo.TaskStatusProcessing) {
		return NewDomainError("cannot start processing task in current status: " + t.status.String())
	}
	
	now := time.Now()
	t.status = vo.TaskStatusProcessing
	t.startedAt = &now
	t.updatedAt = now
	return nil
}

// UpdateProgress 更新进度
func (t *TranscodeTaskEntity) UpdateProgress(progress float64) error {
	if t.status != vo.TaskStatusProcessing {
		return NewDomainError("can only update progress for processing tasks")
	}
	
	if progress < 0 || progress > 100 {
		return NewDomainError("progress must be between 0 and 100")
	}
	
	t.progress = progress
	t.updatedAt = time.Now()
	return nil
}

// Complete 完成任务
func (t *TranscodeTaskEntity) Complete() error {
	if !t.status.CanTransitionTo(vo.TaskStatusCompleted) {
		return NewDomainError("cannot complete task in current status: " + t.status.String())
	}
	
	now := time.Now()
	t.status = vo.TaskStatusCompleted
	t.progress = 100.0
	t.completedAt = &now
	t.updatedAt = now
	
	// 计算实际耗时
	if t.startedAt != nil {
		actualTime := now.Sub(*t.startedAt)
		t.actualTime = &actualTime
	}
	
	return nil
}

// Fail 任务失败
func (t *TranscodeTaskEntity) Fail(errorMessage string) error {
	if !t.status.CanTransitionTo(vo.TaskStatusFailed) {
		return NewDomainError("cannot fail task in current status: " + t.status.String())
	}
	
	now := time.Now()
	t.status = vo.TaskStatusFailed
	t.errorMessage = errorMessage
	t.updatedAt = now
	
	// 如果有开始时间，计算实际耗时
	if t.startedAt != nil {
		actualTime := now.Sub(*t.startedAt)
		t.actualTime = &actualTime
	}
	
	return nil
}

// Retry 重试任务
func (t *TranscodeTaskEntity) Retry() error {
	if t.status != vo.TaskStatusFailed {
		return NewDomainError("can only retry failed tasks")
	}
	
	if t.retryCount >= t.maxRetryCount {
		return NewDomainError("maximum retry count exceeded")
	}
	
	t.retryCount++
	t.status = vo.TaskStatusRetrying
	t.workerID = "" // 清空Worker ID，重新分配
	t.errorMessage = ""
	t.progress = 0.0
	t.startedAt = nil
	t.updatedAt = time.Now()
	
	return nil
}

// Cancel 取消任务
func (t *TranscodeTaskEntity) Cancel() error {
	if t.status.IsFinalStatus() {
		return NewDomainError("cannot cancel task in final status: " + t.status.String())
	}
	
	t.status = vo.TaskStatusCancelled
	t.updatedAt = time.Now()
	return nil
}

// CanRetry 检查是否可以重试
func (t *TranscodeTaskEntity) CanRetry() bool {
	return t.status == vo.TaskStatusFailed && t.retryCount < t.maxRetryCount
}

// IsExpired 检查任务是否过期（超过24小时未完成）
func (t *TranscodeTaskEntity) IsExpired() bool {
	if t.status.IsFinalStatus() {
		return false
	}
	return time.Since(t.createdAt) > 24*time.Hour
}

// SetMetadata 设置元数据
func (t *TranscodeTaskEntity) SetMetadata(key string, value interface{}) {
	t.metadata[key] = value
	t.updatedAt = time.Now()
}

// GetMetadata 获取元数据
func (t *TranscodeTaskEntity) GetMetadata(key string) (interface{}, bool) {
	value, exists := t.metadata[key]
	return value, exists
}

// DomainError 领域错误
type DomainError struct {
	message string
}

func NewDomainError(message string) *DomainError {
	return &DomainError{message: message}
}

func (e *DomainError) Error() string {
	return e.message
}