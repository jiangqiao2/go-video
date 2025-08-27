package http

import "sync"

var (
	videoControllerOnce            sync.Once
	singletonUploadVideoController UploadVideoController
)

type UploadVideoController interface {
}

type uploadVideoController struct {
}
