package dto

// IdentityBrief 身份摘要
type IdentityBrief struct {
	IdentityID       string `json:"identityId"`
	LoginType        string `json:"loginType"`
	MaskedIdentifier string `json:"maskedIdentifier,omitempty"`
	BoundAt          string `json:"boundAt"`
}

// IdentityListResponse 身份列表响应
type IdentityListResponse struct {
	Identities    []IdentityBrief `json:"identities"`
	CanUnbindLast bool            `json:"canUnbindLast"`
}

// IdentityResponse 身份绑定响应
type IdentityResponse struct {
	IdentityID string `json:"identityId"`
	LoginType  string `json:"loginType"`
	BoundAt    string `json:"boundAt"`
}

// BindPhoneRequest 绑定手机号请求
type BindPhoneRequest struct {
	Phone          string `json:"phone" binding:"required"`
	SMSCode        string `json:"smsCode" binding:"required,len=6"`
	SecurityTicket string `json:"securityTicket" binding:"required"`
}

// BindProviderRequest 绑定第三方请求
type BindProviderRequest struct {
	BindTicket string `json:"bindTicket" binding:"required"`
}

// UnbindIdentityRequest 解绑登录方式请求
type UnbindIdentityRequest struct {
	SecurityTicket string `json:"securityTicket"`
}
