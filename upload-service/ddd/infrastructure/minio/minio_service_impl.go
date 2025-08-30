package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
	"upload-service/ddd/domain/gateway"
	"upload-service/ddd/domain/vo"
	"upload-service/internal/resource"
	"upload-service/pkg/assert"
	"upload-service/pkg/errno"
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
	// 获取当前时间用于生成日期路径
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// 从文件名中提取扩展名
	fileName := genStoPathVo.FileName()
	ext := ""
	if dotIndex := strings.LastIndex(fileName, "."); dotIndex != -1 {
		ext = fileName[dotIndex+1:]
	}

	// 生成路径格式: /uploads/{user_uuid}/{yyyy}/{MM}/{dd}/{file_uuid}.{ext}
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

func (m *MinioServiceImpl) GenerateChunkStoragePath(ctx context.Context, uploadVideoUUID string, chunkIndex int) string {
	// 获取当前日期，用于分目录
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// 生成路径格式: chunks/{yyyy}/{MM}/{dd}/{uploadVideoUUID}/chunk_{chunkIndex}
	storagePath := fmt.Sprintf("chunks/%s/%s/%s/%s/chunk_%d",
		year,
		month,
		day,
		uploadVideoUUID,
		chunkIndex,
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
