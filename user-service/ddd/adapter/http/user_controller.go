package http

import (
	"context"
	"sync"
	"user-service/ddd/application/app"
	"user-service/ddd/application/cqe"
	"user-service/pkg/assert"
	"user-service/pkg/errno"
	"user-service/pkg/manager"
	"user-service/pkg/middleware"
	"user-service/pkg/restapi"

	"github.com/gin-gonic/gin"
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
	Refresh(ctx *gin.Context)
	Logout(ctx *gin.Context)
	SaveUser(ctx *gin.Context)
	ChangePassword(ctx *gin.Context)
}

type userControllerImpl struct {
	manager.Controller
	userApp app.UserApp
}

// RegisterOpenApi 注册开放API
func (c *userControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/open/users")
	{
		v1.POST("/register", c.Register)
		v1.POST("/login", c.Login)
		v1.POST("/refresh", c.Refresh)
		v1.POST("/logout", c.Logout)
		v1.GET("/:user_uuid", c.GetUserBasicInfo) // 获取用户基本信息
	}
}

// RegisterInnerApi 注册内部API
func (c *userControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/inner/users")
	{
		v1.GET("/me", middleware.AuthRequired(), c.QueryUserInfo)
		v1.GET("/info/:uuid", middleware.AuthRequired(), c.QueryUserInfo)
		v1.POST("/save", middleware.AuthRequired(), c.SaveUser)
		v1.POST("/password", middleware.AuthRequired(), c.ChangePassword)
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

func (c *userControllerImpl) Refresh(ctx *gin.Context) {
	var req cqe.TokenRefreshReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	result, err := c.userApp.RefreshToken(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

// Logout 注销（删除刷新令牌）
func (c *userControllerImpl) Logout(ctx *gin.Context) {
	var req cqe.TokenRefreshReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	if err := c.userApp.Logout(context.Background(), &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, "ok")
}

func (c *userControllerImpl) QueryUserInfo(ctx *gin.Context) {
	// 从JWT中获取当前用户UUID
	currentUserUUID, exists := ctx.Get("user_uuid")
	if !exists {
		restapi.Failed(ctx, errno.ErrUserInfoNotFound)
		return
	}

	// 获取请求的UUID参数（如果有）
	requestUUID := ctx.Param("uuid")
	if requestUUID == "" {
		// 如果没有UUID参数，返回当前用户信息
		requestUUID = currentUserUUID.(string)
	} else {
		// 如果有UUID参数，验证是否与token中的UUID一致
		if requestUUID != currentUserUUID.(string) {
			restapi.Failed(ctx, errno.ErrUserAccessDenied)
			return
		}
	}

	// 通过应用服务获取用户信息
	userInfo, err := c.userApp.GetUserInfo(context.Background(), requestUUID)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, userInfo)
}

// SaveUser 保存当前用户信息（部分字段）
func (c *userControllerImpl) SaveUser(ctx *gin.Context) {
	var req cqe.UserSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 获取当前用户UUID
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		restapi.Failed(ctx, errno.ErrUnauthorized)
		return
	}

	result, err := c.userApp.SaveUserInfo(context.Background(), userUUID.(string), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

// ChangePassword 修改密码
func (c *userControllerImpl) ChangePassword(ctx *gin.Context) {
	var req cqe.ChangePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		restapi.Failed(ctx, errno.ErrUnauthorized)
		return
	}
	if err := c.userApp.ChangePassword(context.Background(), userUUID.(string), &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, "ok")
}

// GetUserBasicInfo 获取用户基本信息（公开接口）
func (c *userControllerImpl) GetUserBasicInfo(ctx *gin.Context) {
	userUUID := ctx.Param("user_uuid")
	if userUUID == "" {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}

	result, err := c.userApp.GetUserBasicInfo(context.Background(), userUUID)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, result)
}
