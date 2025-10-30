package http

import "upload-service/pkg/manager"

func init() {
	manager.RegisterControllerPlugin(&UploadVideoControllerPlugin{})
	manager.RegisterControllerPlugin(&VideoControllerPlugin{})
}
