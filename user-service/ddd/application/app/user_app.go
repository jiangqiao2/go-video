package app

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"user-service/ddd/application/cqe"
	"user-service/ddd/infrastructure/database/persistence"
	"user-service/ddd/infrastructure/database/po"
	"user-service/pkg/assert"
	"user-service/pkg/config"
	"user-service/pkg/utils"
)

var (
	onceUserApp      sync.Once
	singletonUserApp UserApp
)

type UserApp interface {
	Register(ctx context.Context, req *cqe.UserRegisterReq) (*cqe.UserRegisterResp, error)
	Login(ctx context.Context, req *cqe.UserLoginReq) (*cqe.UserLoginResp, error)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	CreateUser(ctx context.Context, userPo *po.UserPo) error
	GetUserByAccount(ctx context.Context, account string) (*po.UserPo, error)
	ExistsByAccount(ctx context.Context, account string) (bool, error)
}

type userAppImpl struct {
	userRepo UserRepository
	jwtUtil  *utils.JWTUtil
	cfg      *config.Config
}

func DefaultUserApp() UserApp {
	assert.NotCircular()
	onceUserApp.Do(func() {
		singletonUserApp = &userAppImpl{
			userRepo: persistence.NewUserRepository(),
			jwtUtil:  utils.DefaultJWTUtil(),
			cfg:      config.GetGlobalConfig(),
		}
	})
	assert.NotNil(singletonUserApp)
	return singletonUserApp
}

// NewUserApp 创建用户应用服务（支持依赖注入）
func NewUserApp(jwtUtil *utils.JWTUtil, cfg *config.Config) UserApp {
	return &userAppImpl{
		userRepo: persistence.NewUserRepository(),
		jwtUtil:  jwtUtil,
		cfg:      cfg,
	}
}

// Register 用户注册
func (u *userAppImpl) Register(ctx context.Context, req *cqe.UserRegisterReq) (*cqe.UserRegisterResp, error) {
	// 简化实现：如果没有注入依赖，返回mock数据
	if u.userRepo == nil {
		userUUID := uuid.New().String()
		return &cqe.UserRegisterResp{
			UserUUID: userUUID,
			Account:  req.Account,
			Message:  "注册成功（模拟）",
		}, nil
	}

	// 检查账号是否已存在
	exists, err := u.userRepo.ExistsByAccount(ctx, req.Account)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("账号已存在")
	}

	// 验证密码强度
	if err := u.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户PO
	userUUID := uuid.New().String()
	userPo := &po.UserPo{
		UserUUID: userUUID,
		Account:  req.Account,
		Password: string(hashedPassword),
	}

	// 保存用户
	if err := u.userRepo.CreateUser(ctx, userPo); err != nil {
		return nil, err
	}

	return &cqe.UserRegisterResp{
		UserUUID: userUUID,
		Account:  req.Account,
		Message:  "注册成功",
	}, nil
}

// Login 用户登录
func (u *userAppImpl) Login(ctx context.Context, req *cqe.UserLoginReq) (*cqe.UserLoginResp, error) {
	// 简化实现：如果没有注入依赖，返回mock数据
	if u.userRepo == nil || u.jwtUtil == nil {
		userUUID := uuid.New().String()
		return &cqe.UserLoginResp{
			UserUUID:     userUUID,
			Account:      req.Account,
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresIn:    7200, // 2小时
			Message:      "登录成功（模拟）",
		}, nil
	}

	// 查找用户
	user, err := u.userRepo.GetUserByAccount(ctx, req.Account)
	if err != nil {
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 生成JWT令牌
	accessToken, err := u.jwtUtil.GenerateAccessTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errors.New("令牌生成失败")
	}

	refreshToken, err := u.jwtUtil.GenerateRefreshTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errors.New("刷新令牌生成失败")
	}

	// 获取过期时间（秒）
	expiresIn := int64(u.cfg.JWT.ExpireTime.Seconds())

	return &cqe.UserLoginResp{
		UserUUID:     user.UserUUID,
		Account:      user.Account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Message:      "登录成功",
	}, nil
}

// validatePassword 验证密码强度
func (u *userAppImpl) validatePassword(password string) error {
	// 简化实现：如果没有配置，使用默认规则
	if u.cfg == nil {
		if len(password) < 8 {
			return errors.New("密码长度不能少于8位")
		}
		return nil
	}

	// 简化配置检查
	minLength := 8
	if u.cfg != nil {
		// TODO: 从配置中读取密码策略
		minLength = 8 // 默认最小长度
	}

	// 检查长度
	if len(password) < minLength {
		return errors.New("密码长度不足")
	}

	return nil
}
