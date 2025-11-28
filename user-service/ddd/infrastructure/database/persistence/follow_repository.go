package persistence

import (
	"context"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/database/dao"
	"user-service/ddd/infrastructure/database/po"
)

type followRepositoryImpl struct {
	dao *dao.FollowDao
}

func NewFollowRepository() repo.FollowRepository {
	return &followRepositoryImpl{dao: dao.NewFollowDao()}
}

func (r *followRepositoryImpl) Follow(ctx context.Context, userUUID, targetUUID string) error {
	return r.dao.Upsert(ctx, &po.FollowPo{UserUUID: userUUID, TargetUUID: targetUUID, Status: "Following"})
}

func (r *followRepositoryImpl) Unfollow(ctx context.Context, userUUID, targetUUID string) error {
	return r.dao.UpdateStatus(ctx, userUUID, targetUUID, "Unfollowed")
}

func (r *followRepositoryImpl) IsFollowing(ctx context.Context, userUUID, targetUUID string) (bool, error) {
	return r.dao.Exists(ctx, userUUID, targetUUID)
}

func (r *followRepositoryImpl) ListFollowers(ctx context.Context, targetUUID string, offset, limit int) ([]*po.FollowPo, int64, error) {
	return r.dao.QueryFollowers(ctx, targetUUID, offset, limit)
}

func (r *followRepositoryImpl) ListFollowings(ctx context.Context, userUUID string, offset, limit int) ([]*po.FollowPo, int64, error) {
	return r.dao.QueryFollowings(ctx, userUUID, offset, limit)
}
