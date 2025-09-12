package manager

import (
	"github.com/gin-gonic/gin"
	"transcode-service/pkg/config"
	"transcode-service/pkg/utils"

	"gorm.io/gorm"
)

// Dependencies 依赖注入容器
type Dependencies struct {
	DB      *gorm.DB
	Config  *config.Config
	JWTUtil *utils.JWTUtil
}

// ComponentPlugin 组件插件接口
type ComponentPlugin interface {
	// Name 返回插件的名称，不同 ComponentPlugin 的名称不能相同
	Name() string

	// MustCreateComponent 创建 Component，如果创建失败需要 panic
	MustCreateComponent(deps *Dependencies) Component
}

// Component 组件接口
type Component interface {
	// Start 启动组件
	Start() error

	// Stop 停止组件
	Stop() error

	// GetName 获取组件名称
	GetName() string
}

// ServicePlugin 服务插件接口
type ServicePlugin interface {
	// Name 返回插件的名称，不同 ServicePlugin 的名称不能相同
	Name() string

	// MustCreateService 创建 Service，如果创建失败需要 panic
	MustCreateService(deps *Dependencies) Service
}

// Service 服务接口
type Service interface {
	// GetName 获取服务名称
	GetName() string

	// RegisterRoutes 注册路由
	RegisterRoutes(router *gin.Engine)
}

var (
	componentPlugins = map[string]ComponentPlugin{}
	servicePlugins   = map[string]ServicePlugin{}
	components       []Component
	services         []Service
)

// RegisterComponentPlugin 注册组件插件
func RegisterComponentPlugin(p ComponentPlugin) {
	if p.Name() == "" {
		panic("component plugin name cannot be empty")
	}

	if _, existed := componentPlugins[p.Name()]; existed {
		panic("component plugin name already exists: " + p.Name())
	}

	componentPlugins[p.Name()] = p
}

// RegisterServicePlugin 注册服务插件
func RegisterServicePlugin(p ServicePlugin) {
	if p.Name() == "" {
		panic("service plugin name cannot be empty")
	}

	if _, existed := servicePlugins[p.Name()]; existed {
		panic("service plugin name already exists: " + p.Name())
	}

	servicePlugins[p.Name()] = p
}

// MustInitComponents 初始化所有组件
func MustInitComponents(deps *Dependencies) {
	for name, plugin := range componentPlugins {
		component := plugin.MustCreateComponent(deps)
		if err := component.Start(); err != nil {
			panic("failed to start component " + name + ": " + err.Error())
		}
		components = append(components, component)
	}
}

// MustInitServices 初始化所有服务
func MustInitServices(deps *Dependencies) {
	for _, plugin := range servicePlugins {
		service := plugin.MustCreateService(deps)
		services = append(services, service)
	}
}

// RegisterAllRoutes 注册所有路由
func RegisterAllRoutes(router *gin.Engine) {
	// 创建API分组
	openApiGroup := router.Group("/api")
	innerApiGroup := router.Group("/inner")
	debugApiGroup := router.Group("/debug")
	opsApiGroup := router.Group("/ops")

	// 注册控制器路由
	MustInitControllers(openApiGroup, innerApiGroup, debugApiGroup, opsApiGroup)

	// 注册服务路由
	for _, service := range services {
		service.RegisterRoutes(router)
	}
}

// Shutdown 关闭所有组件
func Shutdown() {
	// 逆序关闭组件
	for i := len(components) - 1; i >= 0; i-- {
		if err := components[i].Stop(); err != nil {
			// 记录错误但继续关闭其他组件
			// logger.Error("failed to stop component", map[string]interface{}{"error": err})
		}
	}
	components = nil
	services = nil
}
