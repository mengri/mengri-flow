package repository

import (
	"context"
)

// SMSSender 短信发送接口。
type SMSSender interface {
	SendOTP(ctx context.Context, phone, code string) error
}
