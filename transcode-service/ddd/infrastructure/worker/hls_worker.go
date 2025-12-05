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
	vgrpc "transcode-service/ddd/infrastructure/grpc"
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

	w.wg.Add(w.workerCount)
	for i := 0; i < w.workerCount; i++ {
		go w.workerLoop(workerCtx)
	}
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
		errMsg := truncateError(err.Error(), 480)
		_ = w.hlsRepo.UpdateHLSJobError(ctx, job.JobUUID(), errMsg)
		_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "failed")

		// 计算回调用的任务 ID：优先源任务 UUID，其次当前 HLS 任务 UUID
		taskUUID := job.JobUUID()
		if src := job.SourceJobUUID(); src != nil {
			taskUUID = *src
		}

		// 通知 video-service 失败
		if cli := vgrpc.DefaultVideoServiceClient(); cli != nil {
			if resp, callErr := cli.UpdateTranscodeResult(ctx, job.VideoUUID(), taskUUID, "failed", "", errMsg, 0, 0); callErr != nil {
				logger.Warnf("video-service HLS failure callback failed video_uuid=%s task_uuid=%s error=%s", job.VideoUUID(), taskUUID, callErr.Error())
			} else if resp != nil {
				logger.Infof("video-service HLS failure callback success=%v video_uuid=%s task_uuid=%s", resp.GetSuccess(), job.VideoUUID(), taskUUID)
			}
		} else {
			logger.Warnf("video-service client is nil, skip HLS failure callback video_uuid=%s task_uuid=%s", job.VideoUUID(), taskUUID)
		}

		// 通知 upload-service / 结果上报方失败
		if w.reporter != nil {
			if repErr := w.reporter.ReportFailure(ctx, job.VideoUUID(), taskUUID, errMsg); repErr != nil {
				logger.Warnf("upload-service HLS failure callback failed video_uuid=%s task_uuid=%s error=%s", job.VideoUUID(), taskUUID, repErr.Error())
			}
		}

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
		ct := detectHLSContentType(path)
		obj := gateway.UploadObject{LocalPath: path, ObjectKey: filepath.ToSlash(rel), ContentType: ct}
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
		publicPath = w.buildFileURL(strings.TrimLeft(key, "/"))
	}
	if publicPath != "" {
		_ = w.hlsRepo.UpdateHLSJobOutput(ctx, job.JobUUID(), publicPath)
	}
	_ = w.hlsRepo.UpdateHLSJobProgress(ctx, job.JobUUID(), 100)
	_ = w.hlsRepo.UpdateHLSJobStatus(ctx, job.JobUUID(), "completed")

	// HLS 完成后回调视频服务，传递 master playlist 地址
	if publicPath != "" {
		taskUUID := ""
		if job.SourceJobUUID() != nil {
			taskUUID = *job.SourceJobUUID()
		} else {
			taskUUID = job.JobUUID()
		}

		// 通知 video-service：视频已发布，video_url 为 HLS master 地址
		if cli := vgrpc.DefaultVideoServiceClient(); cli != nil {
			if resp, err := cli.UpdateTranscodeResult(ctx, job.VideoUUID(), taskUUID, "published", publicPath, "", 0, 0); err != nil {
				logger.Warnf("video-service HLS callback failed video_uuid=%s task_uuid=%s error=%s", job.VideoUUID(), taskUUID, err.Error())
			} else if resp != nil {
				logger.Infof("video-service HLS callback success=%v video_uuid=%s task_uuid=%s url=%s", resp.GetSuccess(), job.VideoUUID(), taskUUID, publicPath)
			}
		} else {
			logger.Warnf("video-service client is nil, skip HLS callback video_uuid=%s task_uuid=%s", job.VideoUUID(), taskUUID)
		}

		// 通知 upload-service：最终 Published 状态 + HLS URL（方案 B）
		if w.reporter != nil {
			if err := w.reporter.ReportSuccess(ctx, job.VideoUUID(), taskUUID, publicPath); err != nil {
				logger.Warnf("upload-service HLS callback failed video_uuid=%s task_uuid=%s error=%s", job.VideoUUID(), taskUUID, err.Error())
			} else {
				logger.Infof("upload-service HLS callback success video_uuid=%s task_uuid=%s url=%s", job.VideoUUID(), taskUUID, publicPath)
			}
		}
	}

	if usedExistingLocal {
		_ = os.Remove(localInput)
	}
	if base != "" && base != "." {
		if err := os.RemoveAll(base); err != nil {
			logger.Warnf("failed to clean local HLS dir path=%s error=%s", base, err.Error())
		}
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
	if w.cfg != nil && w.cfg.Transcode.FFmpeg.TempDir != "" {
		tempDir = w.cfg.Transcode.FFmpeg.TempDir
	}
	fileName := filepath.Base(job.InputPath())
	return filepath.Join(tempDir, "inputs", fmt.Sprintf("hls_%s_%s", job.JobUUID(), fileName))
}

func (w *hlsWorkerImpl) deriveLocalCandidate(remoteKey string) string {
	key := strings.TrimPrefix(remoteKey, "/")
	if key == "" {
		return ""
	}
	base := os.TempDir()
	if w.cfg != nil && w.cfg.Transcode.FFmpeg.TempDir != "" {
		base = w.cfg.Transcode.FFmpeg.TempDir
	}
	return filepath.Join(base, key)
}

// truncateError ensures error messages won't overflow downstream DB columns (e.g., VARCHAR(500)).
func truncateError(msg string, max int) string {
	if max <= 0 {
		return msg
	}
	runes := []rune(msg)
	if len(runes) <= max {
		return msg
	}
	return string(runes[:max])
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

func detectHLSContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
