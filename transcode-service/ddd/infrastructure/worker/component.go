package worker

import (
	"context"
	"fmt"

	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/infrastructure/database/persistence"
	grpcClient "transcode-service/ddd/infrastructure/grpc"
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
	hlsRepo := persistence.NewHLSRepository()
	queueInstance := queue.DefaultTaskQueue()
	cfg := deps.Config
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	var storageGateway gateway.StorageGateway
	rustRes := resource.DefaultRustFSResource()
	storageGateway = storage.NewRustFSStorage(
		rustRes.GetEndpoint(),
		rustRes.GetAccessKey(),
		rustRes.GetSecretKey(),
	)
	// 上报转码结果给上传服务（用于更新视频状态/SSE）
	resultReporter := grpcClient.DefaultUploadServiceReporter()

	transcodeSvc := service.NewTranscodeService(repo, hlsRepo, storageGateway, cfg, resultReporter)
	hlsSvc := service.NewHLSService(logger.DefaultLogger(), hlsRepo, cfg)

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
		// HLS Worker 在完成 HLS 后，通过 reporter 通知 upload-service（Published + HLS URL）
		hlsWorker: NewHLSWorker(workerID+"-hls", hlsRepo, hlsSvc, storageGateway, resultReporter, cfg, 1),
	}
}

type transcodeWorkerComponent struct {
	name      string
	queue     queue.TaskQueue
	worker    TranscodeWorker
	hlsWorker HLSWorker
	ctx       context.Context
	cancel    context.CancelFunc
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
	if c.hlsWorker != nil {
		if err := c.hlsWorker.Start(ctx); err != nil {
			return fmt.Errorf("start hls worker failed: %w", err)
		}
	}
	logger.Infof("Transcode worker component started name=%s", c.name)
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
	if c.hlsWorker != nil {
		if err := c.hlsWorker.Stop(); err != nil {
			return fmt.Errorf("stop hls worker failed: %w", err)
		}
	}
	queue.CloseDefaultTaskQueue()
	logger.Infof("Transcode worker component stopped name=%s", c.name)
	return nil
}

func (c *transcodeWorkerComponent) GetName() string {
	return c.name
}
