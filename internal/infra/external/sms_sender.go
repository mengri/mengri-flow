package external

import (
	"context"
	"fmt"
	"log/slog"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"

	"mengri-flow/internal/domain/repository"
	"mengri-flow/internal/infra/config"
	"mengri-flow/pkg/autowire"
)

// AliyunSMSSender 是 SMSSender 的阿里云短信实现。
type AliyunSMSSender struct {
	client       *dysmsapi.Client
	signName     string
	templateCode string
}

var _ repository.SMSSender = (*AliyunSMSSender)(nil)

// GenAliyunSMSSender 创建阿里云短信发送器并注册到 autowire 容器。
func GenAliyunSMSSender(smsCfg *config.SMSConfig) error {
	apiConfig := &openapi.Config{
		AccessKeyId:     tea.String(smsCfg.AccessKeyID),
		AccessKeySecret: tea.String(smsCfg.AccessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}

	client, err := dysmsapi.NewClient(apiConfig)
	if err != nil {
		return fmt.Errorf("create aliyun sms client: %w", err)
	}

	sender := &AliyunSMSSender{
		client:       client,
		signName:     smsCfg.SignName,
		templateCode: smsCfg.TemplateCode,
	}

	autowire.Auto(func() repository.SMSSender { return sender })

	slog.Info("aliyun sms sender initialized",
		"signName", smsCfg.SignName,
		"templateCode", smsCfg.TemplateCode,
	)
	return nil
}

// SendOTP 发送短信验证码。
func (s *AliyunSMSSender) SendOTP(ctx context.Context, phone, code string) error {
	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(s.signName),
		TemplateCode:  tea.String(s.templateCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}

	resp, err := s.client.SendSms(req)
	if err != nil {
		slog.ErrorContext(ctx, "aliyun sms send failed",
			"phone", phone, "error", err,
		)
		return fmt.Errorf("send sms: %w", err)
	}

	if resp.Body == nil || resp.Body.Code == nil || *resp.Body.Code != "OK" {
		errCode := ""
		errMsg := ""
		if resp.Body != nil {
			if resp.Body.Code != nil {
				errCode = *resp.Body.Code
			}
			if resp.Body.Message != nil {
				errMsg = *resp.Body.Message
			}
		}
		slog.ErrorContext(ctx, "aliyun sms send rejected",
			"phone", phone, "code", errCode, "message", errMsg,
		)
		return fmt.Errorf("sms rejected: %s - %s", errCode, errMsg)
	}

	slog.InfoContext(ctx, "sms otp sent", "phone", phone)
	return nil
}
