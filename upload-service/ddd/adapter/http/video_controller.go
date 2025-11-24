package http

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"

	"upload-service/ddd/application/app"
	videoCqe "upload-service/ddd/application/cqe"
	"upload-service/pkg/assert"
	"upload-service/pkg/errno"
	"upload-service/pkg/manager"
	"upload-service/pkg/restapi"
)

var (
	videoControllerOnce      sync.Once
	singletonVideoController VideoController
)

type VideoControllerPlugin struct{}

func (p *VideoControllerPlugin) Name() string {
	return "videoControllerPlugin"
}

func (p *VideoControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	videoControllerOnce.Do(func() {
		singletonVideoController = &videoControllerImpl{
			videoApp: app.DefaultVideoApp(),
			tagApp:   app.DefaultTagApp(),
		}
	})
	assert.NotNil(singletonVideoController)
	return singletonVideoController
}

type VideoController interface {
	manager.Controller
	PublishVideo(ctx *gin.Context)
	ListVideos(ctx *gin.Context)
	ListTags(ctx *gin.Context)
}

type videoControllerImpl struct {
	manager.Controller
	videoApp app.VideoApp
	tagApp   app.TagApp
}

func (c *videoControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
    v1 := router.Group("video/v1/open")
    {
        v1.GET("/tags", c.ListTags)
        v1.GET("/videos", c.ListOpenVideos)
    }
}

func (c *videoControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
    v1 := router.Group("video/v1/inner/videos")
    {
        v1.POST("", c.PublishVideo)
        v1.GET("", c.ListVideos)
    }
}

func (c *videoControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {}

func (c *videoControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {}

func (c *videoControllerImpl) extractUserInfo(ctx *gin.Context) (string, error) {
	userUUID := ctx.GetHeader("X-User-UUID")
	if userUUID == "" {
		return "", errno.ErrUnauthorized
	}
	return userUUID, nil
}

func (c *videoControllerImpl) PublishVideo(ctx *gin.Context) {
	userUUID, err := c.extractUserInfo(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

	var req videoCqe.PublishVideoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}
	req.UserUUID = userUUID

	resp, err := c.videoApp.PublishVideo(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}

func (c *videoControllerImpl) ListVideos(ctx *gin.Context) {
    userUUID, err := c.extractUserInfo(ctx)
    if err != nil {
        restapi.Failed(ctx, err)
        return
    }

	var req videoCqe.ListVideosReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	req.UserUUID = userUUID

	resp, err := c.videoApp.ListUserVideos(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}

    restapi.Success(ctx, resp)
}

func (c *videoControllerImpl) ListOpenVideos(ctx *gin.Context) {
    var req videoCqe.ListOpenVideosReq
    if err := ctx.ShouldBindQuery(&req); err != nil {
        restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
        return
    }
    resp, err := c.videoApp.ListOpenVideos(context.Background(), &req)
    if err != nil {
        restapi.Failed(ctx, err)
        return
    }
    restapi.Success(ctx, resp)
}

func (c *videoControllerImpl) ListTags(ctx *gin.Context) {
	var req videoCqe.ListTagsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "query"))
		return
	}
	resp, err := c.tagApp.ListTags(context.Background(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, resp)
}
