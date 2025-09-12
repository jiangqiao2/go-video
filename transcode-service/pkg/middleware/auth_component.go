package middleware

import (
	"sync"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

var (
	authComponentOnce      sync.Once
	singletonAuthComponent *AuthComponent
)

// AuthComponent 认证组件
type AuthComponent struct {
	jwtUtil *utils.JWTUtil
}

// DefaultAuthComponent 获取认证组件单例
func DefaultAuthComponent() *AuthComponent {
	assert.NotCircular()
	authComponentOnce.Do(func() {
		singletonAuthComponent = &AuthComponent{
			jwtUtil: utils.DefaultJWTUtil(),
		}
	})
	assert.NotNil(singletonAuthComponent)
	return singletonAuthComponent
}

// NewAuthComponent 创建认证组件
func NewAuthComponent(jwtUtil *utils.JWTUtil) *AuthComponent {
	return &AuthComponent{
		jwtUtil: jwtUtil,
	}
}

// Required 必需认证中间件
// 用法: router.POST("/api/v1/videos/upload", auth.Required(), handler)
func (a *AuthComponent) Required() gin.HandlerFunc {
	return AuthMiddleware(a.jwtUtil)
}

// Optional 可选认证中间件
// 用法: router.GET("/api/v1/videos", auth.Optional(), handler)
func (a *AuthComponent) Optional() gin.HandlerFunc {
	return OptionalAuthMiddleware(a.jwtUtil)
}

// GetUserUUID 从上下文获取用户UUID（推荐使用，不暴露内部ID）
func (a *AuthComponent) GetUserUUID(c *gin.Context) (string, bool) {
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		return "", false
	}
	if uuid, ok := userUUID.(string); ok {
		return uuid, true
	}
	return "", false
}

// MustGetUserUUID 从上下文获取用户UUID（如果不存在则panic）
func (a *AuthComponent) MustGetUserUUID(c *gin.Context) string {
	userUUID, exists := a.GetUserUUID(c)
	if !exists {
		panic("用户UUID不存在，请确保使用了认证中间件")
	}
	return userUUID
}

// GetUserID 从上下文获取用户ID（兼容性保留，但不推荐使用）
// 注意：这会暴露数据库内部自增ID，建议使用GetUserUUID
func (a *AuthComponent) GetUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	if id, ok := userID.(uint64); ok {
		return id, true
	}
	return 0, false
}

// MustGetUserID 从上下文获取用户ID（如果不存在则panic）
// 注意：这会暴露数据库内部自增ID，建议使用MustGetUserUUID
func (a *AuthComponent) MustGetUserID(c *gin.Context) uint64 {
	userID, exists := a.GetUserID(c)
	if !exists {
		panic("用户ID不存在，请确保使用了认证中间件")
	}
	return userID
}

// IsAuthenticated 检查用户是否已认证
func (a *AuthComponent) IsAuthenticated(c *gin.Context) bool {
	_, exists := a.GetUserUUID(c)
	if !exists {
		// 兼容性检查：如果没有user_uuid，检查user_id
		_, exists = a.GetUserID(c)
	}
	return exists
}

// 全局便捷函数

// AuthRequired 全局必需认证中间件
func AuthRequired() gin.HandlerFunc {
	return DefaultAuthComponent().Required()
}

// AuthOptional 全局可选认证中间件
func AuthOptional() gin.HandlerFunc {
	return DefaultAuthComponent().Optional()
}

// GetCurrentUserUUID 全局获取当前用户UUID（推荐使用）
func GetCurrentUserUUID(c *gin.Context) (string, bool) {
	return DefaultAuthComponent().GetUserUUID(c)
}

// MustGetCurrentUserUUID 全局获取当前用户UUID（必须存在）
func MustGetCurrentUserUUID(c *gin.Context) string {
	return DefaultAuthComponent().MustGetUserUUID(c)
}

// GetCurrentUserID 全局获取当前用户ID（不推荐，会暴露内部ID）
func GetCurrentUserID(c *gin.Context) (uint64, bool) {
	return DefaultAuthComponent().GetUserID(c)
}

// MustGetCurrentUserID 全局获取当前用户ID（必须存在，不推荐）
func MustGetCurrentUserID(c *gin.Context) uint64 {
	return DefaultAuthComponent().MustGetUserID(c)
}

// IsCurrentUserAuthenticated 全局检查用户是否已认证
func IsCurrentUserAuthenticated(c *gin.Context) bool {
	return DefaultAuthComponent().IsAuthenticated(c)
}
