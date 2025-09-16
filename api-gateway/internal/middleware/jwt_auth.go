package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"

	"api-gateway/pkg/errno"
	"api-gateway/pkg/utils"
)

// JWTAuthMiddleware JWT鉴权中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取Authorization头中的Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errno.ErrUnauthorized.Code,
				"message": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errno.ErrUnauthorized.Code,
				"message": "Invalid Authorization format",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errno.ErrUnauthorized.Code,
				"message": "Empty token",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		jwtUtil := utils.DefaultJWTUtil()
		userUUID, userID, err := jwtUtil.ValidateAccessTokenWithUUID(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errno.ErrUnauthorized.Code,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息注入到请求头中，传递给后端服务
		c.Header("X-User-UUID", userUUID)
		c.Header("X-User-ID", strconv.FormatUint(userID, 10))

		// 也可以设置到上下文中供当前请求使用
		c.Set("user_uuid", userUUID)
		c.Set("user_id", userID)

		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选的JWT鉴权中间件（不强制要求token）
func OptionalJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Next()
			return
		}

		// 尝试验证token
		jwtUtil := utils.DefaultJWTUtil()
		userUUID, userID, err := jwtUtil.ValidateAccessTokenWithUUID(token)
		if err == nil {
			// 验证成功，注入用户信息
			c.Header("X-User-UUID", userUUID)
			c.Header("X-User-ID", strconv.FormatUint(userID, 10))
			c.Set("user_uuid", userUUID)
			c.Set("user_id", userID)
		}

		c.Next()
	}
}
