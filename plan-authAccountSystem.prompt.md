## Plan: 账号体系接口清单与状态机补充

基于前面的用户故事，补充一份可直接用于后端 API 设计评审和前后端联调的"接口清单 + 状态机"。  
约束保持一致：无前台注册、后台建号激活、一个账号多登录方式、第三方首次不自动开户。

### Steps
1. 定义统一枚举与约定（账号状态、登录方式、错误码语义）。
2. 按角色分组接口（管理员端、认证端、账号中心）。
3. 给每个接口补充：用途、鉴权、请求参数、关键返回、关键校验。
4. 输出账号状态机与登录方式绑定状态机。
5. 明确状态迁移触发事件与失败回退策略。

### Relevant files
- 本次为需求设计文档补充，不涉及代码文件修改。

### Verification
1. 覆盖创建、激活、登录、绑定、解绑、审计全链路。
2. 能映射项目统一响应结构 `{ code, data, msg }`。
3. 前端可按接口直接实现页面流程；后端可按状态机写校验逻辑。

---

## 统一约定

- Base URL：`/api/v1`
- 响应结构：`{ code: number, data: T, msg: string }`
- `code = 0` 成功，非 `0` 失败
- 登录方式枚举 `loginType`：
  - `password`
  - `sms`
  - `wechat_qr`
  - `lark_qr`
  - `github_oauth`
- 账号状态枚举 `accountStatus`：
  - `PENDING_ACTIVATION`（未激活）
  - `ACTIVE`（已激活）
  - `LOCKED`（锁定，可选）
  - `DISABLED`（禁用，可选）

---

## 一、接口清单

### A. 管理后台接口（管理员权限）

#### 1) 创建账号
`POST /admin/accounts`

- 用途：通过邮箱创建账号（未激活）
- 鉴权：管理员 token

请求体：
```json
{
  "email": "user@example.com",
  "displayName": "张三",
  "username": "zhangsan"
}
```

响应 data：
```json
{
  "accountId": "acc_xxx",
  "email": "user@example.com",
  "status": "PENDING_ACTIVATION",
  "activationExpireAt": "2026-03-27T12:00:00Z",
  "createdAt": "2026-03-26T12:00:00Z"
}
```

关键校验：
- 邮箱唯一
- 创建成功后生成激活码并发激活邮件
- 写审计日志

---

#### 2) 重发激活邮件
`POST /admin/accounts/{accountId}/activation/resend`

- 用途：重发激活邮件
- 鉴权：管理员 token

请求体：
```json
{
  "reason": "用户未收到邮件"
}
```

响应 data：
```json
{
  "sent": true,
  "activationExpireAt": "2026-03-27T13:00:00Z",
  "throttleSec": 60
}
```

关键校验：
- 仅 `PENDING_ACTIVATION` 可重发
- 旧激活码失效，新激活码生效
- 频控限制（如 60 秒）

---

#### 3) 查询账号详情
`GET /admin/accounts/{accountId}`

- 用途：查询账号状态与绑定方式
- 鉴权：管理员 token

响应 data（示例）：
```json
{
  "accountId": "acc_xxx",
  "email": "user@example.com",
  "status": "ACTIVE",
  "identities": [
    { "loginType": "password", "boundAt": "2026-03-26T13:00:00Z" },
    { "loginType": "github_oauth", "boundAt": "2026-03-26T14:00:00Z" }
  ]
}
```

---

### B. 认证与激活接口（公开/半公开）

#### 4) 激活链接预校验
`GET /auth/activation/validate?token=...`

响应 data：
```json
{
  "valid": true,
  "emailMasked": "u***@example.com",
  "expireAt": "2026-03-27T12:00:00Z",
  "alreadyActivated": false
}
```

---

#### 5) 确认激活并设置密码
`POST /auth/activation/confirm`

请求体：
```json
{
  "token": "act_xxx",
  "password": "Strong@123",
  "confirmPassword": "Strong@123"
}
```

响应 data：
```json
{
  "activated": true,
  "accountId": "acc_xxx",
  "status": "ACTIVE",
  "activatedAt": "2026-03-26T13:00:00Z"
}
```

---

#### 6) 密码登录
`POST /auth/login/password`

请求体：
```json
{
  "account": "user@example.com",
  "password": "Strong@123",
  "deviceInfo": {
    "ua": "Mozilla/5.0",
    "ip": "127.0.0.1",
    "deviceId": "dev_xxx"
  }
}
```

响应 data：
```json
{
  "accessToken": "jwt_xxx",
  "refreshToken": "rft_xxx",
  "expiresIn": 7200,
  "tokenType": "Bearer",
  "account": {
    "accountId": "acc_xxx",
    "email": "user@example.com",
    "displayName": "张三",
    "status": "ACTIVE"
  }
}
```

---

#### 7) 发送短信验证码
`POST /auth/login/sms/send`

请求体：
```json
{
  "phone": "+8613800000000",
  "scene": "login",
  "captchaToken": "captcha_xxx"
}
```

响应 data：
```json
{
  "sent": true,
  "ttlSec": 300,
  "retryAfterSec": 60
}
```

---

#### 8) 短信验证码登录
`POST /auth/login/sms/verify`

请求体：
```json
{
  "phone": "+8613800000000",
  "code": "123456",
  "deviceInfo": {
    "ua": "Mozilla/5.0",
    "ip": "127.0.0.1",
    "deviceId": "dev_xxx"
  }
}
```

响应 data：同密码登录。

---

#### 9) 获取第三方授权地址
`GET /auth/oauth/{provider}/url?scene=login|bind&redirectUri=...`

- `provider`：`wechat | lark | github`

响应 data：
```json
{
  "authUrl": "https://provider.xxx/oauth2/authorize?...",
  "state": "state_xxx",
  "expireAt": "2026-03-26T12:35:00Z"
}
```

---

#### 10) 第三方回调
`GET /auth/oauth/{provider}/callback?code=...&state=...`

响应 data（三态之一）：

1. 已绑定可登录：
```json
{
  "result": "LOGIN_SUCCESS",
  "accessToken": "jwt_xxx",
  "refreshToken": "rft_xxx",
  "expiresIn": 7200,
  "account": {
    "accountId": "acc_xxx",
    "email": "user@example.com",
    "status": "ACTIVE"
  }
}
```

2. 未绑定（登录场景）：
```json
{
  "result": "NEED_BIND_EXISTING_ACCOUNT",
  "provider": "github",
  "bindTicket": "bt_xxx",
  "expireAt": "2026-03-26T12:40:00Z"
}
```

3. 绑定场景成功：
```json
{
  "result": "BIND_SUCCESS",
  "identity": {
    "identityId": "idn_xxx",
    "loginType": "github_oauth",
    "boundAt": "2026-03-26T12:36:00Z"
  }
}
```

---

#### 11) 退出登录
`POST /auth/logout`

响应 data：
```json
{
  "success": true
}
```

---

### C. 账号中心接口（用户已登录）

#### 12) 我的资料
`GET /me/profile`

响应 data（示例）：
```json
{
  "accountId": "acc_xxx",
  "email": "user@example.com",
  "displayName": "张三",
  "accountStatus": "ACTIVE"
}
```

---

#### 13) 我的登录方式列表
`GET /me/identities`

响应 data：
```json
{
  "identities": [
    {
      "identityId": "idn_pwd",
      "loginType": "password",
      "maskedIdentifier": "user@example.com",
      "boundAt": "2026-03-26T13:00:00Z"
    },
    {
      "identityId": "idn_gh",
      "loginType": "github_oauth",
      "maskedIdentifier": "g***b",
      "boundAt": "2026-03-26T14:00:00Z"
    }
  ],
  "canUnbindLast": false
}
```

---

#### 14) 绑定手机号
`POST /me/identities/phone/bind`

请求体：
```json
{
  "phone": "+8613800000000",
  "smsCode": "123456",
  "securityTicket": "sec_xxx"
}
```

响应 data：
```json
{
  "identityId": "idn_sms_xxx",
  "loginType": "sms",
  "boundAt": "2026-03-26T14:10:00Z"
}
```

---

#### 15) 绑定第三方
`POST /me/identities/{provider}/bind`

请求体（推荐）：
```json
{
  "bindTicket": "bt_xxx"
}
```

响应 data：
```json
{
  "identityId": "idn_gh_xxx",
  "loginType": "github_oauth",
  "boundAt": "2026-03-26T14:20:00Z"
}
```

---

#### 16) 解绑登录方式
`DELETE /me/identities/{identityId}`

请求体（可选）：
```json
{
  "securityTicket": "sec_xxx"
}
```

响应 data：
```json
{
  "unbound": true,
  "remainingLoginTypes": ["password", "sms"]
}
```

关键校验：
- 不允许解绑最后一种可用登录方式

---

#### 17) 修改密码
`POST /me/password/change`

请求体：
```json
{
  "oldPassword": "Old@123",
  "newPassword": "New@12345",
  "confirmPassword": "New@12345",
  "revokeOtherSessions": true
}
```

响应 data：
```json
{
  "changed": true,
  "revokedSessions": 3
}
```

---

### D. 审计与安全接口（可选增强）

#### 18) 我的登录记录
`GET /me/security/logins`

#### 19) 审计事件查询（管理员）
`GET /admin/audit/events?accountId=...&eventType=...&from=...&to=...`

---

## 二、状态机设计

### 1) 账号状态机（Account FSM）

状态：
- `PENDING_ACTIVATION`
- `ACTIVE`
- `LOCKED`（可选）
- `DISABLED`（可选）

迁移：
1. `PENDING_ACTIVATION -> ACTIVE`
   - 事件：激活成功并设置密码
2. `PENDING_ACTIVATION -> PENDING_ACTIVATION`
   - 事件：重发激活邮件（token 轮换）
3. `ACTIVE -> LOCKED`
   - 事件：风控触发或管理员锁定
4. `LOCKED -> ACTIVE`
   - 事件：管理员解锁/自动解锁
5. `ACTIVE -> DISABLED`
   - 事件：管理员禁用
6. `DISABLED -> ACTIVE`
   - 事件：管理员恢复

非法规则：
- `PENDING_ACTIVATION` 不允许登录
- `DISABLED` 不允许通过登录流程恢复

---

### 2) 登录方式绑定状态机（Identity FSM）

状态：
- `UNBOUND`（逻辑态）
- `PENDING_BIND`（可选）
- `BOUND`
- `UNBOUND`（解绑后回归）

迁移：
1. `UNBOUND -> PENDING_BIND`
   - 事件：发起绑定（扫码/OAuth）
2. `PENDING_BIND -> BOUND`
   - 事件：回调成功/验证码校验成功
   - 条件：第三方标识未被他人绑定
3. `BOUND -> UNBOUND`
   - 事件：解绑
   - 条件：解绑后仍至少保留一种登录方式
4. `BOUND -> BOUND`
   - 事件：同类型换绑（可选）

---

### 3) 登录决策状态机（Auth Decision）

输入：`loginType + credential`  
输出：`ALLOW / DENY / NEED_BIND`

规则：
1. 找不到对应身份
   - 密码登录：`DENY`（通用文案）
   - 第三方登录：`NEED_BIND`（先登录再绑定）
2. 找到身份但账号非 `ACTIVE`
   - `DENY`（未激活/锁定/禁用）
3. 找到身份且 `ACTIVE`
   - `ALLOW`，签发会话

---

## 三、建议错误码

- `100001` 参数校验失败
- `100002` 资源不存在
- `100003` 无权限

认证类：
- `110001` 凭据错误（通用）
- `110002` 账号未激活
- `110003` 账号锁定
- `110004` 账号禁用
- `110005` 会话无效或过期

激活类：
- `120001` 激活 token 无效
- `120002` 激活 token 过期
- `120003` 激活 token 已使用
- `120004` 账号已激活
- `120005` 激活邮件发送频繁

验证码/风控：
- `130001` 验证码错误
- `130002` 验证码过期
- `130003` 验证码发送频繁
- `130004` 风控拦截（需二次验证）
- `130005` 二次验证失败

绑定类：
- `140001` 第三方身份未绑定
- `140002` 第三方身份已绑定其他账号
- `140003` 手机号已绑定其他账号
- `140004` 解绑失败（最后一种登录方式）
- `140005` bindTicket 无效或过期

---

## 四、关键时序（文本）

### A) 后台建号与激活
1. 管理员创建账号（邮箱）
2. 系统创建 `PENDING_ACTIVATION`
3. 发送激活邮件（带 token）
4. 用户点击链接，前端调用 `activation/validate`
5. 用户设置密码，调用 `activation/confirm`
6. 状态切换 `ACTIVE`，token 失效，返回成功

### B) 密码登录
1. 用户提交账号+密码
2. 校验账号与密码
3. 校验状态为 `ACTIVE`
4. 签发 token，记录日志，返回会话

### C) 第三方登录（未绑定）
1. 用户发起第三方登录（scene=login）
2. 回调换取第三方身份
3. 发现未绑定账号
4. 返回 `NEED_BIND_EXISTING_ACCOUNT + bindTicket`
5. 引导用户先用已绑定方式登录再绑定

### D) 登录后绑定第三方
1. 用户已登录，在账号中心发起绑定（scene=bind）
2. 完成授权并回调
3. 前端提交 `bindTicket` 到绑定接口
4. 后端校验唯一性并绑定
5. 返回绑定成功并刷新身份列表

---

## 五、最小数据模型（建议）

1. `accounts`
   - `id`, `email`, `username`, `display_name`, `status`, `activated_at`, `created_at`, `updated_at`

2. `account_credentials`
   - `account_id`, `password_hash`, `password_updated_at`

3. `account_identities`
   - `id`, `account_id`, `login_type`, `external_id`, `external_meta_json`, `created_at`, `deleted_at`
   - 唯一索引：
     - `(login_type, external_id)` 全局唯一
     - `(account_id, login_type)` 每账号每类型唯一（如不允许同类型多条）

4. `activation_tokens`
   - `token_hash`, `account_id`, `expires_at`, `used_at`, `created_at`

5. `otp_codes`
   - `scene`, `target`, `code_hash`, `expires_at`, `used_at`, `send_count`, `last_sent_at`

6. `audit_events`
   - `id`, `actor_id`, `target_account_id`, `event_type`, `result`, `ip`, `ua`, `metadata_json`, `created_at`

7. `sessions`（可选，如果需要服务端管理会话）
   - `id`, `account_id`, `refresh_token_hash`, `device_info_json`, `ip`, `expires_at`, `created_at`

---

## 六、补充接口

### 20) 刷新 Token
`POST /auth/token/refresh`

请求体：
```json
{
  "refreshToken": "rft_xxx"
}
```

响应 data：
```json
{
  "accessToken": "jwt_new_xxx",
  "refreshToken": "rft_new_xxx",
  "expiresIn": 7200,
  "tokenType": "Bearer"
}
```

关键校验：
- 验证 refreshToken 有效性与过期时间
- 签发新 accessToken + 轮换 refreshToken（旧的失效）

---

### 21) 敏感操作二次验证（获取 SecurityTicket）
`POST /me/security/verify`

- 用途：在执行敏感操作（绑定/解绑/改密码）前，先验证当前身份
- 鉴权：用户 token

请求体：
```json
{
  "password": "Current@123"
}
```

响应 data：
```json
{
  "securityTicket": "sec_xxx",
  "expireAt": "2026-03-26T12:40:00Z",
  "ttlSec": 300
}
```

关键校验：
- 验证密码正确性
- ticket 有效期 5 分钟，单次使用后失效

---

### 22) 管理员账号列表（分页）
`GET /admin/accounts?page=1&pageSize=20&status=ACTIVE&keyword=...`

- 用途：后台管理账号列表
- 鉴权：管理员 token

响应 data：
```json
{
  "items": [
    {
      "accountId": "acc_xxx",
      "email": "user@example.com",
      "username": "zhangsan",
      "displayName": "张三",
      "status": "ACTIVE",
      "identityCount": 2,
      "createdAt": "2026-03-26T12:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "pageSize": 20
}
```

---

### 23) 管理员锁定/解锁/禁用/恢复账号
`PUT /admin/accounts/{accountId}/status`

请求体：
```json
{
  "action": "lock",
  "reason": "异常登录行为"
}
```

- `action` 枚举：`lock` | `unlock` | `disable` | `enable`
- 写审计日志

响应 data：
```json
{
  "accountId": "acc_xxx",
  "status": "LOCKED",
  "updatedAt": "2026-03-26T15:00:00Z"
}
```

---

## 七、技术选型与依赖

### 密码哈希
- 算法：**bcrypt**（`golang.org/x/crypto/bcrypt`）
- Cost 因子：12（可配置）
- 封装为 `HashedPassword` 值对象，提供 `Hash(plain)` 和 `Verify(plain, hash)` 方法

### JWT
- 库：`github.com/golang-jwt/jwt/v5`
- AccessToken 过期时间：2 小时（可配置）
- RefreshToken 过期时间：7 天（可配置）
- Claims 结构：
```go
type Claims struct {
    AccountID   string `json:"accountId"`
    Role        string `json:"role"`        // "user" | "admin"
    jwt.RegisteredClaims
}
```
- 签名算法：HS256（使用配置中的 `jwt.secret`）

### Session / Token 管理
- 纯无状态 JWT（Phase 1）
- RefreshToken 存 DB（`sessions` 表），用于：
  - 轮换时使旧 token 失效
  - `revokeOtherSessions` 功能
  - 管理员强制下线

### 邮件发送
- 使用 SMTP 直连（Phase 1，简单可控）
- 封装为 `EmailSender` 接口（Domain 定义），Infra 层实现
- 后续可替换为消息队列异步发送

### 短信发送（Phase 3）
- 封装为 `SMSSender` 接口
- 实现可选：阿里云 SMS / 腾讯云 SMS
- 通过配置 `sms.provider` 切换

### OAuth 客户端
- 库：`golang.org/x/oauth2`
- 每个 Provider 封装为 `OAuthProvider` 接口的实现
- Provider 枚举与 scope：
  - GitHub：scope=`user:email`，user info endpoint=`https://api.github.com/user`
  - 微信扫码：scope=`snsapi_login`
  - 飞书：scope 参考飞书开放平台文档

### 验证码（OTP）存储
- 存储：**Redis**（TTL 自动过期）
- Key 格式：`otp:{scene}:{target}`（如 `otp:login:+8613800000000`）
- 频控 Key：`otp:rate:{target}`

### 激活 Token
- 生成：`crypto/rand` 生成 32 字节 → base64url 编码
- 存储时对 token 做 SHA-256 哈希（DB 存 hash，不存明文）
- 一次性使用，使用后标记 `used_at`

### 频控/限流
- **Redis + 固定窗口计数器**（Phase 1 简单方案）
- 维度：
  - 激活邮件重发：per accountId, 60 秒
  - 短信发送：per phone, 60 秒
  - 密码登录失败：per accountId, 5 次/30 分钟 → 锁定
- 封装为 `RateLimiter` 接口

### 人机验证（captchaToken）
- Phase 1 不实现，接口保留 `captchaToken` 字段但不做校验
- Phase 2+ 可接入腾讯天御/极验，封装为 `CaptchaVerifier` 接口

### OAuth State 防 CSRF
- 生成：`crypto/rand` 32 字节 → hex 编码
- 存储：Redis，Key=`oauth:state:{state}`，TTL=5 分钟
- 回调时验证 state 存在且未使用，验证后删除

### go.mod 新增依赖
```
github.com/golang-jwt/jwt/v5
golang.org/x/crypto           # bcrypt
golang.org/x/oauth2           # OAuth 客户端
github.com/redis/go-redis/v9  # Redis 客户端
```

---

## 八、领域模型设计

### 1) 聚合根与实体

#### Account（聚合根）

对应表：`accounts` + `account_credentials`

```go
// internal/domain/entity/account.go

type AccountStatus string
const (
    AccountStatusPendingActivation AccountStatus = "PENDING_ACTIVATION"
    AccountStatusActive            AccountStatus = "ACTIVE"
    AccountStatusLocked            AccountStatus = "LOCKED"
    AccountStatusDisabled          AccountStatus = "DISABLED"
)

type Account struct {
    ID            string
    Email         valueobject.Email
    Username      string
    DisplayName   string
    Status        AccountStatus
    PasswordHash  string          // 已哈希密码，PENDING 时为空
    Role          string          // "user" | "admin"
    ActivatedAt   *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// NewAccount 工厂方法 — 创建未激活账号（管理员调用）
func NewAccount(email, username, displayName string) (*Account, error)

// Activate 激活账号并设置密码
// 前置：Status == PENDING_ACTIVATION
// 后置：Status → ACTIVE, PasswordHash 设置, ActivatedAt 设置
func (a *Account) Activate(hashedPassword string) error

// Lock 锁定账号
// 前置：Status == ACTIVE
// 后置：Status → LOCKED
func (a *Account) Lock() error

// Unlock 解锁账号
// 前置：Status == LOCKED
// 后置：Status → ACTIVE
func (a *Account) Unlock() error

// Disable 禁用账号
// 前置：Status == ACTIVE || Status == LOCKED
// 后置：Status → DISABLED
func (a *Account) Disable() error

// Enable 恢复账号
// 前置：Status == DISABLED
// 后置：Status → ACTIVE
func (a *Account) Enable() error

// CanLogin 是否允许登录
func (a *Account) CanLogin() bool  // 仅 ACTIVE 返回 true

// ChangePassword 修改密码
// 前置：Status == ACTIVE
func (a *Account) ChangePassword(newHashedPassword string) error

// IsAdmin 是否管理员
func (a *Account) IsAdmin() bool
```

#### Identity（Account 的子实体）

对应表：`account_identities`

```go
// internal/domain/entity/identity.go

type LoginType string
const (
    LoginTypePassword  LoginType = "password"
    LoginTypeSMS       LoginType = "sms"
    LoginTypeWechatQR  LoginType = "wechat_qr"
    LoginTypeLarkQR    LoginType = "lark_qr"
    LoginTypeGithub    LoginType = "github_oauth"
)

type Identity struct {
    ID            string
    AccountID     string
    LoginType     LoginType
    ExternalID    string    // 密码登录为 email，第三方为 provider user id
    ExternalMeta  string    // JSON，存昵称、头像等
    CreatedAt     time.Time
    DeletedAt     *time.Time
}

// NewIdentity 创建身份绑定记录
func NewIdentity(accountID string, loginType LoginType, externalID string) (*Identity, error)

// CanUnbind 是否可解绑（剩余可用登录方式 > 1）
func CanUnbind(totalActiveCount int) bool
```

#### ActivationToken

对应表：`activation_tokens`

```go
// internal/domain/entity/activation_token.go

type ActivationToken struct {
    TokenHash   string
    AccountID   string
    ExpiresAt   time.Time
    UsedAt      *time.Time
    CreatedAt   time.Time
}

// NewActivationToken 创建激活令牌
func NewActivationToken(accountID string, rawToken string, ttl time.Duration) *ActivationToken

// IsValid 是否有效（未过期 + 未使用）
func (t *ActivationToken) IsValid() bool

// MarkUsed 标记已使用
func (t *ActivationToken) MarkUsed()
```

#### AuditEvent

对应表：`audit_events`

```go
// internal/domain/entity/audit_event.go

type AuditEvent struct {
    ID              string
    ActorID         string    // 操作人（管理员或用户自己）
    TargetAccountID string    // 被操作的账号
    EventType       string    // 见审计事件枚举
    Result          string    // "success" | "failure"
    IP              string
    UA              string
    Metadata        string    // JSON 扩展信息
    CreatedAt       time.Time
}

func NewAuditEvent(actorID, targetID, eventType, result, ip, ua string) *AuditEvent
```

### 2) Value Objects

```go
// internal/domain/valueobject/account_status.go
// AccountStatus 枚举 — 在 entity 中以 type + const 定义即可

// internal/domain/valueobject/login_type.go
// LoginType 枚举 — 同上

// internal/domain/valueobject/hashed_password.go
type HashedPassword struct { hash string }
func NewHashedPassword(plaintext string) (HashedPassword, error) // bcrypt hash
func HashFromStored(hash string) HashedPassword                  // 从 DB 加载
func (p HashedPassword) Verify(plaintext string) bool
func (p HashedPassword) String() string

// internal/domain/valueobject/phone.go
type Phone struct { number string }
func NewPhone(raw string) (Phone, error)  // 校验格式 +86...
func (p Phone) String() string
func (p Phone) Masked() string            // +861380****000

// internal/domain/valueobject/email.go — 复用现有的
```

### 3) Repository 接口

```go
// internal/domain/repository/account_repository.go
type AccountRepository interface {
    Create(ctx context.Context, account *entity.Account) error
    GetByID(ctx context.Context, id string) (*entity.Account, error)
    GetByEmail(ctx context.Context, email string) (*entity.Account, error)
    GetByUsername(ctx context.Context, username string) (*entity.Account, error)
    Update(ctx context.Context, account *entity.Account) error
    List(ctx context.Context, offset, limit int, status *entity.AccountStatus, keyword string) ([]*entity.Account, int64, error)
}

// internal/domain/repository/identity_repository.go
type IdentityRepository interface {
    Create(ctx context.Context, identity *entity.Identity) error
    GetByID(ctx context.Context, id string) (*entity.Identity, error)
    GetByProviderID(ctx context.Context, loginType entity.LoginType, externalID string) (*entity.Identity, error)
    ListByAccountID(ctx context.Context, accountID string) ([]*entity.Identity, error)
    CountActiveByAccountID(ctx context.Context, accountID string) (int, error)
    SoftDelete(ctx context.Context, id string) error
}

// internal/domain/repository/activation_token_repository.go
type ActivationTokenRepository interface {
    Create(ctx context.Context, token *entity.ActivationToken) error
    GetByHash(ctx context.Context, tokenHash string) (*entity.ActivationToken, error)
    InvalidateByAccountID(ctx context.Context, accountID string) error // 使某账号所有旧 token 失效
    MarkUsed(ctx context.Context, tokenHash string) error
}

// internal/domain/repository/audit_event_repository.go
type AuditEventRepository interface {
    Create(ctx context.Context, event *entity.AuditEvent) error
    ListByAccountID(ctx context.Context, accountID string, offset, limit int) ([]*entity.AuditEvent, int64, error)
    ListByFilter(ctx context.Context, filter AuditFilter) ([]*entity.AuditEvent, int64, error)
}

type AuditFilter struct {
    AccountID string
    EventType string
    From      *time.Time
    To        *time.Time
    Offset    int
    Limit     int
}
```

### 4) Domain 层外部服务接口（依赖倒置）

```go
// internal/domain/repository/email_sender.go
type EmailSender interface {
    SendActivationEmail(ctx context.Context, toEmail, activationLink string) error
}

// internal/domain/repository/sms_sender.go
type SMSSender interface {
    SendOTP(ctx context.Context, phone, code string) error
}

// internal/domain/repository/otp_store.go
type OTPStore interface {
    Save(ctx context.Context, scene, target, codeHash string, ttl time.Duration) error
    Get(ctx context.Context, scene, target string) (codeHash string, err error)
    Delete(ctx context.Context, scene, target string) error
    IncrSendCount(ctx context.Context, target string, window time.Duration) (count int, err error)
}

// internal/domain/repository/oauth_provider.go
type OAuthProvider interface {
    GetAuthURL(state, redirectURI string) string
    ExchangeCode(ctx context.Context, code string) (OAuthUserInfo, error)
}

type OAuthUserInfo struct {
    ProviderUserID string
    Email          string
    DisplayName    string
    AvatarURL      string
    RawJSON        string
}

// internal/domain/repository/session_store.go
type SessionStore interface {
    SaveRefreshToken(ctx context.Context, accountID, tokenHash string, ttl time.Duration) error
    ValidateRefreshToken(ctx context.Context, accountID, tokenHash string) (bool, error)
    RevokeRefreshToken(ctx context.Context, tokenHash string) error
    RevokeAllByAccountID(ctx context.Context, accountID string, exceptTokenHash string) (count int, error)
}
```

### 5) Domain Errors

```go
// internal/domain/errors/auth_errors.go

// 通用认证
var (
    ErrCredentialsInvalid   = errors.New("credentials invalid")
    ErrAccountNotActivated  = errors.New("account not activated")
    ErrAccountLocked        = errors.New("account is locked")
    ErrAccountDisabled      = errors.New("account is disabled")
    ErrSessionExpired       = errors.New("session expired or invalid")
)

// 激活
var (
    ErrActivationTokenInvalid = errors.New("activation token invalid")
    ErrActivationTokenExpired = errors.New("activation token expired")
    ErrActivationTokenUsed    = errors.New("activation token already used")
    ErrAlreadyActivated       = errors.New("account already activated")
    ErrActivationTooFrequent  = errors.New("activation email sent too frequently")
)

// 验证码
var (
    ErrOTPInvalid      = errors.New("otp code invalid")
    ErrOTPExpired      = errors.New("otp code expired")
    ErrOTPTooFrequent  = errors.New("otp code sent too frequently")
    ErrCaptchaRequired = errors.New("captcha verification required")
    ErrCaptchaFailed   = errors.New("captcha verification failed")
)

// 绑定
var (
    ErrIdentityNotBound       = errors.New("identity not bound to any account")
    ErrIdentityAlreadyBound   = errors.New("identity already bound to another account")
    ErrPhoneAlreadyBound      = errors.New("phone already bound to another account")
    ErrCannotUnbindLast       = errors.New("cannot unbind the last login method")
    ErrBindTicketInvalid      = errors.New("bind ticket invalid or expired")
    ErrSecurityTicketInvalid  = errors.New("security ticket invalid or expired")
)

// 状态迁移
var (
    ErrInvalidStatusTransition = errors.New("invalid account status transition")
)
```

### 6) Domain Error 到 HTTP 状态码 & 业务码映射

| Domain Error | HTTP Status | 业务 Code |
|-------------|-------------|-----------|
| `ErrCredentialsInvalid` | 401 | 110001 |
| `ErrAccountNotActivated` | 403 | 110002 |
| `ErrAccountLocked` | 403 | 110003 |
| `ErrAccountDisabled` | 403 | 110004 |
| `ErrSessionExpired` | 401 | 110005 |
| `ErrActivationTokenInvalid` | 400 | 120001 |
| `ErrActivationTokenExpired` | 400 | 120002 |
| `ErrActivationTokenUsed` | 400 | 120003 |
| `ErrAlreadyActivated` | 409 | 120004 |
| `ErrActivationTooFrequent` | 429 | 120005 |
| `ErrOTPInvalid` | 400 | 130001 |
| `ErrOTPExpired` | 400 | 130002 |
| `ErrOTPTooFrequent` | 429 | 130003 |
| `ErrCaptchaRequired` | 403 | 130004 |
| `ErrCaptchaFailed` | 403 | 130005 |
| `ErrIdentityNotBound` | 404 | 140001 |
| `ErrIdentityAlreadyBound` | 409 | 140002 |
| `ErrPhoneAlreadyBound` | 409 | 140003 |
| `ErrCannotUnbindLast` | 400 | 140004 |
| `ErrBindTicketInvalid` | 400 | 140005 |
| `ErrSecurityTicketInvalid` | 401 | 140006 |
| `ErrInvalidStatusTransition` | 400 | 100004 |

---

## 九、鉴权中间件设计

### Auth Middleware

```
位置：internal/ports/http/middleware/auth.go

功能：
1. 从请求头 Authorization: Bearer <token> 提取 JWT
2. 验证签名、过期时间
3. 解析 Claims，获取 accountId 和 role
4. 注入到 gin.Context：
   - c.Set("accountId", claims.AccountID)
   - c.Set("role", claims.Role)
5. 失败返回 { code: 110005, msg: "session expired or invalid" }
```

### Admin Middleware

```
位置：internal/ports/http/middleware/admin.go

功能：
1. 前置依赖 Auth Middleware（在路由注册时确保先注册 Auth 再注册 Admin）
2. 从 context 获取 role
3. role != "admin" → 返回 { code: 100003, msg: "no permission" }
```

### SecurityTicket 机制

```
获取 ticket：
1. 用户调用 POST /api/v1/me/security/verify，提交当前密码
2. 验证通过后生成 ticket（crypto/rand 32字节 → hex），存 Redis，TTL=5 分钟
3. 返回 ticket 给前端

使用 ticket：
1. 绑定手机（接口 14）、解绑（接口 16）、改密码（接口 17）的请求体中携带 securityTicket
2. 后端验证 Redis 中 ticket 存在且属于当前用户
3. 验证通过后从 Redis 删除（一次性使用）
```

### 角色模型

```
Phase 1 简单方案：
- accounts 表增加 role 字段，VARCHAR(20)，默认 "user"
- 仅两个角色：user / admin
- 管理员通过数据库直接设置，或通过 seed 脚本创建

Phase 2+ 可扩展为独立 RBAC 表
```

### 路由分组

```
// internal/ports/http/router/router.go

/api/v1/
├── auth/                         # 公开（无中间件）
│   ├── POST   activation/validate    → AuthHandler.ValidateActivation
│   ├── POST   activation/confirm     → AuthHandler.ConfirmActivation
│   ├── POST   login/password         → AuthHandler.LoginByPassword
│   ├── POST   login/sms/send        → AuthHandler.SendSMSCode
│   ├── POST   login/sms/verify      → AuthHandler.LoginBySMS
│   ├── GET    oauth/:provider/url   → AuthHandler.GetOAuthURL
│   ├── GET    oauth/:provider/callback → AuthHandler.OAuthCallback
│   ├── POST   token/refresh          → AuthHandler.RefreshToken
│   └── POST   logout                → AuthHandler.Logout          # 需 Auth
│
├── me/                           # 需 Auth Middleware
│   ├── GET    profile               → MeHandler.GetProfile
│   ├── GET    identities            → MeHandler.ListIdentities
│   ├── POST   identities/phone/bind → MeHandler.BindPhone
│   ├── POST   identities/:provider/bind → MeHandler.BindProvider
│   ├── DELETE identities/:identityId → MeHandler.UnbindIdentity
│   ├── POST   password/change       → MeHandler.ChangePassword
│   ├── POST   security/verify       → MeHandler.SecurityVerify
│   └── GET    security/logins       → MeHandler.LoginHistory
│
└── admin/                        # 需 Auth + Admin Middleware
    ├── POST   accounts              → AccountAdminHandler.Create
    ├── GET    accounts              → AccountAdminHandler.List
    ├── GET    accounts/:accountId   → AccountAdminHandler.GetDetail
    ├── PUT    accounts/:accountId/status → AccountAdminHandler.ChangeStatus
    ├── POST   accounts/:accountId/activation/resend → AccountAdminHandler.ResendActivation
    └── GET    audit/events          → AccountAdminHandler.ListAuditEvents
```

---

## 十、App 层 Service 设计

### AuthService（认证服务）

```go
// internal/app/service/auth_service.go

type AuthService interface {
    // 激活流程
    ValidateActivationToken(ctx context.Context, token string) (*dto.ActivationValidateResponse, error)
    ConfirmActivation(ctx context.Context, req *dto.ActivationConfirmRequest) (*dto.ActivationConfirmResponse, error)

    // 登录
    LoginByPassword(ctx context.Context, req *dto.PasswordLoginRequest) (*dto.LoginResponse, error)
    SendSMSCode(ctx context.Context, req *dto.SMSSendRequest) (*dto.SMSSendResponse, error)
    LoginBySMS(ctx context.Context, req *dto.SMSLoginRequest) (*dto.LoginResponse, error)

    // OAuth
    GetOAuthURL(ctx context.Context, provider, scene, redirectURI string) (*dto.OAuthURLResponse, error)
    HandleOAuthCallback(ctx context.Context, provider, code, state string) (*dto.OAuthCallbackResponse, error)

    // Token
    RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
    Logout(ctx context.Context, accountID string) error
}
```

### AccountAdminService（管理员账号服务）

```go
// internal/app/service/account_admin_service.go

type AccountAdminService interface {
    CreateAccount(ctx context.Context, req *dto.CreateAccountRequest, operatorID string) (*dto.AccountResponse, error)
    GetAccountDetail(ctx context.Context, accountID string) (*dto.AccountDetailResponse, error)
    ListAccounts(ctx context.Context, req *dto.ListAccountsRequest) (*dto.ListAccountsResponse, error)
    ChangeAccountStatus(ctx context.Context, accountID string, req *dto.ChangeStatusRequest, operatorID string) (*dto.AccountResponse, error)
    ResendActivation(ctx context.Context, accountID string, reason string, operatorID string) (*dto.ResendActivationResponse, error)
    ListAuditEvents(ctx context.Context, req *dto.AuditEventFilter) (*dto.AuditEventListResponse, error)
}
```

### MeService（账号中心服务）

```go
// internal/app/service/me_service.go

type MeService interface {
    GetProfile(ctx context.Context, accountID string) (*dto.ProfileResponse, error)
    ListIdentities(ctx context.Context, accountID string) (*dto.IdentityListResponse, error)
    BindPhone(ctx context.Context, accountID string, req *dto.BindPhoneRequest) (*dto.IdentityResponse, error)
    BindProvider(ctx context.Context, accountID string, req *dto.BindProviderRequest) (*dto.IdentityResponse, error)
    UnbindIdentity(ctx context.Context, accountID string, identityID string, ticket string) error
    ChangePassword(ctx context.Context, accountID string, req *dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error)
    SecurityVerify(ctx context.Context, accountID string, password string) (*dto.SecurityTicketResponse, error)
    LoginHistory(ctx context.Context, accountID string, offset, limit int) (*dto.AuditEventListResponse, error)
}
```

---

## 十一、DTO 设计

```go
// internal/app/dto/auth_dto.go

// --- 激活 ---
type ActivationValidateResponse struct {
    Valid            bool   `json:"valid"`
    EmailMasked      string `json:"emailMasked"`
    ExpireAt         string `json:"expireAt"`
    AlreadyActivated bool   `json:"alreadyActivated"`
}

type ActivationConfirmRequest struct {
    Token           string `json:"token" binding:"required"`
    Password        string `json:"password" binding:"required,min=8"`
    ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

type ActivationConfirmResponse struct {
    Activated   bool   `json:"activated"`
    AccountID   string `json:"accountId"`
    Status      string `json:"status"`
    ActivatedAt string `json:"activatedAt"`
}

// --- 登录 ---
type DeviceInfo struct {
    UA       string `json:"ua"`
    IP       string `json:"ip"`
    DeviceID string `json:"deviceId"`
}

type PasswordLoginRequest struct {
    Account    string     `json:"account" binding:"required"`      // 邮箱或用户名
    Password   string     `json:"password" binding:"required"`
    DeviceInfo DeviceInfo `json:"deviceInfo"`
}

type SMSSendRequest struct {
    Phone        string `json:"phone" binding:"required"`
    Scene        string `json:"scene" binding:"required,oneof=login bind"`
    CaptchaToken string `json:"captchaToken"`
}

type SMSSendResponse struct {
    Sent          bool `json:"sent"`
    TTLSec        int  `json:"ttlSec"`
    RetryAfterSec int  `json:"retryAfterSec"`
}

type SMSLoginRequest struct {
    Phone      string     `json:"phone" binding:"required"`
    Code       string     `json:"code" binding:"required,len=6"`
    DeviceInfo DeviceInfo `json:"deviceInfo"`
}

type LoginResponse struct {
    AccessToken  string          `json:"accessToken"`
    RefreshToken string          `json:"refreshToken"`
    ExpiresIn    int             `json:"expiresIn"`
    TokenType    string          `json:"tokenType"`
    Account      AccountBrief    `json:"account"`
}

type AccountBrief struct {
    AccountID   string `json:"accountId"`
    Email       string `json:"email"`
    DisplayName string `json:"displayName"`
    Status      string `json:"status"`
}

// --- OAuth ---
type OAuthURLResponse struct {
    AuthURL  string `json:"authUrl"`
    State    string `json:"state"`
    ExpireAt string `json:"expireAt"`
}

type OAuthCallbackResponse struct {
    Result       string          `json:"result"`       // LOGIN_SUCCESS | NEED_BIND_EXISTING_ACCOUNT | BIND_SUCCESS
    AccessToken  string          `json:"accessToken,omitempty"`
    RefreshToken string          `json:"refreshToken,omitempty"`
    ExpiresIn    int             `json:"expiresIn,omitempty"`
    Account      *AccountBrief   `json:"account,omitempty"`
    Provider     string          `json:"provider,omitempty"`
    BindTicket   string          `json:"bindTicket,omitempty"`
    ExpireAt     string          `json:"expireAt,omitempty"`
    Identity     *IdentityBrief  `json:"identity,omitempty"`
}

// internal/app/dto/account_dto.go

type CreateAccountRequest struct {
    Email       string `json:"email" binding:"required,email"`
    DisplayName string `json:"displayName" binding:"required,min=1,max=50"`
    Username    string `json:"username" binding:"required,min=2,max=50"`
}

type AccountResponse struct {
    AccountID          string `json:"accountId"`
    Email              string `json:"email"`
    Username           string `json:"username"`
    DisplayName        string `json:"displayName"`
    Status             string `json:"status"`
    Role               string `json:"role"`
    ActivationExpireAt string `json:"activationExpireAt,omitempty"`
    ActivatedAt        string `json:"activatedAt,omitempty"`
    CreatedAt          string `json:"createdAt"`
    UpdatedAt          string `json:"updatedAt"`
}

type AccountDetailResponse struct {
    AccountResponse
    Identities []IdentityBrief `json:"identities"`
}

type ListAccountsRequest struct {
    Page     int    `form:"page" binding:"omitempty,min=1"`
    PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
    Status   string `form:"status" binding:"omitempty,oneof=PENDING_ACTIVATION ACTIVE LOCKED DISABLED"`
    Keyword  string `form:"keyword"`
}

type ListAccountsResponse struct {
    Items    []AccountResponse `json:"items"`
    Total    int64             `json:"total"`
    Page     int               `json:"page"`
    PageSize int               `json:"pageSize"`
}

type ChangeStatusRequest struct {
    Action string `json:"action" binding:"required,oneof=lock unlock disable enable"`
    Reason string `json:"reason"`
}

type ResendActivationResponse struct {
    Sent               bool   `json:"sent"`
    ActivationExpireAt string `json:"activationExpireAt"`
    ThrottleSec        int    `json:"throttleSec"`
}

// internal/app/dto/identity_dto.go

type IdentityBrief struct {
    IdentityID       string `json:"identityId"`
    LoginType        string `json:"loginType"`
    MaskedIdentifier string `json:"maskedIdentifier,omitempty"`
    BoundAt          string `json:"boundAt"`
}

type IdentityListResponse struct {
    Identities   []IdentityBrief `json:"identities"`
    CanUnbindLast bool           `json:"canUnbindLast"`
}

type IdentityResponse struct {
    IdentityID string `json:"identityId"`
    LoginType  string `json:"loginType"`
    BoundAt    string `json:"boundAt"`
}

type BindPhoneRequest struct {
    Phone          string `json:"phone" binding:"required"`
    SMSCode        string `json:"smsCode" binding:"required,len=6"`
    SecurityTicket string `json:"securityTicket" binding:"required"`
}

type BindProviderRequest struct {
    BindTicket string `json:"bindTicket" binding:"required"`
}

type ChangePasswordRequest struct {
    OldPassword         string `json:"oldPassword" binding:"required"`
    NewPassword         string `json:"newPassword" binding:"required,min=8"`
    ConfirmPassword     string `json:"confirmPassword" binding:"required,eqfield=NewPassword"`
    RevokeOtherSessions bool   `json:"revokeOtherSessions"`
}

type ChangePasswordResponse struct {
    Changed         bool `json:"changed"`
    RevokedSessions int  `json:"revokedSessions"`
}

type ProfileResponse struct {
    AccountID     string `json:"accountId"`
    Email         string `json:"email"`
    Username      string `json:"username"`
    DisplayName   string `json:"displayName"`
    AccountStatus string `json:"accountStatus"`
    Role          string `json:"role"`
}

type SecurityTicketResponse struct {
    SecurityTicket string `json:"securityTicket"`
    ExpireAt       string `json:"expireAt"`
    TTLSec         int    `json:"ttlSec"`
}

type AuditEventItem struct {
    ID              string `json:"id"`
    EventType       string `json:"eventType"`
    Result          string `json:"result"`
    IP              string `json:"ip"`
    UA              string `json:"ua"`
    CreatedAt       string `json:"createdAt"`
}

type AuditEventFilter struct {
    AccountID string `form:"accountId"`
    EventType string `form:"eventType"`
    From      string `form:"from"`
    To        string `form:"to"`
    Page      int    `form:"page" binding:"omitempty,min=1"`
    PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

type AuditEventListResponse struct {
    Items    []AuditEventItem `json:"items"`
    Total    int64            `json:"total"`
    Page     int              `json:"page"`
    PageSize int              `json:"pageSize"`
}
```

---

## 十二、事务边界与并发控制

### 需要事务的场景

| 场景 | 涉及操作 | 事务策略 |
|------|---------|---------|
| **创建账号** | INSERT accounts + INSERT activation_tokens + 发邮件 | 事务包裹 DB 操作；邮件在事务提交后异步发送（失败可重试） |
| **激活并设密码** | UPDATE accounts(status→ACTIVE) + INSERT account_credentials + UPDATE activation_tokens(used_at) | 同一事务 |
| **绑定第三方** | INSERT account_identities + 唯一索引校验 | 事务；利用 `(login_type, external_id)` 唯一索引防并发重复绑定 |
| **解绑登录方式** | SELECT COUNT + UPDATE account_identities(deleted_at) | 事务 + 悲观锁（`SELECT ... FOR UPDATE`），防并发解绑到 0 |
| **修改密码** | UPDATE account_credentials + 吊销其他 session | 事务包裹密码更新；session 吊销可在事务后操作 Redis |
| **管理员状态变更** | UPDATE accounts(status) + INSERT audit_events | 同一事务 |

### 并发控制策略

| 场景 | 策略 |
|------|------|
| 激活 token 一次性使用 | 乐观锁：`UPDATE activation_tokens SET used_at=NOW() WHERE token_hash=? AND used_at IS NULL`，检查 RowsAffected == 1 |
| 邮箱/用户名唯一 | 数据库唯一索引，捕获 duplicate entry 错误转为 domain error |
| 第三方身份全局唯一 | `(login_type, external_id)` 唯一索引 |
| 密码登录失败计数 | Redis INCR + EXPIRE，原子操作 |
| OTP 验证码使用 | Redis DEL（原子删除），或使用 Lua 脚本 GET+DEL |

### Service 层事务模式

```go
// 推荐在 Service 层通过注入 TransactionManager 来管理事务

// internal/domain/repository/transaction.go
type TransactionManager interface {
    RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// 用法示例（ActivationService.ConfirmActivation）：
func (s *AuthServiceImpl) ConfirmActivation(ctx context.Context, req *dto.ActivationConfirmRequest) error {
    return s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
        // 1. 查询并校验 token
        // 2. 更新 account status → ACTIVE
        // 3. 插入 account_credentials
        // 4. 标记 token 已使用
        return nil
    })
}
```

---

## 十三、审计事件枚举

```go
// event_type 枚举值

// 账号管理
const (
    AuditAccountCreated          = "ACCOUNT_CREATED"           // 管理员创建账号
    AuditActivationEmailSent     = "ACTIVATION_EMAIL_SENT"     // 激活邮件发送
    AuditActivationEmailResent   = "ACTIVATION_EMAIL_RESENT"   // 激活邮件重发
    AuditAccountActivated        = "ACCOUNT_ACTIVATED"         // 账号激活
    AuditAccountLocked           = "ACCOUNT_LOCKED"            // 账号锁定
    AuditAccountUnlocked         = "ACCOUNT_UNLOCKED"          // 账号解锁
    AuditAccountDisabled         = "ACCOUNT_DISABLED"          // 账号禁用
    AuditAccountEnabled          = "ACCOUNT_ENABLED"           // 账号恢复
)

// 认证
const (
    AuditLoginSuccess            = "LOGIN_SUCCESS"             // 登录成功
    AuditLoginFailed             = "LOGIN_FAILED"              // 登录失败
    AuditLogout                  = "LOGOUT"                    // 退出登录
    AuditTokenRefreshed          = "TOKEN_REFRESHED"           // Token 刷新
)

// 身份绑定
const (
    AuditIdentityBound           = "IDENTITY_BOUND"            // 绑定登录方式
    AuditIdentityUnbound         = "IDENTITY_UNBOUND"          // 解绑登录方式
)

// 安全操作
const (
    AuditPasswordChanged         = "PASSWORD_CHANGED"          // 密码修改
    AuditSecurityVerified        = "SECURITY_VERIFIED"         // 二次验证通过
    AuditSessionsRevoked         = "SESSIONS_REVOKED"          // 其他会话吊销
)
```

### 必须记录审计日志的操作

- 创建账号 ✓
- 激活成功 ✓
- 每次登录（成功/失败）✓
- 密码修改 ✓
- 绑定/解绑 ✓
- 管理员状态变更 ✓
- 退出登录 ✓

### result 字段

- `"success"` — 操作成功
- `"failure"` — 操作失败（metadata 中记录原因）

---

## 十四、配置项扩展

在现有 `config.yaml` 基础上新增：

```yaml
# config.yaml 新增部分

auth:
  jwt:
    secret: "${JWT_SECRET}"             # 必填，至少 32 字符
    access_token_expiry: 7200           # 秒，默认 2 小时
    refresh_token_expiry: 604800        # 秒，默认 7 天
  activation:
    token_expiry: 86400                 # 秒，默认 24 小时
    resend_cooldown: 60                 # 秒，默认 60 秒
  password:
    min_length: 8
    bcrypt_cost: 12
  lockout:
    max_attempts: 5                     # 密码错误最大次数
    lock_duration: 1800                 # 锁定时长（秒），默认 30 分钟
  security_ticket_ttl: 300              # 二次验证 ticket 有效期（秒）

oauth:
  github:
    client_id: "${GITHUB_CLIENT_ID}"
    client_secret: "${GITHUB_CLIENT_SECRET}"
    redirect_uri: "${GITHUB_REDIRECT_URI}"
  wechat:
    app_id: "${WECHAT_APP_ID}"
    app_secret: "${WECHAT_APP_SECRET}"
    redirect_uri: "${WECHAT_REDIRECT_URI}"
  lark:
    app_id: "${LARK_APP_ID}"
    app_secret: "${LARK_APP_SECRET}"
    redirect_uri: "${LARK_REDIRECT_URI}"

sms:
  provider: "aliyun"                    # aliyun | tencent
  access_key_id: "${SMS_ACCESS_KEY_ID}"
  access_key_secret: "${SMS_ACCESS_KEY_SECRET}"
  sign_name: "${SMS_SIGN_NAME}"
  template_code: "${SMS_TEMPLATE_CODE}"
  otp_ttl: 300                          # 验证码有效期（秒）
  otp_cooldown: 60                      # 验证码发送间隔（秒）
  otp_length: 6                         # 验证码长度

email:
  smtp:
    host: "${SMTP_HOST}"
    port: 587
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
    from: "${SMTP_FROM}"                # 发件人地址
  activation:
    subject: "激活您的账号"
    base_url: "${ACTIVATION_BASE_URL}"  # 前端激活页面 base URL
```

对应 Go 配置结构体扩展：

```go
// internal/infra/config/config.go 新增

type AuthConfig struct {
    JWT             JWTConfig      `yaml:"jwt"`
    Activation      ActivationConf `yaml:"activation"`
    Password        PasswordConf   `yaml:"password"`
    Lockout         LockoutConf    `yaml:"lockout"`
    SecurityTicketTTL int          `yaml:"security_ticket_ttl"`
}

type JWTConfig struct {
    Secret            string `yaml:"secret"`
    AccessTokenExpiry  int   `yaml:"access_token_expiry"`
    RefreshTokenExpiry int   `yaml:"refresh_token_expiry"`
}

type ActivationConf struct {
    TokenExpiry    int `yaml:"token_expiry"`
    ResendCooldown int `yaml:"resend_cooldown"`
}

type PasswordConf struct {
    MinLength  int `yaml:"min_length"`
    BcryptCost int `yaml:"bcrypt_cost"`
}

type LockoutConf struct {
    MaxAttempts  int `yaml:"max_attempts"`
    LockDuration int `yaml:"lock_duration"`
}

type OAuthConfig struct {
    GitHub OAuthProviderConf `yaml:"github"`
    WeChat OAuthProviderConf `yaml:"wechat"`
    Lark   OAuthProviderConf `yaml:"lark"`
}

type OAuthProviderConf struct {
    ClientID     string `yaml:"client_id"`
    ClientSecret string `yaml:"client_secret"`
    RedirectURI  string `yaml:"redirect_uri"`
}

type SMSConfig struct {
    Provider        string `yaml:"provider"`
    AccessKeyID     string `yaml:"access_key_id"`
    AccessKeySecret string `yaml:"access_key_secret"`
    SignName        string `yaml:"sign_name"`
    TemplateCode    string `yaml:"template_code"`
    OTPTTL          int    `yaml:"otp_ttl"`
    OTPCooldown     int    `yaml:"otp_cooldown"`
    OTPLength       int    `yaml:"otp_length"`
}

type EmailConfig struct {
    SMTP       SMTPConfig       `yaml:"smtp"`
    Activation ActivationEmail  `yaml:"activation"`
}

type SMTPConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    From     string `yaml:"from"`
}

type ActivationEmail struct {
    Subject string `yaml:"subject"`
    BaseURL string `yaml:"base_url"`
}

// Config 主结构新增字段：
// Auth  AuthConfig  `yaml:"auth"`
// OAuth OAuthConfig `yaml:"oauth"`
// SMS   SMSConfig   `yaml:"sms"`
// Email EmailConfig `yaml:"email"`
```

---

## 十五、数据库迁移 SQL

### 000002_create_accounts_table.up.sql

```sql
CREATE TABLE IF NOT EXISTS accounts (
    id          VARCHAR(36)  PRIMARY KEY,
    email       VARCHAR(255) NOT NULL,
    username    VARCHAR(50)  NOT NULL,
    display_name VARCHAR(100) NOT NULL DEFAULT '',
    status      VARCHAR(30)  NOT NULL DEFAULT 'PENDING_ACTIVATION',
    role        VARCHAR(20)  NOT NULL DEFAULT 'user',
    activated_at DATETIME(3) NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_username (username),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000003_create_account_credentials.up.sql

```sql
CREATE TABLE IF NOT EXISTS account_credentials (
    account_id       VARCHAR(36) PRIMARY KEY,
    password_hash    VARCHAR(255) NOT NULL,
    password_updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT fk_credentials_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000004_create_account_identities.up.sql

```sql
CREATE TABLE IF NOT EXISTS account_identities (
    id          VARCHAR(36)  PRIMARY KEY,
    account_id  VARCHAR(36)  NOT NULL,
    login_type  VARCHAR(30)  NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    external_meta_json TEXT,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)  NULL,
    UNIQUE KEY uk_login_type_external_id (login_type, external_id),
    UNIQUE KEY uk_account_login_type (account_id, login_type),
    KEY idx_account_id (account_id),
    CONSTRAINT fk_identity_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000005_create_activation_tokens.up.sql

```sql
CREATE TABLE IF NOT EXISTS activation_tokens (
    token_hash  VARCHAR(64)  PRIMARY KEY,
    account_id  VARCHAR(36)  NOT NULL,
    expires_at  DATETIME(3)  NOT NULL,
    used_at     DATETIME(3)  NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_account_id (account_id),
    KEY idx_expires_at (expires_at),
    CONSTRAINT fk_activation_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000006_create_audit_events.up.sql

```sql
CREATE TABLE IF NOT EXISTS audit_events (
    id                VARCHAR(36)  PRIMARY KEY,
    actor_id          VARCHAR(36)  NOT NULL DEFAULT '',
    target_account_id VARCHAR(36)  NOT NULL DEFAULT '',
    event_type        VARCHAR(50)  NOT NULL,
    result            VARCHAR(10)  NOT NULL DEFAULT 'success',
    ip                VARCHAR(45)  NOT NULL DEFAULT '',
    ua                VARCHAR(500) NOT NULL DEFAULT '',
    metadata_json     TEXT,
    created_at        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_target_account (target_account_id),
    KEY idx_event_type (event_type),
    KEY idx_created_at (created_at),
    KEY idx_actor (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000007_create_sessions.up.sql

```sql
CREATE TABLE IF NOT EXISTS sessions (
    id                 VARCHAR(36) PRIMARY KEY,
    account_id         VARCHAR(36) NOT NULL,
    refresh_token_hash VARCHAR(64) NOT NULL,
    device_info_json   TEXT,
    ip                 VARCHAR(45) NOT NULL DEFAULT '',
    expires_at         DATETIME(3) NOT NULL,
    created_at         DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_account_id (account_id),
    KEY idx_refresh_token (refresh_token_hash),
    KEY idx_expires_at (expires_at),
    CONSTRAINT fk_session_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

注：`otp_codes` 不建表，使用 Redis 存储（Key: `otp:{scene}:{target}`，TTL 自动过期）。

---

## 十六、SQLC Queries

### sqlc/queries/account.sql

```sql
-- name: GetAccountByID :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE id = ?;

-- name: GetAccountByEmail :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE email = ?;

-- name: GetAccountByUsername :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE username = ?;

-- name: CreateAccount :execresult
INSERT INTO accounts (id, email, username, display_name, status, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAccountStatus :exec
UPDATE accounts SET status = ?, updated_at = ? WHERE id = ?;

-- name: UpdateAccountActivation :exec
UPDATE accounts SET status = 'ACTIVE', activated_at = ?, updated_at = ? WHERE id = ?;

-- name: ListAccounts :many
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountAccounts :one
SELECT COUNT(*) FROM accounts;
```

### sqlc/queries/identity.sql

```sql
-- name: CreateIdentity :execresult
INSERT INTO account_identities (id, account_id, login_type, external_id, external_meta_json, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetIdentityByProviderID :one
SELECT id, account_id, login_type, external_id, external_meta_json, created_at, deleted_at
FROM account_identities WHERE login_type = ? AND external_id = ? AND deleted_at IS NULL;

-- name: ListIdentitiesByAccountID :many
SELECT id, account_id, login_type, external_id, external_meta_json, created_at, deleted_at
FROM account_identities WHERE account_id = ? AND deleted_at IS NULL;

-- name: CountActiveIdentities :one
SELECT COUNT(*) FROM account_identities WHERE account_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteIdentity :exec
UPDATE account_identities SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
```

### sqlc/queries/activation_token.sql

```sql
-- name: CreateActivationToken :execresult
INSERT INTO activation_tokens (token_hash, account_id, expires_at, created_at) VALUES (?, ?, ?, ?);

-- name: GetActivationTokenByHash :one
SELECT token_hash, account_id, expires_at, used_at, created_at
FROM activation_tokens WHERE token_hash = ?;

-- name: MarkActivationTokenUsed :exec
UPDATE activation_tokens SET used_at = ? WHERE token_hash = ? AND used_at IS NULL;

-- name: InvalidateTokensByAccountID :exec
UPDATE activation_tokens SET used_at = NOW() WHERE account_id = ? AND used_at IS NULL;
```

### sqlc/queries/credential.sql

```sql
-- name: CreateCredential :execresult
INSERT INTO account_credentials (account_id, password_hash, password_updated_at) VALUES (?, ?, ?);

-- name: GetCredentialByAccountID :one
SELECT account_id, password_hash, password_updated_at
FROM account_credentials WHERE account_id = ?;

-- name: UpdateCredentialPassword :exec
UPDATE account_credentials SET password_hash = ?, password_updated_at = ? WHERE account_id = ?;
```

### sqlc/queries/audit_event.sql

```sql
-- name: CreateAuditEvent :execresult
INSERT INTO audit_events (id, actor_id, target_account_id, event_type, result, ip, ua, metadata_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListAuditEventsByAccount :many
SELECT id, actor_id, target_account_id, event_type, result, ip, ua, metadata_json, created_at
FROM audit_events WHERE target_account_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountAuditEventsByAccount :one
SELECT COUNT(*) FROM audit_events WHERE target_account_id = ?;
```

### sqlc/queries/session.sql

```sql
-- name: CreateSession :execresult
INSERT INTO sessions (id, account_id, refresh_token_hash, device_info_json, ip, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetSessionByRefreshTokenHash :one
SELECT id, account_id, refresh_token_hash, device_info_json, ip, expires_at, created_at
FROM sessions WHERE refresh_token_hash = ? AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE refresh_token_hash = ?;

-- name: DeleteSessionsByAccountID :exec
DELETE FROM sessions WHERE account_id = ? AND refresh_token_hash != ?;

-- name: CountSessionsByAccountID :one
SELECT COUNT(*) FROM sessions WHERE account_id = ? AND expires_at > NOW();
```

---

## 十七、文件生成清单

按 DDD 分层列出所有需要创建的文件路径：

### Domain 层 — `internal/domain/`

```
entity/
  account.go                    # Account 聚合根 + AccountStatus 枚举 + 工厂 + 业务方法
  identity.go                   # Identity 子实体 + LoginType 枚举 + CanUnbind
  activation_token.go           # ActivationToken 实体
  audit_event.go                # AuditEvent 实体 + 事件类型常量

valueobject/
  hashed_password.go            # 密码哈希值对象 (bcrypt)
  phone.go                      # 手机号值对象
  # email.go                    # 复用现有

repository/
  account_repository.go         # AccountRepository 接口
  identity_repository.go        # IdentityRepository 接口
  activation_token_repository.go # ActivationTokenRepository 接口
  audit_event_repository.go     # AuditEventRepository 接口 + AuditFilter
  email_sender.go               # EmailSender 接口
  sms_sender.go                 # SMSSender 接口
  otp_store.go                  # OTPStore 接口 (Redis)
  oauth_provider.go             # OAuthProvider 接口 + OAuthUserInfo
  session_store.go              # SessionStore 接口
  transaction.go                # TransactionManager 接口

errors/
  auth_errors.go                # 认证/激活/绑定相关 sentinel errors
```

### App 层 — `internal/app/`

```
dto/
  auth_dto.go                   # 登录/激活/OAuth/Token 相关 DTO
  account_dto.go                # 账号管理 DTO（创建/列表/状态变更）
  identity_dto.go               # 身份绑定 DTO

service/
  auth_service_iface.go         # AuthService 接口 + autowire 注册
  auth_service.go               # AuthServiceImpl 实现
  account_admin_service_iface.go # AccountAdminService 接口 + autowire 注册
  account_admin_service.go      # AccountAdminServiceImpl 实现
  me_service_iface.go           # MeService 接口 + autowire 注册
  me_service.go                 # MeServiceImpl 实现
```

### Infra 层 — `internal/infra/`

```
persistence/mysql/
  account_repository/
    model.go                    # AccountModel + GORM 表映射 + Auto 注册
    account_repository.go       # AccountRepository GORM 实现
  identity_repository/
    model.go                    # IdentityModel + Auto 注册
    identity_repository.go      # IdentityRepository GORM 实现
  activation_token_repository/
    model.go                    # ActivationTokenModel + Auto 注册
    activation_token_repository.go
  audit_event_repository/
    model.go                    # AuditEventModel + Auto 注册
    audit_event_repository.go
  session_repository/
    model.go                    # SessionModel + Auto 注册
    session_repository.go
  credential_repository/
    model.go                    # CredentialModel + Auto 注册
    credential_repository.go
  transaction.go                # TransactionManager GORM 实现

auth/
  jwt.go                        # JWT 签发 / 解析 / Claims 定义
  password.go                   # bcrypt 哈希 / 验证（HashedPassword 实现）

cache/
  otp_store.go                  # OTPStore Redis 实现
  rate_limiter.go               # RateLimiter Redis 实现
  security_ticket.go            # SecurityTicket Redis 存取
  oauth_state.go                # OAuth state Redis 存取
  bind_ticket.go                # BindTicket Redis 存取

external/
  email_sender.go               # SMTP EmailSender 实现
  sms_sender.go                 # SMS 发送实现（阿里云/腾讯云）
  oauth/
    github.go                   # GitHub OAuthProvider 实现
    wechat.go                   # 微信 OAuthProvider 实现
    lark.go                     # 飞书 OAuthProvider 实现
    provider_factory.go         # 根据 provider name 返回对应实现

config/
  # config.go 已有，扩展 AuthConfig/OAuthConfig/SMSConfig/EmailConfig
```

### Ports 层 — `internal/ports/http/`

```
handler/
  auth_handler_iface.go         # AuthHandler 接口 + autowire 注册
  auth_handler.go               # Gin handler 实现（登录/激活/OAuth/Token）
  account_admin_handler_iface.go # AccountAdminHandler 接口 + autowire 注册
  account_admin_handler.go      # Gin handler 实现（管理员操作）
  me_handler_iface.go           # MeHandler 接口 + autowire 注册
  me_handler.go                 # Gin handler 实现（账号中心）

middleware/
  auth.go                       # JWT 认证中间件
  admin.go                      # 管理员权限中间件

router/
  # router.go 扩展，新增 auth/me/admin 路由组
```

### 迁移文件 — `migrations/`

```
000002_create_accounts_table.up.sql
000002_create_accounts_table.down.sql
000003_create_account_credentials.up.sql
000003_create_account_credentials.down.sql
000004_create_account_identities.up.sql
000004_create_account_identities.down.sql
000005_create_activation_tokens.up.sql
000005_create_activation_tokens.down.sql
000006_create_audit_events.up.sql
000006_create_audit_events.down.sql
000007_create_sessions.up.sql
000007_create_sessions.down.sql
```

### SQLC Queries — `sqlc/queries/`

```
account.sql
identity.sql
activation_token.sql
credential.sql
audit_event.sql
session.sql
```

### 前端（Phase 2+）— `web/src/`

```
types/auth.ts                   # 类型定义
api/auth.ts                     # API 调用
stores/auth.ts                  # Pinia store
composables/useAuth.ts          # 组合式函数
views/login/index.vue           # 登录页
views/activation/index.vue      # 激活页
views/account/index.vue         # 账号中心
views/admin/accounts/index.vue  # 管理后台-账号管理
```

---

## 十八、前端页面流转

### 激活流程

```
1. 用户收到激活邮件，链接格式：
   ${ACTIVATION_BASE_URL}/activation?token=xxx

2. 前端激活页面：
   - 页面加载时调用 GET /api/v1/auth/activation/validate?token=xxx
   - 有效 → 显示"设置密码"表单
   - 无效/过期 → 显示错误提示 + "联系管理员重发"按钮
   - 已激活 → 跳转登录页

3. 用户提交密码后调用 POST /api/v1/auth/activation/confirm
   - 成功 → 跳转登录页，提示"激活成功，请登录"
```

### OAuth 回调流程

```
1. 前端调用 GET /api/v1/auth/oauth/{provider}/url?scene=login&redirectUri=...
2. 前端跳转到 authUrl（或弹窗）
3. 第三方回调到后端 GET /api/v1/auth/oauth/{provider}/callback?code=...&state=...
4. 后端处理后通过 302 重定向到前端页面，URL 携带结果参数：
   - 登录成功：/oauth/callback?result=LOGIN_SUCCESS&token=...
   - 需要绑定：/oauth/callback?result=NEED_BIND&bindTicket=...&provider=...
5. 前端 /oauth/callback 页面根据 result 参数：
   - LOGIN_SUCCESS → 存储 token，跳转首页
   - NEED_BIND → 跳转到登录页，提示"请先登录已有账号再绑定"
```

### 密码登录 `account` 字段

- 同时支持**邮箱**和**用户名**登录
- 后端按格式判断：包含 `@` 则按邮箱查询，否则按用户名查询

---

## 十九、开发分期

### Phase 1（核心可用）

- 账号 CRUD（管理员创建、查询、列表、状态变更）
- 激活流程（发邮件、校验、设密码）
- 密码登录
- JWT 签发 / 刷新 / 退出
- Auth + Admin 中间件
- 审计日志（写入 + 管理员查询）
- 账号中心（我的资料、修改密码）

**产出：一个可用的"邮箱邀请制 + 密码登录"系统**

### Phase 2（第三方登录）

- GitHub OAuth 登录 + 绑定/解绑
- 微信扫码登录 + 绑定/解绑
- 飞书登录 + 绑定/解绑
- identities 管理界面
- SecurityTicket 二次验证

### Phase 3（短信登录）

- SMS 验证码发送
- 短信登录
- 绑定手机号
- 人机验证接入

### Phase 4（增强安全）

- 登录失败自动锁定
- 异地登录检测（DeviceInfo 比对）
- 频控/限流完善
- 登录记录页面
- 管理员强制下线

---

## 二十、密码校验规则

```
规则（可通过配置调整）：
- 最少 8 个字符
- 至少包含 1 个大写字母
- 至少包含 1 个小写字母
- 至少包含 1 个数字
- 至少包含 1 个特殊字符（!@#$%^&*...）
- 不能与旧密码相同（修改密码时）

校验位置：
- 前端：DTO binding tag + 前端表单校验（即时反馈）
- 后端：HashedPassword.Hash() 内或 Service 层校验（安全保障）
```
