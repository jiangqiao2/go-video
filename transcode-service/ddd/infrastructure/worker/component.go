package worker

import (
	"context"
	"fmt"

	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/infrastructure/database/persistence"
	grpcclient "transcode-service/ddd/infrastructure/grpc"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/ddd/infrastructure/storage"
	"transcode-service/internal/resource"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
)

// TranscodeWorkerComponentPlugin 负责启动转码Worker
type TranscodeWorkerComponentPlugin struct{}

func (p *TranscodeWorkerComponentPlugin) Name() string {
	return "transcodeWorkerComponent"
}

func (p *TranscodeWorkerComponentPlugin) MustCreateComponent(deps *manager.Dependencies) manager.Component {
	repo := persistence.NewTranscodeRepository()
	queueInstance := queue.DefaultTaskQueue()
	minioResource := resource.DefaultMinioResource()
	storageGateway := storage.NewMinioStorage(minioResource)
	cfg := deps.Config
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	resultReporter := grpcclient.DefaultUploadServiceReporter()
	transcodeSvc := service.NewTranscodeService(repo, storageGateway, cfg, resultReporter)

	workerCount := 1
	workerID := "transcode-worker"
	if cfg != nil {
		if cfg.Worker.MaxConcurrentTasks > 0 {
			workerCount = cfg.Worker.MaxConcurrentTasks
		}
		if cfg.Worker.WorkerID != "" {
			workerID = cfg.Worker.WorkerID
		}
	}

	return &transcodeWorkerComponent{
		name:   "transcodeWorker",
		queue:  queueInstance,
		worker: NewTranscodeWorker(workerID, queueInstance, transcodeSvc, repo, workerCount),
	}
}

type transcodeWorkerComponent struct {
	name   string
	queue  queue.TaskQueue
	worker TranscodeWorker
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *transcodeWorkerComponent) Start() error {
	if c.worker == nil {
		return fmt.Errorf("transcode worker not initialized")
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	if err := c.worker.Start(ctx); err != nil {
		return fmt.Errorf("start transcode worker failed: %w", err)
	}
	logger.Info("Transcode worker component started", map[string]interface{}{"name": c.name})
	return nil
}

func (c *transcodeWorkerComponent) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.worker != nil {
		if err := c.worker.Stop(); err != nil {
			return fmt.Errorf("stop transcode worker failed: %w", err)
		}
	}
	queue.CloseDefaultTaskQueue()
	logger.Info("Transcode worker component stopped", map[string]interface{}{"name": c.name})
	return nil
}

func (c *transcodeWorkerComponent) GetName() string {
	return c.name
}
