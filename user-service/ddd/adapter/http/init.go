package http

import (
	"user-service/pkg/manager"
)

// init 包初始化函数，自动注册所有控制器插件
func init() {
	// 注册用户控制器插件
	manager.RegisterControllerPlugin(&UserControllerPlugin{})
}
