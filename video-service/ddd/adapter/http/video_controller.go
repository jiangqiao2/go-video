package http

import (
	"context"
	"net/http"
	"sync"

	apppkg "video-service/ddd/application/app"
	cqe "video-service/ddd/application/cqe"
	dto "video-service/ddd/application/dto"
	"video-service/pkg/assert"
	"video-service/pkg/errno"
	"video-service/pkg/manager"
	middleware "video-service/pkg/middleware"
	"video-service/pkg/restapi"

	"github.com/gin-gonic/gin"
)

type VideoControllerPlugin struct{}

func (p *VideoControllerPlugin) Name() string { return "videoControllerPlugin" }

var (
	onceCtl  sync.Once
	instance manager.Controller
)

func (p *VideoControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	onceCtl.Do(func() {
		instance = &videoControllerImpl{videoApp: apppkg.DefaultVideoApp()}
	})
	assert.NotNil(instance)
	return instance
}

type videoControllerImpl struct {
	manager.Controller
	videoApp apppkg.VideoApp
}

func (c *videoControllerImpl) extractUserInfo(ctx *gin.Context) (string, error) {
	if uuid, ok := middleware.GetCurrentUserUUID(ctx); ok && uuid != "" {
		return uuid, nil
	}
	if v := ctx.GetHeader("X-User-UUID"); v != "" {
		return v, nil
	}
	return "", errno.ErrUnauthorized
}

func (c *videoControllerImpl) RegisterOpenApi(group *gin.RouterGroup) {
	v1 := group.Group("video/v1/open")
	v1.GET("/ping", func(ctx *gin.Context) { restapi.Success(ctx, gin.H{"message": "pong"}) })
	v1.GET("/get/:videoUUID", c.Get)
	v1.GET("/list", c.List)
	v1.POST("/play/:videoUUID", c.Play)
	v1.GET("/comments/:videoUUID", c.ListComments)
	v1.POST("/fullscreen", c.ToggleFullscreen)
}

func (c *videoControllerImpl) RegisterInnerApi(group *gin.RouterGroup) {
	v1 := group.Group("video/v1/inner")
	v1.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "service": "video-service"})
	})
	v1.POST("/publish", middleware.AuthRequired(), c.Publish)
	v1.POST("/like", middleware.AuthRequired(), c.Like)
	v1.POST("/unlike", middleware.AuthRequired(), c.Unlike)
	v1.POST("/comment", middleware.AuthRequired(), c.AddComment)
	v1.POST("/precreate", middleware.AuthRequired(), c.Precreate)
	v1.POST("/transcode/update", c.UpdateTranscodeResult)
}

func (c *videoControllerImpl) RegisterDebugApi(group *gin.RouterGroup) {}
func (c *videoControllerImpl) RegisterOpsApi(group *gin.RouterGroup)   {}

func (c *videoControllerImpl) Publish(ctx *gin.Context) {
	var req cqe.PublishVideoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	if req.UserUUID == "" {
		uuid, err := c.extractUserInfo(ctx)
		if err != nil {
			restapi.Failed(ctx, err)
			return
		}
		req.UserUUID = uuid
	}
	res, err := c.videoApp.Publish(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Precreate(ctx *gin.Context) {
	var req cqe.PrecreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	if req.UserUUID == "" {
		uuid, err := c.extractUserInfo(ctx)
		if err != nil {
			restapi.Failed(ctx, err)
			return
		}
		req.UserUUID = uuid
	}
	res, err := c.videoApp.Precreate(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) UpdateTranscodeResult(ctx *gin.Context) {
	var req cqe.UpdateTranscodeResultReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	res, err := c.videoApp.UpdateTranscodeResult(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Get(ctx *gin.Context) {
	req := cqe.GetVideoReq{VideoUUID: ctx.Param("videoUUID")}
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.Get(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) List(ctx *gin.Context) {
	var req cqe.ListVideosReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid query"))
		return
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.List(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Like(ctx *gin.Context) {
	var req cqe.LikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	uuid, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	req.UserUUID = uuid
	req.Normalize()
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.Like(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Unlike(ctx *gin.Context) {
	var req cqe.LikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	uuid, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	req.UserUUID = uuid
	req.Normalize()
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.Unlike(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Play(ctx *gin.Context) {
	req := cqe.PlayReq{VideoUUID: ctx.Param("videoUUID")}
	req.Normalize()
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.Play(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) AddComment(ctx *gin.Context) {
	var req cqe.CommentCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	if req.UserUUID == "" {
		uuid, err := c.extractUserInfo(ctx)
		if err != nil {
			restapi.Failed(ctx, err)
			return
		}
		req.UserUUID = uuid
	}
	res, err := c.videoApp.AddComment(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) ListComments(ctx *gin.Context) {
	var req cqe.ListCommentsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid query"))
		return
	}
	req.VideoUUID = ctx.Param("videoUUID")
	req.Normalize()
	if err := req.Validate(); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	res, err := c.videoApp.ListComments(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) ToggleFullscreen(ctx *gin.Context) {
	restapi.Success(ctx, &dto.FullscreenDto{Fullscreen: true})
}
