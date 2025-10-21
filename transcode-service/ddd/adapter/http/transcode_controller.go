package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync"
	"transcode-service/ddd/application/app"
	"transcode-service/ddd/application/cqe"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/restapi"
)

var (
	transcodeControllerOnce      sync.Once
	singletonTranscodeController TranscodeController
)

type TranscodeControllerPlugin struct {
}

func (p *TranscodeControllerPlugin) Name() string {
	return "transcodeControllerPlugin"
}
func (p *TranscodeControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	transcodeControllerOnce.Do(func() {
		singletonTranscodeController = &transcodeControllerImpl{
			transcodeApp: app.DefaultTranscodeApp(),
		}
	})
	assert.NotNil(singletonTranscodeController)
	return singletonTranscodeController
}

type TranscodeController interface {
	manager.Controller
}

type transcodeControllerImpl struct {
	manager.Controller
	transcodeApp app.TranscodeApp
}

// RegisterOpenApi 注册开放API
func (t *transcodeControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	// 开放API实现
	v1 := router.Group("v1/transcode")
	{
		v1.POST("/tasks", t.CreateTranscodeTask)
	}
}

// RegisterInnerApi 注册内部API
func (t *transcodeControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {

}

// RegisterDebugApi 注册调试API
func (t *transcodeControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {
	// 调试API实现
}

// RegisterOpsApi 注册运维API
func (t *transcodeControllerImpl) RegisterOpsApi(router *gin.RouterGroup) {
	// 运维API实现
}

func (t *transcodeControllerImpl) CreateTranscodeTask(c *gin.Context) {
	var req cqe.CreateTranscodeTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		restapi.Failed(c, err)
		return
	}
	res, err := t.transcodeApp.CreateTranscodeTask(context.Background(), &req)
	if err != nil {
		restapi.Failed(c, err)
		return
	}
	restapi.Success(c, res)
}
