package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gateway-service/pkg/auth"
)

// Authenticator validates JWT tokens and enriches downstream requests.
type Authenticator struct {
	jwt    *auth.JWTUtil
	logger *logrus.Entry
}

// NewAuthenticator creates a reusable authenticator middleware helper.
func NewAuthenticator(jwtUtil *auth.JWTUtil, logger *logrus.Logger) *Authenticator {
	return &Authenticator{
		jwt:    jwtUtil,
		logger: logrus.NewEntry(logger).WithField("component", "auth"),
	}
}

// Wrap wraps a handler with authentication logic.
func (a *Authenticator) Wrap(handler gin.HandlerFunc, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			handler(c)
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if required {
				abortUnauthorized(c, "missing authorization header")
				return
			}
			handler(c)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			if required {
				abortUnauthorized(c, "invalid authorization format")
				return
			}
			handler(c)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			if required {
				abortUnauthorized(c, "empty access token")
				return
			}
			handler(c)
			return
		}

		userUUID, userID, err := a.jwt.ValidateAccessTokenWithUUID(token)
		if err != nil {
			a.logger.WithError(err).Warn("token validation failed")
			if required {
				abortUnauthorized(c, "invalid access token")
				return
			}
			handler(c)
			return
		}

		if userUUID != "" {
			c.Set("user_uuid", userUUID)
			c.Request.Header.Set("X-User-UUID", userUUID)
		}
		if userID != 0 {
			c.Set("user_id", userID)
			c.Request.Header.Set("X-User-ID", strconv.FormatUint(userID, 10))
		}

		handler(c)
	}
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": "未授权",
		"error":   message,
	})
}

