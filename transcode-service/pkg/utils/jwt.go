package utils

import (
	"errors"
	"sync"
	"time"
	"transcode-service/pkg/config"

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

// Claims JWT声明（兼容性保留，但不推荐使用）
type Claims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// UUIDClaims JWT声明（推荐使用，基于UUID）
type UUIDClaims struct {
	UserUUID string `json:"user_uuid"`
	UserID   uint64 `json:"user_id,omitempty"` // 兼容性字段，可选
	jwt.RegisteredClaims
}

// DefaultJWTUtil 返回JWT工具单例
func DefaultJWTUtil() *JWTUtil {
	jwtUtilOnce.Do(func() {
		// 从全局配置获取JWT配置
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			panic("JWT工具未初始化")
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

// GenerateAccessTokenWithUUID 生成访问令牌（推荐使用，基于UUID）
func (j *JWTUtil) GenerateAccessTokenWithUUID(userUUID string, userID uint64) (string, error) {
	claims := &UUIDClaims{
		UserUUID: userUUID,
		UserID:   userID, // 兼容性字段
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshTokenWithUUID 生成刷新令牌（推荐使用，基于UUID）
func (j *JWTUtil) GenerateRefreshTokenWithUUID(userUUID string, userID uint64) (string, error) {
	claims := &UUIDClaims{
		UserUUID: userUUID,
		UserID:   userID, // 兼容性字段
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateAccessToken 生成访问令牌（兼容性保留，不推荐使用）
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

// GenerateRefreshToken 生成刷新令牌（兼容性保留，不推荐使用）
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

// ValidateAccessTokenWithUUID 验证访问令牌并返回UUID（推荐使用）
func (j *JWTUtil) ValidateAccessTokenWithUUID(tokenString string) (string, uint64, error) {
	return j.validateTokenWithUUID(tokenString)
}

// ValidateRefreshTokenWithUUID 验证刷新令牌并返回UUID（推荐使用）
func (j *JWTUtil) ValidateRefreshTokenWithUUID(tokenString string) (string, uint64, error) {
	return j.validateTokenWithUUID(tokenString)
}

// validateTokenWithUUID 验证令牌并返回UUID
func (j *JWTUtil) validateTokenWithUUID(tokenString string) (string, uint64, error) {
	// 首先尝试解析UUID格式的token
	token, err := jwt.ParseWithClaims(tokenString, &UUIDClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return j.secretKey, nil
	})

	if err == nil {
		if claims, ok := token.Claims.(*UUIDClaims); ok && token.Valid {
			return claims.UserUUID, claims.UserID, nil
		}
	}

	// 如果UUID格式解析失败，尝试兼容性解析（旧格式）
	token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return "", 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 旧格式token，只有UserID，没有UUID
		return "", claims.UserID, nil
	}

	return "", 0, errors.New("无效的令牌")
}

// ValidateAccessToken 验证访问令牌（兼容性保留，不推荐使用）
func (j *JWTUtil) ValidateAccessToken(tokenString string) (uint64, error) {
	return j.validateToken(tokenString)
}

// ValidateRefreshToken 验证刷新令牌（兼容性保留，不推荐使用）
func (j *JWTUtil) ValidateRefreshToken(tokenString string) (uint64, error) {
	return j.validateToken(tokenString)
}

// validateToken 验证令牌（兼容性保留，不推荐使用）
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
