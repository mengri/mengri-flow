package router

import (
	"log/slog"
	"mengri-flow/internal/app/service"
	"mengri-flow/internal/infra/auth"
	"mengri-flow/internal/ports/http/handler"
	"mengri-flow/internal/ports/http/middleware"
	"mengri-flow/web"

	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler         handler.AuthHandler         `autowired:""`
	accountAdminHandler handler.AccountAdminHandler `autowired:""`
	meHandler           handler.MeHandler           `autowired:""`
	workspaceHandler    *handler.WorkspaceHandler   `autowired:""`
	environmentHandler  *handler.EnvironmentHandler `autowired:""`
	clusterHandler      *handler.ClusterHandler     `autowired:""`
	resourceHandler     *handler.ResourceHandler    `autowired:""`
	toolHandler         *handler.ToolHandler        `autowired:""`
	flowHandler         *handler.FlowHandler        `autowired:""`
	triggerHandler      *handler.TriggerHandler     `autowired:""`
	runHandler          *handler.RunHandler         `autowired:""`
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
// --- Auth 公开接口（无需认证） ---
	authGroup := v1.Group("/auth")
	{
		authGroup.GET("/activation/validate", r.authHandler.ValidateActivation)
		authGroup.POST("/activation/confirm", r.authHandler.ConfirmActivation)
		authGroup.POST("/login/password", r.authHandler.LoginByPassword)
		authGroup.POST("/login/sms/send", r.authHandler.SendSMSCode)
		authGroup.POST("/login/sms/verify", r.authHandler.LoginBySMS)
		authGroup.GET("/oauth/:provider/url", r.authHandler.GetOAuthURL)
		authGroup.GET("/oauth/:provider/callback", r.authHandler.OAuthCallback)
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
		meGroup.POST("/identities/phone/bind", r.meHandler.BindPhone)
		meGroup.POST("/identities/:provider/bind", r.meHandler.BindProvider)
		meGroup.DELETE("/identities/:identityId", r.meHandler.UnbindIdentity)
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

		// --- Workspace 工作空间管理（需要认证） ---
		workspaceGroup := v1.Group("/workspaces")
		workspaceGroup.Use(middleware.Auth(r.jwtManager))
		{
			workspaceGroup.GET("", r.workspaceHandler.ListWorkspaces)
			workspaceGroup.POST("", r.workspaceHandler.CreateWorkspace)
			workspaceGroup.GET("/:id", r.workspaceHandler.GetWorkspace)
			workspaceGroup.PUT("/:id", r.workspaceHandler.UpdateWorkspace)
			workspaceGroup.DELETE("/:id", r.workspaceHandler.DeleteWorkspace)
			workspaceGroup.POST("/:id/members", r.workspaceHandler.AddMember)
			workspaceGroup.DELETE("/:id/members/:userId", r.workspaceHandler.RemoveMember)
		}

		// --- Environment 环境管理（需要认证） ---
		environmentGroup := v1.Group("/environments")
		environmentGroup.Use(middleware.Auth(r.jwtManager))
		{
			environmentGroup.GET("", r.environmentHandler.ListEnvironments)
			environmentGroup.POST("", r.environmentHandler.CreateEnvironment)
			environmentGroup.GET("/:id", r.environmentHandler.GetEnvironment)
			environmentGroup.PUT("/:id", r.environmentHandler.UpdateEnvironment)
			environmentGroup.DELETE("/:id", r.environmentHandler.DeleteEnvironment)
		}

		// --- Cluster 集群管理（需要认证） ---
		clusterGroup := v1.Group("/clusters")
		clusterGroup.Use(middleware.Auth(r.jwtManager))
		{
			clusterGroup.GET("", r.clusterHandler.ListClusters)
			clusterGroup.POST("", r.clusterHandler.CreateCluster)
			clusterGroup.GET("/:id", r.clusterHandler.GetClusterDetail)
			clusterGroup.PUT("/:id", r.clusterHandler.UpdateCluster)
			clusterGroup.DELETE("/:id", r.clusterHandler.DeleteCluster)
			clusterGroup.POST("/:id/test-etcd", r.clusterHandler.TestEtcdConnection)
		}

		// --- Resource 资源管理（需要认证） ---
		resourceGroup := v1.Group("/resources")
		resourceGroup.Use(middleware.Auth(r.jwtManager))
		{
			resourceGroup.GET("", r.resourceHandler.ListResources)
			resourceGroup.POST("", r.resourceHandler.CreateResource)
			resourceGroup.GET("/:id", r.resourceHandler.GetResource)
			resourceGroup.PUT("/:id", r.resourceHandler.UpdateResource)
			resourceGroup.DELETE("/:id", r.resourceHandler.DeleteResource)
			resourceGroup.POST("/test-connection", r.resourceHandler.TestConnection)
		}

		// --- Tool 工具管理（需要认证） ---
		toolGroup := v1.Group("/tools")
		toolGroup.Use(middleware.Auth(r.jwtManager))
		{
			toolGroup.GET("", r.toolHandler.ListTools)
			toolGroup.POST("", r.toolHandler.CreateTool)
			toolGroup.GET("/:id", r.toolHandler.GetTool)
			toolGroup.PUT("/:id", r.toolHandler.UpdateTool)
			toolGroup.POST("/test", r.toolHandler.TestTool)
			toolGroup.POST("/import", r.toolHandler.ImportTools)
			toolGroup.POST("/:id/publish", r.toolHandler.PublishTool)
			toolGroup.POST("/:id/deprecate", r.toolHandler.DeprecateTool)
			toolGroup.GET("/:id/versions", r.toolHandler.ListVersions)
		}

		// --- Flow 流程管理（需要认证） ---
		flowGroup := v1.Group("/flows")
		flowGroup.Use(middleware.Auth(r.jwtManager))
		{
			flowGroup.GET("", r.flowHandler.ListFlows)
			flowGroup.POST("", r.flowHandler.CreateFlow)
			flowGroup.GET("/:id", r.flowHandler.GetFlow)
			flowGroup.PUT("/:id", r.flowHandler.UpdateFlow)
			flowGroup.DELETE("/:id", r.flowHandler.DeleteFlow)
			flowGroup.POST("/test", r.flowHandler.TestFlow)
			flowGroup.POST("/:id/publish", r.flowHandler.PublishFlow)
			flowGroup.GET("/:id/versions", r.flowHandler.ListVersions)
			flowGroup.POST("/:id/rollback", r.flowHandler.RollbackVersion)
		}

		// --- Trigger 触发器管理（需要认证） ---
		triggerGroup := v1.Group("/triggers")
		triggerGroup.Use(middleware.Auth(r.jwtManager))
		{
			triggerGroup.GET("", r.triggerHandler.ListTriggers)
			triggerGroup.POST("", r.triggerHandler.CreateTrigger)
			triggerGroup.GET("/:id", r.triggerHandler.GetTrigger)
			triggerGroup.PUT("/:id", r.triggerHandler.UpdateTrigger)
			triggerGroup.DELETE("/:id", r.triggerHandler.DeleteTrigger)
			triggerGroup.POST("/:id/enable", r.triggerHandler.EnableTrigger)
			triggerGroup.POST("/:id/disable", r.triggerHandler.DisableTrigger)
			triggerGroup.POST("/:id/publish", r.triggerHandler.PublishToCluster)
		}

		// --- Run 运行记录（需要认证） ---
		runGroup := v1.Group("/runs")
		runGroup.Use(middleware.Auth(r.jwtManager))
		{
			runGroup.GET("", r.runHandler.ListRuns)
			runGroup.GET("/:id", r.runHandler.GetRunDetail)
			runGroup.GET("/:id/timeline", r.runHandler.GetExecutionTimeline)
			runGroup.POST("/:id/retry", r.runHandler.RetryRun)
			runGroup.GET("/stats", r.runHandler.GetRunStats)
		}
	}

	engine.NoRoute(frontendFS)
	slog.Info("frontend SPA embedded and served at /")
	return nil
}
