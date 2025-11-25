package resource

import "transcode-service/pkg/manager"

func init() {
	// 注册资源插件
	manager.RegisterResourcePlugin(&MySqlResourcePlugin{})
	manager.RegisterResourcePlugin(&RedisResourcePlugin{})
	manager.RegisterResourcePlugin(&RustFSResourcePlugin{})
}
