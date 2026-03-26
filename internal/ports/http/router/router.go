package router

import (
	"log/slog"
	"mengri-flow/internal/ports/http/handler"
	"mengri-flow/internal/ports/http/middleware"
	"mengri-flow/web"

	"github.com/gin-gonic/gin"
)

type Router struct {
	userHandler handler.UserHandler `autowired:""`
}

// Setup 定义所有路由和中间件。Handler 参数均为接口类型，便于测试和依赖注入。
// Setup 初始化路由。所有 Handler 参数均为接口类型，便于测试和依赖注入。
func (r *Router) Setup(engine *gin.Engine) error {
	// 加载内嵌的前端产物
	frontendFS, err := web.DistFS()
	if err != nil {
		slog.Warn("failed to load embedded frontend", "error", err)
		return err
	}
	// 全局中间件
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", r.userHandler.Create)
			users.GET("", r.userHandler.List)
			users.GET("/:id", r.userHandler.GetByID)
			users.PUT("/:id", r.userHandler.Update)
			users.DELETE("/:id", r.userHandler.Delete)
		}

	}

	engine.NoRoute(frontendFS)
	slog.Info("frontend SPA embedded and served at /")
	return nil
}
