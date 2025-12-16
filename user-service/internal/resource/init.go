package resource

import (
	"user-service/pkg/manager"
)

// init 包初始化函数，自动注册所有资源插件
func init() {
	// 注册MySQL资源插件
	manager.RegisterResourcePlugin(&MySqlResourcePlugin{})

	// 注册Logger资源插件
	manager.RegisterResourcePlugin(&LoggerResourcePlugin{})

	// 注册Redis资源插件
	manager.RegisterResourcePlugin(&RedisResourcePlugin{})

	// 注册Kafka资源插件（用于关注事件异步化等场景）
	manager.RegisterResourcePlugin(&KafkaResourcePlugin{})
}
