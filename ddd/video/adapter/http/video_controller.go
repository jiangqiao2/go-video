package http

import (
	"context"
	"go-video/pkg/logger"
	"sync"

	"go-video/ddd/video/application/app"
	"go-video/ddd/video/application/cqe"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
	"go-video/pkg/manager"
	"go-video/pkg/middleware"
	"go-video/pkg/restapi"

	"github.com/gin-gonic/gin"
)

var (
	videoControllerOnce      sync.Once
	singletonVideoController VideoController
)

type VideoControllerPlugin struct {
}

func (p *VideoControllerPlugin) Name() string {
	return "videoControllerPlugin"
}

func (p *VideoControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	videoControllerOnce.Do(func() {
		videoApp := app.DefaultVideoApp()
		singletonVideoController = &videoControllerImpl{
			Controller: nil, // 这里可以设置为nil，因为我们实现了manager.Controller接口
			videoApp:   videoApp,
		}
	})
	assert.NotNil(singletonVideoController)
	return singletonVideoController
}

type VideoController interface {
	manager.Controller
}

type videoControllerImpl struct {
	manager.Controller
	videoApp app.VideoApp
}

func DefaultVideoController() VideoController {
	assert.NotCircular()
	videoControllerOnce.Do(func() {
		videoApp := app.DefaultVideoApp()
		singletonVideoController = &videoControllerImpl{
			videoApp: videoApp,
		}
	})
	assert.NotNil(singletonVideoController)
	return singletonVideoController
}

// RegisterOpenApi 注册开放API
func (c *videoControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		// 视频上传需要认证
		v1.POST("/videos/upload", middleware.AuthRequired(), c.UploadVideo)
		// 视频查看可以不需要认证（公开访问）
		v1.GET("/videos/:id", c.GetVideo)
		v1.GET("/videos", c.GetVideoList)
	}
}

// RegisterInnerApi 注册内部API
func (c *videoControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	// 内部API实现
}

// RegisterDebugApi 注册调试API
func (c *videoControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	// 调试API实现
}

// RegisterOpsApi 注册运维API
func (c *videoControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

// UploadVideo 上传视频
func (c *videoControllerImpl) UploadVideo(ctx *gin.Context) {
	logger.Info("upload video")
	// 处理multipart/form-data请求【
	var cmd cqe.UploadVideoCommand

	// 从认证中间件获取用户UUID（必须存在）
	cmd.UserUUID = middleware.MustGetCurrentUserUUID(ctx)

	// 获取表单字段
	cmd.Title = ctx.PostForm("title")
	cmd.Description = ctx.PostForm("description")
	cmd.Format = ctx.PostForm("format")

	// 获取文件
	file, err := ctx.FormFile("file")
	if err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "file"))
		return
	}
	cmd.File = file
	cmd.FileSize = file.Size
	result, err := c.videoApp.Create(context.Background(), &cmd)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)

}

// GetVideo 获取视频
func (c *videoControllerImpl) GetVideo(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "get video endpoint"})
}

// GetVideoList 获取视频列表
func (c *videoControllerImpl) GetVideoList(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "get video list endpoint"})
}
