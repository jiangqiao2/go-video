package http

import "video-service/pkg/manager"

func init() {
	manager.RegisterControllerPlugin(&VideoControllerPlugin{})
}
