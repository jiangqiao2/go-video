package manager

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"notification-service/pkg/config"
)

// Dependencies 依赖注入容器（预留扩展，目前仅持有 DB 和 Config）。
type Dependencies struct {
	DB     *gorm.DB
	Config *config.Config
}

// RegisterAllRoutes 注册所有路由。
// 对通知服务来说，我们只依赖 Controller 插件，不使用组件和服务插件。
func RegisterAllRoutes(router *gin.Engine) {
	openApiGroup := router.Group("/api")
	innerApiGroup := router.Group("/api")
	debugApiGroup := router.Group("/debug")
	opsApiGroup := router.Group("/ops")

	MustInitControllers(openApiGroup, innerApiGroup, debugApiGroup, opsApiGroup)
}
