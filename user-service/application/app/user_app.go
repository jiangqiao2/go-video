package app

import (
	"context"
	"sync"
	"time"

	"go-video/ddd/internal/resource"
	"go-video/ddd/user/application/cqe"
	"go-video/ddd/user/application/dto"
	"go-video/ddd/user/domain/entity"
	"go-video/ddd/user/domain/service"
	"go-video/ddd/user/domain/vo"
	"go-video/pkg/assert"
	"go-video/pkg/errno"
	"go-video/pkg/utils"
)

var (
	userAppOnce      sync.Once
	singletonUserApp *UserApp
)

// UserApp 用户应用服务
type UserApp struct {
	userDomainService *service.UserService
	jwtUtil           *utils.JWTUtil
}

// DefaultUserApp 获取用户应用服务单例
func DefaultUserApp() *UserApp {
	assert.NotCircular()
	userAppOnce.Do(func() {
		userDomainService := service.DefaultUserService()
		jwtUtil := resource.DefaultJWTResource()
		singletonUserApp = &UserApp{
			userDomainService: userDomainService,
			jwtUtil:           jwtUtil.GetJWTUtil(),
		}
	})
	assert.NotNil(singletonUserApp)
	return singletonUserApp
}

// NewUserApp 创建用户应用服务实例（支持依赖注入）
func NewUserApp(userDomainService *service.UserService, jwtUtil *utils.JWTUtil) *UserApp {
	return &UserApp{
		userDomainService: userDomainService,
		jwtUtil:           jwtUtil,
	}
}

// CreateUser 创建用户
func (a *UserApp) CreateUser(ctx context.Context, cmd *cqe.CreateUserCommand) (*dto.CreateUserResponse, error) {
	// 创建用户
	user, err := a.userDomainService.CreateUser(ctx, cmd.Username, cmd.Password, cmd.Email)
	if err != nil {
		return nil, err
	}

	// 转换为响应DTO
	return &dto.CreateUserResponse{
		UUID:     user.UUID(),
		Username: user.Username(),
		Email:    user.Email(),
		Status:   int(user.Status()),
	}, nil
}

// Login 用户登录
func (a *UserApp) Login(ctx context.Context, cmd *cqe.LoginCommand) (*dto.LoginResponse, error) {
	// 用户认证
	user, err := a.userDomainService.AuthenticateUser(ctx, cmd.Username, cmd.Password)
	if err != nil {
		return nil, err
	}

	// 生成JWT令牌（使用UUID格式）
	if a.jwtUtil == nil {
		return nil, errno.NewSimpleBizError(errno.ErrInternalServer, nil, "JWT工具未初始化")
	}

	// 使用新方法生成包含UUID的token
	token, err := a.jwtUtil.GenerateAccessTokenWithUUID(user.UUID(), user.ID())
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrInternalServer, err, "生成令牌失败")
	}

	// 计算过期时间
	expiresAt := time.Now().Add(24 * time.Hour) // 默认24小时过期

	// 转换为响应DTO
	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      a.toUserInfo(user),
	}, nil
}

// GetUserInfo 获取用户信息
func (a *UserApp) GetUserInfo(ctx context.Context, userID uint64) (*dto.UserInfo, error) {
	user, err := a.userDomainService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return a.toUserInfo(user), nil
}

// UpdateProfile 更新用户资料
func (a *UserApp) UpdateProfile(ctx context.Context, cmd *cqe.UpdateProfileCommand) error {
	return a.userDomainService.UpdateProfile(ctx, cmd.UserID, cmd.Nickname, cmd.Avatar)
}

// ChangePassword 修改密码
func (a *UserApp) ChangePassword(ctx context.Context, cmd *cqe.ChangePasswordCommand) error {
	return a.userDomainService.ChangePassword(ctx, cmd.UserID, cmd.OldPassword, cmd.NewPassword)
}

// GetUserList 获取用户列表
func (a *UserApp) GetUserList(ctx context.Context, query *cqe.GetUserListQuery) (*dto.GetUserListResponse, error) {
	// 创建分页对象
	page, err := vo.NewPage(query.Page, query.PageSize)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "分页参数无效")
	}

	// 获取用户列表
	users, total, err := a.userDomainService.GetUserList(ctx, page)
	if err != nil {
		return nil, err
	}

	// 转换为响应DTO
	userInfos := make([]*dto.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, a.toUserInfo(user))
	}

	return &dto.GetUserListResponse{
		Users:    userInfos,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// ActivateUser 激活用户
func (a *UserApp) ActivateUser(ctx context.Context, userID uint64) error {
	return a.userDomainService.ActivateUser(ctx, userID)
}

// DisableUser 禁用用户
func (a *UserApp) DisableUser(ctx context.Context, userID uint64) error {
	return a.userDomainService.DisableUser(ctx, userID)
}

// DeleteUser 删除用户
func (a *UserApp) DeleteUser(ctx context.Context, userID uint64) error {
	return a.userDomainService.DeleteUser(ctx, userID)
}

// ValidateToken 验证令牌
func (a *UserApp) ValidateToken(ctx context.Context, token string) (*dto.UserInfo, error) {
	if a.jwtUtil == nil {
		return nil, errno.NewSimpleBizError(errno.ErrInternalServer, nil, "JWT工具未初始化")
	}

	// 验证令牌并获取用户信息（优先使用UUID）
	userUUID, userID, err := a.jwtUtil.ValidateAccessTokenWithUUID(token)
	if err != nil {
		return nil, errno.NewSimpleBizError(errno.ErrUnauthorized, err, "令牌无效")
	}

	// 获取用户信息（优先使用UUID查找）
	var user *entity.User
	if userUUID != "" {
		// 使用UUID查找用户
		user, err = a.userDomainService.GetUserByUUID(ctx, userUUID)
	} else {
		// 兼容性：使用ID查找用户
		user, err = a.userDomainService.GetUserByID(ctx, userID)
	}
	if err != nil {
		return nil, err
	}

	return a.toUserInfo(user), nil
}

// toUserInfo 将用户实体转换为用户信息DTO
func (a *UserApp) toUserInfo(user *entity.User) *dto.UserInfo {
	return &dto.UserInfo{
		UUID:     user.UUID(),
		Username: user.Username(),
		Email:    user.Email(),
		Nickname: user.Nickname(),
		Avatar:   user.Avatar(),
		Status:   int(user.Status()),
	}
}
