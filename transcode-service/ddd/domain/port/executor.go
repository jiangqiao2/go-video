package port

import (
	"context"

	"transcode-service/ddd/domain/entity"
)

// ProgressCallback is invoked by executors to report percentage progress (0-100).
type ProgressCallback func(progress int)

// TranscodeExecutor executes a full transcode job (typically MP4 output) and returns
// the object key and public URL of the generated asset. Implementations may choose
// to skip uploading based on the provided options.
type TranscodeExecutor interface {
	Execute(ctx context.Context, task *entity.TranscodeTaskEntity, opts TranscodeOptions) (objectKey string, publicURL string, err error)
}

// HLSExecutor performs HLS slicing for a job and returns the master playlist public URL.
type HLSExecutor interface {
	Slice(ctx context.Context, job *entity.HLSJobEntity, opts HLSOptions) (masterURL string, err error)
}

// TranscodeOptions controls executor behaviour.
type TranscodeOptions struct {
	SkipUpload  bool
	ProgressCb  ProgressCallback
	RequestID   string
	TraceID     string
	TempDir     string
	TimeoutSecs int
}

// HLSOptions controls HLS slicing behaviour.
type HLSOptions struct {
	ProgressCb  ProgressCallback
	RequestID   string
	TraceID     string
	TempDir     string
	TimeoutSecs int
}
