package redisclient

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"

	"transcode-service/pkg/config"
)

// Client wraps the go-redis client to allow tailored helpers.
type Client struct {
	native *redis.Client
}

// New builds a redis client using service configuration and validates the connection.
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

// Raw exposes the underlying go-redis client for advanced use cases.
func (c *Client) Raw() *redis.Client {
	return c.native
}

// Close stops the redis client and releases pooled connections.
func (c *Client) Close() error {
	return c.native.Close()
}

func pickDuration(v time.Duration, fallback time.Duration) time.Duration {
	if v <= 0 {
		return fallback
	}
	return v
}
