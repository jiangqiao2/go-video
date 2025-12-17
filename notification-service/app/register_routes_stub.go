package app

import (
	"github.com/gin-gonic/gin"
	"notification-service/pkg/manager"
)

// userServiceRegisterAllRoutes is a thin wrapper to call the local manager's RegisterAllRoutes.
// It exists to keep app.go focused on bootstrapping while still reusing the existing route system.
func userServiceRegisterAllRoutes(router *gin.Engine) {
	manager.RegisterAllRoutes(router)
}
