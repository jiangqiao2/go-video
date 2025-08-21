package http

import (
	"go-video/pkg/manager"
	"strconv"
	"sync"

	"go-video/ddd/user/application/app"
	"go-video/ddd/user/application/cqe"
	"go-video/ddd/user/application/dto"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
	"go-video/pkg/restapi"

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
	userController := DefaultUserController()
	// 确保类型转换正确
	if controller, ok := userController.(manager.Controller); ok {
		return controller
	}
	panic("UserController does not implement manager.Controller")
}

// UserController 用户控制器
type UserController interface {
}

type userControllerImpl struct {
	manager.Controller
	userApp *app.UserApp
}

// DefaultUserController 获取用户控制器单例
func DefaultUserController() UserController {
	assert.NotCircular()
	userControllerOnce.Do(func() {
		userApp := app.DefaultUserApp()
		singletonUserController = &userControllerImpl{
			userApp: userApp,
		}
	})
	assert.NotNil(singletonUserController)
	return singletonUserController
}

// RegisterOpenApi 注册开放API
func (c *userControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		// 用户注册
		v1.POST("/users/register", c.Register)
		// 用户登录
		v1.POST("/users/login", c.Login)
	}
}

// RegisterInnerApi 注册内部API
func (c *userControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	v1.Use(c.AuthMiddleware()) // 需要认证
	{
		// 获取当前用户信息
		v1.GET("/users/me", c.GetCurrentUser)
		// 更新用户资料
		v1.PUT("/users/me", c.UpdateProfile)
		// 修改密码
		v1.PUT("/users/me/password", c.ChangePassword)
	}
}

// RegisterOpsApi 注册运维API
func (c *userControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	ops := router.Group("/v1")
	ops.Use(c.AdminAuthMiddleware()) // 需要管理员权限
	{
		// 获取用户列表
		ops.GET("/users", c.GetUserList)
		// 获取用户详情
		ops.GET("/users/:id", c.GetUserDetail)
		// 激活用户
		ops.PUT("/users/:id/activate", c.ActivateUser)
		// 禁用用户
		ops.PUT("/users/:id/disable", c.DisableUser)
		// 删除用户
		ops.DELETE("/users/:id", c.DeleteUser)
	}
}

// RegisterDebugApi 注册调试API
func (c *userControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	debug := router.Group("/v1")
	{
		// 验证令牌
		debug.POST("/users/validate-token", c.ValidateToken)
	}
}

// Register 用户注册
func (c *userControllerImpl) Register(ctx *gin.Context) {
	var cmd cqe.CreateUserCommand
	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}

	resp, err := c.userApp.CreateUser(ctx.Request.Context(), &cmd)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, resp)
}

// Login 用户登录
func (c *userControllerImpl) Login(ctx *gin.Context) {
	var cmd cqe.LoginCommand
	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}

	resp, err := c.userApp.Login(ctx.Request.Context(), &cmd)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, resp)
}

// GetCurrentUser 获取当前用户信息
func (c *userControllerImpl) GetCurrentUser(ctx *gin.Context) {
	userUUID := c.getCurrentUserUUID(ctx)
	if userUUID == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrUnauthorized, nil, nil))
		return
	}

	// 通过UUID获取用户信息
	userInfo, err := c.userApp.ValidateToken(ctx.Request.Context(), c.getTokenFromContext(ctx))
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, userInfo)
}

// UpdateProfile 更新用户资料
func (c *userControllerImpl) UpdateProfile(ctx *gin.Context) {
	userUUID := c.getCurrentUserUUID(ctx)
	if userUUID == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrUnauthorized, nil, nil))
		return
	}

	var cmd cqe.UpdateProfileCommand
	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// 验证用户身份并获取内部ID
	_, err := c.userApp.ValidateToken(ctx.Request.Context(), c.getTokenFromContext(ctx))
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// TODO: 重构为基于UUID的实现，暂时跳过UserID设置

	if err := c.userApp.UpdateProfile(ctx.Request.Context(), &cmd); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, &dto.CommonResponse{Message: "更新成功"})
}

// ChangePassword 修改密码
func (c *userControllerImpl) ChangePassword(ctx *gin.Context) {
	userUUID := c.getCurrentUserUUID(ctx)
	if userUUID == "" {
		restapi.Failed(ctx, errno.NewBizError(errno.ErrUnauthorized, nil))
		return
	}

	var cmd cqe.ChangePasswordCommand
	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}

	// 验证用户身份
	_, err := c.userApp.ValidateToken(ctx.Request.Context(), c.getTokenFromContext(ctx))
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	// TODO: 重构为基于UUID的实现，暂时跳过UserID设置

	if err := c.userApp.ChangePassword(ctx.Request.Context(), &cmd); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, &dto.CommonResponse{Message: "密码修改成功"})
}

// GetUserList 获取用户列表
func (c *userControllerImpl) GetUserList(ctx *gin.Context) {
	var query cqe.GetUserListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}

	// 设置默认值
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	resp, err := c.userApp.GetUserList(ctx.Request.Context(), &query)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, map[string]interface{}{
		"users":     resp.Users,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

// GetUserDetail 获取用户详情
func (c *userControllerImpl) GetUserDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "id"))
		return
	}

	userInfo, err := c.userApp.GetUserInfo(ctx.Request.Context(), id)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, userInfo)
}

// ActivateUser 激活用户
func (c *userControllerImpl) ActivateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "id"))
		return
	}

	if err := c.userApp.ActivateUser(ctx.Request.Context(), id); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, &dto.CommonResponse{Message: "用户激活成功"})
}

// DisableUser 禁用用户
func (c *userControllerImpl) DisableUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "id"))
		return
	}

	if err := c.userApp.DisableUser(ctx.Request.Context(), id); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, &dto.CommonResponse{Message: "用户禁用成功"})
}

// DeleteUser 删除用户
func (c *userControllerImpl) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "id"))
		return
	}

	if err := c.userApp.DeleteUser(ctx.Request.Context(), id); err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, &dto.CommonResponse{Message: "用户删除成功"})
}

// ValidateToken 验证令牌
func (c *userControllerImpl) ValidateToken(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "Authorization header missing"))
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userInfo, err := c.userApp.ValidateToken(ctx.Request.Context(), token)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, userInfo)
}

// AuthMiddleware 认证中间件
func (c *userControllerImpl) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			restapi.Failed(ctx, errno.NewBizError(errno.ErrUnauthorized, nil))
			ctx.Abort()
			return
		}

		// 移除Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userInfo, err := c.userApp.ValidateToken(ctx.Request.Context(), token)
		if err != nil {
			restapi.Failed(ctx, err)
			ctx.Abort()
			return
		}

		// 将用户信息存储到上下文
		ctx.Set("user_uuid", userInfo.UUID)
		ctx.Set("username", userInfo.Username)
		ctx.Next()
	}
}

// AdminAuthMiddleware 管理员认证中间件
func (c *userControllerImpl) AdminAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 先进行普通认证
		c.AuthMiddleware()(ctx)
		if ctx.IsAborted() {
			return
		}

		// TODO: 添加管理员权限检查逻辑
		// 这里可以检查用户角色或权限
		ctx.Next()
	}
}

// getCurrentUserUUID 获取当前用户UUID
func (c *userControllerImpl) getCurrentUserUUID(ctx *gin.Context) string {
	if userUUID, exists := ctx.Get("user_uuid"); exists {
		if uuid, ok := userUUID.(string); ok {
			return uuid
		}
	}
	return ""
}

// getTokenFromContext 从上下文获取JWT令牌
func (c *userControllerImpl) getTokenFromContext(ctx *gin.Context) string {
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:]
	}
	return token
}
