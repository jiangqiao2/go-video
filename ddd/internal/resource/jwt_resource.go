package resource

import (
	"sync"

	"go-video/pkg/assert"
	"go-video/pkg/utils"
)

var (
	jwtResourceOnce      sync.Once
	singletonJWTResource *utils.JWTUtil
)

// DefaultJWTResource 获取JWT资源单例
func DefaultJWTResource() *utils.JWTUtil {
	assert.NotCircular()
	jwtResourceOnce.Do(func() {
		singletonJWTResource = utils.DefaultJWTUtil()
	})
	assert.NotNil(singletonJWTResource)
	return singletonJWTResource
}

// NewJWTResource 创建JWT资源实例（支持依赖注入）
func NewJWTResource(jwtUtil *utils.JWTUtil) *utils.JWTUtil {
	return jwtUtil
}