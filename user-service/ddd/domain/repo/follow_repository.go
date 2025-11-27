package repo

import (
	"context"
	"user-service/ddd/infrastructure/database/po"
)

type FollowRepository interface {
	Follow(ctx context.Context, userUUID, targetUUID string) error
	Unfollow(ctx context.Context, userUUID, targetUUID string) error
	IsFollowing(ctx context.Context, userUUID, targetUUID string) (bool, error)
	ListFollowers(ctx context.Context, targetUUID string, offset, limit int) ([]*po.FollowPo, int64, error)
	ListFollowings(ctx context.Context, userUUID string, offset, limit int) ([]*po.FollowPo, int64, error)
}
