package resource

import (
	"sync"

	"github.com/redis/go-redis/v9"

	"user-service/pkg/assert"
	"user-service/pkg/config"
	"user-service/pkg/manager"
	"user-service/pkg/redisclient"
)

var (
	redisResourceOnce sync.Once
	redisSingleton    *RedisResource
)

// RedisResource manages the lifecycle of the redis client.
type RedisResource struct {
	client *redisclient.Client
}

// DefaultRedisResource returns the singleton redis resource instance.
func DefaultRedisResource() *RedisResource {
	assert.NotCircular()
	redisResourceOnce.Do(func() {
		redisSingleton = &RedisResource{}
	})
	assert.NotNil(redisSingleton)
	return redisSingleton
}

// MustOpen connects to redis using global configuration.
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

// Close terminates the redis connection pool.
func (r *RedisResource) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}
}

// Client exposes the raw go-redis client.
func (r *RedisResource) Client() *redis.Client {
	if r.client == nil {
		return nil
	}
	return r.client.Raw()
}

// RedisResourcePlugin registers the redis resource with the manager.
type RedisResourcePlugin struct{}

// Name returns the plugin name.
func (p *RedisResourcePlugin) Name() string {
	return "redis"
}

// MustCreateResource provides the redis resource instance.
func (p *RedisResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultRedisResource()
}
