package router

import (
	"log/slog"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/ports/http/handler"
	"mengri-flow/internal/ports/http/middleware"
	"mengri-flow/web"

	"github.com/gin-gonic/gin"
)

type Router struct {
	userHandler         handler.UserHandler         `autowired:""`
	authHandler         handler.AuthHandler         `autowired:""`
	accountAdminHandler handler.AccountAdminHandler `autowired:""`
	meHandler           handler.MeHandler           `autowired:""`
	jwtManager          *auth.JWTManager            `autowired:""`
}

func (r *Router) Setup(engine *gin.Engine) error {
	frontendFS, err := web.DistFS()
	if err != nil {
		slog.Warn("failed to load embedded frontend", "error", err)
		return err
	}

	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")
	{
		// --- 已有的 users CRUD ---
		users := v1.Group("/users")
		{
			users.POST("", r.userHandler.Create)
			users.GET("", r.userHandler.List)
			users.GET("/:id", r.userHandler.GetByID)
			users.PUT("/:id", r.userHandler.Update)
			users.DELETE("/:id", r.userHandler.Delete)
		}

		// --- Auth 公开接口（无需认证） ---
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/activation/validate", r.authHandler.ValidateActivation)
			authGroup.POST("/activation/confirm", r.authHandler.ConfirmActivation)
			authGroup.POST("/login/password", r.authHandler.LoginByPassword)
			authGroup.POST("/token/refresh", r.authHandler.RefreshToken)

			// logout 需要认证
			authGroup.POST("/logout", middleware.Auth(r.jwtManager), r.authHandler.Logout)
		}

		// --- Me 账号中心（需要认证） ---
		meGroup := v1.Group("/me")
		meGroup.Use(middleware.Auth(r.jwtManager))
		{
			meGroup.GET("/profile", r.meHandler.GetProfile)
			meGroup.GET("/identities", r.meHandler.ListIdentities)
			meGroup.POST("/password/change", r.meHandler.ChangePassword)
			meGroup.POST("/security/verify", r.meHandler.SecurityVerify)
			meGroup.GET("/security/logins", r.meHandler.LoginHistory)
		}

		// --- Admin 管理后台（需要认证 + 管理员角色） ---
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.Auth(r.jwtManager))
		adminGroup.Use(middleware.Admin())
		{
			adminGroup.POST("/accounts", r.accountAdminHandler.Create)
			adminGroup.GET("/accounts", r.accountAdminHandler.List)
			adminGroup.GET("/accounts/:accountId", r.accountAdminHandler.GetDetail)
			adminGroup.PUT("/accounts/:accountId/status", r.accountAdminHandler.ChangeStatus)
			adminGroup.POST("/accounts/:accountId/activation/resend", r.accountAdminHandler.ResendActivation)
			adminGroup.GET("/audit/events", r.accountAdminHandler.ListAuditEvents)
		}
	}

	engine.NoRoute(frontendFS)
	slog.Info("frontend SPA embedded and served at /")
	return nil
}
