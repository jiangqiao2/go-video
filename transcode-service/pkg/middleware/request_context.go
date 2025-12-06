package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"transcode-service/pkg/grpcutil"
	"transcode-service/pkg/restapi"
)

// RequestContextMiddleware 注入 user_uuid 和 request_id，便于下游和日志使用。
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID := c.GetHeader("X-User-UUID")
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		if userUUID != "" {
			c.Set("user_uuid", userUUID)
		}
		c.Set("request_id", reqID)
		c.Set(restapi.HeaderKeyRequestId, reqID)
		ctxWithReqID, _ := grpcutil.ContextWithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctxWithReqID)
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}
