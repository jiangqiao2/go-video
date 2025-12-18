package main

import (
	"upload-service/app"
	"upload-service/pkg/observability"
)

func main() {
	observability.StartProfiling("upload-service")
	app.Run()
}
