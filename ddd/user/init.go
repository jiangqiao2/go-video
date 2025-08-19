package user

import (
	"sync"

	"go-video/ddd/user/adapter/http"
	"go-video/ddd/user/application/app"
	"go-video/ddd/user/domain/service"
	"go-video/ddd/user/infrastructure/database/persistence"
	"go-video/pkg/assert"
	"go-video/pkg/manager"
	"go-video/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	userPluginOnce      sync.Once
	singletonUserPlugin *UserPlugin
)

// UserPlugin 用户插件
type UserPlugin struct {
	userController *http.UserController
}

// DefaultUserPlugin 获取用户插件单例
func DefaultUserPlugin() *UserPlugin {
	assert.NotCircular()
	userPluginOnce.Do(func() {
		userController := http.DefaultUserController()
		singletonUserPlugin = &UserPlugin{
			userController: userController,
		}
	})
	assert.NotNil(singletonUserPlugin)
	return singletonUserPlugin
}

// NewUserPlugin 创建用户插件实例（支持依赖注入）
func NewUserPlugin(db *gorm.DB, jwtUtil *utils.JWTUtil) *UserPlugin {
	// 创建仓储
	userRepo := persistence.NewUserRepository()

	// 创建领域服务
	userDomainService := service.NewUserService(userRepo)

	// 创建应用服务
	userApp := app.NewUserApp(userDomainService, jwtUtil)

	// 创建控制器
	userController := http.NewUserController(userApp)

	return &UserPlugin{
		userController: userController,
	}
}

// Name 插件名称
func (p *UserPlugin) Name() string {
	return "user"
}

// MustCreateService 创建服务（实现manager.ServicePlugin接口）
func (p *UserPlugin) MustCreateService(deps *manager.Dependencies) manager.Service {
	// 使用依赖注入创建插件
	plugin := NewUserPlugin(deps.DB, deps.JWTUtil)

	return &UserService{
		userController: plugin.userController,
	}
}

// UserService 用户服务（实现manager.Service接口）
type UserService struct {
	userController *http.UserController
}

// GetName 获取服务名称
func (s *UserService) GetName() string {
	return "user"
}

// RegisterRoutes 注册路由
func (s *UserService) RegisterRoutes(router *gin.Engine, jwtUtil *utils.JWTUtil) {
	// 注册开放API（无需认证）
	s.userController.RegisterOpenApi(router)

	// 注册内部API（需要认证）
	s.userController.RegisterInnerApi(router)

	// 注册运维API（需要管理员权限）
	s.userController.RegisterOpsApi(router)

	// 注册调试API
	s.userController.RegisterDebugApi(router)
}
