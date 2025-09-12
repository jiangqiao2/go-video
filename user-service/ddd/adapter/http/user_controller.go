package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync"
	"user-service/ddd/application/app"
	"user-service/ddd/application/cqe"
	"user-service/pkg/assert"
	"user-service/pkg/manager"
	"user-service/pkg/restapi"
)

var (
	userControllerOnce      sync.Once
	singletonUserController UserController
)

type UserControllerPlugin struct {
}

func (p *UserControllerPlugin) Name() string {
	return "userControllerPlugin"
}

func (p *UserControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	userControllerOnce.Do(func() {
		singletonUserController = &userControllerImpl{
			userApp: app.DefaultUserApp(),
		}
	})
	assert.NotNil(singletonUserController)
	return singletonUserController
}

type UserController interface {
	manager.Controller
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
}

type userControllerImpl struct {
	manager.Controller
	userApp app.UserApp
}

// RegisterOpenApi 注册开放API
func (c *userControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	// 开放API实现
	v1 := router.Group("open/v1/users")
	{
		v1.POST("/register", c.Register)
		v1.POST("/login", c.Login)
	}
}

// RegisterInnerApi 注册内部API
func (c *userControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	// 内部API实现
	v1 := router.Group("inner/v1/users")
	{
		v1.GET("/me", c.QueryUserInfo)
	}
}

// RegisterDebugApi 注册调试API
func (c *userControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {

}

// RegisterOpsApi 注册运维API
func (c *userControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

// Register 用户注册
func (c *userControllerImpl) Register(ctx *gin.Context) {
	var req cqe.UserRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	result, err := c.userApp.Register(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

// Login 用户登录
func (c *userControllerImpl) Login(ctx *gin.Context) {
	var req cqe.UserLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	result, err := c.userApp.Login(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

func (c *userControllerImpl) QueryUserInfo(ctx *gin.Context) {

}
