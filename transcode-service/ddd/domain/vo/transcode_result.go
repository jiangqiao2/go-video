package vo

import (
	"context"

	"transcode-service/ddd/domain/gateway"
)

// TranscodeResult 封装通知上传服务的公共字段。
type TranscodeResult struct {
	VideoUUID string
	TaskUUID  string
}

// NewTranscodeResult 构造转码结果上下文。
func NewTranscodeResult(taskUUID, videoUUID string) TranscodeResult {
	return TranscodeResult{
		TaskUUID:  taskUUID,
		VideoUUID: videoUUID,
	}
}

// ReportSuccess 统一封装成功上报。
func (r TranscodeResult) ReportSuccess(ctx context.Context, reporter gateway.TranscodeResultReporter, publicURL string) error {
	if reporter == nil {
		return nil
	}
	return reporter.ReportSuccess(ctx, r.VideoUUID, r.TaskUUID, publicURL)
}

// ReportFailure 统一封装失败上报。
func (r TranscodeResult) ReportFailure(ctx context.Context, reporter gateway.TranscodeResultReporter, errMsg string) error {
	if reporter == nil {
		return nil
	}
	return reporter.ReportFailure(ctx, r.VideoUUID, r.TaskUUID, errMsg)
}
