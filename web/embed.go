package web

import (
	"embed"
	"io/fs"
)

// 内嵌编译后的前端产物。
// go:embed 要求目录在编译时存在，构建前需先执行前端构建 (make build-web)。
// 开发时如果 web/dist 为空，会内嵌空目录，前端请求返回 404 即可。
//
//go:embed all:dist
var distFS embed.FS

// DistFS 返回去掉 "dist" 前缀后的文件系统，使路径从 / 开始直接匹配。
// 例如 dist/index.html → /index.html，dist/assets/xxx.js → /assets/xxx.js
func DistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
