package main

import (
	"go-video/ddd/app"

	// 导入模块以触发插件注册
	_ "go-video/ddd/user"
	_ "go-video/ddd/video"
)

func main() {
	app.Run()
}
