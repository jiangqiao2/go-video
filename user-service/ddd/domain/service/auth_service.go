package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"

	"user-service/ddd/application/cqe"
	"user-service/ddd/application/dto"
	"user-service/ddd/domain/repo"
	"user-service/ddd/domain/vo"
	"user-service/ddd/infrastructure/database/persistence"
	"user-service/pkg/errno"
	"user-service/pkg/revocation"
	"user-service/pkg/utils"
)

type AuthService struct {
	userRepo repo.UserRepository
	jwtUtil  *utils.JWTUtil
}

func NewAuthService() *AuthService {
	return &AuthService{userRepo: persistence.NewUserRepository(), jwtUtil: utils.DefaultJWTUtil()}
}

func (s *AuthService) Login(ctx context.Context, req *cqe.UserLoginReq, opts vo.AuthOptions) (*dto.UserLoginDto, error) {
	user, err := s.userRepo.GetUserByAccount(ctx, req.Account)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}
	accessToken, err := s.jwtUtil.GenerateAccessTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errno.ErrTokenGenerate
	}
	refreshToken, err := s.jwtUtil.GenerateRefreshTokenWithUUID(user.UserUUID, user.Id)
	if err != nil {
		return nil, errno.ErrRefreshTokenGenerate
	}
	h := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(h[:])
	store := revocation.DefaultRevocationStore()
	if store != nil {
		if err := store.StoreRefreshToken(ctx, user.UserUUID, tokenHash, opts.RefreshTTL); err != nil {
			return nil, errno.ErrRefreshTokenGenerate
		}
	}
	expiresIn := int64(opts.AccessTTL.Seconds())
	return &dto.UserLoginDto{
		UserUUID:     user.UserUUID,
		Account:      user.Account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		AvatarURL:    user.AvatarUrl,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *cqe.TokenRefreshReq, opts vo.AuthOptions) (*dto.TokenRefreshDto, error) {
	userUUID, _, err := s.jwtUtil.ValidateRefreshTokenWithUUID(req.RefreshToken)
	if err != nil || userUUID == "" {
		return nil, errno.ErrUnauthorized
	}
	store := revocation.DefaultRevocationStore()
	if store != nil {
		h := sha256.Sum256([]byte(req.RefreshToken))
		oldHash := hex.EncodeToString(h[:])
		ok, err := store.ExistsRefreshToken(ctx, userUUID, oldHash)
		if err != nil || !ok {
			return nil, errno.ErrUnauthorized
		}
	}
	userPo, err := s.userRepo.GetUserByUUID(ctx, userUUID)
	if err != nil || userPo == nil {
		return nil, errno.ErrUserNotFound
	}
	accessToken, err := s.jwtUtil.GenerateAccessTokenWithUUID(userPo.UserUUID, userPo.Id)
	if err != nil {
		return nil, errno.ErrTokenGenerate
	}
	refreshToken, err := s.jwtUtil.GenerateRefreshTokenWithUUID(userPo.UserUUID, userPo.Id)
	if err != nil {
		return nil, errno.ErrRefreshTokenGenerate
	}
	if store != nil {
		oh := sha256.Sum256([]byte(req.RefreshToken))
		_ = store.DeleteRefreshToken(ctx, userUUID, hex.EncodeToString(oh[:]))
		nh := sha256.Sum256([]byte(refreshToken))
		if err := store.StoreRefreshToken(ctx, userUUID, hex.EncodeToString(nh[:]), opts.RefreshTTL); err != nil {
			return nil, errno.ErrRefreshTokenGenerate
		}
	}
	expiresIn := int64(opts.AccessTTL.Seconds())
	return &dto.TokenRefreshDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *cqe.TokenRefreshReq) error {
	userUUID, _, err := s.jwtUtil.ValidateRefreshTokenWithUUID(req.RefreshToken)
	if err != nil || userUUID == "" {
		return errno.ErrUnauthorized
	}
	store := revocation.DefaultRevocationStore()
	if store != nil {
		h := sha256.Sum256([]byte(req.RefreshToken))
		tokenHash := hex.EncodeToString(h[:])
		if err := store.DeleteRefreshToken(ctx, userUUID, tokenHash); err != nil {
			return errno.ErrUnauthorized
		}
	}
	return nil
}
