package utils

import (
	"errors"
	"sync"
	"time"

	"go-video/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWT工具单例
	jwtUtilOnce      sync.Once
	singletonJWTUtil *JWTUtil
)

// JWTUtil JWT工具类
type JWTUtil struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// Claims JWT声明
type Claims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// DefaultJWTUtil 返回JWT工具单例
func DefaultJWTUtil() *JWTUtil {
	jwtUtilOnce.Do(func() {
		cfg, err := config.Load("configs/config.dev.yaml")
		if err != nil {
			panic("failed to load config: " + err.Error())
		}

		singletonJWTUtil = &JWTUtil{
			secretKey:       []byte(cfg.JWT.Secret),
			accessTokenTTL:  cfg.JWT.ExpireTime,
			refreshTokenTTL: cfg.JWT.RefreshExpireTime,
		}
	})
	if singletonJWTUtil == nil {
		panic("failed to create JWT util singleton")
	}
	return singletonJWTUtil
}

// NewJWTUtil 创建JWT工具实例
func NewJWTUtil(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTUtil {
	return &JWTUtil{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateAccessToken 生成访问令牌
func (j *JWTUtil) GenerateAccessToken(userID uint64) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTUtil) GenerateRefreshToken(userID uint64) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateAccessToken 验证访问令牌
func (j *JWTUtil) ValidateAccessToken(tokenString string) (uint64, error) {
	return j.validateToken(tokenString)
}

// ValidateRefreshToken 验证刷新令牌
func (j *JWTUtil) ValidateRefreshToken(tokenString string) (uint64, error) {
	return j.validateToken(tokenString)
}

// validateToken 验证令牌
func (j *JWTUtil) validateToken(tokenString string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("无效的令牌")
}
