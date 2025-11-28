package component

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	appsvc "transcode-service/ddd/application/app"
	cqe "transcode-service/ddd/application/cqe"
	pkgkafka "transcode-service/pkg/kafka"
	"transcode-service/pkg/logger"
	"transcode-service/pkg/manager"
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
	return &transcodeTaskConsumer{app: app}
}

type transcodeTaskConsumer struct {
	app    appsvc.TranscodeApp
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *transcodeTaskConsumer) Start() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	reader := pkgkafka.DefaultClient().Reader("transcode.tasks", "transcode-service-group")
	go func() {
		defer reader.Close()
		logger.Infof("Kafka consumer started topic=%s group=%s", "transcode.tasks", "transcode-service-group")
		for {
			msg, err := reader.ReadMessage(c.ctx)
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
			var m struct {
				UserUUID         string `json:"user_uuid"`
				VideoUUID        string `json:"video_uuid"`
				InputPath        string `json:"input_path"`
				TargetResolution string `json:"target_resolution"`
				TargetBitrate    string `json:"target_bitrate"`
			}
			if err := json.Unmarshal(msg.Value, &m); err != nil {
				logger.Warnf("Kafka message unmarshal error error=%s", err.Error())
				continue
			}
			logger.Infof("Kafka message received video_uuid=%s user_uuid=%s", m.VideoUUID, m.UserUUID)
			req := &cqe.CreateTranscodeTaskReq{
				UserUUID:     m.UserUUID,
				VideoUUID:    m.VideoUUID,
				OriginalPath: m.InputPath,
				Resolution:   m.TargetResolution,
				Bitrate:      m.TargetBitrate,
			}
			if _, err := c.app.CreateTranscodeTask(context.Background(), req); err != nil {
				logger.Warnf("CreateTranscodeTask failed error=%s video_uuid=%s", err.Error(), m.VideoUUID)
			}
		}
	}()
	return nil
}

func (c *transcodeTaskConsumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
func (c *transcodeTaskConsumer) GetName() string { return "transcodeTaskConsumer" }
