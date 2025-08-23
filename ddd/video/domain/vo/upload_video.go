package vo

import "mime/multipart"

type VideoUploadVO struct {
	userUUID    string
	videoUUID   string
	taskUUID    string
	storagePath string
	file        *multipart.FileHeader
}

func NewUploadVideo(userUUID string, videoUUID string, taskUUID string, storagePath string, file *multipart.FileHeader) *VideoUploadVO {
	return &VideoUploadVO{
		userUUID:    userUUID,
		videoUUID:   videoUUID,
		taskUUID:    taskUUID,
		file:        file,
		storagePath: storagePath,
	}
}
func (v *VideoUploadVO) UserUUID() string {
	return v.userUUID
}

func (v *VideoUploadVO) VideoUUID() string {
	return v.videoUUID
}

func (v *VideoUploadVO) TaskUUID() string {
	return v.taskUUID
}

func (v *VideoUploadVO) StoragePath() string {
	return v.storagePath
}

func (v *VideoUploadVO) File() *multipart.FileHeader {
	return v.file
}

type VideoUploadTaskStatus struct {
	value string
}

var (

	// 初始化
	VideoUploadTaskStatusInit = VideoUploadTaskStatus{
		"init",
	}
	// 准备中
	VideoUploadTaskStatusInProgress = VideoUploadTaskStatus{
		"in_progress",
	}

	VideoUploadTaskStatusCompleted = VideoUploadTaskStatus{
		"completed",
	}
	VideoUploadTaskStatusFailed = VideoUploadTaskStatus{
		"failed",
	}
)

var VideoUploadTaskStatuses = []VideoUploadTaskStatus{

	VideoUploadTaskStatusInit,
	VideoUploadTaskStatusInProgress,
	VideoUploadTaskStatusCompleted,
	VideoUploadTaskStatusFailed,
}

func NewVideoUploadTaskStatus(value string) VideoUploadTaskStatus {
	for _, status := range VideoUploadTaskStatuses {
		if status.value == value {
			return status
		}
	}
	return VideoUploadTaskStatusInit
}

// String 返回状态的字符串值（用于PO转换）
func (s VideoUploadTaskStatus) String() string {
	return s.value
}

// Value 返回状态的字符串值（别名方法）
func (s VideoUploadTaskStatus) Value() string {
	return s.value
}

// Equals 比较两个状态是否相等
func (s VideoUploadTaskStatus) Equals(other VideoUploadTaskStatus) bool {
	return s.value == other.value
}

// IsInit 检查是否为初始化状态
func (s VideoUploadTaskStatus) IsInit() bool {
	return s.value == VideoUploadTaskStatusInit.value
}

// IsInProgress 检查是否为等待状态
func (s VideoUploadTaskStatus) IsInProgress() bool {
	return s.value == VideoUploadTaskStatusInProgress.value
}

// IsCompleted 检查是否为完成状态
func (s VideoUploadTaskStatus) IsCompleted() bool {
	return s.value == VideoUploadTaskStatusCompleted.value
}

// IsFailed 检查是否为失败状态
func (s VideoUploadTaskStatus) IsFailed() bool {
	return s.value == VideoUploadTaskStatusFailed.value
}

type VideoStatus struct {
	value string
}

var (
	VideoStatusInit = VideoStatus{
		"init",
	}
	VideoStatusInProgress = VideoStatus{
		"in_progress",
	}
	VideoStatusCompleted = VideoStatus{
		"completed",
	}
	VideoStatusFailed = VideoStatus{
		"failed",
	}
)
var VideoStatuses = []VideoStatus{
	VideoStatusInit,
	VideoStatusInProgress,
	VideoStatusCompleted,
	VideoStatusFailed,
}

func NewVideoStatus(value string) VideoStatus {
	for _, status := range VideoStatuses {
		if status.value == value {
			return status
		}
	}
	return VideoStatusInit
}

func (s VideoStatus) Value() string {
	return s.value
}
