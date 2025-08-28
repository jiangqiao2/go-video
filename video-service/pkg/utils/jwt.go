package utils

import (
	"errors"
	"time"
)

// JWTUtil JWT工具类（简化版，移除jwt依赖）
type JWTUtil struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// Claims 简化的声明结构
type Claims struct {
	UserID uint64 `json:"user_id"`
}

// UUIDClaims 基于UUID的声明结构
type UUIDClaims struct {
	UserUUID string `json:"user_uuid"`
	UserID   uint64 `json:"user_id"`
}

// NewJWTUtil 创建JWT工具实例
func NewJWTUtil(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTUtil {
	return &JWTUtil{
		secretKey:       secretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateAccessTokenWithUUID 生成基于UUID的访问令牌（简化版）
func (j *JWTUtil) GenerateAccessTokenWithUUID(userUUID string, userID uint64) (string, error) {
	// 简化实现，返回固定格式的token
	return "mock-token-" + userUUID, nil
}

// GenerateRefreshTokenWithUUID 生成基于UUID的刷新令牌（简化版）
func (j *JWTUtil) GenerateRefreshTokenWithUUID(userUUID string, userID uint64) (string, error) {
	// 简化实现，返回固定格式的token
	return "mock-refresh-token-" + userUUID, nil
}

// ValidateAccessTokenWithUUID 验证基于UUID的访问令牌（简化版）
func (j *JWTUtil) ValidateAccessTokenWithUUID(tokenString string) (string, uint64, error) {
	// 简化实现，总是返回成功
	if tokenString == "" {
		return "", 0, errors.New("token is empty")
	}
	return "mock-user-uuid", 1, nil
}

// ValidateRefreshTokenWithUUID 验证基于UUID的刷新令牌（简化版）
func (j *JWTUtil) ValidateRefreshTokenWithUUID(tokenString string) (string, uint64, error) {
	// 简化实现，总是返回成功
	if tokenString == "" {
		return "", 0, errors.New("token is empty")
	}
	return "mock-user-uuid", 1, nil
}
