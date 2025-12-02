package utils

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	pkcs8 "github.com/youmark/pkcs8"

	"user-service/pkg/config"
	"user-service/pkg/revocation"
)

var (
	// JWT工具单例
	jwtUtilOnce      sync.Once
	singletonJWTUtil *JWTUtil
)

// JWTUtil JWT工具类
type JWTUtil struct {
	secretKey       []byte
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	revoker         revocation.Store
}

// Claims JWT声明（兼容性保留，但不推荐使用）
type Claims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// UUIDClaims JWT声明（推荐使用，基于UUID）
type UUIDClaims struct {
	UserUUID  string `json:"user_uuid"`
	UserID    uint64 `json:"user_id,omitempty"`
	TokenType string `json:"token_type"`
	Version   int64  `json:"ver,omitempty"`
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

		j := &JWTUtil{
			secretKey:       []byte(cfg.JWT.Secret),
			issuer:          cfg.JWT.Issuer,
			accessTokenTTL:  cfg.JWT.ExpireTime,
			refreshTokenTTL: cfg.JWT.RefreshExpireTime,
		}

		// 加载RSA密钥
		if cfg.JWT.RSAPrivateKeyPath != "" {
			if pk, err := loadRSAPrivateKeyFromPEM(cfg.JWT.RSAPrivateKeyPath, cfg.JWT.RSAPrivateKeyPassword); err == nil {
				j.privateKey = pk
				fmt.Printf("[JWT] Loaded RSA private key from %s\n", cfg.JWT.RSAPrivateKeyPath)
			} else {
				fmt.Printf("[JWT] Failed to load RSA private key from %s: %v\n", cfg.JWT.RSAPrivateKeyPath, err)
			}
		}
		if cfg.JWT.RSAPublicKeyPath != "" {
			if pub, err := loadRSAPublicKeyFromPEM(cfg.JWT.RSAPublicKeyPath); err == nil {
				j.publicKey = pub
				fmt.Printf("[JWT] Loaded RSA public key from %s\n", cfg.JWT.RSAPublicKeyPath)
			} else {
				fmt.Printf("[JWT] Failed to load RSA public key from %s: %v\n", cfg.JWT.RSAPublicKeyPath, err)
			}
		}
		if j.privateKey != nil {
			fmt.Printf("[JWT] Private key modulus size: %d bits\n", j.privateKey.N.BitLen())
		}
		if j.publicKey != nil {
			fmt.Printf("[JWT] Public key modulus size: %d bits\n", j.publicKey.N.BitLen())
		}

		j.revoker = revocation.DefaultRevocationStore()
		singletonJWTUtil = j
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
	var ver int64
	if j.revoker != nil {
		if v, err := j.revoker.GetVersion(context.Background(), userUUID); err == nil {
			ver = v
		}
	}
	claims := &UUIDClaims{
		UserUUID:  userUUID,
		UserID:    userID,
		TokenType: "access",
		Version:   ver,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
		},
	}

	if j.privateKey != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(j.privateKey)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshTokenWithUUID 生成刷新令牌（推荐使用，基于UUID）
func (j *JWTUtil) GenerateRefreshTokenWithUUID(userUUID string, userID uint64) (string, error) {
	var ver int64
	if j.revoker != nil {
		if v, err := j.revoker.GetVersion(context.Background(), userUUID); err == nil {
			ver = v
		}
	}
	claims := &UUIDClaims{
		UserUUID:  userUUID,
		UserID:    userID,
		TokenType: "refresh",
		Version:   ver,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
		},
	}

	if j.privateKey != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(j.privateKey)
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
	uuid, uid, _, err := j.validateTokenWithUUID(tokenString)
	if err != nil {
		return "", 0, err
	}
	if j.publicKey != nil || len(j.secretKey) > 0 {
		token, err := jwt.ParseWithClaims(tokenString, &UUIDClaims{}, func(token *jwt.Token) (interface{}, error) {
			if j.publicKey != nil {
				return j.publicKey, nil
			}
			return j.secretKey, nil
		})
		if err == nil {
			if claims, ok := token.Claims.(*UUIDClaims); ok && token.Valid {
				if claims.TokenType != "access" {
					return "", 0, errors.New("错误的令牌类型")
				}
			}
		}
	}
	return uuid, uid, nil
}

// ValidateRefreshTokenWithUUID 验证刷新令牌并返回UUID（推荐使用）
func (j *JWTUtil) ValidateRefreshTokenWithUUID(tokenString string) (string, uint64, error) {
	uuid, uid, ver, err := j.validateTokenWithUUID(tokenString)
	if err != nil {
		return "", 0, err
	}
	if j.publicKey != nil || len(j.secretKey) > 0 {
		token, err := jwt.ParseWithClaims(tokenString, &UUIDClaims{}, func(token *jwt.Token) (interface{}, error) {
			if j.publicKey != nil {
				return j.publicKey, nil
			}
			return j.secretKey, nil
		})
		if err == nil {
			if claims, ok := token.Claims.(*UUIDClaims); ok && token.Valid {
				if claims.TokenType != "refresh" {
					return "", 0, errors.New("错误的令牌类型")
				}
			}
		}
	}
	if j.revoker != nil {
		if current, err := j.revoker.GetVersion(context.Background(), uuid); err == nil && current > ver {
			return "", 0, errors.New("token revoked")
		}
	}
	return uuid, uid, nil
}

// validateTokenWithUUID 验证令牌并返回UUID
func (j *JWTUtil) validateTokenWithUUID(tokenString string) (string, uint64, int64, error) {
	// 优先使用RSA公钥验证
	token, err := jwt.ParseWithClaims(tokenString, &UUIDClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return j.publicKey, nil
	})
	if err == nil {
		if claims, ok := token.Claims.(*UUIDClaims); ok && token.Valid {
			return claims.UserUUID, claims.UserID, claims.Version, nil
		}
	}
	return "", 0, 0, errors.New("无效的签名方法")
}

// loadRSAPrivateKeyFromPEM 加载RSA私钥（支持PKCS#1、PKCS#8、加密PKCS#8）
func loadRSAPrivateKeyFromPEM(path string, password string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("私钥PEM解析失败")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		if k, ok := pk.(*rsa.PrivateKey); ok {
			return k, nil
		}
		return nil, errors.New("不是RSA私钥")
	case "ENCRYPTED PRIVATE KEY":
		if password == "" {
			return nil, errors.New("检测到加密私钥，但未提供密码")
		}
		key, err := pkcs8.ParsePKCS8PrivateKey(block.Bytes, []byte(password))
		if err != nil {
			return nil, err
		}
		if k, ok := key.(*rsa.PrivateKey); ok {
			return k, nil
		}
		return nil, errors.New("不是RSA私钥")
	default:
		return nil, errors.New("未知的私钥PEM类型")
	}
}

// loadRSAPublicKeyFromPEM 加载RSA公钥
func loadRSAPublicKeyFromPEM(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("公钥PEM解析失败")
	}
	if block.Type == "PUBLIC KEY" {
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		if k, ok := pub.(*rsa.PublicKey); ok {
			return k, nil
		}
	}
	if block.Type == "RSA PUBLIC KEY" {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}
	return nil, errors.New("不是RSA公钥")
}
