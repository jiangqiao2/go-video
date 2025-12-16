package persistence

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/cache"
	"user-service/ddd/infrastructure/database/dao"
	"user-service/ddd/infrastructure/database/po"
	"user-service/internal/resource"
)

type followRepositoryImpl struct {
	dao   *dao.FollowDao
	cache *cache.FollowCache
}

func NewFollowRepository() repo.FollowRepository {
	var followCache *cache.FollowCache
	if cli := resource.DefaultRedisResource().Client(); cli != nil {
		followCache = cache.NewFollowCache(cli)
	}
	return &followRepositoryImpl{
		dao:   dao.NewFollowDao(),
		cache: followCache,
	}
}

func (r *followRepositoryImpl) Follow(ctx context.Context, userUUID, targetUUID string) error {
	if err := r.dao.Upsert(ctx, &po.FollowPo{UserUUID: userUUID, TargetUUID: targetUUID, Status: "Following"}); err != nil {
		return err
	}
	r.afterFollow(ctx, userUUID, targetUUID)
	r.setEdge(ctx, userUUID, targetUUID, true)
	return nil
}

func (r *followRepositoryImpl) Unfollow(ctx context.Context, userUUID, targetUUID string) error {
	if err := r.dao.UpdateStatus(ctx, userUUID, targetUUID, "Unfollowed"); err != nil {
		return err
	}
	r.afterUnfollow(ctx, userUUID, targetUUID)
	r.deleteEdge(ctx, userUUID, targetUUID)
	return nil
}

func (r *followRepositoryImpl) IsFollowing(ctx context.Context, userUUID, targetUUID string) (bool, error) {
	if r.cache != nil {
		if following, ok, err := r.cache.GetEdge(ctx, userUUID, targetUUID); err == nil && ok {
			return following, nil
		}
	}
	following, err := r.dao.Exists(ctx, userUUID, targetUUID)
	if err != nil {
		return false, err
	}
	r.setEdge(ctx, userUUID, targetUUID, following)
	return following, nil
}

func (r *followRepositoryImpl) ListFollowers(ctx context.Context, targetUUID string, cursor string, limit int) ([]*po.FollowPo, int64, error) {
	if r.cache != nil && cursor == "" {
		if cached, ok, err := r.cache.GetFollowerList(ctx, targetUUID); err == nil && ok {
			if len(cached) == 0 {
				return []*po.FollowPo{}, 0, nil
			}
			list := toFollowPoList(cached, limit, func(item cache.CachedFollowItem) *po.FollowPo {
				return &po.FollowPo{BaseModel: po.BaseModel{Id: item.ID, CreatedAt: time.Unix(0, item.CreatedAt)}, UserUUID: item.UserUUID, TargetUUID: targetUUID, Status: "Following"}
			})
			total, _ := r.CountFollowers(ctx, targetUUID)
			return list, total, nil
		}
	}
	cursorTime, cursorID := parseCursor(cursor)
	list, total, err := r.dao.QueryFollowers(ctx, targetUUID, cursorTime, cursorID, limit)
	if err == nil && r.cache != nil && cursor == "" {
		_ = r.cache.SetFollowerList(ctx, targetUUID, toCachedItems(list))
	}
	return list, total, err
}

func (r *followRepositoryImpl) ListFollowings(ctx context.Context, userUUID string, cursor string, limit int) ([]*po.FollowPo, int64, error) {
	if r.cache != nil && cursor == "" {
		if cached, ok, err := r.cache.GetFollowingList(ctx, userUUID); err == nil && ok {
			if len(cached) == 0 {
				return []*po.FollowPo{}, 0, nil
			}
			list := toFollowPoList(cached, limit, func(item cache.CachedFollowItem) *po.FollowPo {
				return &po.FollowPo{BaseModel: po.BaseModel{Id: item.ID, CreatedAt: time.Unix(0, item.CreatedAt)}, UserUUID: userUUID, TargetUUID: item.UserUUID, Status: "Following"}
			})
			total, _ := r.CountFollowings(ctx, userUUID)
			return list, total, nil
		}
	}
	cursorTime, cursorID := parseCursor(cursor)
	list, total, err := r.dao.QueryFollowings(ctx, userUUID, cursorTime, cursorID, limit)
	if err == nil && r.cache != nil && cursor == "" {
		_ = r.cache.SetFollowingList(ctx, userUUID, toCachedItems(list))
	}
	return list, total, err
}

func (r *followRepositoryImpl) CountFollowers(ctx context.Context, targetUUID string) (int64, error) {
	if r.cache != nil {
		if v, ok, err := r.cache.GetFollowerCount(ctx, targetUUID); err == nil && ok {
			return v, nil
		}
	}
	count, err := r.dao.CountFollowers(ctx, targetUUID)
	if err == nil {
		r.setFollowerCount(ctx, targetUUID, count)
	}
	return count, err
}

func (r *followRepositoryImpl) CountFollowings(ctx context.Context, userUUID string) (int64, error) {
	if r.cache != nil {
		if v, ok, err := r.cache.GetFollowingCount(ctx, userUUID); err == nil && ok {
			return v, nil
		}
	}
	count, err := r.dao.CountFollowings(ctx, userUUID)
	if err == nil {
		r.setFollowingCount(ctx, userUUID, count)
	}
	return count, err
}

func (r *followRepositoryImpl) invalidateCounts(ctx context.Context, userUUID, targetUUID string) {
	if r.cache == nil {
		return
	}
	r.cache.InvalidateCounts(ctx, userUUID, targetUUID)
	r.cache.InvalidateLists(ctx, userUUID, targetUUID)
}

// afterFollow updates cached counters and lists for a successful follow.
// Counters are updated incrementally to avoid forcing DB COUNT on every write.
func (r *followRepositoryImpl) afterFollow(ctx context.Context, userUUID, targetUUID string) {
	if r.cache == nil {
		return
	}
	// Increment counts; occasional drift is corrected by Count* when cache miss happens.
	_ = r.cache.IncrFollowingCount(ctx, userUUID, 1)
	_ = r.cache.IncrFollowerCount(ctx, targetUUID, 1)
	// Lists are still invalidated and rebuilt lazily on read.
	r.cache.InvalidateLists(ctx, userUUID, targetUUID)
}

// afterUnfollow updates cached counters and lists for a successful unfollow.
func (r *followRepositoryImpl) afterUnfollow(ctx context.Context, userUUID, targetUUID string) {
	if r.cache == nil {
		return
	}
	_ = r.cache.IncrFollowingCount(ctx, userUUID, -1)
	_ = r.cache.IncrFollowerCount(ctx, targetUUID, -1)
	r.cache.InvalidateLists(ctx, userUUID, targetUUID)
}

func (r *followRepositoryImpl) setEdge(ctx context.Context, userUUID, targetUUID string, following bool) {
	if r.cache == nil {
		return
	}
	_ = r.cache.SetEdge(ctx, userUUID, targetUUID, following)
}

func (r *followRepositoryImpl) deleteEdge(ctx context.Context, userUUID, targetUUID string) {
	if r.cache == nil {
		return
	}
	_ = r.cache.DeleteEdge(ctx, userUUID, targetUUID)
}

func (r *followRepositoryImpl) setFollowerCount(ctx context.Context, userUUID string, count int64) {
	if r.cache == nil {
		return
	}
	_ = r.cache.SetFollowerCount(ctx, userUUID, count)
}

func (r *followRepositoryImpl) setFollowingCount(ctx context.Context, userUUID string, count int64) {
	if r.cache == nil {
		return
	}
	_ = r.cache.SetFollowingCount(ctx, userUUID, count)
}

func toCachedItems(list []*po.FollowPo) []cache.CachedFollowItem {
	items := make([]cache.CachedFollowItem, 0, len(list))
	for _, v := range list {
		if v == nil {
			continue
		}
		items = append(items, cache.CachedFollowItem{
			UserUUID:  v.UserUUID,
			CreatedAt: v.CreatedAt.UnixNano(),
			ID:        v.Id,
		})
	}
	return items
}

func toFollowPoList(items []cache.CachedFollowItem, limit int, builder func(item cache.CachedFollowItem) *po.FollowPo) []*po.FollowPo {
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	res := make([]*po.FollowPo, 0, len(items))
	for _, item := range items {
		res = append(res, builder(item))
	}
	return res
}

// cursor format: unixnano:id (both int). Empty means no cursor.
func parseCursor(cursor string) (time.Time, uint64) {
	if cursor == "" {
		return time.Time{}, 0
	}
	parts := strings.Split(cursor, ":")
	if len(parts) != 2 {
		return time.Time{}, 0
	}
	ns, err1 := strconv.ParseInt(parts[0], 10, 64)
	id, err2 := strconv.ParseUint(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return time.Time{}, 0
	}
	return time.Unix(0, ns), id
}

func makeCursor(p *po.FollowPo) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d:%d", p.CreatedAt.UnixNano(), p.Id)
}
