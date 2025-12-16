package kafka

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"user-service/pkg/config"
	"user-service/pkg/grpcutil"

	kafka "github.com/segmentio/kafka-go"
)

// Client is a thin wrapper around segmentio/kafka-go with lazy writer caching.
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

// DefaultClient returns the global Kafka client singleton.
func DefaultClient() *Client {
	once.Do(func() {
		singleton = &Client{}
	})
	return singleton
}

// MustOpen initializes the client using global config.
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

// Close closes all cached writers.
func (c *Client) Close() {
	c.writers.Range(func(key, value interface{}) bool {
		if w, ok := value.(*kafka.Writer); ok {
			_ = w.Close()
		}
		return true
	})
}

// Writer returns a cached writer for a topic.
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

// Produce writes a single message to the given topic.
func (c *Client) Produce(ctx context.Context, topic string, key, value []byte) error {
	w := c.Writer(topic)
	reqID := grpcutil.RequestIDFromContext(ctx)
	headers := []kafka.Header{
		{Key: "request-id", Value: []byte(reqID)},
	}
	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Time:    time.Now(),
		Headers: headers,
	}
	return w.WriteMessages(ctx, msg)
}

// Reader builds a new Kafka reader for the given topic/group.
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

// EnsureTopic creates the topic if it does not exist.
func (c *Client) EnsureTopic(topic string, numPartitions, replicationFactor int) error {
	if len(c.brokers) == 0 {
		return nil
	}
	conn, err := kafka.Dial("tcp", c.brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	cc, err := kafka.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer cc.Close()
	return cc.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	})
}
