package persistence

import (
	"context"
	"go-video/ddd/user/domain/entity"
	"go-video/ddd/user/domain/repo"
	"go-video/ddd/user/domain/vo"
	"go-video/ddd/user/infrastructure/database/convertor"
	"go-video/ddd/user/infrastructure/database/dao"
)

// userRepositoryImpl 用户仓储实现
type userRepositoryImpl struct {
	userDao       *dao.UserDao
	userConvertor *convertor.UserConvertor
}

// NewUserRepository 创建用户仓储实例（支持依赖注入）
func NewUserRepository() repo.UserRepository {
	userDao := dao.NewUserDao()
	userConvertor := convertor.NewUserConvertor()
	return &userRepositoryImpl{
		userDao:       userDao,
		userConvertor: userConvertor,
	}
}

// Save 保存用户
func (r *userRepositoryImpl) Save(ctx context.Context, user *entity.User) error {
	userPO := r.userConvertor.ToPO(user)
	if err := r.userDao.Create(ctx, userPO); err != nil {
		return err
	}
	// 设置生成的ID和UUID
	user.SetID(userPO.Id)
	user.SetUUID(userPO.UUID)
	return nil
}

// FindByID 根据内部ID查找用户（仅内部使用）
func (r *userRepositoryImpl) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	userPO, err := r.userDao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntity(userPO), nil
}

// FindByUUID 根据UUID查找用户
func (r *userRepositoryImpl) FindByUUID(ctx context.Context, uuid string) (*entity.User, error) {
	userPO, err := r.userDao.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntity(userPO), nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	userPO, err := r.userDao.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntity(userPO), nil
}

// FindByEmail 根据邮箱查找用户
func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	userPO, err := r.userDao.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntity(userPO), nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.userDao.ExistsByUsername(ctx, username)
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.userDao.ExistsByEmail(ctx, email)
}

// FindByPage 分页查询用户
func (r *userRepositoryImpl) FindByPage(ctx context.Context, page *vo.Page) ([]*entity.User, error) {
	userPOs, err := r.userDao.GetByPage(ctx, page.Offset(), page.Limit())
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntities(userPOs), nil
}

// Count 统计用户总数
func (r *userRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.userDao.Count(ctx)
}

// Update 更新用户
func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	userPO := r.userConvertor.ToPO(user)
	return r.userDao.Update(ctx, userPO)
}

// Delete 删除用户
func (r *userRepositoryImpl) Delete(ctx context.Context, id uint64) error {
	return r.userDao.Delete(ctx, id)
}

// FindActiveUsers 查找激活用户
func (r *userRepositoryImpl) FindActiveUsers(ctx context.Context, page *vo.Page) ([]*entity.User, error) {
	userPOs, err := r.userDao.GetActiveUsersByPage(ctx, page.Offset(), page.Limit())
	if err != nil {
		return nil, err
	}
	return r.userConvertor.ToEntities(userPOs), nil
}

// CountActiveUsers 统计激活用户数量
func (r *userRepositoryImpl) CountActiveUsers(ctx context.Context) (int64, error) {
	return r.userDao.CountActiveUsers(ctx)
}
