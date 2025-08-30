package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"upload-service/ddd/application/app"

	"upload-service/pkg/restapi"

	"sync"
	uploadCqe "upload-service/ddd/application/cqe"
	"upload-service/pkg/assert"
	"upload-service/pkg/manager"
)

var (
	uploadVideoControllerOnce      sync.Once
	singletonUploadVideoController UploadVideoController
)

type UploadVideoControllerPlugin struct {
}

func (p *UploadVideoControllerPlugin) Name() string {
	return "uploadVideoControllerPlugin"
}
func (p *UploadVideoControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	uploadVideoControllerOnce.Do(func() {
		singletonUploadVideoController = &uploadVideoControllerImpl{
			uploadVideoApp: app.DefaultUploadVideoApp(),
		}
	})
	assert.NotNil(singletonUploadVideoController)
	return singletonUploadVideoController
}

type UploadVideoController interface {
	manager.Controller
	Init(ctx *gin.Context)
}

type uploadVideoControllerImpl struct {
	manager.Controller
	uploadVideoApp app.UploadVideoApp
}

// RegisterOpenApi 注册开放API
func (c *uploadVideoControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {

}

// RegisterInnerApi 注册内部API
func (c *uploadVideoControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	// 内部API实现
	v1 := router.Group("api/v1/upload")
	{
		v1.POST("/init", c.Init)
	}
}

// RegisterDebugApi 注册调试API
func (c *uploadVideoControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	// 调试API实现
}

// RegisterOpsApi 注册运维API
func (c *uploadVideoControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

func (c *uploadVideoControllerImpl) Init(ctx *gin.Context) {
	var cqe uploadCqe.UploadVideoInitReq
	if err := ctx.ShouldBindJSON(&cqe); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	result, err := c.uploadVideoApp.UploadVideoInit(context.Background(), &cqe)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, result)
}

func (c *uploadVideoControllerImpl) UploadVideoChunk(ctx *gin.Context) {
	var cqe uploadCqe.UploadChunkReq
	if err := ctx.ShouldBindJSON(&cqe); err != nil {
		restapi.Failed(ctx, err)
		return
	}

}
