package middleware

import (
	"net/http"
	"strings"
	"transcode-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtUtil *utils.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未授权",
				"error":   "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未授权",
				"error":   "无效的Authorization格式",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未授权",
				"error":   "缺少访问令牌",
			})
			c.Abort()
			return
		}

		// 验证token（优先使用UUID格式）
		userUUID, userID, err := jwtUtil.ValidateAccessTokenWithUUID(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未授权",
				"error":   "无效的访问令牌",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		if userUUID != "" {
			c.Set("user_uuid", userUUID) // 优先存储UUID
		}
		c.Set("user_id", userID) // 兼容性支持
		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware(jwtUtil *utils.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Next()
			return
		}

		// 验证token（优先使用UUID格式）
		userUUID, userID, err := jwtUtil.ValidateAccessTokenWithUUID(token)
		if err != nil {
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		if userUUID != "" {
			c.Set("user_uuid", userUUID) // 优先存储UUID
		}
		c.Set("user_id", userID) // 兼容性支持
		c.Next()
	}
}
