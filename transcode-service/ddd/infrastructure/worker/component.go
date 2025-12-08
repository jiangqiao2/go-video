package worker

import (
	"context"
	"fmt"

	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/infrastructure/database/persistence"
	"transcode-service/ddd/infrastructure/executor"
	grpcClient "transcode-service/ddd/infrastructure/grpc"
	"transcode-service/ddd/infrastructure/progress"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/ddd/infrastructure/storage"
	"transcode-service/internal/resource"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/task"
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

	ffExecutor := executor.NewFFmpegExecutor(cfg, storageGateway)
	progressSink := progress.NewDBSink(repo)
	transcodeSvc := service.NewTranscodeService(repo, hlsRepo, storageGateway, cfg, resultReporter, ffExecutor, progressSink)
	hlsSvc := service.DefaultHLSService()

	workerCount := 1
	hlsWorkerCount := 1
	workerID := "transcode-worker"
	if cfg != nil {
		if cfg.Worker.MaxConcurrentTasks > 0 {
			workerCount = cfg.Worker.MaxConcurrentTasks
		}
		if cfg.Worker.HLSMaxConcurrentTasks > 0 {
			hlsWorkerCount = cfg.Worker.HLSMaxConcurrentTasks
		} else {
			hlsWorkerCount = workerCount
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
		hlsWorker: NewHLSWorker(workerID+"-hls", hlsRepo, hlsSvc, storageGateway, resultReporter, cfg, hlsWorkerCount),
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

	// 注册后台任务，让应用启动时统一管理
	task.Register(&backgroundTaskAdapter{name: c.name, startFunc: c.worker.Start, stopFunc: c.worker.Stop})
	if c.hlsWorker != nil {
		task.Register(&backgroundTaskAdapter{name: c.name + "-hls", startFunc: c.hlsWorker.Start, stopFunc: c.hlsWorker.Stop})
	}
	logger.Infof("Transcode worker component registered background tasks name=%s", c.name)
	return nil
}

func (c *transcodeWorkerComponent) Stop() error {
	// 背景任务由 task.Manager 控制停止，这里保持幂等
	if c.cancel != nil {
		c.cancel()
	}
	queue.CloseDefaultTaskQueue()
	logger.Infof("Transcode worker component stopped name=%s", c.name)
	return nil
}

func (c *transcodeWorkerComponent) GetName() string {
	return c.name
}

// backgroundTaskAdapter adapts Start/Stop functions to the BackgroundTask interface.
type backgroundTaskAdapter struct {
	name      string
	startFunc func(ctx context.Context) error
	stopFunc  func() error
}

func (b *backgroundTaskAdapter) Name() string                    { return b.name }
func (b *backgroundTaskAdapter) Start(ctx context.Context) error { return b.startFunc(ctx) }
func (b *backgroundTaskAdapter) Stop() error                     { return b.stopFunc() }
