package entity

import (
	"github.com/google/uuid"
	"time"
	"transcode-service/ddd/domain/vo"
)

// TranscodeTaskEntity 转码任务实体
type TranscodeTaskEntity struct {
	id            uint64 // 数据库主键ID
	taskUUID      string
	userUUID      string
	videoUUID     string
	videoPushUUID string
	originalPath  string
	outputPath    string
	status        vo.TaskStatus
	progress      int
	errorMessage  string
	params        vo.TranscodeParams
	createdAt     time.Time
	updatedAt     time.Time
}

// NewTranscodeTaskEntity 创建转码任务实体
func NewTranscodeTaskEntity(
	taskUUID, userUUID, videoUUID, originalPath, outputPath string,
) *TranscodeTaskEntity {
	now := time.Now()
	return &TranscodeTaskEntity{
		taskUUID:     taskUUID,
		userUUID:     userUUID,
		videoUUID:    videoUUID,
		originalPath: originalPath,
		outputPath:   outputPath,
		status:       vo.TaskStatusPending,
		progress:     0,
		errorMessage: "",
		createdAt:    now,
		updatedAt:    now,
	}
}

// DefaultTranscodeTaskEntity 创建默认转码任务实体（自动生成UUID）
func DefaultTranscodeTaskEntity(userUUID, videoUUID, videoPushUUID, originalPath string, params vo.TranscodeParams) *TranscodeTaskEntity {
	taskUUID := uuid.New().String()
	now := time.Now()

	// 生成输出路径
	outputPath := generateOutputPath(userUUID, videoUUID, params)

	return &TranscodeTaskEntity{
		taskUUID:      taskUUID,
		userUUID:      userUUID,
		videoUUID:     videoUUID,
		videoPushUUID: videoPushUUID,
		originalPath:  originalPath,
		outputPath:    outputPath,
		status:        vo.TaskStatusPending,
		progress:      0,
		errorMessage:  "",
		params:        params,
		createdAt:     now,
		updatedAt:     now,
	}
}

// NewTranscodeTaskEntityWithDetails 创建带详细信息的转码任务实体
func NewTranscodeTaskEntityWithDetails(
	id uint64,
	taskUUID, userUUID, videoUUID, originalPath, outputPath string,
	status vo.TaskStatus, progress int, errorMessage string,
	params vo.TranscodeParams, createdAt, updatedAt time.Time,
) *TranscodeTaskEntity {
	return &TranscodeTaskEntity{
		id:            id,
		taskUUID:      taskUUID,
		userUUID:      userUUID,
		videoUUID:     videoUUID,
		videoPushUUID: "",
		originalPath:  originalPath,
		outputPath:    outputPath,
		status:        status,
		progress:      progress,
		errorMessage:  errorMessage,
		params:        params,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// generateOutputPath 生成输出路径
func generateOutputPath(userUUID, videoUUID string, params vo.TranscodeParams) string {
	return "/transcoded/" + userUUID + "/" + videoUUID + "_" + params.Resolution + "_" + params.Bitrate + ".mp4"
}

// ID 获取数据库主键ID
func (t *TranscodeTaskEntity) ID() uint64 {
	return t.id
}

// SetID 设置数据库主键ID
func (t *TranscodeTaskEntity) SetID(id uint64) {
	t.id = id
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

// VideoPushUUID 获取上传服务侧的UUID
func (t *TranscodeTaskEntity) VideoPushUUID() string {
	return t.videoPushUUID
}

// SetVideoPushUUID 设置上传服务侧的UUID
func (t *TranscodeTaskEntity) SetVideoPushUUID(uuid string) {
	t.videoPushUUID = uuid
}

// OriginalPath 获取原始路径
func (t *TranscodeTaskEntity) OriginalPath() string {
	return t.originalPath
}

// InputPath 获取原始路径（别名）
func (t *TranscodeTaskEntity) InputPath() string {
	return t.originalPath
}

// OutputPath 获取输出路径
func (t *TranscodeTaskEntity) OutputPath() string {
	return t.outputPath
}

// SetOutputPath 设置输出路径
func (t *TranscodeTaskEntity) SetOutputPath(path string) {
	t.outputPath = path
	t.updatedAt = time.Now()
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

// GetParams 获取转码参数
func (t *TranscodeTaskEntity) GetParams() vo.TranscodeParams {
	return t.params
}

// CreatedAt 获取创建时间
func (t *TranscodeTaskEntity) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt 获取更新时间
func (t *TranscodeTaskEntity) UpdatedAt() time.Time {
	return t.updatedAt
}

// SetStatus 设置状态
func (t *TranscodeTaskEntity) SetStatus(status vo.TaskStatus) {
	t.status = status
	t.updatedAt = time.Now()
}

// SetProgress 设置进度
func (t *TranscodeTaskEntity) SetProgress(progress int) {
	t.progress = progress
	t.updatedAt = time.Now()
}

// SetErrorMessage 设置错误信息
func (t *TranscodeTaskEntity) SetErrorMessage(message string) {
	t.errorMessage = message
	t.updatedAt = time.Now()
}

// SetTimestamps 设置创建和更新时间（用于持久化还原）
func (t *TranscodeTaskEntity) SetTimestamps(createdAt, updatedAt time.Time) {
	t.createdAt = createdAt
	t.updatedAt = updatedAt
}

// SetParams 设置转码参数
func (t *TranscodeTaskEntity) SetParams(params vo.TranscodeParams) {
	t.params = params
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
