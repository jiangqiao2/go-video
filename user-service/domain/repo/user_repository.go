package repo

import (
	"context"

	"go-video/ddd/user/domain/entity"
	"go-video/ddd/user/domain/vo"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Save 保存用户
	Save(ctx context.Context, user *entity.User) error

	// FindByID 根据内部ID查找用户（仅内部使用）
	FindByID(ctx context.Context, id uint64) (*entity.User, error)

	// FindByUUID 根据UUID查找用户
	FindByUUID(ctx context.Context, uuid string) (*entity.User, error)

	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// FindByPage 分页查询用户
	FindByPage(ctx context.Context, page *vo.Page) ([]*entity.User, error)

	// Count 统计用户总数
	Count(ctx context.Context) (int64, error)

	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error

	// Delete 删除用户
	Delete(ctx context.Context, id uint64) error

	// FindActiveUsers 查找激活用户
	FindActiveUsers(ctx context.Context, page *vo.Page) ([]*entity.User, error)

	// CountActiveUsers 统计激活用户数量
	CountActiveUsers(ctx context.Context) (int64, error)
}