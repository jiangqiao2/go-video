package resource

import "transcode-service/pkg/manager"

func init() {
	// 注册MySQL资源插件
	manager.RegisterResourcePlugin(&MySqlResourcePlugin{})
}
