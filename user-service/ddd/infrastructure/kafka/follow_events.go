package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"user-service/pkg/config"
	pkgkafka "user-service/pkg/kafka"
	"user-service/pkg/logger"
)

const (
	// FollowOpFollow represents a follow operation.
	FollowOpFollow = "follow"
	// FollowOpUnfollow represents an unfollow operation.
	FollowOpUnfollow = "unfollow"
)

// FollowEvent is the payload sent to Kafka for follow/unfollow commands.
type FollowEvent struct {
	UserUUID   string `json:"user_uuid"`
	TargetUUID string `json:"target_uuid"`
	Op         string `json:"op"`
	TS         int64  `json:"ts"` // unix millis
}

// PublishFollowEvent sends a follow/unfollow command to Kafka.
// If Kafka is disabled or not configured, it returns a non-nil error so callers can fall back.
func PublishFollowEvent(ctx context.Context, op, userUUID, targetUUID string) error {
	if userUUID == "" || targetUUID == "" {
		return errors.New("userUUID/targetUUID cannot be empty")
	}
	cfg := config.GetGlobalConfig()
	if cfg == nil || !cfg.Kafka.Enabled {
		return errors.New("kafka not enabled for user-service")
	}
	ev := FollowEvent{
		UserUUID:   userUUID,
		TargetUUID: targetUUID,
		Op:         op,
		TS:         time.Now().UnixMilli(),
	}
	data, err := json.Marshal(&ev)
	if err != nil {
		return err
	}
	key := []byte(targetUUID)
	if len(key) == 0 {
		key = []byte(userUUID)
	}
	if err := pkgkafka.DefaultClient().Produce(ctx, pkgkafka.FollowEventsTopic, key, data); err != nil {
		logger.WithContext(ctx).Warnf("PublishFollowEvent failed op=%s user=%s target=%s err=%v", op, userUUID, targetUUID, err)
		return err
	}
	return nil
}
