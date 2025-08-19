package cqe

import (
	"go-video/pkg/errno"
	"mime/multipart"

	"go-video/ddd/video/domain/entity"
)

// UploadVideoCommand 上传视频命令
// 包含上传视频所需的所有参数
// 实现Validate方法进行参数校验
type UploadVideoCommand struct {
	UserUUID    string                `json:"user_uuid" validate:"required,uuid4"`     // 用户UUID
	Title       string                `json:"title" validate:"required,min=1,max=100"` // 视频标题
	Description string                `json:"description" validate:"max=500"`          // 视频描述
	File        *multipart.FileHeader `json:"-" validate:"required"`                   // 视频文件
	Thumbnail   *multipart.FileHeader `json:"-"`                                       // 缩略图文件
	Format      string                `json:"format"`                                  // 视频格式
	FileSize    int64                 `json:"file_size"`                               // 文件大小(字节)
}

// ToEntity 转换为视频实体
func (c *UploadVideoCommand) ToEntity() (*entity.Video, error) {
	return entity.NewVideo(
		c.UserUUID,
		c.Title,
		c.Description,
		c.File.Filename,
		c.FileSize,
		c.Format,
	)
}

// Validate 实现Command接口的校验方法
func (c *UploadVideoCommand) Validate() error {
	if len(c.UserUUID) <= 0 {
		return errno.ErrUserNotFound
	}
	return nil
}
