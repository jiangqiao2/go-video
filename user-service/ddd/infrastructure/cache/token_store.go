package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"user-service/pkg/revocation"
)

type RedisRevocationStore struct {
	cli redis.Cmdable
}

func NewRedisRevocationStore(cli redis.Cmdable) revocation.Store {
	return &RedisRevocationStore{cli: cli}
}

func versionKey(userUUID string) string {
	return fmt.Sprintf("auth:ver:%s", userUUID)
}

func refreshKey(userUUID string, tokenHash string) string {
	return fmt.Sprintf("auth:refresh:%s:%s", userUUID, tokenHash)
}

func (r *RedisRevocationStore) GetVersion(ctx context.Context, userUUID string) (int64, error) {
	s, err := r.cli.Get(ctx, versionKey(userUUID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var v int64
	_, _ = fmt.Sscan(s, &v)
	return v, nil
}

func (r *RedisRevocationStore) IncrementVersion(ctx context.Context, userUUID string) (int64, error) {
	return r.cli.Incr(ctx, versionKey(userUUID)).Val(), nil
}

func (r *RedisRevocationStore) StoreRefreshToken(ctx context.Context, userUUID string, tokenHash string, ttl time.Duration) error {
	return r.cli.Set(ctx, refreshKey(userUUID, tokenHash), "1", ttl).Err()
}

func (r *RedisRevocationStore) DeleteRefreshToken(ctx context.Context, userUUID string, tokenHash string) error {
	return r.cli.Del(ctx, refreshKey(userUUID, tokenHash)).Err()
}

func (r *RedisRevocationStore) ExistsRefreshToken(ctx context.Context, userUUID string, tokenHash string) (bool, error) {
	n, err := r.cli.Exists(ctx, refreshKey(userUUID, tokenHash)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
