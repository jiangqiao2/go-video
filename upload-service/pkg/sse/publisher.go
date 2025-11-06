package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultStreamMaxLen = int64(1000)
	streamKeyTemplate   = "sse:event:%s"
)

// Message mirrors the SSE message schema stored in Redis Streams.
type Message struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	UserID    string          `json:"user_id"`
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	Attempts  int             `json:"attempts"`
}

// Publisher appends SSE messages to Redis Streams.
type Publisher struct {
	client redis.Cmdable
	maxLen int64
}

// NewPublisher constructs a Publisher with sensible defaults.
func NewPublisher(client redis.Cmdable) *Publisher {
	return &Publisher{
		client: client,
		maxLen: defaultStreamMaxLen,
	}
}

// WithMaxLen allows overriding the configured approximate stream length.
func (p *Publisher) WithMaxLen(maxLen int64) *Publisher {
	if maxLen > 0 {
		p.maxLen = maxLen
	}
	return p
}

// Publish writes an SSE message for the given topic and user.
func (p *Publisher) Publish(ctx context.Context, topic, userID, event string, payload any) (string, error) {
	if p == nil || p.client == nil {
		return "", fmt.Errorf("sse publisher not initialised")
	}
	if topic == "" {
		return "", fmt.Errorf("topic cannot be empty")
	}
	if event == "" {
		return "", fmt.Errorf("event cannot be empty")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode payload: %w", err)
	}

	msg := &Message{
		Topic:     topic,
		UserID:    userID,
		Event:     event,
		Payload:   body,
		CreatedAt: time.Now().UTC(),
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("encode message: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: streamKey(topic),
		ID:     "*",
		MaxLen: p.maxLen,
		Approx: true,
		Values: map[string]any{
			"body": encoded,
		},
	}

	messageID, err := p.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("append sse event: %w", err)
	}
	return messageID, nil
}

func streamKey(topic string) string {
	return fmt.Sprintf(streamKeyTemplate, topic)
}
