package manager

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-video/pkg/config"
	"go-video/pkg/utils"
)

type (
	// ServicePlugin 服务插件接口
	ServicePlugin interface {
		// Name 返回插件的名称，不同 ServicePlugin 的名称不能相同
		Name() string

		// MustCreateService 创建服务，如果创建失败需要 panic
		MustCreateService(deps *Dependencies) Service
	}

	// Service 服务接口
	Service interface {
		// RegisterRoutes 注册路由
		RegisterRoutes(router *gin.Engine, jwtUtil *utils.JWTUtil)

		// GetName 获取服务名称
		GetName() string
	}

	// ControllerPlugin 控制器插件接口
	ControllerPlugin interface {
		Name() string
		RegisterOpenApi(router *gin.Engine)
		RegisterInnerApi(router *gin.Engine)
		RegisterOpsApi(router *gin.Engine)
		RegisterDebugApi(router *gin.Engine)
	}

	// ComponentPlugin 组件插件接口
	ComponentPlugin interface {
		Name() string
		MustCreateComponent(deps *Dependencies) Component
	}

	// Component 组件接口
	Component interface {
		GetName() string
		Init() error
		Destroy() error
	}

	// Dependencies 依赖注入容器
	Dependencies struct {
		DB      *gorm.DB
		Config  *config.Config
		JWTUtil *utils.JWTUtil
	}
)

var (
	servicePlugins    = map[string]ServicePlugin{}
	controllerPlugins = map[string]ControllerPlugin{}
	componentPlugins  = map[string]ComponentPlugin{}
	services          = map[string]Service{}
	controllers       = map[string]ControllerPlugin{}
	components        = map[string]Component{}
	mutex             sync.RWMutex
)

// RegisterServicePlugin 注册服务插件
func RegisterServicePlugin(p ServicePlugin) {
	mutex.Lock()
	defer mutex.Unlock()
	
	if p.Name() == "" {
		panic(fmt.Errorf("%T: empty name", p))
	}

	existedPlugin, existed := servicePlugins[p.Name()]
	if existed {
		panic(fmt.Errorf("%T and %T got same name: %s", p, existedPlugin, p.Name()))
	}

	servicePlugins[p.Name()] = p
	log.Printf("Registered service plugin: %s", p.Name())
}

// RegisterControllerPlugin 注册控制器插件
func RegisterControllerPlugin(p ControllerPlugin) {
	mutex.Lock()
	defer mutex.Unlock()
	
	if p.Name() == "" {
		panic(fmt.Errorf("%T: empty name", p))
	}

	existedPlugin, existed := controllerPlugins[p.Name()]
	if existed {
		panic(fmt.Errorf("%T and %T got same name: %s", p, existedPlugin, p.Name()))
	}

	controllerPlugins[p.Name()] = p
	log.Printf("Registered controller plugin: %s", p.Name())
}

// RegisterComponentPlugin 注册组件插件
func RegisterComponentPlugin(p ComponentPlugin) {
	mutex.Lock()
	defer mutex.Unlock()
	
	if p.Name() == "" {
		panic(fmt.Errorf("%T: empty name", p))
	}

	existedPlugin, existed := componentPlugins[p.Name()]
	if existed {
		panic(fmt.Errorf("%T and %T got same name: %s", p, existedPlugin, p.Name()))
	}

	componentPlugins[p.Name()] = p
	log.Printf("Registered component plugin: %s", p.Name())
}

// MustInitServices 初始化已注册的服务，包括相关服务的创建，如果失败则 panic
func MustInitServices(deps *Dependencies) {
	mutex.Lock()
	defer mutex.Unlock()
	
	log.Println("Initializing services...")
	for name, plugin := range servicePlugins {
		log.Printf("Creating service: %s", name)
		service := plugin.MustCreateService(deps)
		services[name] = service
		log.Printf("Initialized service: plugin=%s, service=%s", name, service.GetName())
	}
	log.Printf("Total services initialized: %d", len(services))
}

// MustInitComponents 初始化已注册的组件，包括相关组件的创建，如果失败则 panic
func MustInitComponents(deps *Dependencies) {
	mutex.Lock()
	defer mutex.Unlock()
	
	log.Println("Initializing components...")
	for name, plugin := range componentPlugins {
		log.Printf("Creating component: %s", name)
		component := plugin.MustCreateComponent(deps)
		if err := component.Init(); err != nil {
			log.Fatalf("Failed to initialize component %s: %v", component.GetName(), err)
		}
		components[name] = component
		log.Printf("Initialized component: plugin=%s, component=%s", name, component.GetName())
	}
	log.Printf("Total components initialized: %d", len(components))
}

// RegisterAllRoutes 为所有已初始化的服务和控制器注册路由
func RegisterAllRoutes(router *gin.Engine, jwtUtil *utils.JWTUtil) {
	mutex.RLock()
	defer mutex.RUnlock()
	
	log.Println("Registering routes...")
	
	// 注册服务路由
	for name, service := range services {
		log.Printf("Registering routes for service: %s", name)
		service.RegisterRoutes(router, jwtUtil)
	}
	
	// 注册控制器路由
	for name, controller := range controllers {
		log.Printf("Registering routes for controller: %s", name)
		controller.RegisterOpenApi(router)
		controller.RegisterInnerApi(router)
		controller.RegisterOpsApi(router)
		controller.RegisterDebugApi(router)
	}
	
	log.Println("All routes registered")
}

// GetService 获取指定名称的服务
func GetService(name string) (Service, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	service, exists := services[name]
	return service, exists
}

// GetController 获取指定名称的控制器
func GetController(name string) (ControllerPlugin, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	controller, exists := controllers[name]
	return controller, exists
}

// GetComponent 获取指定名称的组件
func GetComponent(name string) (Component, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	component, exists := components[name]
	return component, exists
}

// GetAllServices 获取所有已初始化的服务
func GetAllServices() map[string]Service {
	mutex.RLock()
	defer mutex.RUnlock()
	result := make(map[string]Service)
	for k, v := range services {
		result[k] = v
	}
	return result
}

// GetAllControllers 获取所有已初始化的控制器
func GetAllControllers() map[string]ControllerPlugin {
	mutex.RLock()
	defer mutex.RUnlock()
	result := make(map[string]ControllerPlugin)
	for k, v := range controllers {
		result[k] = v
	}
	return result
}

// GetAllComponents 获取所有已初始化的组件
func GetAllComponents() map[string]Component {
	mutex.RLock()
	defer mutex.RUnlock()
	result := make(map[string]Component)
	for k, v := range components {
		result[k] = v
	}
	return result
}

// Shutdown 优雅关闭所有组件
func Shutdown() {
	mutex.Lock()
	defer mutex.Unlock()
	
	log.Println("Shutting down components...")
	for name, component := range components {
		log.Printf("Destroying component: %s", name)
		if err := component.Destroy(); err != nil {
			log.Printf("Error destroying component %s: %v", name, err)
		}
	}
	log.Println("All components shut down")
}

// Reset 重置所有服务（仅用于测试）
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()
	
	servicePlugins = map[string]ServicePlugin{}
	controllerPlugins = map[string]ControllerPlugin{}
	componentPlugins = map[string]ComponentPlugin{}
	services = map[string]Service{}
	controllers = map[string]ControllerPlugin{}
	components = map[string]Component{}
	log.Println("Reset all services, controllers, components and plugins")
}
