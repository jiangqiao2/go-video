package redisclient

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"

	"notification-service/pkg/config"
)

// Client wraps a go-redis client instance.
type Client struct {
	native *redis.Client
}

// New creates a new redis client based on the provided configuration.
// It performs an initial PING to fail fast if redis is unavailable.
func New(cfg config.RedisConfig) (*Client, error) {
	opts := &redis.Options{
		Addr: cfg.GetRedisAddr(),
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB != 0 {
		opts.DB = cfg.DB
	}
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}

	opts.DialTimeout = pickDuration(cfg.DialTimeout, 5*time.Second)
	opts.ReadTimeout = pickDuration(cfg.ReadTimeout, 3*time.Second)
	opts.WriteTimeout = pickDuration(cfg.WriteTimeout, 3*time.Second)

	if cfg.EnableTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	cli := redis.NewClient(opts)
	if err := cli.Ping(context.Background()).Err(); err != nil {
		_ = cli.Close()
		return nil, err
	}

	return &Client{native: cli}, nil
}

// Raw returns the underlying client for direct command execution.
func (c *Client) Raw() *redis.Client {
	return c.native
}

// Close releases all resources used by the client.
func (c *Client) Close() error {
	return c.native.Close()
}

func pickDuration(v time.Duration, fallback time.Duration) time.Duration {
	if v <= 0 {
		return fallback
	}
	return v
}
