package entity

import (
	"go-video/ddd/video/domain/vo"
	"time"

	"github.com/google/uuid"
)

// Video 视频实体
type Video struct {
	uuid        string
	userUuid    string
	title       string
	description string
	filename    string
	fileSize    int64
	format      string
	storagePath string
	status      vo.VideoStatus
}

// VideoStatus 视频状态

func DefaultVideo(userUuid, title, description, filename string, fileSize int64, format, storagePath string, status vo.VideoStatus) *Video {
	return &Video{
		uuid:        uuid.New().String(),
		userUuid:    userUuid,
		title:       title,
		description: description,
		filename:    filename,
		fileSize:    fileSize,
		format:      format,
		storagePath: storagePath,
		status:      status,
	}
}

// NewVideo 创建新视频
func NewVideo(uuid, userUuid string, title, description, filename string, fileSize int64, format string, status vo.VideoStatus) *Video {
	return &Video{
		uuid:        uuid,
		userUuid:    userUuid,
		title:       title,
		description: description,
		filename:    filename,
		fileSize:    fileSize,
		format:      format,
		status:      status,
	}
}

// UUID 获取视频UUID
func (v *Video) UUID() string {
	return v.uuid
}

// UserID 获取用户ID
func (v *Video) UserUuid() string {
	return v.userUuid
}

// Title 获取标题
func (v *Video) Title() string {
	return v.title
}

// Description 获取描述
func (v *Video) Description() string {
	return v.description
}

// Filename 获取文件名
func (v *Video) Filename() string {
	return v.filename
}

// FileSize 获取文件大小
func (v *Video) FileSize() int64 {
	return v.fileSize
}

// Format 获取格式
func (v *Video) Format() string {
	return v.format
}

// StoragePath 获取存储路径
func (v *Video) StoragePath() string {
	return v.storagePath
}

// Status 获取状态
func (v *Video) Status() vo.VideoStatus {
	return v.status
}

// SetUUID 设置UUID
func (v *Video) SetUUID(uuid string) {
	v.uuid = uuid
}

// SetTitle 设置标题
func (v *Video) SetTitle(title string) {
	v.title = title
}

// SetDescription 设置描述
func (v *Video) SetDescription(description string) {
	v.description = description
}

// SetStoragePath 设置存储路径
func (v *Video) SetStoragePath(path string) {
	v.storagePath = path
}

// SetStatus 设置状态
func (v *Video) SetStatus(status vo.VideoStatus) {
	v.status = status
}

type VideoUploadTaskEntity struct {
	uuid        string
	userUuid    string
	videoUuid   string
	status      vo.VideoUploadTaskStatus
	errorMsg    string
	completedAt *time.Time
	storagePath string
}

func DefaultVideoUploadTaskEntity(userUuid, videoUuid string,
	status vo.VideoUploadTaskStatus,
	errorMsg string,
	completdAt *time.Time,
	storagePath string) *VideoUploadTaskEntity {
	return &VideoUploadTaskEntity{
		uuid:        uuid.New().String(),
		videoUuid:   videoUuid,
		userUuid:    userUuid,
		status:      status,
		errorMsg:    errorMsg,
		completedAt: completdAt,
		storagePath: storagePath,
	}
}

// NewVideoUploadTask 创建新的视频上传任务
func NewVideoUploadTask(uuid string,
	userUuid string,
	status vo.VideoUploadTaskStatus,
	errorMsg string,
	completedAt *time.Time,
	objectName string) *VideoUploadTaskEntity {
	return &VideoUploadTaskEntity{
		uuid:        uuid,
		userUuid:    userUuid,
		status:      status,
		errorMsg:    errorMsg,
		completedAt: completedAt,
		storagePath: objectName,
	}
}

// UUID 获取任务UUID
func (v *VideoUploadTaskEntity) UUID() string {
	return v.uuid
}

// UserUuid 获取用户UUID
func (v *VideoUploadTaskEntity) UserUuid() string {
	return v.userUuid
}

// Status 获取任务状态
func (v *VideoUploadTaskEntity) Status() vo.VideoUploadTaskStatus {
	return v.status
}

// ErrorMsg 获取错误信息
func (v *VideoUploadTaskEntity) ErrorMsg() string {
	return v.errorMsg
}

// CompletedAt 获取完成时间
func (v *VideoUploadTaskEntity) CompletedAt() *time.Time {
	return v.completedAt
}

// ObjectName 获取对象名称
func (v *VideoUploadTaskEntity) ObjectName() string {
	return v.storagePath
}

// SetUUID 设置任务UUID（仅用于从数据库加载）
func (v *VideoUploadTaskEntity) SetUUID(uuid string) {
	v.uuid = uuid
}

// SetObjectName 设置对象名称
func (v *VideoUploadTaskEntity) SetObjectName(objectName string) {
	v.storagePath = objectName
}

// SetCompletedAt 设置完成时间（仅用于从数据库加载）
func (v *VideoUploadTaskEntity) SetCompletedAt(completedAt *time.Time) {
	v.completedAt = completedAt
}

func (v *VideoUploadTaskEntity) SetStatus(status vo.VideoUploadTaskStatus) {
	v.status = status
}

// IsCompleted 检查是否已完成
func (v *VideoUploadTaskEntity) IsCompleted() bool {
	return v.status == vo.VideoUploadTaskStatusCompleted
}

// IsFailed 检查是否失败
func (v *VideoUploadTaskEntity) IsFailed() bool {
	return v.status == vo.VideoUploadTaskStatusFailed
}

// IsPending 检查是否进行中
func (v *VideoUploadTaskEntity) IsInProgress() bool {
	return v.status == vo.VideoUploadTaskStatusInProgress
}

type VideoEntity struct {
	videoUploadTask *VideoUploadTaskEntity
	video           *Video
}

func NewVideoEntity(video *Video, videoUploadTask *VideoUploadTaskEntity) *VideoEntity {
	return &VideoEntity{
		videoUploadTask: videoUploadTask,
		video:           video,
	}
}
func (v *VideoEntity) Video() *Video {
	return v.video
}
func (v *VideoEntity) VideoUploadTask() *VideoUploadTaskEntity {
	return v.videoUploadTask
}
