package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Video 视频实体
type Video struct {
	id          uint64
	uuid        string
	userUuid    string
	title       string
	description string
	filename    string
	fileSize    int64
	duration    int // 秒
	format      string
	storagePath string
	status      VideoStatus
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// VideoStatus 视频状态
type VideoStatus int

const (
	VideoStatusUploading  VideoStatus = 1 // 上传中
	VideoStatusProcessing VideoStatus = 2 // 处理中
	VideoStatusReady      VideoStatus = 3 // 就绪
	VideoStatusFailed     VideoStatus = 4 // 失败
	VideoStatusDeleted    VideoStatus = 5 // 已删除
)

// NewVideo 创建新视频
func NewVideo(userUuid string, title, description, filename string, fileSize int64, format string) (*Video, error) {
	video := &Video{
		uuid:        uuid.New().String(),
		userUuid:    userUuid,
		title:       title,
		description: description,
		filename:    filename,
		fileSize:    fileSize,
		format:      format,
		status:      VideoStatusUploading,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	// 验证标题
	if err := video.ValidateTitle(); err != nil {
		return nil, err
	}

	// 验证文件名
	if err := video.ValidateFilename(); err != nil {
		return nil, err
	}

	return video, nil
}

// ID 获取视频内部ID
func (v *Video) ID() uint64 {
	return v.id
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

// Duration 获取时长
func (v *Video) Duration() int {
	return v.duration
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
func (v *Video) Status() VideoStatus {
	return v.status
}

// CreatedAt 获取创建时间
func (v *Video) CreatedAt() time.Time {
	return v.createdAt
}

// UpdatedAt 获取更新时间
func (v *Video) UpdatedAt() time.Time {
	return v.updatedAt
}

// DeletedAt 获取删除时间
func (v *Video) DeletedAt() *time.Time {
	return v.deletedAt
}

// SetID 设置内部ID
func (v *Video) SetID(id uint64) {
	v.id = id
}

// SetUUID 设置UUID
func (v *Video) SetUUID(uuid string) {
	v.uuid = uuid
}

// SetTitle 设置标题
func (v *Video) SetTitle(title string) error {
	v.title = title
	return v.ValidateTitle()
}

// SetDescription 设置描述
func (v *Video) SetDescription(description string) {
	v.description = description
	v.updatedAt = time.Now()
}

// SetDuration 设置时长
func (v *Video) SetDuration(duration int) {
	v.duration = duration
	v.updatedAt = time.Now()
}

// SetStoragePath 设置存储路径
func (v *Video) SetStoragePath(path string) {
	v.storagePath = path
	v.updatedAt = time.Now()
}

// SetStatus 设置状态
func (v *Video) SetStatus(status VideoStatus) {
	v.status = status
	v.updatedAt = time.Now()
}

// SetTimestamps 设置时间戳
func (v *Video) SetTimestamps(createdAt, updatedAt time.Time, deletedAt *time.Time) {
	v.createdAt = createdAt
	v.updatedAt = updatedAt
	v.deletedAt = deletedAt
}

// MarkAsReady 标记为就绪
func (v *Video) MarkAsReady() {
	v.SetStatus(VideoStatusReady)
}

// MarkAsProcessing 标记为处理中
func (v *Video) MarkAsProcessing() {
	v.SetStatus(VideoStatusProcessing)
}

// MarkAsFailed 标记为失败
func (v *Video) MarkAsFailed() {
	v.SetStatus(VideoStatusFailed)
}

// Delete 软删除
func (v *Video) Delete() {
	now := time.Now()
	v.deletedAt = &now
	v.status = VideoStatusDeleted
	v.updatedAt = now
}

// IsDeleted 是否已删除
func (v *Video) IsDeleted() bool {
	return v.deletedAt != nil || v.status == VideoStatusDeleted
}

// IsReady 是否就绪
func (v *Video) IsReady() bool {
	return v.status == VideoStatusReady
}

// ValidateTitle 验证标题
func (v *Video) ValidateTitle() error {
	if len(v.title) == 0 {
		return errors.New("标题不能为空")
	}
	if len(v.title) > 255 {
		return errors.New("标题长度不能超过255个字符")
	}
	return nil
}

// ValidateFilename 验证文件名
func (v *Video) ValidateFilename() error {
	if len(v.filename) == 0 {
		return errors.New("文件名不能为空")
	}
	if len(v.filename) > 255 {
		return errors.New("文件名长度不能超过255个字符")
	}
	return nil
}
