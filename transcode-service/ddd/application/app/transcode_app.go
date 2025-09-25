package app

import (
	"context"
	"sync"
	"transcode-service/ddd/application/cqe"
	"transcode-service/ddd/application/dto"
	"transcode-service/pkg/assert"
)

var (
	singleTranscodeApp TranscodeApp
	onceTranscodeApp   sync.Once
)

type TranscodeApp interface {
	// CreateTranscodeTask 创建转码任务
	CreateTranscodeTask(ctx context.Context, req *cqe.TranscodeTaskCqe) (*dto.TranscodeTaskDTO, error)
	// GetTranscodeTask 获取转码任务详情
	GetTranscodeTask(ctx context.Context, taskUUID string) (*dto.TranscodeTaskDTO, error)
	// ListTranscodeTasks 获取转码任务列表
	ListTranscodeTasks(ctx context.Context, userUUID string, page, size int) ([]*dto.TranscodeTaskDTO, int64, error)
	// UpdateTranscodeTaskStatus 更新转码任务状态
	UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, errorMessage string) error
	// CancelTranscodeTask 取消转码任务
	CancelTranscodeTask(ctx context.Context, taskUUID string) error
	// GetTranscodeProgress 获取转码进度
	GetTranscodeProgress(ctx context.Context, taskUUID string) (float64, error)
}

type transcodeAppImpl struct {
}

func DefaultTranscodeApp() TranscodeApp {
	assert.NotCircular()
	onceTranscodeApp.Do(func() {
		singleTranscodeApp = &transcodeAppImpl{}

	})
	assert.NotNil(singleTranscodeApp)
	return singleTranscodeApp
}

func (t *transcodeAppImpl) CreateTranscodeTask(ctx context.Context, req *cqe.TranscodeTaskCqe) (*dto.TranscodeTaskDTO, error) {
	// TODO: 实现转码任务创建逻辑
	return nil, nil
}

func (t *transcodeAppImpl) GetTranscodeTask(ctx context.Context, taskUUID string) (*dto.TranscodeTaskDTO, error) {
	// TODO: 实现获取转码任务逻辑
	return nil, nil
}

func (t *transcodeAppImpl) ListTranscodeTasks(ctx context.Context, userUUID string, page, size int) ([]*dto.TranscodeTaskDTO, int64, error) {
	// TODO: 实现获取转码任务列表逻辑
	return nil, 0, nil
}

func (t *transcodeAppImpl) UpdateTranscodeTaskStatus(ctx context.Context, taskUUID, status, errorMessage string) error {
	// TODO: 实现更新转码任务状态逻辑
	return nil
}

func (t *transcodeAppImpl) CancelTranscodeTask(ctx context.Context, taskUUID string) error {
	// TODO: 实现取消转码任务逻辑
	return nil
}

func (t *transcodeAppImpl) GetTranscodeProgress(ctx context.Context, taskUUID string) (float64, error) {
	// TODO: 实现获取转码进度逻辑
	return 0.0, nil
}
