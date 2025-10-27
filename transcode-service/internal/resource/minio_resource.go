package resource

import (
	"context"
	"fmt"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"transcode-service/pkg/assert"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
)

var (
	minioResourceOnce      sync.Once
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
	minioResourceOnce.Do(func() {
		singletonMinioResource = &MinioResource{}
	})
	assert.NotNil(singletonMinioResource)
	return singletonMinioResource
}

// MustOpen 初始化MinIO资源
func (r *MinioResource) MustOpen() {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		panic("global config not initialized before MinioResource")
	}

	minioCfg := cfg.Minio
	if minioCfg.Endpoint == "" {
		panic("minio endpoint is required")
	}
	if minioCfg.BucketName == "" {
		panic("minio bucket_name is required")
	}

	endpoint := minioCfg.Endpoint
	accessKey := minioCfg.AccessKeyID
	secretKey := minioCfg.SecretAccessKey

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: minioCfg.UseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create minio client: %v", err))
	}

	r.client = client
	r.bucketName = minioCfg.BucketName

	r.ensureBucket()

	logger.Info("MinIO resource initialized", map[string]interface{}{
		"endpoint":    endpoint,
		"bucket_name": r.bucketName,
	})
}

// ensureBucket 确保桶存在
func (r *MinioResource) ensureBucket() {
	ctx := context.Background()
	exists, err := r.client.BucketExists(ctx, r.bucketName)
	if err != nil {
		panic(fmt.Sprintf("failed to check minio bucket: %v", err))
	}
	if exists {
		return
	}
	if err := r.client.MakeBucket(ctx, r.bucketName, minio.MakeBucketOptions{}); err != nil {
		panic(fmt.Sprintf("failed to create minio bucket: %v", err))
	}
}

// GetClient 获取MinIO客户端
func (r *MinioResource) GetClient() *minio.Client {
	return r.client
}

// GetBucketName 获取桶名称
func (r *MinioResource) GetBucketName() string {
	return r.bucketName
}

// Close 释放资源
func (r *MinioResource) Close() {
	// minio-go客户端无需关闭连接
}

// MinioResourcePlugin MinIO资源插件
type MinioResourcePlugin struct{}

func (p *MinioResourcePlugin) Name() string {
	return "minioResource"
}

func (p *MinioResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultMinioResource()
}
