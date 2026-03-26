package repository

import (
	"context"
)

// EmailSender 邮件发送接口（Domain 定义，Infra 实现）。
type EmailSender interface {
	SendActivationEmail(ctx context.Context, toEmail, activationLink string) error
}
