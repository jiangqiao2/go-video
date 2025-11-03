package gateway

import "context"

// TranscodeResultReporter notifies downstream services about task outcomes.
type TranscodeResultReporter interface {
	ReportSuccess(ctx context.Context, videoUUID, taskUUID, outputPath string) error
	ReportFailure(ctx context.Context, videoUUID, taskUUID, errorMessage string) error
}
