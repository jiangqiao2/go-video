package main

import (
	"notification-service/app"
	"notification-service/pkg/observability"
)

func main() {
	observability.StartProfiling("notification-service")
	app.Run()
}
