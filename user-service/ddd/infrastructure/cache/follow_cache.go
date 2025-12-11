package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultEdgeTTL   = 10 * time.Minute
	defaultCountTTL  = 5 * time.Minute
	defaultListTTL   = 2 * time.Minute
	followingStatus  = "1"
	notFollowingMark = "0"
)

// FollowCache caches follow relationships and relation counts in Redis.
type FollowCache struct {
	cli      redis.Cmdable
	edgeTTL  time.Duration
	countTTL time.Duration
	listTTL  time.Duration
}

type CachedFollowItem struct {
	UserUUID  string `json:"user_uuid"`
	CreatedAt int64  `json:"created_at"` // unix nano
	ID        uint64 `json:"id"`
}

func NewFollowCache(cli redis.Cmdable) *FollowCache {
	return &FollowCache{
		cli:      cli,
		edgeTTL:  defaultEdgeTTL,
		countTTL: defaultCountTTL,
		listTTL:  defaultListTTL,
	}
}

func (c *FollowCache) edgeKey(userUUID, targetUUID string) string {
	return fmt.Sprintf("follow:edge:%s:%s", userUUID, targetUUID)
}

func (c *FollowCache) followerCountKey(userUUID string) string {
	return fmt.Sprintf("follow:count:follower:%s", userUUID)
}

func (c *FollowCache) followingCountKey(userUUID string) string {
	return fmt.Sprintf("follow:count:following:%s", userUUID)
}

// GetEdge returns (following, found, error).
func (c *FollowCache) GetEdge(ctx context.Context, userUUID, targetUUID string) (bool, bool, error) {
	val, err := c.cli.Get(ctx, c.edgeKey(userUUID, targetUUID)).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return val == followingStatus, true, nil
}

func (c *FollowCache) SetEdge(ctx context.Context, userUUID, targetUUID string, following bool) error {
	val := notFollowingMark
	if following {
		val = followingStatus
	}
	return c.cli.Set(ctx, c.edgeKey(userUUID, targetUUID), val, c.edgeTTL).Err()
}

func (c *FollowCache) DeleteEdge(ctx context.Context, userUUID, targetUUID string) error {
	return c.cli.Del(ctx, c.edgeKey(userUUID, targetUUID)).Err()
}

// GetFollowerCount returns (count, found, error).
func (c *FollowCache) GetFollowerCount(ctx context.Context, userUUID string) (int64, bool, error) {
	return c.getCount(ctx, c.followerCountKey(userUUID))
}

// GetFollowingCount returns (count, found, error).
func (c *FollowCache) GetFollowingCount(ctx context.Context, userUUID string) (int64, bool, error) {
	return c.getCount(ctx, c.followingCountKey(userUUID))
}

func (c *FollowCache) SetFollowerCount(ctx context.Context, userUUID string, count int64) error {
	return c.setCount(ctx, c.followerCountKey(userUUID), count)
}

func (c *FollowCache) SetFollowingCount(ctx context.Context, userUUID string, count int64) error {
	return c.setCount(ctx, c.followingCountKey(userUUID), count)
}

// InvalidateCounts deletes both follower and following counters for given users.
func (c *FollowCache) InvalidateCounts(ctx context.Context, userUUIDs ...string) {
	keys := make([]string, 0, len(userUUIDs)*2)
	for _, u := range userUUIDs {
		if u == "" {
			continue
		}
		keys = append(keys, c.followerCountKey(u), c.followingCountKey(u))
	}
	if len(keys) == 0 {
		return
	}
	_ = c.cli.Del(ctx, keys...).Err()
}

func (c *FollowCache) followerListKey(userUUID string) string {
	return fmt.Sprintf("follow:list:follower:%s", userUUID)
}

func (c *FollowCache) followingListKey(userUUID string) string {
	return fmt.Sprintf("follow:list:following:%s", userUUID)
}

func (c *FollowCache) GetFollowerList(ctx context.Context, targetUUID string) ([]CachedFollowItem, bool, error) {
	return c.getList(ctx, c.followerListKey(targetUUID))
}

func (c *FollowCache) SetFollowerList(ctx context.Context, targetUUID string, items []CachedFollowItem) error {
	return c.setList(ctx, c.followerListKey(targetUUID), items)
}

func (c *FollowCache) GetFollowingList(ctx context.Context, userUUID string) ([]CachedFollowItem, bool, error) {
	return c.getList(ctx, c.followingListKey(userUUID))
}

func (c *FollowCache) SetFollowingList(ctx context.Context, userUUID string, items []CachedFollowItem) error {
	return c.setList(ctx, c.followingListKey(userUUID), items)
}

func (c *FollowCache) InvalidateLists(ctx context.Context, userUUIDs ...string) {
	keys := make([]string, 0, len(userUUIDs)*2)
	for _, u := range userUUIDs {
		if u == "" {
			continue
		}
		keys = append(keys, c.followerListKey(u), c.followingListKey(u))
	}
	if len(keys) == 0 {
		return
	}
	_ = c.cli.Del(ctx, keys...).Err()
}

func (c *FollowCache) getCount(ctx context.Context, key string) (int64, bool, error) {
	val, err := c.cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	count, convErr := strconv.ParseInt(val, 10, 64)
	if convErr != nil {
		_ = c.cli.Del(ctx, key).Err() // cleanup corrupted value
		return 0, false, convErr
	}
	return count, true, nil
}

func (c *FollowCache) setCount(ctx context.Context, key string, count int64) error {
	return c.cli.Set(ctx, key, count, c.countTTL).Err()
}

func (c *FollowCache) getList(ctx context.Context, key string) ([]CachedFollowItem, bool, error) {
	val, err := c.cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var items []CachedFollowItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		_ = c.cli.Del(ctx, key).Err()
		return nil, false, err
	}
	return items, true, nil
}

func (c *FollowCache) setList(ctx context.Context, key string, items []CachedFollowItem) error {
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.cli.Set(ctx, key, b, c.listTTL).Err()
}
