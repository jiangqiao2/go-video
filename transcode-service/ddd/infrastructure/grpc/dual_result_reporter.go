package grpc

import (
	"context"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/pkg/logger"
)

type dualResultReporter struct {
	upload *UploadServiceClient
	video  *VideoServiceClient
}

func DefaultDualResultReporter() gateway.TranscodeResultReporter {
	return &dualResultReporter{
		upload: DefaultUploadServiceClient(),
		video:  DefaultVideoServiceClient(),
	}
}

func (r *dualResultReporter) ReportSuccess(ctx context.Context, videoUUID, taskUUID, videoURL string) error {
	if r.upload != nil {
		// 将最终可播放的 URL 传递给 upload-service
		_, _ = r.upload.UpdateTranscodeStatus(ctx, videoUUID, taskUUID, "Published", videoURL, "")
	}
	if r.video != nil {
		_, _ = r.video.UpdateTranscodeResult(ctx, videoUUID, taskUUID, "Published", videoURL, "", 0, 0)
	}
	logger.WithContext(ctx).Infof("transcode result success reported video_uuid=%s task_uuid=%s", videoUUID, taskUUID)
	return nil
}

func (r *dualResultReporter) ReportFailure(ctx context.Context, videoUUID, taskUUID, errorMessage string) error {
	if r.upload != nil {
		_, _ = r.upload.UpdateTranscodeStatus(ctx, videoUUID, taskUUID, "Failed", "", errorMessage)
	}
	if r.video != nil {
		_, _ = r.video.UpdateTranscodeResult(ctx, videoUUID, taskUUID, "Failed", "", errorMessage, 0, 0)
	}
	logger.WithContext(ctx).Warnf("transcode result failure reported video_uuid=%s task_uuid=%s error=%s", videoUUID, taskUUID, errorMessage)
	return nil
}
