package dto

// --- 激活 ---

// ActivationValidateResponse 激活链接预校验响应
type ActivationValidateResponse struct {
	Valid            bool   `json:"valid"`
	EmailMasked      string `json:"emailMasked"`
	ExpireAt         string `json:"expireAt"`
	AlreadyActivated bool   `json:"alreadyActivated"`
}

// ActivationConfirmRequest 确认激活并设置密码请求
type ActivationConfirmRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

// ActivationConfirmResponse 确认激活响应
type ActivationConfirmResponse struct {
	Activated   bool   `json:"activated"`
	AccountID   string `json:"accountId"`
	Status      string `json:"status"`
	ActivatedAt string `json:"activatedAt"`
}

// --- 登录 ---

// DeviceInfo 设备信息
type DeviceInfo struct {
	UA       string `json:"ua"`
	IP       string `json:"ip"`
	DeviceID string `json:"deviceId"`
}

// PasswordLoginRequest 密码登录请求
type PasswordLoginRequest struct {
	Account    string     `json:"account" binding:"required"`
	Password   string     `json:"password" binding:"required"`
	DeviceInfo DeviceInfo `json:"deviceInfo"`
}

// SMSSendRequest 发送短信验证码请求
type SMSSendRequest struct {
	Phone        string `json:"phone" binding:"required"`
	Scene        string `json:"scene" binding:"required,oneof=login bind"`
	CaptchaToken string `json:"captchaToken"`
}

// SMSSendResponse 发送短信验证码响应
type SMSSendResponse struct {
	Sent          bool `json:"sent"`
	TTLSec        int  `json:"ttlSec"`
	RetryAfterSec int  `json:"retryAfterSec"`
}

// SMSLoginRequest 短信验证码登录请求
type SMSLoginRequest struct {
	Phone      string     `json:"phone" binding:"required"`
	Code       string     `json:"code" binding:"required,len=6"`
	DeviceInfo DeviceInfo `json:"deviceInfo"`
}

// LoginResponse 登录响应（密码登录、短信登录、OAuth 登录共用）
type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int          `json:"expiresIn"`
	TokenType    string       `json:"tokenType"`
	Account      AccountBrief `json:"account"`
}

// AccountBrief 账号摘要信息
type AccountBrief struct {
	AccountID   string `json:"accountId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
}

// --- OAuth ---

// OAuthURLResponse 获取第三方授权地址响应
type OAuthURLResponse struct {
	AuthURL  string `json:"authUrl"`
	State    string `json:"state"`
	ExpireAt string `json:"expireAt"`
}

// OAuthCallbackResponse 第三方回调响应（三态：LOGIN_SUCCESS / NEED_BIND_EXISTING_ACCOUNT / BIND_SUCCESS）
type OAuthCallbackResponse struct {
	Result       string         `json:"result"`
	AccessToken  string         `json:"accessToken,omitempty"`
	RefreshToken string         `json:"refreshToken,omitempty"`
	ExpiresIn    int            `json:"expiresIn,omitempty"`
	Account      *AccountBrief  `json:"account,omitempty"`
	Provider     string         `json:"provider,omitempty"`
	BindTicket   string         `json:"bindTicket,omitempty"`
	ExpireAt     string         `json:"expireAt,omitempty"`
	Identity     *IdentityBrief `json:"identity,omitempty"`
}

// --- Token ---

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
