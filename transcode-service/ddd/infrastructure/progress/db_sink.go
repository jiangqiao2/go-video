package progress

import (
	"context"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/port"
	"transcode-service/ddd/domain/repo"
)

// DBSink writes progress to the repository.
type DBSink struct {
	repo repo.TranscodeJobRepository
}

func NewDBSink(r repo.TranscodeJobRepository) port.ProgressSink {
	return &DBSink{repo: r}
}

func (s *DBSink) SaveProgress(ctx context.Context, task *entity.TranscodeTaskEntity, progress int) error {
	if s.repo == nil || task == nil {
		return nil
	}
	return s.repo.UpdateTranscodeJobProgress(ctx, task.TaskUUID(), progress)
}
