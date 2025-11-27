package http

import (
	"net/http"
	"strconv"
	"sync"

	apppkg "video-service/ddd/application/app"
	cqe "video-service/ddd/application/cqe"
	"video-service/pkg/assert"
	"video-service/pkg/errno"
	"video-service/pkg/manager"
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
		instance = &videoControllerImpl{videoApp: apppkg.NewVideoApp()}
	})
	return instance
}

type videoControllerImpl struct {
	manager.Controller
	videoApp apppkg.VideoApp
}

func (c *videoControllerImpl) RegisterOpenApi(group *gin.RouterGroup) {
	v1 := group.Group("video/v1/open")
	v1.GET("/ping", func(ctx *gin.Context) { restapi.Success(ctx, gin.H{"message": "pong"}) })
	v1.POST("/publish", c.Publish)
	v1.GET("/get/:videoUUID", c.Get)
	v1.GET("/list", c.List)
	v1.POST("/like", c.Like)
	v1.POST("/unlike", c.Unlike)
	v1.POST("/play/:videoUUID", c.Play)
	v1.POST("/comment", c.AddComment)
	v1.GET("/comments/:videoUUID", c.ListComments)
	v1.POST("/fullscreen", c.ToggleFullscreen)
}

func (c *videoControllerImpl) RegisterInnerApi(group *gin.RouterGroup) {
	v1 := group.Group("video/v1/inner")
	v1.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "service": "video-service"})
	})
}

func (c *videoControllerImpl) RegisterDebugApi(group *gin.RouterGroup) {}
func (c *videoControllerImpl) RegisterOpsApi(group *gin.RouterGroup)   {}

func (c *videoControllerImpl) Publish(ctx *gin.Context) {
	var req cqe.PublishVideoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	res, err := c.videoApp.Publish(&req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) Get(ctx *gin.Context) {
	videoUUID := ctx.Param("videoUUID")
	res, err := c.videoApp.Get(videoUUID)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) List(ctx *gin.Context) {
	page := 1
	size := 20
	if v := ctx.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := ctx.Query("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 && s <= 200 {
			size = s
		}
	}
	res, err := c.videoApp.List(page, size)
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
	userUUID := ctx.Query("user_uuid")
	if userUUID == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "missing user_uuid"))
		return
	}
	if err := c.videoApp.Like(userUUID, req.VideoUUID); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, gin.H{"liked": true})
}

func (c *videoControllerImpl) Unlike(ctx *gin.Context) {
	var req cqe.LikeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	userUUID := ctx.Query("user_uuid")
	if userUUID == "" {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "missing user_uuid"))
		return
	}
	if err := c.videoApp.Unlike(userUUID, req.VideoUUID); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, gin.H{"liked": false})
}

func (c *videoControllerImpl) Play(ctx *gin.Context) {
	videoUUID := ctx.Param("videoUUID")
	if err := c.videoApp.Play(videoUUID); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, gin.H{"video_uuid": videoUUID, "autoplay": true, "muted": false})
}

func (c *videoControllerImpl) AddComment(ctx *gin.Context) {
	var req cqe.CommentCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "invalid body"))
		return
	}
	res, err := c.videoApp.AddComment(&req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) ListComments(ctx *gin.Context) {
	videoUUID := ctx.Param("videoUUID")
	page := 1
	size := 20
	if v := ctx.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := ctx.Query("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 && s <= 200 {
			size = s
		}
	}
	res, err := c.videoApp.ListComments(videoUUID, page, size)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}

func (c *videoControllerImpl) ToggleFullscreen(ctx *gin.Context) {
	restapi.Success(ctx, gin.H{"fullscreen": true})
}
