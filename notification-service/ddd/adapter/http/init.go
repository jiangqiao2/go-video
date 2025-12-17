package http

import "notification-service/pkg/manager"

func init() {
	manager.RegisterControllerPlugin(&NotificationControllerPlugin{})
}
