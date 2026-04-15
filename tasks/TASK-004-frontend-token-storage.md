# TASK-004 前端 Token 存储安全加固

**优先级**: 高
**模块**: web
**涉及层**: Frontend

---

## 背景

当前前端将 `accessToken` 和 `refreshToken` 存储在 `localStorage` 中。`localStorage` 可被同源的任何 JavaScript 代码访问，一旦发生 XSS 漏洞，攻击者可直接窃取 token。

## 问题定位

```
web/src/utils/request.ts (或 web/src/api/client.ts)
localStorage.setItem('token', accessToken)
localStorage.setItem('refreshToken', refreshToken)
```

## 方案对比

| 方案 | 安全性 | 改动范围 | 兼容性 |
|------|--------|----------|--------|
| A. httpOnly Cookie | 最高 | 后端需改造 | 需 CSRF 防护 |
| B. sessionStorage | 中等 | 前端仅改存储位置 | 关闭标签即失效 |
| C. 内存 + Silent Refresh | 较高 | 前端改存储 + 刷新逻辑 | 较复杂 |

## 推荐方案：A. httpOnly Cookie

### 后端改动

1. 登录/刷新接口将 token 写入 `Set-Cookie` 响应头，而非 body
2. Cookie 属性设置：`HttpOnly; Secure; SameSite=Strict; Path=/api`
3. `accessToken` 和 `refreshToken` 分别设置不同的过期时间
4. 登录响应 body 仍返回用户信息（`AccountBrief`），不再返回 token 字符串

```go
// 登录成功后设置 Cookie
c.SetCookie("access_token", accessToken, jwtMgr.AccessTokenExpiry(), "/api", "", true, true)
c.SetCookie("refresh_token", refreshToken, int(jwtMgr.RefreshTokenExpiry().Seconds()), "/api/auth", "", true, true)
c.SetSameSite(http.SameSiteStrictMode)
```

### 前端改动

1. 移除 `localStorage.setItem('token', ...)` 逻辑
2. Axios 配置 `withCredentials: true`
3. 移除请求拦截器中的 `Authorization` 头注入（Cookie 自动携带）
4. 刷新 token 改为 `POST /api/v1/auth/refresh`（Cookie 自动携带 refresh_token）

### CSRF 防护

由于使用 Cookie 认证，需要增加 CSRF 防护：

1. 后端签发 `XSRF-TOKEN` Cookie（非 httpOnly）
2. 前端读取 `XSRF-TOKEN` 并在请求头 `X-XSRF-TOKEN` 中回传
3. 后端中间件校验 `X-XSRF-TOKEN` 与 Cookie 值匹配

## 验收标准

- [ ] token 不再存储在 localStorage
- [ ] Cookie 设置 HttpOnly + Secure + SameSite=Strict
- [ ] 前端请求自动携带 Cookie，无需手动注入 Authorization 头
- [ ] CSRF 防护机制生效
- [ ] 登出接口清除 Cookie
- [ ] 多标签页共享登录状态

## 相关文件

- `web/src/api/client.ts` — Axios 实例配置
- `web/src/stores/auth.ts` — Token 存储逻辑
- `web/src/utils/request.ts` — 请求/响应拦截器
- `internal/ports/http/handler/auth_handler.go` — 登录响应设置 Cookie
- `internal/ports/http/middleware/auth.go` — 认证中间件改为读 Cookie
- `internal/ports/http/router/router.go` — CORS 配置 withCredentials
