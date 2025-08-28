package user

import (
	"go-video/ddd/user/adapter/http"
	"go-video/pkg/manager"
)

// init 包初始化函数，注册用户控制器插件
func init() {
	// 注册用户控制器插件到管理器
	manager.RegisterControllerPlugin(&http.UserControllerPlugin{})
}
