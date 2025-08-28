package gateway

import (
	"context"
	"go-video/ddd/video/domain/vo"
	"mime/multipart"
)

// MinioService MinIO服务接口
type MinioService interface {
	// UploadVideo 上传视频文件
	UploadVideo(ctx context.Context, userUUID string, file *multipart.FileHeader) (string, error)

	// DownloadVideo 下载视频文件
	DownloadVideo(ctx context.Context, objectName string) ([]byte, error)

	// DeleteVideo 删除视频文件
	DeleteVideo(ctx context.Context, objectName string) error

	// GetVideoURL 获取视频访问URL
	GetVideoURL(ctx context.Context, objectName string) (string, error)

	// GenerateObjectName 生成文件路径
	GenerateObjectName(userUUID, filename string) string

	SyncUploadVideo(ctx context.Context, videoUploadVo *vo.VideoUploadVO)
}
