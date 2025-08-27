package video

import (
	"go-video/ddd/video/adapter/http"
	"go-video/pkg/manager"
)

// init 包初始化函数，注册视频控制器插件
func init() {
	// 注册视频控制器插件到管理器
	manager.RegisterControllerPlugin(&http.VideoControllerPlugin{})
}
