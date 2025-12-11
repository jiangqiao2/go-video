package repo

import (
	"context"
	"user-service/ddd/infrastructure/database/po"
)

type FollowRepository interface {
	Follow(ctx context.Context, userUUID, targetUUID string) error
	Unfollow(ctx context.Context, userUUID, targetUUID string) error
	IsFollowing(ctx context.Context, userUUID, targetUUID string) (bool, error)
	ListFollowers(ctx context.Context, targetUUID string, cursor string, limit int) ([]*po.FollowPo, int64, error)
	ListFollowings(ctx context.Context, userUUID string, cursor string, limit int) ([]*po.FollowPo, int64, error)
	CountFollowers(ctx context.Context, targetUUID string) (int64, error)
	CountFollowings(ctx context.Context, userUUID string) (int64, error)
}
