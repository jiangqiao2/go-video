package repo

import (
	"context"
	"user-service/ddd/infrastructure/database/po"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	CreateUser(ctx context.Context, userPo *po.UserPo) error
	GetUserByAccount(ctx context.Context, account string) (*po.UserPo, error)
	GetUserByUUID(ctx context.Context, userUUID string) (*po.UserPo, error)
	ExistsByAccount(ctx context.Context, account string) (bool, error)
}
