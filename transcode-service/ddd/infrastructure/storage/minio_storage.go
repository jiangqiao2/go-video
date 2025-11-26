package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"

	"transcode-service/ddd/domain/gateway"
	"transcode-service/internal/resource"
	"transcode-service/pkg/logger"
)

// MinioStorage MinIO存储实现
type MinioStorage struct {
	minioResource *resource.MinioResource
}

// NewMinioStorage 创建MinIO存储实例
func NewMinioStorage(minioResource *resource.MinioResource) gateway.StorageGateway {
	return &MinioStorage{
		minioResource: minioResource,
	}
}

// UploadTranscodedFile 上传转码后的文件，返回可访问的对象路径
func (s *MinioStorage) UploadTranscodedFile(ctx context.Context, localPath, objectKey, contentType string) (string, error) {
	client := s.minioResource.GetClient()
	bucketName := s.minioResource.GetBucketName()

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		logger.Error("Failed to open local file", map[string]interface{}{
			"local_path": localPath,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("open local file failed: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error("Failed to get file info", map[string]interface{}{
			"local_path": localPath,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("get file info failed: %w", err)
	}

	if contentType == "" {
		contentType = getContentTypeFromExtension(objectKey)
	}

	// 上传文件到MinIO
	_, err = client.PutObject(ctx, bucketName, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Error("Failed to upload transcoded file to MinIO", map[string]interface{}{
			"local_path": localPath,
			"object_key": objectKey,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("upload transcoded file to minio failed: %w", err)
	}

	logger.Info("Transcoded file uploaded successfully", map[string]interface{}{
		"local_path": localPath,
		"object_key": objectKey,
		"size":       fileInfo.Size(),
	})

	// 返回对象路径
	return objectKey, nil
}

// UploadObjects 批量上传对象
func (s *MinioStorage) UploadObjects(ctx context.Context, objects []gateway.UploadObject) error {
	if len(objects) == 0 {
		return nil
	}

	client := s.minioResource.GetClient()
	bucketName := s.minioResource.GetBucketName()

	for _, obj := range objects {
		file, err := os.Open(obj.LocalPath)
		if err != nil {
			logger.Error("Failed to open local file for batch upload", map[string]interface{}{
				"local_path": obj.LocalPath,
				"error":      err.Error(),
			})
			return fmt.Errorf("open local file failed: %w", err)
		}

		fileInfo, err := file.Stat()
		if err != nil {
			file.Close()
			logger.Error("Failed to stat local file for batch upload", map[string]interface{}{
				"local_path": obj.LocalPath,
				"error":      err.Error(),
			})
			return fmt.Errorf("get file info failed: %w", err)
		}

		contentType := obj.ContentType
		if contentType == "" {
			contentType = getContentTypeFromExtension(obj.ObjectKey)
		}

		_, err = client.PutObject(ctx, bucketName, obj.ObjectKey, file, fileInfo.Size(), minio.PutObjectOptions{
			ContentType: contentType,
		})
		file.Close()
		if err != nil {
			logger.Error("Failed to upload object during batch upload", map[string]interface{}{
				"local_path": obj.LocalPath,
				"object_key": obj.ObjectKey,
				"error":      err.Error(),
			})
			return fmt.Errorf("upload object to minio failed: %w", err)
		}

		logger.Info("Uploaded object", map[string]interface{}{
			"object_key": obj.ObjectKey,
			"local_path": obj.LocalPath,
			"size":       fileInfo.Size(),
		})
	}

	return nil
}

// DownloadFile 从MinIO下载文件到本地路径
func (s *MinioStorage) DownloadFile(ctx context.Context, objectKey, localPath string) error {
	client := s.minioResource.GetClient()
	bucketName := s.minioResource.GetBucketName()

	// 确保本地目录存在
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		logger.Error("Failed to create local directory", map[string]interface{}{
			"local_path": localPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("create local directory failed: %w", err)
	}

	// 从MinIO获取对象
	object, err := client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("Failed to get object from MinIO", map[string]interface{}{
			"object_key": objectKey,
			"error":      err.Error(),
		})
		return fmt.Errorf("get object from minio failed: %w", err)
	}
	defer object.Close()

	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		logger.Error("Failed to create local file", map[string]interface{}{
			"local_path": localPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("create local file failed: %w", err)
	}
	defer localFile.Close()

	// 复制数据到本地文件
	_, err = localFile.ReadFrom(object)
	if err != nil {
		logger.Error("Failed to download file from MinIO", map[string]interface{}{
			"object_key": objectKey,
			"local_path": localPath,
			"error":      err.Error(),
		})
		return fmt.Errorf("download file from minio failed: %w", err)
	}

	logger.Info("File downloaded successfully", map[string]interface{}{
		"object_key": objectKey,
		"local_path": localPath,
	})

	return nil
}

// getContentTypeFromExtension 根据文件扩展名获取内容类型
func getContentTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}
