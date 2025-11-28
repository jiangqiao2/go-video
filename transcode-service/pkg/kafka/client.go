package kafka

import (
	"context"
	"sync"
	"time"

	"transcode-service/pkg/config"

	kafka "github.com/segmentio/kafka-go"
)

type Client struct {
	brokers  []string
	clientID string
	dialer   *kafka.Dialer
	writers  sync.Map // topic -> *kafka.Writer
}

var (
	once      sync.Once
	singleton *Client
)

func DefaultClient() *Client {
	once.Do(func() {
		singleton = &Client{}
	})
	return singleton
}

func (c *Client) MustOpen() {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		panic("global config not initialized before Kafka client")
	}
	c.brokers = cfg.Kafka.BootstrapServers
	c.clientID = cfg.Kafka.ClientID
	c.dialer = &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: c.clientID,
	}
}

func (c *Client) Close() {
	c.writers.Range(func(key, value interface{}) bool {
		if w, ok := value.(*kafka.Writer); ok {
			_ = w.Close()
		}
		return true
	})
}

func (c *Client) Writer(topic string) *kafka.Writer {
	if v, ok := c.writers.Load(topic); ok {
		return v.(*kafka.Writer)
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(c.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	actual, _ := c.writers.LoadOrStore(topic, w)
	return actual.(*kafka.Writer)
}

func (c *Client) Produce(ctx context.Context, topic string, key, value []byte) error {
	w := c.Writer(topic)
	msg := kafka.Message{Key: key, Value: value, Time: time.Now()}
	return w.WriteMessages(ctx, msg)
}

func (c *Client) Reader(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  groupID,
		Topic:    topic,
		Dialer:   c.dialer,
		MinBytes: 1,
		MaxBytes: 10 << 20,
	})
}
