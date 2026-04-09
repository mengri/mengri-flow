package auth

import (
	"time"
)

// IJWTManager JWT 管理器接口
type IJWTManager interface {
	GenerateAccessToken(accountID, role string) (string, error)
	GenerateRefreshToken(accountID, role string) (string, error)
	ParseToken(tokenStr string) (*Claims, error)
	AccessTokenExpiry() int
	RefreshTokenExpiry() time.Duration
}
