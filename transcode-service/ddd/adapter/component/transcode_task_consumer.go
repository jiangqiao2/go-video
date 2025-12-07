package component

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"
	appsvc "transcode-service/ddd/application/app"
	cqe "transcode-service/ddd/application/cqe"
	"transcode-service/ddd/domain/repo"
	"transcode-service/ddd/domain/vo"
	"transcode-service/ddd/infrastructure/database/persistence"
	"transcode-service/ddd/infrastructure/queue"
	"transcode-service/pkg/config"
	"transcode-service/pkg/grpcutil"
	pkgkafka "transcode-service/pkg/kafka"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/task"

	kafka "github.com/segmentio/kafka-go"
)

type TranscodeTaskConsumerPlugin struct{}

func (p *TranscodeTaskConsumerPlugin) Name() string { return "transcodeTaskConsumer" }

func (p *TranscodeTaskConsumerPlugin) MustCreateComponent(deps *manager.Dependencies) manager.Component {
	var app appsvc.TranscodeApp
	if deps != nil {
		if v, ok := deps.TranscodeAppService.(appsvc.TranscodeApp); ok {
			app = v
		}
	}
	if app == nil {
		app = appsvc.DefaultTranscodeApp()
	}
	return &transcodeTaskConsumer{app: app, repo: persistence.NewTranscodeRepository()}
}

type transcodeTaskConsumer struct {
	app                  appsvc.TranscodeApp
	ctx                  context.Context
	cancel               context.CancelFunc
	repo                 repo.TranscodeJobRepository
	reader               *kafka.Reader
	msgCh                chan kafka.Message
	wgRead               sync.WaitGroup
	wgProc               sync.WaitGroup
	max                  int
	interval             time.Duration
	topic                string
	group                string
	commitOnDecodeError  bool
	commitOnProcessError bool
}

func (c *transcodeTaskConsumer) Start() error {
	task.Register(&backgroundTaskAdapter{name: "kafka-consumer", startFunc: c.startInternal, stopFunc: c.Stop})
	// 由 TaskManager 统一启动
	return nil
}

func (c *transcodeTaskConsumer) startInternal(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)
	cfg := config.GetGlobalConfig()
	if cfg != nil {
		if cfg.Worker.MaxConcurrentTasks > 0 {
			c.max = cfg.Worker.MaxConcurrentTasks
		}
		if cfg.Worker.TaskPollInterval > 0 {
			c.interval = cfg.Worker.TaskPollInterval
		}
		if cfg.Kafka.GroupID != "" {
			c.group = cfg.Kafka.GroupID
		}
		if cfg.Kafka.Topics.TranscodeTasks != "" {
			c.topic = cfg.Kafka.Topics.TranscodeTasks
		}
		c.commitOnDecodeError = cfg.Kafka.CommitOnDecodeError
		c.commitOnProcessError = cfg.Kafka.CommitOnProcessError
	}
	if c.max <= 0 {
		c.max = 1
	}
	if c.interval <= 0 {
		c.interval = time.Second
	}
	c.reader = pkgkafka.DefaultClient().Reader(c.topic, c.group)
	c.msgCh = make(chan kafka.Message, c.max)
	logger.Infof("Kafka consumer started topic=%s group=%s", c.topic, c.group)
	c.wgRead.Add(1)
	go c.consumeLoop()
	for i := 0; i < c.max; i++ {
		c.wgProc.Add(1)
		go c.processLoop(i)
	}
	return nil
}

func (c *transcodeTaskConsumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wgRead.Wait()
	if c.msgCh != nil {
		close(c.msgCh)
	}
	c.wgProc.Wait()
	if c.reader != nil {
		_ = c.reader.Close()
	}
	return nil
}
func (c *transcodeTaskConsumer) GetName() string { return "transcodeTaskConsumer" }

func (c *transcodeTaskConsumer) shouldPause(max int) bool {
	if max <= 0 {
		max = 1
	}
	running := 0
	window := 5 * time.Second
	if cfg := config.GetGlobalConfig(); cfg != nil {
		if cfg.Worker.HeartbeatInterval > 0 {
			window = cfg.Worker.HeartbeatInterval
		}
		if cfg.Worker.TaskPollInterval > 0 && cfg.Worker.TaskPollInterval > window {
			window = cfg.Worker.TaskPollInterval
		}
	}
	now := time.Now()
	if c.repo != nil {
		if list, err := c.repo.QueryTranscodeJobsByStatus(context.Background(), vo.TaskStatusProcessing, max*2); err == nil {
			for _, t := range list {
				if t == nil {
					continue
				}
				ts := t.UpdatedAt()
				if !ts.IsZero() && now.Sub(ts) <= window {
					running++
				}
				if running >= max {
					break
				}
			}
		}
	}
	size := queue.DefaultTaskQueue().Size()
	if running >= max {
		return true
	}
	if size >= max {
		return true
	}
	return false
}

func (c *transcodeTaskConsumer) consumeLoop() {
	defer c.wgRead.Done()
	for {
		if c.ctx.Err() != nil {
			return
		}
		if c.shouldPause(c.max) {
			logger.Debug("Kafka consumer paused", map[string]interface{}{"max": c.max, "size": queue.DefaultTaskQueue().Size()})
			time.Sleep(c.interval)
			continue
		}
		msg, err := c.reader.FetchMessage(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") {
				logger.Debug("Kafka reader EOF")
			} else {
				logger.Warnf("Kafka read error error=%s", err.Error())
			}
			continue
		}
		select {
		case c.msgCh <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *transcodeTaskConsumer) processLoop(workerID int) {
	defer c.wgProc.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.msgCh:
			if !ok {
				return
			}
			msgCtx := c.ctx
			if rid := headerValue(msg.Headers, "request-id"); rid != "" {
				if ctxWithReq, _ := grpcutil.ContextWithRequestID(msgCtx, rid); ctxWithReq != nil {
					msgCtx = ctxWithReq
				}
			}
			req, err := c.decodeKafkaMessage(&msg)
			if err != nil {
				if c.commitOnDecodeError {
					if e := c.reader.CommitMessages(c.ctx, msg); e != nil {
						logger.Warnf("Kafka commit error error=%s partition=%d offset=%d worker=%d", e.Error(), msg.Partition, msg.Offset, workerID)
					} else {
						logger.Infof("Kafka commit done partition=%d offset=%d worker=%d", msg.Partition, msg.Offset, workerID)
					}
				}
				continue
			}
			if _, err := c.app.CreateTranscodeTask(msgCtx, req); err != nil {
				if c.commitOnProcessError {
					if e := c.reader.CommitMessages(c.ctx, msg); e != nil {
						logger.Warnf("Kafka commit error error=%s partition=%d offset=%d worker=%d", e.Error(), msg.Partition, msg.Offset, workerID)
					} else {
						logger.Infof("Kafka commit done partition=%d offset=%d worker=%d", msg.Partition, msg.Offset, workerID)
					}
				}
				time.Sleep(c.interval)
				continue
			}
			if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
				logger.Warnf("Kafka commit error error=%s partition=%d offset=%d worker=%d", err.Error(), msg.Partition, msg.Offset, workerID)
			} else {
				logger.Infof("Kafka commit done partition=%d offset=%d worker=%d", msg.Partition, msg.Offset, workerID)
			}
		}
	}
}

func (c *transcodeTaskConsumer) decodeKafkaMessage(msg *kafka.Message) (*cqe.CreateTranscodeTaskReq, error) {
	var m struct {
		UserUUID         string `json:"user_uuid"`
		VideoUUID        string `json:"video_uuid"`
		VideoPushUUID    string `json:"video_push_uuid"`
		InputPath        string `json:"input_path"`
		TargetResolution string `json:"target_resolution"`
		TargetBitrate    string `json:"target_bitrate"`
	}
	if err := json.Unmarshal(msg.Value, &m); err != nil {
		return nil, err
	}
	req := &cqe.CreateTranscodeTaskReq{
		UserUUID:      m.UserUUID,
		VideoUUID:     m.VideoUUID,
		VideoPushUUID: m.VideoPushUUID,
		OriginalPath:  m.InputPath,
		Resolution:    m.TargetResolution,
		Bitrate:       m.TargetBitrate,
	}
	return req, nil
}

func headerValue(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

type backgroundTaskAdapter struct {
	name      string
	startFunc func(ctx context.Context) error
	stopFunc  func() error
}

func (b *backgroundTaskAdapter) Name() string                    { return b.name }
func (b *backgroundTaskAdapter) Start(ctx context.Context) error { return b.startFunc(ctx) }
func (b *backgroundTaskAdapter) Stop() error                     { return b.stopFunc() }
