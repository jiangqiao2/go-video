package main

import (
	"video-service/app"
	"video-service/pkg/observability"
)

func main() {
	observability.StartProfiling("video-service")
	app.Run()
}
