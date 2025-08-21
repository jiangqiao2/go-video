package minio

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-video/pkg/logger"
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
		logger.Info("MinioServiceImpl file open error: " + err.Error())
		return "", err
	}
	defer src.Close()

	// 生成对象名称
	fileUuid := uuid.NewString()
	objectName := m.generateObjectName(userUUID, fileUuid)

	// 获取内容类型
	contentType := m.getContentType()

	// 上传文件到MinIO
	client := m.minioClient.GetClient()
	bucketName := m.minioClient.GetBucketName()
	_, err = client.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Info("MinioServiceImpl bucketName upload err: " + err.Error())
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

// generateObjectName 生成对象名称（按年-月-日）
func (m *MinioServiceImpl) generateObjectName(userUUID, filename string) string {
	// 格式化为 yyyy-MM-dd
	dateStr := time.Now().Format("2006-01-02")

	// 保留后缀
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	return fmt.Sprintf("videos/%s/%s/%s%s", userUUID, dateStr, baseName, ext)
}

func (m *MinioServiceImpl) getContentType() string {
	return "application/octet-stream"
}
