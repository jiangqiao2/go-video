package worker

import "transcode-service/pkg/manager"

func init() {
	manager.RegisterComponentPlugin(&TranscodeWorkerComponentPlugin{})
}
