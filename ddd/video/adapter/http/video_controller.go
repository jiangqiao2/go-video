package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-video/ddd/video/application/app"
	"go-video/ddd/video/application/cqe"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
	"go-video/pkg/restapi"
	"sync"
)

var (
	singletonVideoController VideoController
	onceVideoController      sync.Once
)

type VideoController interface {
	UploadVideo(ctx *gin.Context)
}

type videoControllerImpl struct {
	videoApp app.VideoApp
}

func DefaultVideoController() VideoController {
	assert.NotCircular()
	onceVideoController.Do(func() {
		singletonVideoController = &videoControllerImpl{
			videoApp: app.DefaultVideoApp(),
		}
	})
	assert.NotNil(singletonVideoController)
	return singletonVideoController
}

func (v *videoControllerImpl) UploadVideo(c *gin.Context) {
	var (
		cmd cqe.UploadVideoCommand
	)
	err := c.ShouldBindJSON(&cmd)
	if err != nil {
		restapi.Failed(c, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "body"))
		return
	}
	err = v.videoApp.Create(context.Background(), &cmd)
	if err != nil {
		restapi.Failed(c, err)
		return
	}
	restapi.Success(c, nil)
}
