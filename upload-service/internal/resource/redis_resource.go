package resource

import (
	"sync"

	"github.com/redis/go-redis/v9"

	"upload-service/pkg/assert"
	"upload-service/pkg/config"
	"upload-service/pkg/manager"
	"upload-service/pkg/redisclient"
)

var (
	redisResourceOnce sync.Once
	redisSingleton    *RedisResource
)

// RedisResource keeps a singleton redis client for the application.
type RedisResource struct {
	client *redisclient.Client
}

// DefaultRedisResource returns the singleton instance.
func DefaultRedisResource() *RedisResource {
	assert.NotCircular()
	redisResourceOnce.Do(func() {
		redisSingleton = &RedisResource{}
	})
	assert.NotNil(redisSingleton)
	return redisSingleton
}

// MustOpen dials redis using the global configuration.
func (r *RedisResource) MustOpen() {
	if r.client != nil {
		return
	}

	cfg := config.GetGlobalConfig()
	if cfg == nil {
		panic("global config not initialized")
	}

	client, err := redisclient.New(cfg.Redis)
	if err != nil {
		panic("failed to connect redis: " + err.Error())
	}

	r.client = client
}

// Close shuts down the redis client if it was created.
func (r *RedisResource) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}
}

// Client exposes the underlying go-redis client.
func (r *RedisResource) Client() *redis.Client {
	if r.client == nil {
		return nil
	}
	return r.client.Raw()
}

// RedisResourcePlugin registers the Redis resource with the manager.
type RedisResourcePlugin struct{}

// Name returns the unique plugin name.
func (p *RedisResourcePlugin) Name() string {
	return "redis"
}

// MustCreateResource returns the shared redis resource instance.
func (p *RedisResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultRedisResource()
}
