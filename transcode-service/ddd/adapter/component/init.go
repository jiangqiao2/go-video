package component

import "transcode-service/pkg/manager"

func init() {
	manager.RegisterComponentPlugin(&TranscodeTaskConsumerPlugin{})
}
