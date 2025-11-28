package app

import (
	"context"
	"sync"
	"time"
	"user-service/ddd/application/cqe"
	"user-service/ddd/application/dto"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/database/persistence"
	"user-service/ddd/infrastructure/database/po"
	"user-service/pkg/assert"
	"user-service/pkg/errno"
	"user-service/pkg/logger"
)

type SocialApp interface {
	Follow(ctx context.Context, req *cqe.FollowReq) error
	Unfollow(ctx context.Context, req *cqe.FollowReq) error
	FollowStatus(ctx context.Context, req *cqe.FollowStatusReq) (*dto.FollowStatusDto, error)
	ListFollowers(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error)
	ListFollowings(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error)
	GetUserRelationStat(ctx context.Context, targetUserUUID string, currentUserUUID string) (*dto.UserRelationStatDto, error)
}

type socialAppImpl struct {
	userRepo   repo.UserRepository
	followRepo repo.FollowRepository
}

var (
	onceSocialApp      sync.Once
	singletonSocialApp SocialApp
)

func DefaultSocialApp() SocialApp {
	assert.NotCircular()
	onceSocialApp.Do(func() {
		singletonSocialApp = &socialAppImpl{
			userRepo:   persistence.NewUserRepository(),
			followRepo: persistence.NewFollowRepository(),
		}
	})
	assert.NotNil(singletonSocialApp)
	return singletonSocialApp
}

func NewSocialApp(userRepo repo.UserRepository, followRepo repo.FollowRepository) SocialApp {
	return &socialAppImpl{
		userRepo:   userRepo,
		followRepo: followRepo,
	}
}

func (u *socialAppImpl) Follow(ctx context.Context, req *cqe.FollowReq) error {
	if req == nil || req.UserUUID == "" || req.TargetUUID == "" {
		return errno.ErrParameterInvalid
	}
	if req.UserUUID == req.TargetUUID {
		return errno.ErrFollowSelf
	}
	exists, err := u.userRepo.ExistsByUUID(ctx, req.TargetUUID)
	if err != nil {
		logger.Errorf("Follow exists is err %v", err)
		return err
	}
	if !exists {
		logger.Errorf("Follow exists is exist")
		return errno.ErrUserNotFound
	}
	return u.followRepo.Follow(ctx, req.UserUUID, req.TargetUUID)
}

func (u *socialAppImpl) Unfollow(ctx context.Context, req *cqe.FollowReq) error {
	if req == nil || req.UserUUID == "" || req.TargetUUID == "" {
		return errno.ErrParameterInvalid
	}
	return u.followRepo.Unfollow(ctx, req.UserUUID, req.TargetUUID)
}

func (u *socialAppImpl) FollowStatus(ctx context.Context, req *cqe.FollowStatusReq) (*dto.FollowStatusDto, error) {
	if req == nil || req.UserUUID == "" || req.TargetUUID == "" {
		return nil, errno.ErrParameterInvalid
	}
	following, err := u.followRepo.IsFollowing(ctx, req.UserUUID, req.TargetUUID)
	if err != nil {
		return nil, err
	}
	return &dto.FollowStatusDto{Following: following}, nil
}

func (u *socialAppImpl) ListFollowers(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error) {
	if req == nil || req.TargetUUID == "" {
		return nil, errno.ErrParameterInvalid
	}
	offset, limit := normalizePage(req.Page, req.Size)
	list, total, err := u.followRepo.ListFollowers(ctx, req.TargetUUID, offset, limit)
	if err != nil {
		return nil, err
	}
	return buildFollowListResp(list, req.Page, limit, total), nil
}

func (u *socialAppImpl) ListFollowings(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error) {
	if req == nil || req.TargetUUID == "" {
		return nil, errno.ErrParameterInvalid
	}
	offset, limit := normalizePage(req.Page, req.Size)
	list, total, err := u.followRepo.ListFollowings(ctx, req.TargetUUID, offset, limit)
	if err != nil {
		return nil, err
	}
	return buildFollowListResp(list, req.Page, limit, total), nil
}

func buildFollowListResp(list []*po.FollowPo, page, size int, total int64) *dto.FollowListDto {
	resp := &dto.FollowListDto{
		Page:  page,
		Size:  size,
		Total: total,
		List:  make([]dto.FollowUser, 0, len(list)),
	}
	for _, v := range list {
		if v == nil {
			continue
		}
		resp.List = append(resp.List, dto.FollowUser{
			UserUUID:  v.UserUUID,
			CreatedAt: v.CreatedAt.Format(time.RFC3339),
		})
	}
	return resp
}

// GetUserRelationStat 获取用户关系统计（粉丝数、关注数、关注状态）
func (u *socialAppImpl) GetUserRelationStat(ctx context.Context, targetUserUUID string, currentUserUUID string) (*dto.UserRelationStatDto, error) {
	if targetUserUUID == "" {
		return nil, errno.ErrParameterInvalid
	}

	// 检查目标用户是否存在
	exists, err := u.userRepo.ExistsByUUID(ctx, targetUserUUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errno.ErrUserNotFound
	}

	// 获取粉丝数
	followers, followerTotal, err := u.followRepo.ListFollowers(ctx, targetUserUUID, 0, 1)
	if err != nil {
		return nil, err
	}
	_ = followers // 不需要具体列表，只需要总数

	// 获取关注数
	followings, followingTotal, err := u.followRepo.ListFollowings(ctx, targetUserUUID, 0, 1)
	if err != nil {
		return nil, err
	}
	_ = followings

	// 判断当前用户是否已关注目标用户
	isFollowed := false
	if currentUserUUID != "" && currentUserUUID != targetUserUUID {
		isFollowed, err = u.followRepo.IsFollowing(ctx, currentUserUUID, targetUserUUID)
		if err != nil {
			// 如果查询关注状态失败，不影响整体返回，默认为未关注
			isFollowed = false
		}
	}

	return &dto.UserRelationStatDto{
		UserUUID:       targetUserUUID,
		FollowerCount:  followerTotal,
		FollowingCount: followingTotal,
		IsFollowed:     isFollowed,
	}, nil
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size
	return offset, size
}
