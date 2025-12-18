package main

import (
	"user-service/app"
	"user-service/pkg/observability"
)

func main() {
	observability.StartProfiling("user-service")
	app.Run()
}
