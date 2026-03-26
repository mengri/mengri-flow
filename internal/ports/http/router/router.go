package router

import (
	"io/fs"
	"log/slog"
	"mengri-flow/internal/ports/http/handler"
	"mengri-flow/internal/ports/http/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Setup 初始化路由
func Setup(engine *gin.Engine, userHandler *handler.UserHandler, frontendFS fs.FS) {
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
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}

	// ---- 前端静态文件服务 (内嵌 SPA) ----
	if frontendFS != nil {
		setupFrontend(engine, frontendFS)
	}
}

// setupFrontend 将内嵌的前端产物以 / 为根路径提供服务。
// 核心逻辑：
//  1. 请求路径能匹配到静态文件（JS/CSS/图片等） → 直接返回文件
//  2. 请求路径匹配不到文件（如 /users, /about） → 返回 index.html（SPA fallback）
//  3. /api/* 和 /health 已在上方注册，不会走到这里
func setupFrontend(engine *gin.Engine, frontendFS fs.FS) {
	fileServer := http.FileServer(http.FS(frontendFS))

	// 读取 index.html 内容用于 SPA fallback
	indexHTML, err := fs.ReadFile(frontendFS, "index.html")
	if err != nil {
		slog.Warn("frontend index.html not found, SPA fallback disabled", "error", err)
		return
	}

	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳过 API 路径，返回 JSON 404
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"data": nil,
				"msg":  "endpoint not found",
			})
			return
		}

		// 尝试打开静态文件
		// 去掉开头的 / 以匹配 fs.FS 的相对路径
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath != "" {
			if f, err := frontendFS.Open(cleanPath); err == nil {
				f.Close()
				// 文件存在，交给 FileServer 处理（自动设置 Content-Type、缓存等）
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// 文件不存在 → SPA fallback：返回 index.html
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	slog.Info("frontend SPA embedded and served at /")
}
