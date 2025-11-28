package http

import (
	"context"
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
	FollowStatus(ctx *gin.Context)
	ListFollowers(ctx *gin.Context)
	ListFollowings(ctx *gin.Context)
}

type socialControllerImpl struct {
	manager.Controller
	socialApp app.SocialApp
}

func (c *socialControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {}

func (c *socialControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/inner/social")
	{
		v1.POST("/follow", middleware.AuthRequired(), c.Follow)
		v1.POST("/unfollow", middleware.AuthRequired(), c.Unfollow)
		v1.GET("/follow/status", middleware.AuthRequired(), c.FollowStatus)
		v1.GET("/followers", middleware.AuthRequired(), c.ListFollowers)
		v1.GET("/followings", middleware.AuthRequired(), c.ListFollowings)
	}
	// 兼容别名：relation 路径
	rel := router.Group("user/v1/inner/relation")
	{
		rel.POST("/follow", middleware.AuthRequired(), c.Follow)
		rel.POST("/unfollow", middleware.AuthRequired(), c.Unfollow)
		rel.GET("/follow/status", middleware.AuthRequired(), c.FollowStatus)
		rel.GET("/followers", middleware.AuthRequired(), c.ListFollowers)
		rel.GET("/followings", middleware.AuthRequired(), c.ListFollowings)
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
	req.UserUUID = userUUID
	if err := c.socialApp.Follow(context.Background(), &req); err != nil {
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
	req.UserUUID = userUUID
	if err := c.socialApp.Unfollow(context.Background(), &req); err != nil {
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
	status, err := c.socialApp.FollowStatus(context.Background(), &req)
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
	resp, err := c.socialApp.ListFollowers(context.Background(), &query)
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
	resp, err := c.socialApp.ListFollowings(context.Background(), &query)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}
