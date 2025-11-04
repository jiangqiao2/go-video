package resource

import (
    "context"
    "sync"
    "upload-service/pkg/assert"
    "upload-service/pkg/config"
    "upload-service/pkg/logger"
    "upload-service/pkg/manager"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
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
        cfg := config.GetGlobalConfig()
        if cfg == nil {
            panic("global config not initialized")
        }
        client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
            Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
            Secure: cfg.Minio.UseSSL,
        })
        if err != nil {
            panic("failed to create minio client: " + err.Error())
        }
        singletonMinioResource = &MinioResource{
            client:     client,
            bucketName: cfg.Minio.BucketName,
        }
    })
    assert.NotNil(singletonMinioResource)
    return singletonMinioResource
}

// MustOpen 打开MinIO连接
func (r *MinioResource) MustOpen() {
    if r.client == nil {
        cfg := config.GetGlobalConfig()
        if cfg == nil {
            panic("global config not initialized")
        }
        client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
            Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
            Secure: cfg.Minio.UseSSL,
        })
        if err != nil {
            panic("failed to create minio client: " + err.Error())
        }
        r.client = client
        r.bucketName = cfg.Minio.BucketName
    }
    assert.NotNil(r.client)

    // 确保存储桶存在
    r.ensureBucket()
}

// 使用全局配置创建客户端，无需单独加载配置文件

// ensureBucket 确保存储桶存在
func (r *MinioResource) ensureBucket() {
	ctx := context.Background()
	exists, err := r.client.BucketExists(ctx, r.bucketName)
	if err != nil {
		logger.DefaultLogger().Errorf("check bucket exists failed: %v", err)
		return
	}

	if !exists {
		err = r.client.MakeBucket(ctx, r.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			logger.DefaultLogger().Errorf("create bucket failed: %v", err)
		} else {
			logger.DefaultLogger().Infof("bucket '%s' created successfully", r.bucketName)
		}
	} else {
		logger.DefaultLogger().Infof("bucket '%s' already exists", r.bucketName)
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
