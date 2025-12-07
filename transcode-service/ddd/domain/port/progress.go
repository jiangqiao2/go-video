package port

import (
	"context"

	"transcode-service/ddd/domain/entity"
)

// ProgressSink persists or forwards task progress updates.
type ProgressSink interface {
	SaveProgress(ctx context.Context, task *entity.TranscodeTaskEntity, progress int) error
}
