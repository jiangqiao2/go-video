package persistence

import (
	"context"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/database/dao"
	"user-service/ddd/infrastructure/database/po"
	"user-service/pkg/errno"
)

// userRepositoryImpl 用户仓储实现
type userRepositoryImpl struct {
	userDao *dao.UserDao
}

// NewUserRepository 创建用户仓储
func NewUserRepository() repo.UserRepository {
	return &userRepositoryImpl{
		userDao: dao.NewUserDao(),
	}
}

// CreateUser 创建用户
func (r *userRepositoryImpl) CreateUser(ctx context.Context, userPo *po.UserPo) error {
	return r.userDao.Create(ctx, userPo)
}

// GetUserByAccount 根据账号获取用户
func (r *userRepositoryImpl) GetUserByAccount(ctx context.Context, account string) (*po.UserPo, error) {
	user, err := r.userDao.QueryByAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errno.ErrUserNotFound
	}
	return user, nil
}

// GetUserByUUID 根据UUID获取用户
func (r *userRepositoryImpl) GetUserByUUID(ctx context.Context, userUUID string) (*po.UserPo, error) {
	user, err := r.userDao.QueryByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errno.ErrUserNotFound
	}
	return user, nil
}

// GetUserByID 根据ID获取用户
func (r *userRepositoryImpl) GetUserByID(ctx context.Context, id uint64) (*po.UserPo, error) {
	user, err := r.userDao.QueryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errno.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser 更新用户
func (r *userRepositoryImpl) UpdateUser(ctx context.Context, userPo *po.UserPo) error {
	return r.userDao.Update(ctx, userPo)
}

// DeleteUser 删除用户
func (r *userRepositoryImpl) DeleteUser(ctx context.Context, userUUID string) error {
	return r.userDao.DeleteByUUID(ctx, userUUID)
}

// ExistsByAccount 检查账号是否存在
func (r *userRepositoryImpl) ExistsByAccount(ctx context.Context, account string) (bool, error) {
	return r.userDao.ExistsByAccount(ctx, account)
}

// ExistsByUUID 检查UUID是否存在
func (r *userRepositoryImpl) ExistsByUUID(ctx context.Context, userUUID string) (bool, error) {
	return r.userDao.ExistsByUUID(ctx, userUUID)
}
