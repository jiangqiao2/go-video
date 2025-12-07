package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"transcode-service/ddd/domain/entity"
	"transcode-service/ddd/domain/gateway"
	"transcode-service/ddd/domain/port"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/executor"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"
)

// TranscodeService 转码领域服务
type TranscodeService interface {
	// ExecuteTranscode 执行转码任务
	ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error
}

type transcodeServiceImpl struct {
	transcodeRepo  repo.TranscodeJobRepository
	hlsRepo        repo.HLSJobRepository
	storageGateway gateway.StorageGateway
	cfg            *config.Config
	resultReporter gateway.TranscodeResultReporter
	executor       port.TranscodeExecutor
	progressSink   port.ProgressSink
	progressMu     sync.Mutex
	lastPersist    map[string]time.Time
}

// NewTranscodeService 创建转码领域服务
func NewTranscodeService(transcodeRepo repo.TranscodeJobRepository, hlsRepo repo.HLSJobRepository, storage gateway.StorageGateway, cfg *config.Config, reporter gateway.TranscodeResultReporter, executor port.TranscodeExecutor, sink port.ProgressSink) TranscodeService {
	return &transcodeServiceImpl{
		transcodeRepo:  transcodeRepo,
		hlsRepo:        hlsRepo,
		storageGateway: storage,
		cfg:            cfg,
		resultReporter: reporter,
		executor:       executor,
		progressSink:   sink,
		lastPersist:    make(map[string]time.Time),
	}
}

// ExecuteTranscode 执行转码任务
func (s *transcodeServiceImpl) ExecuteTranscode(ctx context.Context, task *entity.TranscodeTaskEntity) error {
	logger.Infof("start transcode task task_uuid=%s video_uuid=%s resolution=%s bitrate=%s",
		task.TaskUUID(), task.VideoUUID(), task.GetParams().Resolution, task.GetParams().Bitrate)

	if s.cfg == nil {
		s.cfg = config.GetGlobalConfig()
	}
	if s.executor == nil {
		// 延迟获取默认 executor（ffmpeg）
		s.executor = getDefaultTranscodeExecutor(s.cfg, s.storageGateway)
	}

	// 更新任务状态为处理中
	if err := task.TransitionTo(vo.TaskStatusProcessing); err != nil {
		return err
	}
	task.SetProgress(0)
	task.SetErrorMessage("")
	if err := s.updateJobStatus(ctx, task, vo.TaskStatusProcessing, ""); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}
	defer s.clearProgressThrottle(task.TaskUUID())

	opt := port.TranscodeOptions{
		SkipUpload: s.cfg != nil && s.cfg.Transcode.SkipFullUpload,
		ProgressCb: func(p int) {
			s.setProgress(task, float64(p), 100)
		},
	}
	uploadedKey, _, err := s.executor.Execute(ctx, task, opt)
	if err != nil {
		_ = task.TransitionTo(vo.TaskStatusFailed)
		task.SetErrorMessage(err.Error())
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
		return fmt.Errorf("转码执行失败: %w", err)
	}

	if !opt.SkipUpload {
		task.SetOutputPath(uploadedKey)
	} else {
		task.SetOutputPath("")
	}
	_ = task.TransitionTo(vo.TaskStatusCompleted)
	task.SetProgress(100)
	task.SetErrorMessage("")

	if err := s.transcodeRepo.UpdateTranscodeJob(ctx, task); err != nil {
		errorMsg := fmt.Sprintf("更新任务完成状态失败: %v", err)
		_ = task.TransitionTo(vo.TaskStatusFailed)
		task.SetErrorMessage(errorMsg)
		_ = s.updateJobStatus(ctx, task, vo.TaskStatusFailed, task.ErrorMessage())
		return fmt.Errorf("更新任务完成状态失败: %w", err)
	}

	variants := make([]vo.ResolutionConfig, 0, 4)
	existed := map[string]struct{}{}
	if s.cfg != nil && len(s.cfg.Transcode.OutputFormats) > 0 {
		for _, of := range s.cfg.Transcode.OutputFormats {
			name := strings.TrimSpace(of.Name)
			br := strings.TrimSpace(of.Bitrate)
			if name == "" || br == "" {
				continue
			}
			if rc, err := vo.NewResolutionConfig(name, br); err == nil {
				if _, ok := existed[rc.Resolution]; !ok {
					variants = append(variants, *rc)
					existed[rc.Resolution] = struct{}{}
				}
			}
		}
	}
	defaults := map[string]string{"1080p": "4000k", "720p": "2000k", "480p": "1000k"}
	for res, br := range defaults {
		if _, ok := existed[res]; !ok {
			if rc, err := vo.NewResolutionConfig(res, br); err == nil {
				variants = append(variants, *rc)
			}
		}
	}

	inputForHLS := uploadedKey
	if opt.SkipUpload || strings.TrimSpace(uploadedKey) == "" {
		inputForHLS = task.OriginalPath()
	}
	if len(variants) > 0 && s.hlsRepo != nil {
		if hcfg, err2 := vo.NewHLSConfig(true, variants); err2 == nil {
			hJobUUID := uuid.New().String()
			outputDir := filepath.ToSlash(filepath.Join("storage/hls", task.UserUUID(), task.VideoUUID(), hJobUUID))
			hJob := entity.NewHLSJobEntity(hJobUUID, task.UserUUID(), task.VideoUUID(), inputForHLS, outputDir, *hcfg)
			src := task.TaskUUID()
			hJob.SetSource(&src, "transcoded")
			_ = s.hlsRepo.CreateHLSJob(ctx, hJob)
			_ = queue.DefaultHLSJobQueue().Enqueue(ctx, hJob)
		}
	}

	logger.Infof("transcode task finished task_uuid=%s output_path=%s skip_upload=%t",
		task.TaskUUID(), uploadedKey, opt.SkipUpload)

	return nil
}

// updateJobStatus 封装状态更新，统一使用任务当前的输出路径与进度。
func (s *transcodeServiceImpl) updateJobStatus(ctx context.Context, task *entity.TranscodeTaskEntity, status vo.TaskStatus, message string) error {
	if s.transcodeRepo == nil {
		return errors.New("transcodeRepo is nil")
	}
	return s.transcodeRepo.UpdateTranscodeJobStatus(ctx, task.TaskUUID(), status, message, task.OutputPath(), task.Progress())
}

func (s *transcodeServiceImpl) setProgress(task *entity.TranscodeTaskEntity, currentSec, totalSec float64) {
	if totalSec <= 0 {
		return
	}
	pct := int((currentSec / totalSec) * 100)
	if pct > 99 {
		pct = 99
	}
	if pct < 0 {
		pct = 0
	}
	task.SetProgress(pct)
	shouldPersist := false
	now := time.Now()
	s.progressMu.Lock()
	last := s.lastPersist[task.TaskUUID()]
	if last.IsZero() || now.Sub(last) >= time.Minute {
		s.lastPersist[task.TaskUUID()] = now
		shouldPersist = true
	}
	s.progressMu.Unlock()
	if shouldPersist {
		if sink := s.progressWriter(); sink != nil {
			if err := sink.SaveProgress(context.Background(), task, pct); err != nil {
				logger.Errorf("update transcode progress failed task_uuid=%s progress=%d error=%s", task.TaskUUID(), pct, err.Error())
			}
		} else if err := s.transcodeRepo.UpdateTranscodeJobProgress(context.Background(), task.TaskUUID(), pct); err != nil {
			logger.Errorf("update transcode progress failed task_uuid=%s progress=%d error=%s", task.TaskUUID(), pct, err.Error())
		}
	}
}

func getDefaultTranscodeExecutor(cfg *config.Config, storage gateway.StorageGateway) port.TranscodeExecutor {
	return executor.NewFFmpegExecutor(cfg, storage)
}

func (s *transcodeServiceImpl) progressWriter() port.ProgressSink {
	if s.progressSink != nil {
		return s.progressSink
	}
	return nil
}

func detectHLSContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
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

func (s *transcodeServiceImpl) clearProgressThrottle(taskUUID string) {
	s.progressMu.Lock()
	delete(s.lastPersist, taskUUID)
	s.progressMu.Unlock()
}
