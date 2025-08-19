package service

import (
	"context"
	"sync"

	"go-video/ddd/user/domain/entity"
	"go-video/ddd/user/domain/repo"
	"go-video/ddd/user/domain/vo"
	"go-video/ddd/user/infrastructure/database/persistence"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
)

var (
	userDomainServiceOnce      sync.Once
	singletonUserDomainService *UserService
)

// UserService 用户领域服务
type UserService struct {
	userRepo repo.UserRepository
}

// DefaultUserService 获取用户领域服务单例
func DefaultUserService() *UserService {
	assert.NotCircular()
	userDomainServiceOnce.Do(func() {
		userRepo := persistence.NewUserRepository()
		singletonUserDomainService = &UserService{
			userRepo: userRepo,
		}
	})
	assert.NotNil(singletonUserDomainService)
	return singletonUserDomainService
}

// NewUserService 创建用户领域服务实例（支持依赖注入）
func NewUserService(userRepo repo.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, username, password, email string) (*entity.User, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "检查用户名失败")
	}
	if exists {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "用户名已存在")
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "检查邮箱失败")
	}
	if exists {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "邮箱已被使用")
	}

	// 创建用户实体
	user, err := entity.NewUser(username, password, email)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "用户信息无效")
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "保存用户失败")
	}

	return user, nil
}

// AuthenticateUser 用户认证
func (s *UserService) AuthenticateUser(ctx context.Context, username, password string) (*entity.User, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "查找用户失败")
	}
	if user == nil {
		return nil, errno.NewSimpleBizError(errno.ErrUnauthorized, nil, "用户名或密码错误")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errno.NewSimpleBizError(errno.ErrUnauthorized, nil, "用户未激活或已被禁用")
	}

	// 验证密码
	if !user.VerifyPassword(password) {
		return nil, errno.NewSimpleBizError(errno.ErrUnauthorized, nil, "用户名或密码错误")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id uint64) (*entity.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "查找用户失败")
	}
	if user == nil {
		return nil, errno.NewSimpleBizError(errno.ErrNotFound, nil, "用户不存在")
	}
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrDatabase, err, "查找用户失败")
	}
	if user == nil {
		return nil, errno.NewSimpleBizError(errno.ErrNotFound, nil, "用户不存在")
	}
	return user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	// 获取用户
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !user.VerifyPassword(oldPassword) {
		return errno.NewSimpleBizError(errno.ErrUnauthorized, nil, "原密码错误")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "新密码无效")
	}

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errno.NewSimpleBizError(errno.ErrDatabase, err, "更新用户失败")
	}

	return nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, nickname, avatar string) error {
	// 获取用户
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// 更新昵称
	if nickname != "" {
		if err := user.SetNickname(nickname); err != nil {
			return errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "昵称无效")
		}
	}

	// 更新头像
	if avatar != "" {
		user.SetAvatar(avatar)
	}

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errno.NewSimpleBizError(errno.ErrDatabase, err, "更新用户失败")
	}

	return nil
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(ctx context.Context, page *vo.Page) ([]*entity.User, int64, error) {
	// 获取用户列表
	users, err := s.userRepo.FindByPage(ctx, page)
	if err != nil {
		return nil, 0, errno.NewSimpleBizError(errno.ErrDatabase, err, "查询用户列表失败")
	}

	// 获取总数
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, errno.NewSimpleBizError(errno.ErrDatabase, err, "查询用户总数失败")
	}

	return users, total, nil
}

// ActivateUser 激活用户
func (s *UserService) ActivateUser(ctx context.Context, userID uint64) error {
	// 获取用户
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// 激活用户
	user.Activate()

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errno.NewSimpleBizError(errno.ErrDatabase, err, "激活用户失败")
	}

	return nil
}

// DisableUser 禁用用户
func (s *UserService) DisableUser(ctx context.Context, userID uint64) error {
	// 获取用户
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// 禁用用户
	user.Disable()

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errno.NewSimpleBizError(errno.ErrDatabase, err, "禁用用户失败")
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID uint64) error {
	// 获取用户
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// 软删除用户
	user.Delete()

	// 更新用户
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errno.NewSimpleBizError(errno.ErrDatabase, err, "删除用户失败")
	}

	return nil
}
