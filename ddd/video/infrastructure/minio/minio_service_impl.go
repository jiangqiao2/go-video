package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"go-video/ddd/internal/resource"
	"go-video/ddd/video/domain/gateway"
	"go-video/pkg/assert"
)

var (
	minioServiceOnce      sync.Once
	singletonMinioService gateway.MinioService
)

type MinioServiceImpl struct {
	minioClient *resource.MinioResource
}

func DefaultMinioService() gateway.MinioService {
	assert.NotCircular()
	minioServiceOnce.Do(func() {
		singletonMinioService = &MinioServiceImpl{
			minioClient: resource.DefaultMinioResource(),
		}
	})
	return singletonMinioService
}

// UploadVideo 上传视频文件
func (m *MinioServiceImpl) UploadVideo(ctx context.Context, userUUID string, file *multipart.FileHeader) (string, error) {
	// 确保MinIO资源已初始化
	m.minioClient.MustOpen()
	
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	
	// 生成对象名称
	objectName := m.generateObjectName(userUUID, file.Filename)
	
	// 获取内容类型
	contentType := m.getContentType(file.Filename)
	
	// 上传文件到MinIO
	client := m.minioClient.GetClient()
	bucketName := m.minioClient.GetBucketName()
	_, err = client.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	
	return objectName, nil
}

// DownloadVideo 下载视频文件
func (m *MinioServiceImpl) DownloadVideo(ctx context.Context, objectName string) ([]byte, error) {
	// 确保MinIO资源已初始化
	m.minioClient.MustOpen()
	
	// 从MinIO下载文件
	client := m.minioClient.GetClient()
	bucketName := m.minioClient.GetBucketName()
	object, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	
	// 读取文件内容
	data := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := object.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}
	
	return data, nil
}

// DeleteVideo 删除视频文件
func (m *MinioServiceImpl) DeleteVideo(ctx context.Context, objectName string) error {
	// 确保MinIO资源已初始化
	m.minioClient.MustOpen()
	
	// 从MinIO删除文件
	client := m.minioClient.GetClient()
	bucketName := m.minioClient.GetBucketName()
	return client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// GetVideoURL 获取视频访问URL
func (m *MinioServiceImpl) GetVideoURL(ctx context.Context, objectName string) (string, error) {
	// 确保MinIO资源已初始化
	m.minioClient.MustOpen()
	
	// 生成预签名URL（1小时有效期）
	client := m.minioClient.GetClient()
	bucketName := m.minioClient.GetBucketName()
	url, err := client.PresignedGetObject(ctx, bucketName, objectName, 3600*time.Second, nil)
	if err != nil {
		return "", err
	}
	
	return url.String(), nil
}

// generateObjectName 生成对象名称
func (m *MinioServiceImpl) generateObjectName(userUUID, filename string) string {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("videos/%s/%d_%s%s", userUUID, timestamp, baseName, ext)
}

// getContentType 获取内容类型
func (m *MinioServiceImpl) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/avi"
	case ".mov":
		return "video/mov"
	case ".wmv":
		return "video/wmv"
	case ".flv":
		return "video/flv"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/mkv"
	default:
		return "application/octet-stream"
	}
}
