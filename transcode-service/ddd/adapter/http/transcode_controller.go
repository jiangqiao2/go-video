package http

import (
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
		singletonTranscodeController = transcodeControllerImpl{}
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

func (t *transcodeControllerImpl) CreateTranscodeTask(ctx *gin.Context) {
	var req cqe.TranscodeTaskCqe
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}

}
