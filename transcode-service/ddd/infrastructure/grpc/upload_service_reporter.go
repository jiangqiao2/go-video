package grpcclient

import (
	"context"
	"fmt"
	"sync"

	uploadpb "go-vedio-1/proto/upload"

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
		return fmt.Errorf("upload service client is not initialised")
	}
	req := &uploadpb.UpdateTranscodeStatusRequest{
		VideoUuid:         videoUUID,
		TranscodeTaskUuid: taskUUID,
		Status:            uploadStatusPublished,
		VideoUrl:          outputPath,
	}
	resp, err := r.client.UpdateTranscodeStatus(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil || !resp.GetSuccess() {
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
	req := &uploadpb.UpdateTranscodeStatusRequest{
		VideoUuid:         videoUUID,
		TranscodeTaskUuid: taskUUID,
		Status:            uploadStatusFailed,
		ErrorMessage:      errorMessage,
	}
	resp, err := r.client.UpdateTranscodeStatus(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil || !resp.GetSuccess() {
		return fmt.Errorf("upload-service returned failure: %s", resp.GetMessage())
	}
	return nil
}
