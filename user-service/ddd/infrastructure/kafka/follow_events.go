package kafka

import (
	"context"
	"encoding/json"
	"time"

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
