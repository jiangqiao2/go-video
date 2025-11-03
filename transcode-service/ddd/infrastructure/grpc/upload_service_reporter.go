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

func (r *uploadServiceReporter) ReportSuccess(ctx context.Context, videoUUID, taskUUID, outputPath string) error {
	if r.client == nil {
		logger.Info("ReportSuccess r.client is nil")
		return fmt.Errorf("upload service client is not initialised")
	}
	
	resp, err := r.client.UpdateTranscodeStatus(ctx, videoUUID, taskUUID, uploadStatusPublished, outputPath, "")
	if err != nil {
		logger.Error("ReportSuccess failed", map[string]interface{}{
			"video_uuid": videoUUID,
			"task_uuid":  taskUUID,
			"error":      err.Error(),
		})
		return err
	}
	if resp == nil || !resp.GetSuccess() {
		logger.Error("ReportSuccess resp.success is false", map[string]interface{}{
			"message": resp.GetMessage(),
		})
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
		logger.Error("ReportFailure failed", map[string]interface{}{
			"video_uuid":    videoUUID,
			"task_uuid":     taskUUID,
			"error_message": errorMessage,
			"error":         err.Error(),
		})
		return err
	}
	if resp == nil || !resp.GetSuccess() {
		logger.Error("ReportFailure resp.success is false", map[string]interface{}{
			"message": resp.GetMessage(),
		})
		return fmt.Errorf("upload-service returned failure: %s", resp.GetMessage())
	}
	return nil
}
