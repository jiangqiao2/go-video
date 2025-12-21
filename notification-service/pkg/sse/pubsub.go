package sse

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"notification-service/pkg/logger"
)

const (
	// defaultRedisChannel is the shared channel used for cross-instance notification events.
	defaultRedisChannel = "go-video:notification:sse"
)

// redisEnvelope is the message shape stored in Redis Pub/Sub.
// It wraps the user-specific SSE Event so all instances can fan it back
// into their local in-memory Hub.
type redisEnvelope struct {
	UserUUID string      `json:"user_uuid"`
	Type     string      `json:"type"`
	Data     interface{} `json:"data,omitempty"`
	SentAt   time.Time   `json:"sent_at"`
}

// redisPubSubBridge connects the local in-process Hub with a Redis Pub/Sub channel.
// It lets any instance publish a notification that will be fanned out to all
// other instances, while keeping adapters bound only to the Hub abstraction.
type redisPubSubBridge struct {
	// client must be a *redis.Client (or compatible type) that supports both
	// Publish and Subscribe. We intentionally do not use redis.Cmdable here
	// because Cmdable does not declare Subscribe.
	client  *redis.Client
	channel string
}

var globalBridge *redisPubSubBridge

// InitRedisPubSub wires the global Hub to a Redis Pub/Sub channel.
// It should be called once during service startup after the Redis client
// has been initialised. If redis is unavailable, the service still works
// in single-instance mode via the in-memory Hub.
func InitRedisPubSub(client *redis.Client, channel string) {
	if client == nil {
		return
	}
	if channel == "" {
		channel = defaultRedisChannel
	}

	globalBridge = &redisPubSubBridge{
		client:  client,
		channel: channel,
	}

	go globalBridge.runSubscriber()
	logger.Infof("sse: redis pubsub bridge initialised channel=%s", channel)
}

// PublishNotification dispatches an SSE notification event.
//   - In single-instance/dev mode (no redis bridge), it writes directly to the local Hub.
//   - In multi-instance mode (redis bridge enabled), it publishes to Redis so that
//     every instance receives the event and replays it into its own Hub.
func PublishNotification(userUUID string, ev Event) {
	if userUUID == "" || ev.Type == "" {
		return
	}

	if globalBridge != nil {
		globalBridge.publish(userUUID, ev)
		return
	}

	// Fallback: process-local only.
	DefaultHub().Publish(userUUID, ev)
}

// publish sends the event to the shared Redis channel.
func (b *redisPubSubBridge) publish(userUUID string, ev Event) {
	env := &redisEnvelope{
		UserUUID: userUUID,
		Type:     ev.Type,
		Data:     ev.Data,
		SentAt:   time.Now().UTC(),
	}

	body, err := json.Marshal(env)
	if err != nil {
		logger.Errorf("sse: encode redis envelope failed error=%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := b.client.Publish(ctx, b.channel, body).Err(); err != nil {
		logger.Errorf("sse: publish redis message failed channel=%s error=%v", b.channel, err)
	}
}

// runSubscriber listens on the shared Redis channel and forwards events
// into the local in-memory Hub. This lets SSE streams on any instance
// receive notifications regardless of where they were produced.
func (b *redisPubSubBridge) runSubscriber() {
	ctx := context.Background()

	pubsub := b.client.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	// Ensure subscription is established before reading messages.
	if _, err := pubsub.Receive(ctx); err != nil {
		logger.Errorf("sse: failed to subscribe to redis channel=%s error=%v", b.channel, err)
		return
	}

	ch := pubsub.Channel()
	for msg := range ch {
		var env redisEnvelope
		if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
			logger.Errorf("sse: failed to decode redis message channel=%s error=%v", b.channel, err)
			continue
		}
		if env.UserUUID == "" || env.Type == "" {
			continue
		}
		// Fan-in back to the local hub; adapters stay unaware of redis.
		DefaultHub().Publish(env.UserUUID, Event{
			Type: env.Type,
			Data: env.Data,
		})
	}
}
