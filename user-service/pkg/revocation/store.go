package revocation

import (
	"context"
	"sync"
	"time"
)

type Store interface {
	GetVersion(ctx context.Context, userUUID string) (int64, error)
	IncrementVersion(ctx context.Context, userUUID string) (int64, error)
	StoreRefreshToken(ctx context.Context, userUUID string, tokenHash string, ttl time.Duration) error
	DeleteRefreshToken(ctx context.Context, userUUID string, tokenHash string) error
	// DeleteAllRefreshTokens 删除该用户的所有刷新令牌记录（用于简单单端登录）
	DeleteAllRefreshTokens(ctx context.Context, userUUID string) error
	ExistsRefreshToken(ctx context.Context, userUUID string, tokenHash string) (bool, error)
}

var (
	once     sync.Once
	instance Store
)

func DefaultRevocationStore() Store {
	return instance
}

func Init(store Store) {
	once.Do(func() {
		instance = store
	})
}
