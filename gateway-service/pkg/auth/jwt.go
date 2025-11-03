package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenInvalid     = errors.New("token is invalid")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTUtil provides JWT token operations
type JWTUtil struct {
	secret            []byte
	expireTime        time.Duration
	refreshExpireTime time.Duration
}

// NewJWTUtil creates a new JWT utility instance
func NewJWTUtil(secret string, expireTime, refreshExpireTime time.Duration) *JWTUtil {
	return &JWTUtil{
		secret:            []byte(secret),
		expireTime:        expireTime,
		refreshExpireTime: refreshExpireTime,
	}
}

// GenerateToken generates a new JWT token
func (j *JWTUtil) GenerateToken(userID uint64, username string) (string, error) {
	return j.GenerateTokenWithUUID(userID, "", username)
}

// GenerateTokenWithUUID generates a new JWT token with UUID
func (j *JWTUtil) GenerateTokenWithUUID(userID uint64, userUUID, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserUUID: userUUID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken generates a refresh token
func (j *JWTUtil) GenerateRefreshToken(userID uint64, username string) (string, error) {
	return j.GenerateRefreshTokenWithUUID(userID, "", username)
}

// GenerateRefreshTokenWithUUID generates a refresh token with UUID
func (j *JWTUtil) GenerateRefreshTokenWithUUID(userID uint64, userUUID, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserUUID: userUUID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken parses and validates a JWT token
func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken refreshes an existing token
func (j *JWTUtil) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	return j.GenerateToken(claims.UserID, claims.Username)
}

// ValidateAccessTokenWithUUID validates a token and returns user UUID and ID
func (j *JWTUtil) ValidateAccessTokenWithUUID(tokenString string) (string, uint64, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", 0, err
	}

	return claims.UserUUID, claims.UserID, nil
}