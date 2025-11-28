package kafka

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"transcode-service/pkg/config"
	"transcode-service/pkg/logger"

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

type rewriteResolver struct{}

func (rewriteResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return net.DefaultResolver.LookupHost(ctx, "host.docker.internal")
	default:
		return net.DefaultResolver.LookupHost(ctx, host)
	}
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
	c.dialer.Resolver = rewriteResolver{}
	logger.Infof("Kafka client opened brokers=%v client_id=%s", c.brokers, c.clientID)
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
	logger.Infof("Kafka reader created topic=%s group=%s brokers=%v", topic, groupID, c.brokers)
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
