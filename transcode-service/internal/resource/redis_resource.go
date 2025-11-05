package resource

import (
	"sync"

	"github.com/redis/go-redis/v9"

	"transcode-service/pkg/assert"
	"transcode-service/pkg/config"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/redisclient"
)

var (
	redisResourceOnce sync.Once
	redisSingleton    *RedisResource
)

// RedisResource manages the lifecycle of the shared Redis client.
type RedisResource struct {
	client *redisclient.Client
}

// DefaultRedisResource returns the global Redis resource instance.
func DefaultRedisResource() *RedisResource {
	assert.NotCircular()
	redisResourceOnce.Do(func() {
		redisSingleton = &RedisResource{}
	})
	assert.NotNil(redisSingleton)
	return redisSingleton
}

// MustOpen establishes the Redis connection using global configuration.
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

// Close tidy ups the underlying Redis client.
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

// RedisResourcePlugin wires the resource into the manager.
type RedisResourcePlugin struct{}

// Name identifies the plugin slot.
func (p *RedisResourcePlugin) Name() string {
	return "redis"
}

// MustCreateResource returns the singleton Redis resource for registration.
func (p *RedisResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultRedisResource()
}
