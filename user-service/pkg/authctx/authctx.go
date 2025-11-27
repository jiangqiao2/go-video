package authctx

import (
	"user-service/pkg/errno"

	"github.com/gin-gonic/gin"
)

// GetUserUUID returns the user_uuid from context.
func GetUserUUID(ctx *gin.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Get("user_uuid")
	if !ok {
		return "", false
	}
	if s, ok := v.(string); ok && s != "" {
		return s, true
	}
	return "", false
}

// MustGetUserUUID returns user_uuid or an error suitable for API failure.
func MustGetUserUUID(ctx *gin.Context) (string, error) {
	if uid, ok := GetUserUUID(ctx); ok {
		return uid, nil
	}
	return "", errno.ErrUnauthorized
}
