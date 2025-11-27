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
)

type SocialApp interface {
	Follow(ctx context.Context, req *cqe.FollowReq) error
	Unfollow(ctx context.Context, req *cqe.FollowReq) error
	FollowStatus(ctx context.Context, req *cqe.FollowStatusReq) (*dto.FollowStatusDto, error)
	ListFollowers(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error)
	ListFollowings(ctx context.Context, req *cqe.FollowListQuery) (*dto.FollowListDto, error)
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
		return err
	}
	if !exists {
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
