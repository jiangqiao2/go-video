package http

import "transcode-service/pkg/manager"

func init() {
	manager.RegisterControllerPlugin(&TranscodeControllerPlugin{})
}
