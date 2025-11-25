package resource

import (
	"sync"
	"user-service/pkg/assert"
	"user-service/pkg/manager"
)

var (
	minioOnce              sync.Once
	singletonMinioResource *MinioResource
)

// MinioResource MinIO资源管理器（简化版）
type MinioResource struct {
	// TODO: 实现MinIO客户端
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
	// TODO: 实现MinIO连接逻辑
}

// Close 关闭MinIO连接
func (r *MinioResource) Close() {
	// TODO: 实现MinIO关闭逻辑
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
