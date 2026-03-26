package external

import (
	"context"
	"fmt"
	"log/slog"
	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"
	"net/smtp"
	"strings"
)

// SMTPEmailSender 是 EmailSender 的 SMTP 实现。
type SMTPEmailSender struct {
	host    string
	port    int
	user    string
	pass    string
	from    string
	subject string
	baseURL string
}

var _ repository.EmailSender = (*SMTPEmailSender)(nil)

// GenSMTPEmailSender 创建 SMTP 邮件发送器。
func GenSMTPEmailSender(emailCfg *config.EmailConfig) {
	emailSender := &SMTPEmailSender{
		host:    emailCfg.SMTP.Host,
		port:    emailCfg.SMTP.Port,
		user:    emailCfg.SMTP.Username,
		pass:    emailCfg.SMTP.Password,
		from:    emailCfg.SMTP.From,
		subject: emailCfg.Activation.Subject,
		baseURL: emailCfg.Activation.BaseURL,
	}
	autowire.Auto(func() repository.EmailSender { return emailSender })

}

// SendActivationEmail 发送激活邮件。
func (s *SMTPEmailSender) SendActivationEmail(ctx context.Context, toEmail, activationLink string) error {
	body := buildActivationBody(s.subject, activationLink)
	msg := buildMIME(s.from, toEmail, s.subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	err := smtp.SendMail(addr, auth, s.from, []string{toEmail}, []byte(msg))
	if err != nil {
		slog.Error("failed to send activation email", "to", toEmail, "error", err)
		return fmt.Errorf("send activation email: %w", err)
	}

	slog.Info("activation email sent", "to", toEmail)
	return nil
}

func buildMIME(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

func buildActivationBody(subject, activationLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>%s</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2>%s</h2>
  <p>请点击下方链接激活您的账号：</p>
  <p><a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #1890ff; color: #fff; text-decoration: none; border-radius: 4px;">激活账号</a></p>
  <p style="color: #999; font-size: 12px;">如果按钮无法点击，请复制以下链接到浏览器地址栏：<br/>%s</p>
  <p style="color: #999; font-size: 12px;">此链接有效期为 24 小时，过期后需管理员重新发送。</p>
</body>
</html>`, subject, subject, activationLink, activationLink)
}
