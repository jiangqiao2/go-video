package bus

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"
    "strings"

    "github.com/redis/go-redis/v9"
)

// Message describes an SSE event persisted in Redis Streams.
type Message struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	UserID    string          `json:"user_id"`
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	Attempts  int             `json:"attempts"`
}

// RedisStreamBus publishes and consumes SSE events using Redis Streams.
type RedisStreamBus struct {
	client redis.Cmdable
	maxLen int64
}

// NewRedisStreamBus creates a new RedisStreamBus with an optional max length for streams.
func NewRedisStreamBus(client redis.Cmdable, maxLen int64) *RedisStreamBus {
	if maxLen <= 0 {
		maxLen = 1000
	}
	return &RedisStreamBus{client: client, maxLen: maxLen}
}

func (b *RedisStreamBus) streamKey(topic string) string {
	return fmt.Sprintf("sse:event:%s", topic)
}

// Publish appends a message to the Redis stream for the topic.
func (b *RedisStreamBus) Publish(ctx context.Context, topic string, msg Message) (string, error) {
	msg.Topic = topic
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}

	body, err := json.Marshal(&msg)
	if err != nil {
		return "", err
	}

    args := &redis.XAddArgs{
        Stream: b.streamKey(topic),
        ID:     "*",
        MaxLen: b.maxLen,
        Approx: true,
        Values: map[string]any{
            "body": body,
        },
    }

	return b.client.XAdd(ctx, args).Result()
}

// CreateGroup ensures a consumer group exists for the topic.
func (b *RedisStreamBus) CreateGroup(ctx context.Context, topic, group string) error {
	if group == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	stream := b.streamKey(topic)
    if err := b.client.XGroupCreateMkStream(ctx, stream, group, "$").Err(); err != nil {
        // go-redis v9 does not expose ErrBusyGroup; detect by message content.
        if strings.Contains(err.Error(), "BUSYGROUP") {
            return nil
        }
        return err
    }
    return nil
}

// ConsumeOptions configures the stream consumption loop.
type ConsumeOptions struct {
	Topic    string
	Group    string
	Consumer string
	Block    time.Duration
	Count    int64
}

// Consume starts a blocking loop that forwards messages to the handler.
func (b *RedisStreamBus) Consume(ctx context.Context, opt ConsumeOptions, handler func(context.Context, Message) error) error {
	if opt.Topic == "" || opt.Group == "" || opt.Consumer == "" {
		return fmt.Errorf("invalid consume options")
	}

	stream := b.streamKey(opt.Topic)
	block := opt.Block
	if block <= 0 {
		block = time.Second
	}
	count := opt.Count
	if count <= 0 {
		count = 10
	}

	for {
		streams, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    opt.Group,
			Consumer: opt.Consumer,
			Streams:  []string{stream, ">"},
			Count:    count,
			Block:    block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		for _, streamResult := range streams {
			for _, rawMsg := range streamResult.Messages {
				body, ok := rawMsg.Values["body"].(string)
				if !ok {
					// Malformed entry, acknowledge to avoid retries.
					_ = b.client.XAck(ctx, stream, opt.Group, rawMsg.ID).Err()
					continue
				}

				var message Message
				if err := json.Unmarshal([]byte(body), &message); err != nil {
					_ = b.client.XAck(ctx, stream, opt.Group, rawMsg.ID).Err()
					continue
				}

				message.ID = rawMsg.ID

				if err := handler(ctx, message); err != nil {
					message.Attempts++
					// Leave the message pending for retry by another consumer.
					continue
				}

				_ = b.client.XAck(ctx, stream, opt.Group, rawMsg.ID).Err()
			}
		}
	}
}
