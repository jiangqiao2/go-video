package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/service"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

type HLSWorker interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	GetStats() WorkerStats
}

type hlsWorkerImpl struct {
	id          string
	hlsRepo     repo.HLSJobRepository
	hlsService  service.HLSService
	storage     gateway.StorageGateway
	reporter    gateway.TranscodeResultReporter
	cfg         *config.Config
	workerCount int
	running     bool
	cancel      context.CancelFunc
	stats       WorkerStats
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

func NewHLSWorker(id string, hlsRepo repo.HLSJobRepository, hlsService service.HLSService, storage gateway.StorageGateway, reporter gateway.TranscodeResultReporter, cfg *config.Config, workerCount int) HLSWorker {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &hlsWorkerImpl{
		id:          id,
		hlsRepo:     hlsRepo,
		hlsService:  hlsService,
		storage:     storage,
		reporter:    reporter,
		cfg:         cfg,
		workerCount: workerCount,
		stats:       WorkerStats{StartTime: time.Now()},
	}
}

func (w *hlsWorkerImpl) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return fmt.Errorf("worker %s is already running", w.id)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.running = true
	w.stats.StartTime = time.Now()
	go func() {
		jobs, err := w.hlsRepo.QueryHLSJobsByStatus(workerCtx, "pending", 100)
		if err == nil {
			for _, j := range jobs {
				_ = queue.DefaultHLSJobQueue().Enqueue(workerCtx, j)
			}
		}
	}()

	w.wg.Add(1)
	go w.workerLoop(workerCtx)
	return nil
}

func (w *hlsWorkerImpl) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running {
		return nil
	}
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	w.running = false
	return nil
}

func (w *hlsWorkerImpl) IsRunning() bool       { w.mu.RLock(); defer w.mu.RUnlock(); return w.running }
func (w *hlsWorkerImpl) GetStats() WorkerStats { w.mu.RLock(); defer w.mu.RUnlock(); return w.stats }

func (w *hlsWorkerImpl) workerLoop(ctx context.Context) {
	defer w.wg.Done()
	for {
		job, err := queue.DefaultHLSJobQueue().Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		if job == nil {
			continue
		}
		_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "processing")
		w.processJob(ctx, job)
	}
}

func (w *hlsWorkerImpl) processJob(ctx context.Context, job *entity.HLSJobEntity) {
	w.updateStats(func(s *WorkerStats) { s.CurrentlyRunning++; s.LastTaskTime = time.Now() })
	defer w.updateStats(func(s *WorkerStats) { s.CurrentlyRunning--; s.ProcessedTasks++ })

	usedExistingLocal := false
	localInput := ""
	candidate := w.deriveLocalCandidate(job.InputPath())
	if candidate != "" {
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			localInput = candidate
			usedExistingLocal = true
		}
	}
	if !usedExistingLocal {
		localInput = w.getLocalInputPath(job)
		if err := os.MkdirAll(filepath.Dir(localInput), 0o755); err != nil {
			_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), err.Error())
			_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")
			w.updateStats(func(s *WorkerStats) { s.FailedTasks++ })
			return
		}
		if err := w.storage.DownloadFile(ctx, job.InputPath(), localInput); err != nil {
			_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), err.Error())
			_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")
			w.updateStats(func(s *WorkerStats) { s.FailedTasks++ })
			return
		}
		defer os.Remove(localInput)
	}

	if err := w.hlsService.GenerateHLSSlices(ctx, job, localInput); err != nil {
		_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), err.Error())
		_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")
		w.updateStats(func(s *WorkerStats) { s.FailedTasks++ })
		return
	}

	objects := make([]gateway.UploadObject, 0, 32)
	base := filepath.Clean(job.OutputDir())
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel := path
		if strings.HasPrefix(rel, "storage"+string(filepath.Separator)) {
			if r, e := filepath.Rel("storage", rel); e == nil {
				rel = r
			}
		}
		obj := gateway.UploadObject{LocalPath: path, ObjectKey: filepath.ToSlash(rel), ContentType: ""}
		objects = append(objects, obj)
		return nil
	})
	if len(objects) == 0 {
		_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), "no hls files generated")
		_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")
		w.updateStats(func(s *WorkerStats) { s.FailedTasks++ })
		return
	}
	if err := w.storage.UploadObjects(ctx, objects); err != nil {
		_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), err.Error())
		_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")
		w.updateStats(func(s *WorkerStats) { s.FailedTasks++ })
		return
	}

	master := job.MasterPlaylist()
	publicPath := ""
	if master != nil {
		m := *master
		if strings.HasPrefix(m, "storage"+string(filepath.Separator)) {
			if r, e := filepath.Rel("storage", m); e == nil {
				m = r
			}
		}
		key := filepath.ToSlash(m) // e.g. hls/uid/vid/job/master.m3u8
		publicPath = "/storage/transcode/" + strings.TrimLeft(key, "/")
	}
	if publicPath != "" {
		_ = w.hlsRepo.UpdateHLSJobOutput(ctx, job.JobUUID(), publicPath)
	}
	_ = w.hlsRepo.UpdateHLSJobProgress(ctx, job.JobUUID(), 100)
	_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "completed")

	// 通知上传服务仅更新状态，不回写地址
	if w.reporter != nil && publicPath != "" {
		// 尝试获取关联的转码任务UUID (SourceID)
		taskUUID := ""
		if job.SourceJobUUID() != nil {
			taskUUID = *job.SourceJobUUID()
		} else {
			// 如果没有 SourceID，可能无法关联回原任务，这是一个潜在问题
			// 但通常 HLS 任务是由转码任务触发的
			taskUUID = job.JobUUID() // 降级方案
		}

		result := vo.NewTranscodeResult(taskUUID, job.VideoUUID())
		if err := result.ReportSuccess(ctx, w.reporter, publicPath); err != nil {
			logger.Warnf("report HLS success failed task_uuid=%s video_uuid=%s error=%s",
				taskUUID, job.VideoUUID(), err.Error())
		}
	}

	if usedExistingLocal {
		_ = os.Remove(localInput)
	}
	w.updateStats(func(s *WorkerStats) { s.SuccessfulTasks++ })
}

func (w *hlsWorkerImpl) updateStats(f func(*WorkerStats)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	f(&w.stats)
}

func (w *hlsWorkerImpl) getLocalInputPath(job *entity.HLSJobEntity) string {
	tempDir := os.TempDir()
	fileName := filepath.Base(job.InputPath())
	return filepath.Join(tempDir, "inputs", fmt.Sprintf("hls_%s_%s", job.JobUUID(), fileName))
}

func (w *hlsWorkerImpl) deriveLocalCandidate(remoteKey string) string {
	key := strings.TrimPrefix(remoteKey, "/")
	if key == "" {
		return ""
	}
	return filepath.Join(os.TempDir(), key)
}

func (w *hlsWorkerImpl) buildFileURL(objectKey string) string {
	if strings.TrimSpace(objectKey) == "" {
		return ""
	}
	cfg := w.cfg
	if cfg == nil {
		cfg = config.GetGlobalConfig()
	}
	if cfg == nil {
		return objectKey
	}

	key := strings.TrimLeft(objectKey, "/")
	if strings.HasPrefix(key, "transcode/") {
		key = strings.TrimPrefix(key, "transcode/")
	}
	path := fmt.Sprintf("/storage/transcode/%s", key)
	publicBase := strings.TrimSpace(cfg.Public.StorageBase)
	if publicBase != "" {
		if !strings.HasPrefix(publicBase, "http://") && !strings.HasPrefix(publicBase, "https://") {
			publicBase = "http://" + publicBase
		}
		return strings.TrimRight(publicBase, "/") + path
	}
	return path
}
