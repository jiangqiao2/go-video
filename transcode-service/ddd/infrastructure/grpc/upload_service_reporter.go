package grpc

import (
	"context"
	"fmt"
	"sync"
	"transcode-service/pkg/logger"

	"transcode-service/ddd/domain/gateway"
)

const (
	uploadStatusPublished = "Published"
	uploadStatusFailed    = "Failed"
)

type uploadServiceReporter struct {
	client *UploadServiceClient
}

var (
	reporterOnce      sync.Once
	singletonReporter gateway.TranscodeResultReporter
)

// DefaultUploadServiceReporter returns a singleton reporter using the default client.
func DefaultUploadServiceReporter() gateway.TranscodeResultReporter {
	reporterOnce.Do(func() {
		singletonReporter = NewUploadServiceReporter(DefaultUploadServiceClient())
	})
	return singletonReporter
}

// NewUploadServiceReporter builds a reporter with the provided client.
func NewUploadServiceReporter(client *UploadServiceClient) gateway.TranscodeResultReporter {
	return &uploadServiceReporter{client: client}
}

func (r *uploadServiceReporter) ReportSuccess(ctx context.Context, videoUUID, taskUUID, videoURL string) error {
	if r.client == nil {
		logger.Infof("ReportSuccess r.client is nil")
		return fmt.Errorf("upload service client is not initialised")
	}

	// 转码成功：将最终可播放的 URL（通常是 HLS master.m3u8）回写给 upload-service
	resp, err := r.client.UpdateTranscodeStatus(ctx, videoUUID, taskUUID, uploadStatusPublished, videoURL, "")
	if err != nil {
		logger.Errorf("ReportSuccess failed video_uuid=%s task_uuid=%s error=%v", videoUUID, taskUUID, err)
		return err
	}
	if resp == nil || !resp.GetSuccess() {
		logger.Errorf("ReportSuccess resp.success is false message=%s", resp.GetMessage())
		return fmt.Errorf("upload-service returned failure: %s", resp.GetMessage())
	}
	return nil
}

func (r *uploadServiceReporter) ReportFailure(ctx context.Context, videoUUID, taskUUID, errorMessage string) error {
	if r.client == nil {
		return fmt.Errorf("upload service client is not initialised")
	}
	if errorMessage == "" {
		errorMessage = "transcode failed"
	}

	resp, err := r.client.UpdateTranscodeStatus(ctx, videoUUID, taskUUID, uploadStatusFailed, "", errorMessage)
	if err != nil {
		logger.Errorf("ReportFailure failed video_uuid=%s task_uuid=%s error_message=%s error=%v", videoUUID, taskUUID, errorMessage, err)
		return err
	}
	if resp == nil || !resp.GetSuccess() {
		logger.Errorf("ReportFailure resp.success is false message=%s", resp.GetMessage())
		return fmt.Errorf("upload-service returned failure: %s", resp.GetMessage())
	}
	return nil
}
