package app

import (
	"context"
	"sync"
	"user-service/ddd/application/cqe"
	"user-service/ddd/application/dto"
	"user-service/ddd/domain/entity"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/database/persistence"
	"user-service/ddd/infrastructure/database/po"
	"user-service/pkg/assert"
	"user-service/pkg/config"
	"user-service/pkg/errno"
	"user-service/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	onceUserApp      sync.Once
	singletonUserApp UserApp
)

type UserApp interface {
	Register(ctx context.Context, req *cqe.UserRegisterReq) (*dto.UserRegisterDto, error)
	Login(ctx context.Context, req *cqe.UserLoginReq) (*dto.UserLoginDto, error)
	GetUserInfo(ctx context.Context, userUUID string) (*dto.UserInfoDto, error)
	GetUserBasicInfo(ctx context.Context, userUUID string) (*dto.UserBasicInfoDto, error)
	SaveUserInfo(ctx context.Context, userUUID string, req *cqe.UserSaveReq) (*dto.UserInfoDto, error)
	RefreshToken(ctx context.Context, req *cqe.TokenRefreshReq) (*dto.TokenRefreshDto, error)
	ChangePassword(ctx context.Context, userUUID string, req *cqe.ChangePasswordReq) error
}

type userAppImpl struct {
	userRepo repo.UserRepository
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
func (u *userAppImpl) Register(ctx context.Context, req *cqe.UserRegisterReq) (*dto.UserRegisterDto, error) {
	// 检查账号是否已存在
	exists, err := u.userRepo.ExistsByAccount(ctx, req.Account)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errno.ErrAccountExists
	}

	// 验证密码强度
	if err := u.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errno.ErrPasswordEncrypt
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

	return &dto.UserRegisterDto{
		UserUUID: userUUID,
		Account:  req.Account,
	}, nil
}

// Login 用户登录
func (u *userAppImpl) Login(ctx context.Context, req *cqe.UserLoginReq) (*dto.UserLoginDto, error) {

	// 查找用户
	user, err := u.userRepo.GetUserByAccount(ctx, req.Account)
	if err != nil {
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// 生成JWT令牌
	accessToken, err := u.jwtUtil.GenerateAccessTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errno.ErrTokenGenerate
	}

	refreshToken, err := u.jwtUtil.GenerateRefreshTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errno.ErrRefreshTokenGenerate
	}

	// 获取过期时间（秒）
	expiresIn := int64(u.cfg.JWT.ExpireTime.Seconds())

	return &dto.UserLoginDto{
		UserUUID:     user.UserUUID,
		Account:      user.Account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		AvatarURL:    user.AvatarUrl,
	}, nil
}

func (u *userAppImpl) RefreshToken(ctx context.Context, req *cqe.TokenRefreshReq) (*dto.TokenRefreshDto, error) {
	userUUID, _, err := u.jwtUtil.ValidateRefreshTokenWithUUID(req.RefreshToken)
	if err != nil || userUUID == "" {
		return nil, errno.ErrUnauthorized
	}
	userPo, err := u.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil || userPo == nil {
		return nil, errno.ErrUserNotFound
	}
	accessToken, err := u.jwtUtil.GenerateAccessTokenWithUUID(userPo.UserUUID, userPo.Id)
	if err != nil {
		return nil, errno.ErrTokenGenerate
	}
	refreshToken, err := u.jwtUtil.GenerateRefreshTokenWithUUID(userPo.UserUUID, userPo.Id)
	if err != nil {
		return nil, errno.ErrRefreshTokenGenerate
	}
	expiresIn := int64(u.cfg.JWT.ExpireTime.Seconds())
	return &dto.TokenRefreshDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// validatePassword 验证密码强度
func (u *userAppImpl) validatePassword(password string) error {
	// 简化实现：如果没有配置，使用默认规则
	if u.cfg == nil {
		if len(password) < 8 {
			return errno.ErrPasswordWeak
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
		return errno.ErrPasswordWeak
	}

	return nil
}

// GetUserInfo 获取用户信息
func (u *userAppImpl) GetUserInfo(ctx context.Context, userUUID string) (*dto.UserInfoDto, error) {
	// 从数据库获取用户PO
	userPo, err := u.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 将PO转换为领域实体
	userEntity := entity.DefaultUserEntity(userPo.UserUUID, userPo.Account, userPo.Password)

	// 将实体转换为响应DTO
	return &dto.UserInfoDto{
		UserUUID:  userEntity.GetUserUUID(),
		Nickname:  userPo.Nickname,
		AvatarUrl: userPo.AvatarUrl,
	}, nil
}

// SaveUserInfo 保存用户信息（部分字段）
func (u *userAppImpl) SaveUserInfo(ctx context.Context, userUUID string, req *cqe.UserSaveReq) (*dto.UserInfoDto, error) {
	// 获取当前用户
	userPo, err := u.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if userPo == nil {
		return nil, errno.ErrUserNotFound
	}
	// 更新账号（如果变更）
	if req.Account != "" && req.Account != userPo.Account {
		exists, err := u.userRepo.ExistsByAccount(ctx, req.Account)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errno.ErrAccountExists
		}
		userPo.Account = req.Account
	}
	if req.Nickname != "" {
		userPo.Nickname = req.Nickname
	}
	if req.AvatarUrl != "" {
		userPo.AvatarUrl = req.AvatarUrl
	}
	if err := u.userRepo.UpdateUser(ctx, userPo); err != nil {
		return nil, err
	}
	return &dto.UserInfoDto{
		UserUUID:  userPo.UserUUID,
		Nickname:  userPo.Nickname,
		AvatarUrl: userPo.AvatarUrl,
	}, nil
}

// ChangePassword 修改密码
func (u *userAppImpl) ChangePassword(ctx context.Context, userUUID string, req *cqe.ChangePasswordReq) error {
	userPo, err := u.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil {
		return err
	}
	if userPo == nil {
		return errno.ErrUserNotFound
	}
	// 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(userPo.Password), []byte(req.OldPassword)); err != nil {
		return errno.ErrPasswordIncorrect
	}
	// 校验新密码
	if err := u.validatePassword(req.NewPassword); err != nil {
		return err
	}
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errno.ErrPasswordEncrypt
	}
	userPo.Password = string(hashedPassword)
	return u.userRepo.UpdateUser(ctx, userPo)
}

// GetUserBasicInfo 获取用户基本信息（公开接口）
func (u *userAppImpl) GetUserBasicInfo(ctx context.Context, userUUID string) (*dto.UserBasicInfoDto, error) {
	// 从数据库获取用户PO
	userPo, err := u.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 将PO转换为公开DTO
	return &dto.UserBasicInfoDto{
		UserUUID:    userPo.UserUUID,
		Nickname:    userPo.Nickname,
		AvatarUrl:   userPo.AvatarUrl,
		Description: userPo.Description,
		CoverUrl:    userPo.CoverUrl,
		CreatedAt:   userPo.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
