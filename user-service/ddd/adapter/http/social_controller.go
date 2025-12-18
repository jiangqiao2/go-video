package http

import (
	"sync"
	"user-service/ddd/application/app"
	"user-service/ddd/application/cqe"
	"user-service/pkg/assert"
	"user-service/pkg/authctx"
	"user-service/pkg/errno"
	"user-service/pkg/manager"
	"user-service/pkg/middleware"
	"user-service/pkg/restapi"

	"github.com/gin-gonic/gin"
)

var (
	socialControllerOnce      sync.Once
	singletonSocialController SocialController
)

type SocialControllerPlugin struct{}

func (p *SocialControllerPlugin) Name() string {
	return "socialControllerPlugin"
}

func (p *SocialControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	socialControllerOnce.Do(func() {
		singletonSocialController = &socialControllerImpl{
			socialApp: app.DefaultSocialApp(),
		}
	})
	assert.NotNil(singletonSocialController)
	return singletonSocialController
}

type SocialController interface {
	manager.Controller
	Follow(ctx *gin.Context)
	Unfollow(ctx *gin.Context)
	ToggleFollow(ctx *gin.Context)
	FollowStatus(ctx *gin.Context)
	ListFollowers(ctx *gin.Context)
	ListFollowings(ctx *gin.Context)
}

type socialControllerImpl struct {
	manager.Controller
	socialApp app.SocialApp
}

func (c *socialControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1Open := router.Group("user/v1/open/users")
	{
		// 使用可选认证中间件，允许未登录访问，但如果登录了会解析token
		v1Open.GET("/:user_uuid/relation", middleware.AuthOptional(), c.GetUserRelation)
	}
}

func (c *socialControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/inner/relation")
	{
		v1.POST("/follow", middleware.AuthRequired(), c.Follow)
		v1.POST("/unfollow", middleware.AuthRequired(), c.Unfollow)
		v1.POST("/follow/toggle", middleware.AuthRequired(), c.ToggleFollow)
		v1.GET("/status", middleware.AuthRequired(), c.FollowStatus)
		v1.GET("/followers", middleware.AuthRequired(), c.ListFollowers)
		v1.GET("/followings", middleware.AuthRequired(), c.ListFollowings)
		// legacy alias for clients calling /follow/status
		v1.GET("/follow/status", middleware.AuthRequired(), c.FollowStatus)
	}
}

func (c *socialControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {}
func (c *socialControllerImpl) RegisterOpsApi(router *gin.RouterGroup)   {}

func (c *socialControllerImpl) Follow(ctx *gin.Context) {
	userUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.FollowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "target_uuid"))
		return
	}
	if req.TargetUUID == "" && req.TargetUserUUID != "" {
		req.TargetUUID = req.TargetUserUUID
	}
	if req.TargetUUID == "" {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}
	// 为了兼容老接口，将请求转换为 ToggleFollow 语义，统一走应用层的 ToggleFollow。
	toggleReq := cqe.FollowToggleReq{
		UserUUID:   userUUID,
		TargetUUID: req.TargetUUID,
		Action:     "follow",
	}
	if err := c.socialApp.ToggleFollow(ctx.Request.Context(), &toggleReq); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, nil)
}

func (c *socialControllerImpl) Unfollow(ctx *gin.Context) {
	userUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.FollowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "target_uuid"))
		return
	}
	if req.TargetUUID == "" && req.TargetUserUUID != "" {
		req.TargetUUID = req.TargetUserUUID
	}
	if req.TargetUUID == "" {
		restapi.Failed(ctx, errno.ErrParameterInvalid)
		return
	}
	toggleReq := cqe.FollowToggleReq{
		UserUUID:   userUUID,
		TargetUUID: req.TargetUUID,
		Action:     "unfollow",
	}
	if err := c.socialApp.ToggleFollow(ctx.Request.Context(), &toggleReq); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, nil)
}

// ToggleFollow 统一的关注/取关接口，通过 follow 字段控制是关注还是取消关注。
func (c *socialControllerImpl) ToggleFollow(ctx *gin.Context) {
	userUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.FollowToggleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "target_uuid"))
		return
	}
	req.UserUUID = userUUID
	if err := c.socialApp.ToggleFollow(ctx.Request.Context(), &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, nil)
}

func (c *socialControllerImpl) FollowStatus(ctx *gin.Context) {
	userUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.FollowStatusReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	req.UserUUID = userUUID
	if req.TargetUUID == "" && req.TargetUserUUID != "" {
		req.TargetUUID = req.TargetUserUUID
	}
	if req.TargetUUID == "" {
		req.TargetUUID = userUUID
	}
	status, err := c.socialApp.FollowStatus(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, status)
}

func (c *socialControllerImpl) ListFollowers(ctx *gin.Context) {
	currentUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var query cqe.FollowListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	if query.TargetUUID == "" && query.TargetUserUUID != "" {
		query.TargetUUID = query.TargetUserUUID
	}
	query.Normalize(currentUUID)
	resp, err := c.socialApp.ListFollowers(ctx.Request.Context(), &query)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}

func (c *socialControllerImpl) ListFollowings(ctx *gin.Context) {
	currentUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var query cqe.FollowListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	if query.TargetUUID == "" && query.TargetUserUUID != "" {
		query.TargetUUID = query.TargetUserUUID
	}
	query.Normalize(currentUUID)
	resp, err := c.socialApp.ListFollowings(ctx.Request.Context(), &query)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}

func (c *socialControllerImpl) GetUserRelation(ctx *gin.Context) {
	targetUserUUID := ctx.Param("user_uuid")
	if targetUserUUID == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "user_uuid"))
		return
	}

	currentUUID := ""
	if uuid, ok := middleware.GetCurrentUserUUID(ctx); ok && uuid != "" {
		currentUUID = uuid
	} else {
		if h := ctx.GetHeader("X-User-UUID"); h != "" {
			currentUUID = h
		}
	}

	stat, err := c.socialApp.GetUserRelationStat(ctx.Request.Context(), targetUserUUID, currentUUID)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	restapi.Success(ctx, stat)
}
