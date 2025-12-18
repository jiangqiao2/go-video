package component

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	drepo "user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/database/persistence"
	kafkainfra "user-service/ddd/infrastructure/kafka"
	"user-service/pkg/config"
	pkgkafka "user-service/pkg/kafka"
	"user-service/pkg/logger"
	"user-service/pkg/manager"

	kafka "github.com/segmentio/kafka-go"
)

// FollowEventConsumerPlugin wires the follow-event Kafka consumer into the component system.
type FollowEventConsumerPlugin struct{}

func (p *FollowEventConsumerPlugin) Name() string { return "followEventConsumer" }

func (p *FollowEventConsumerPlugin) MustCreateComponent(deps *manager.Dependencies) manager.Component {
	return &followEventConsumer{
		repo: persistence.NewFollowRepository(),
	}
}

type followEventConsumer struct {
	repo   drepo.FollowRepository
	ctx    context.Context
	cancel context.CancelFunc
	reader *kafka.Reader
	wg     sync.WaitGroup
}

func (c *followEventConsumer) Start() error {
	cfg := config.GetGlobalConfig()
	if cfg == nil || !cfg.Kafka.Enabled {
		logger.Info("FollowEventConsumer skipped because Kafka is disabled")
		return nil
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	group := cfg.Kafka.GroupID
	if group == "" {
		group = "user-service-follow-consumer"
	}
	c.reader = pkgkafka.DefaultClient().Reader(pkgkafka.FollowEventsTopic, group)
	c.wg.Add(1)
	go c.consumeLoop()
	logger.Infof("FollowEventConsumer started topic=%s group=%s", pkgkafka.FollowEventsTopic, group)
	return nil
}

func (c *followEventConsumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	if c.reader != nil {
		_ = c.reader.Close()
	}
	return nil
}

func (c *followEventConsumer) GetName() string { return "followEventConsumer" }

func (c *followEventConsumer) consumeLoop() {
	defer c.wg.Done()
	for {
		if c.ctx.Err() != nil {
			return
		}
		msg, err := c.reader.FetchMessage(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") {
				logger.Debug("FollowEventConsumer Kafka EOF")
			} else {
				logger.Warnf("FollowEventConsumer read error error=%v", err)
			}
			time.Sleep(time.Second)
			continue
		}

		logger.WithFields(map[string]interface{}{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
		}).Debug("FollowEventConsumer received message")

		if err := c.handleMessage(msg); err != nil {
			logger.Warnf("FollowEventConsumer handle error error=%v partition=%d offset=%d", err, msg.Partition, msg.Offset)
			// 简单重试：暂不提交 offset，稍后重试
			time.Sleep(time.Second)
			continue
		}
		if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
			logger.Warnf("FollowEventConsumer commit error error=%v partition=%d offset=%d", err, msg.Partition, msg.Offset)
		}
	}
}

func (c *followEventConsumer) handleMessage(msg kafka.Message) error {
	var ev kafkainfra.FollowEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		logger.Warnf("FollowEventConsumer unmarshal error error=%v partition=%d offset=%d value=%s", err, msg.Partition, msg.Offset, string(msg.Value))
		return err
	}
	op := strings.ToLower(ev.Op)

	logger.WithFields(map[string]interface{}{
		"user_uuid":   ev.UserUUID,
		"target_uuid": ev.TargetUUID,
		"op":          op,
		"partition":   msg.Partition,
		"offset":      msg.Offset,
	}).Info("FollowEventConsumer handling event")

	switch op {
	case kafkainfra.FollowOpFollow:
		return c.repo.Follow(c.ctx, ev.UserUUID, ev.TargetUUID)
	case kafkainfra.FollowOpUnfollow:
		return c.repo.Unfollow(c.ctx, ev.UserUUID, ev.TargetUUID)
	default:
		logger.Warnf("FollowEventConsumer ignore unknown op op=%s user=%s target=%s", ev.Op, ev.UserUUID, ev.TargetUUID)
		return nil
	}
}

func init() {
	manager.RegisterComponentPlugin(&FollowEventConsumerPlugin{})
}
