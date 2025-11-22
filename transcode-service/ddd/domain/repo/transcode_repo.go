package repo

import (
    "context"

    "transcode-service/ddd/domain/entity"
    "transcode-service/ddd/domain/vo"
)

type TranscodeJobRepository interface {
    CreateTranscodeJob(ctx context.Context, job *entity.TranscodeTaskEntity) error
    UpdateTranscodeJobProgress(ctx context.Context, jobUUID string, progress int) error
    UpdateTranscodeJob(ctx context.Context, job *entity.TranscodeTaskEntity) error
    GetTranscodeJob(ctx context.Context, jobUUID string) (*entity.TranscodeTaskEntity, error)
    UpdateTranscodeJobStatus(ctx context.Context, jobUUID string, status vo.TaskStatus, message, outputPath string, progress int) error
    QueryTranscodeJobsByStatus(ctx context.Context, status vo.TaskStatus, limit int) ([]*entity.TranscodeTaskEntity, error)
}

type HLSJobRepository interface {
    CreateHLSJob(ctx context.Context, job *entity.HLSJobEntity) error
    UpdateHLSJobProgress(ctx context.Context, jobUUID string, progress int) error
    UpdateHLSJobStatus(ctx context.Context, jobUUID string, status string) error
    UpdateHLSJobOutput(ctx context.Context, jobUUID string, masterPlaylist string) error
    UpdateHLSJobError(ctx context.Context, jobUUID string, errorMessage string) error
    GetHLSJob(ctx context.Context, jobUUID string) (*entity.HLSJobEntity, error)
    QueryHLSJobsByStatus(ctx context.Context, status string, limit int) ([]*entity.HLSJobEntity, error)
}
