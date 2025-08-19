package resource

import (
	"context"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go-video/pkg/assert"
	"go-video/pkg/config"
	"go-video/pkg/logger"
	"go-video/pkg/manager"
)

var (
	minioOnce              sync.Once
	singletonMinioResource *MinioResource
)

// MinioResource MinIO资源管理器
type MinioResource struct {
	client     *minio.Client
	bucketName string
}

// DefaultMinioResource 获取MinIO资源单例
func DefaultMinioResource() *MinioResource {
	assert.NotCircular()
	minioOnce.Do(func() {
		singletonMinioResource = &MinioResource{}
	})
	assert.NotNil(singletonMinioResource)
	return singletonMinioResource
}

// MustOpen 打开MinIO连接
func (r *MinioResource) MustOpen() {
	if r.client == nil {
		r.client, r.bucketName = newMinioClient()
		if r.client == nil {
			panic("failed to create minio client")
		}
	}
	assert.NotNil(r.client)

	// 确保存储桶存在
	r.ensureBucket()
}

// newMinioClient 创建MinIO客户端
func newMinioClient() (*minio.Client, string) {
	cfg, err := config.Load("configs/config.dev.yaml")
	if err != nil {
		logger.DefaultLogger().Error("load config failed")
		return nil, ""
	}

	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		logger.DefaultLogger().Error("create minio client failed")
		return nil, ""
	}

	return client, cfg.Minio.BucketName
}

// ensureBucket 确保存储桶存在
func (r *MinioResource) ensureBucket() {
	ctx := context.Background()
	exists, err := r.client.BucketExists(ctx, r.bucketName)
	if err != nil {
		logger.DefaultLogger().Error("check bucket exists failed")
		return
	}

	if !exists {
		err = r.client.MakeBucket(ctx, r.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			logger.DefaultLogger().Error("create bucket failed")
		}
	}
}

// GetClient 获取MinIO客户端
func (r *MinioResource) GetClient() *minio.Client {
	return r.client
}

// GetBucketName 获取存储桶名称
func (r *MinioResource) GetBucketName() string {
	return r.bucketName
}

// Close 关闭MinIO连接
func (r *MinioResource) Close() {
	// MinIO客户端无需显式关闭
}

// MinioResourcePlugin MinIO资源插件
type MinioResourcePlugin struct{}

// Name 返回插件名称
func (p *MinioResourcePlugin) Name() string {
	return "minio"
}

// MustCreateResource 创建MinIO资源
func (p *MinioResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultMinioResource()
}
