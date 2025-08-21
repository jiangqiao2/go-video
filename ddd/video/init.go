package video

import (
	"sync"

	"go-video/ddd/video/adapter/http"
	"go-video/pkg/assert"
	"go-video/pkg/manager"
)

var (
	videoPluginOnce      sync.Once
	singletonVideoPlugin *VideoPlugin
)

// VideoPlugin 视频插件
type VideoPlugin struct {
	videoController http.VideoController
}

// DefaultVideoPlugin 获取视频插件单例
func DefaultVideoPlugin() *VideoPlugin {
	assert.NotCircular()
	videoPluginOnce.Do(func() {
		videoController := http.DefaultVideoController()
		singletonVideoPlugin = &VideoPlugin{
			videoController: videoController,
		}
	})
	assert.NotNil(singletonVideoPlugin)
	return singletonVideoPlugin
}

// init 包初始化函数，注册视频控制器插件
func init() {
	// 注册视频控制器插件到管理器
	manager.RegisterControllerPlugin(&http.VideoControllerPlugin{})
}
