package resource

import (
	"sync"

	"go-video/pkg/assert"
	"go-video/pkg/manager"
	"go-video/pkg/utils"
)

var (
	jwtResourceOnce      sync.Once
	singletonJWTResource *JWTResource
)

// JWTResource JWT资源管理器
type JWTResource struct {
	jwtUtil *utils.JWTUtil
}

// DefaultJWTResource 获取JWT资源单例
func DefaultJWTResource() *JWTResource {
	assert.NotCircular()
	jwtResourceOnce.Do(func() {
		singletonJWTResource = &JWTResource{}
	})
	assert.NotNil(singletonJWTResource)
	return singletonJWTResource
}

// MustOpen 初始化JWT工具
func (r *JWTResource) MustOpen() {
	if r.jwtUtil == nil {
		r.jwtUtil = utils.DefaultJWTUtil()
	}
	assert.NotNil(r.jwtUtil)
}

// Close 关闭JWT资源（JWT工具无需关闭）
func (r *JWTResource) Close() {
	// JWT工具无需关闭操作
}

// GetJWTUtil 获取JWT工具
func (r *JWTResource) GetJWTUtil() *utils.JWTUtil {
	return r.jwtUtil
}

// JWTResourcePlugin JWT资源插件
type JWTResourcePlugin struct{}

// Name 返回插件名称
func (p *JWTResourcePlugin) Name() string {
	return "jwt"
}

// MustCreateResource 创建JWT资源
func (p *JWTResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultJWTResource()
}

// NewJWTResource 创建JWT资源实例（支持依赖注入）
func NewJWTResource(jwtUtil *utils.JWTUtil) *utils.JWTUtil {
	return jwtUtil
}
