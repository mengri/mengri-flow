package auth

import (
	"fmt"
	"time"

	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 自定义声明。
type Claims struct {
	AccountID string `json:"accountId"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT 签发与验证管理器。
type JWTManager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// GenerateJWTManager 创建 JWT 管理器。
func GenerateJWTManager(cfg *config.JWTConfig) {

	autowire.Auto(func() *JWTManager {
		return &JWTManager{
			secret:             []byte(cfg.Secret),
			accessTokenExpiry:  time.Duration(cfg.AccessTokenExpiry) * time.Second,
			refreshTokenExpiry: time.Duration(cfg.RefreshTokenExpiry) * time.Second,
		}
	})
}

// GenerateAccessToken 签发 AccessToken。
func (m *JWTManager) GenerateAccessToken(accountID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "mengri-flow",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken 签发 RefreshToken。
func (m *JWTManager) GenerateRefreshToken(accountID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "mengri-flow",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken 解析并验证 JWT token。
func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// AccessTokenExpiry 返回 AccessToken 过期秒数。
func (m *JWTManager) AccessTokenExpiry() int {
	return int(m.accessTokenExpiry.Seconds())
}

// RefreshTokenExpiry 返回 RefreshToken 过期时长。
func (m *JWTManager) RefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}
