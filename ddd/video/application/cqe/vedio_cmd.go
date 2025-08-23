package cqe

import (
	"go-video/pkg/errno"
	"mime/multipart"
)

// UploadVideoCommand 上传视频命令
// 包含上传视频所需的所有参数
// 实现Validate方法进行参数校验
type UploadVideoCommand struct {
	UserUUID    string                `json:"user_uuid"`   // 用户UUID
	Title       string                `json:"title"`       // 视频标题
	Description string                `json:"description"` // 视频描述
	File        *multipart.FileHeader `json:"-"`           // 视频文件
	Format      string                `json:"format"`      // 视频格式
	FileSize    int64                 `json:"file_size"`   // 文件大小(字节)
}

// Validate 实现Command接口的校验方法
func (c *UploadVideoCommand) Validate() error {
	// 验证用户UUID
	if len(c.UserUUID) <= 0 {
		return errno.ErrMissingParam
	}

	// 验证标题
	if len(c.Title) == 0 {
		return errno.ErrMissingParam
	}
	if len(c.Title) > 100 {
		return errno.ErrParamTooLong
	}

	// 验证描述长度
	if len(c.Description) > 500 {
		return errno.ErrParamTooLong
	}

	// 验证文件是否存在
	if c.File == nil {
		return errno.ErrMissingParam
	}

	// 验证文件大小（限制为100MB）
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if c.File.Size > maxFileSize {
		return errno.ErrVideoTooLarge
	}

	// 验证视频文件格式
	if !c.isValidVideoFormat() {
		return errno.ErrVideoFormatInvalid
	}

	return nil
}

// isValidVideoFormat 检查是否为有效的视频格式
func (c *UploadVideoCommand) isValidVideoFormat() bool {
	validFormats := []string{
		"video/mp4",
		"video/avi",
		"video/mov",
		"video/wmv",
		"video/flv",
		"video/webm",
		"video/mkv",
	}

	// 检查MIME类型
	contentType := c.File.Header.Get("Content-Type")
	for _, format := range validFormats {
		if contentType == format {
			return true
		}
	}

	// 检查文件扩展名
	filename := c.File.Filename
	if len(filename) == 0 {
		return false
	}

	validExtensions := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
	for _, ext := range validExtensions {
		if len(filename) >= len(ext) && filename[len(filename)-len(ext):] == ext {
			return true
		}
	}

	return false
}
