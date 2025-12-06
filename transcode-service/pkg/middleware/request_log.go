package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"transcode-service/pkg/logger"
	"transcode-service/pkg/restapi"
)

// RequestLogMiddleware logs HTTP requests with request_id and basic metrics.
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start).Milliseconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		fields := map[string]interface{}{
			"request_id": restapi.GetRequestId(c),
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": latency,
		}
		logger.WithFields(fields).Info("http request")
	}
}
