package web

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/** 内嵌编译后的前端产物。
/** go:embed 要求目录在编译时存在，构建前需先执行前端构建 (make build-web)。
/**  开发时如果 web/dist 为空，会内嵌空目录，前端请求返回 404 即可。
*/
//go:embed all:dist
var distFS embed.FS

// DistFS 返回一个 Gin HandlerFunc，用于提供内嵌的前端产物服务。
func DistFS() (gin.HandlerFunc, error) {
	//返回去掉 "dist" 前缀后的文件系统，使路径从 / 开始直接匹配。
	// 例如 dist/index.html → /index.html，dist/assets/xxx.js → /assets/xxx.js
	frontendFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(frontendFS))

	// 读取 index.html 内容用于 SPA fallback
	indexHTML, err := fs.ReadFile(frontendFS, "index.html")
	if err != nil {
		slog.Warn("frontend index.html not found, SPA fallback disabled", "error", err)
		return nil, err
	}
	return func(c *gin.Context) {
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
				return // 直接返回文件响应，不继续执行后续逻辑
			}
		}

		// 文件不存在 → SPA fallback：返回 index.html
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}, nil
}
