package main

import (
	"transcode-service/app"
	"transcode-service/pkg/observability"
)

func main() {
	observability.StartProfiling("transcode-service")
	app.Run()
}
