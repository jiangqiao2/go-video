package minio

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"

    "github.com/minio/minio-go/v7"
    log "github.com/sirupsen/logrus"
    "github.com/google/uuid"

	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/vo"
	"upload-service/internal/resource"
	"upload-service/pkg/assert"
	"upload-service/pkg/errno"
    "upload-service/pkg/logger"
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

func (m *MinioServiceImpl) GenerateStoragePath(ctx context.Context, genStoPathVo *vo.GenerateStoragePathVO) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	fileName := genStoPathVo.FileName()
	ext := ""
	if dotIndex := strings.LastIndex(fileName, "."); dotIndex != -1 {
		ext = fileName[dotIndex+1:]
	}

	storagePath := fmt.Sprintf("uploads/%s/%s/%s/%s/%s.%s",
		genStoPathVo.UserUUID(),
		year,
		month,
		day,
		genStoPathVo.UploadVideoUUID(),
		ext,
	)
	return storagePath
}

func (m *MinioServiceImpl) GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	storagePath := fmt.Sprintf("chunks/%s/%s/%s/%s/chunk_",
		year,
		month,
		day,
		uploadVideoUUID,
	)
	return storagePath
}

func (m *MinioServiceImpl) UploadChunk(ctx context.Context, minIoChunkVo *vo.MinIoUploadChunkVo) error {
	exists, err := m.minioClient.GetClient().BucketExists(ctx, minIoChunkVo.BucketName())
	if err != nil {
		return err
	}
	if !exists {
		return errno.NewSimpleBizError(errno.ErrMinIoBuckNameNotExist, nil, "")
	}

	_, err = m.minioClient.GetClient().PutObject(ctx, minIoChunkVo.BucketName(), minIoChunkVo.StoragePath(), minIoChunkVo.Reader(), minIoChunkVo.FileSize(),
		minio.PutObjectOptions{ContentType: minIoChunkVo.ContentType()})
	if err != nil {
		log.Errorf("minio put object error: %v", err)
		return errno.NewSimpleBizError(errno.ErrMinIoBuckNameNotExist, nil, "")
	}
	return nil
}

func (m *MinioServiceImpl) MergeChunk(ctx context.Context, mergeChunkVo *vo.MergeChunkVo) error {
	var srcs []minio.CopySrcOptions
	for i := int64(0); i < mergeChunkVo.TotalChunks(); i++ {
		chunkObject := fmt.Sprintf("%s%d", mergeChunkVo.ChunkStoragePath(), i)
		src := minio.CopySrcOptions{
			Bucket: "uploads",
			Object: chunkObject,
		}
		srcs = append(srcs, src)
	}

	dst := minio.CopyDestOptions{
		Bucket: "uploads",
		Object: mergeChunkVo.StoragePath(),
	}

	_, err := m.minioClient.GetClient().ComposeObject(ctx, dst, srcs...)
	if err != nil {
		logger.Errorf("MergeChunk merge %v, err:%v", mergeChunkVo, err)
		return fmt.Errorf("compose object error: %w", err)
	}
	return nil
}

func (m *MinioServiceImpl) DeleteChunks(ctx context.Context, chunkStoragePath string, totalChunks int64) error {
	if chunkStoragePath == "" || totalChunks <= 0 {
		return nil
	}

	bucket := m.minioClient.GetBucketName()
	var firstErr error

	for i := int64(0); i < totalChunks; i++ {
		objectName := fmt.Sprintf("%s%d", chunkStoragePath, i)
		err := m.minioClient.GetClient().RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			logger.Errorf("DeleteChunks remove object failed", map[string]interface{}{
				"bucket": bucket,
				"object": objectName,
				"error":  err,
			})
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (m *MinioServiceImpl) GenerateImagePath(ctx context.Context, ivo *vo.GenerateImagePathVO) string {
    now := time.Now()
    year := now.Format("2006")
    month := now.Format("01")
    day := now.Format("02")
    fileName := ivo.FileName()
    ext := ""
    if dot := strings.LastIndex(fileName, "."); dot != -1 {
        ext = fileName[dot+1:]
    }
    category := ivo.Category()
    if category == "" {
        category = "images"
    }
    name := uuid.NewString()
    return fmt.Sprintf("%s/%s/%s/%s/%s.%s", category, ivo.UserUUID(), year, month, day, name, ext)
}

func (m *MinioServiceImpl) PresignPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
    if expires <= 0 {
        expires = 15 * time.Minute
    }
    u, err := m.minioClient.GetClient().PresignedPutObject(ctx, bucket, key, expires)
    if err != nil {
        return "", err
    }
    return u.String(), nil
}

func (m *MinioServiceImpl) PresignGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
    if expires <= 0 {
        expires = 24 * time.Hour
    }
    u, err := m.minioClient.GetClient().PresignedGetObject(ctx, bucket, key, expires, nil)
    if err != nil {
        return "", err
    }
    return u.String(), nil
}
