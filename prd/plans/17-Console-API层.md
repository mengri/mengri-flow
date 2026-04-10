# 任务 17: Console API层

## 任务概述
整合所有HTTP处理器，注册路由，实现统一的请求处理、响应格式、错误处理和Swagger文档。

## 上下文依赖
- 任务 10-16: 所有业务模块
- 任务 01: 基础框架（中间件、响应格式）

## 涉及文件
- `internal/ports/http/router/router.go` - 路由注册
- `internal/ports/http/handler/` - 所有handler整合
- `docs/swagger/` - Swagger文档
- `cmd/server/main.go` - 服务入口

## 详细步骤

### 17.1 路由注册整合
**文件：`internal/ports/http/router/router.go`**
```go
package router

import (
    "github.com/gin-gonic/gin"
    "backend/internal/app/service"
    "backend/internal/infra/auth"
    "backend/internal/infra/plugin"
    "backend/internal/ports/http/handler"
    "backend/internal/ports/http/middleware"
)

func SetupRouter(
    authService *service.AuthService,
    workspaceService *service.WorkspaceService,
    envService *service.EnvironmentService,
    clusterService *service.ClusterService,
    resourceService *service.ResourceService,
    toolService *service.ToolService,
    flowService *service.FlowService,
    triggerService *service.TriggerService,
    runService *service.RunService,
    jwtManager *auth.JWTManager,
    pluginRegistry *plugin.Registry,
) *gin.Engine {
    router := gin.New()
    
    // 全局中间件
    router.Use(middleware.Logger())
    router.Use(middleware.Recovery())
    router.Use(middleware.CORS())
    
    // 认证中间件
    authMiddleware := middleware.NewAuthMiddleware(jwtManager)
    rbacMiddleware := middleware.NewRBACMiddleware()
    
    // API v1
    v1 := router.Group("/api/v1")
    {
        // 公开路由
        public := v1.Group("")
        {
            authHandler := handler.NewAuthHandler(authService, jwtManager)
            public.POST("/auth/login", authHandler.Login)
            public.POST("/auth/refresh", authHandler.RefreshToken)
        }
        
        // 需要认证的路由
        authenticated := v1.Group("")
        authenticated.Use(authMiddleware.ValidateToken())
        {
            // 个人中心
            profileHandler := handler.NewProfileHandler()
            authenticated.GET("/profile", profileHandler.GetProfile)
            authenticated.PUT("/profile", profileHandler.UpdateProfile)
            
            // 工作空间
            workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
            authenticated.GET("/workspaces", workspaceHandler.ListWorkspaces)
            authenticated.POST("/workspaces", workspaceHandler.CreateWorkspace)
            authenticated.GET("/workspaces/:id", workspaceHandler.GetWorkspace)
            authenticated.PUT("/workspaces/:id", workspaceHandler.UpdateWorkspace)
            authenticated.DELETE("/workspaces/:id", workspaceHandler.DeleteWorkspace)
            authenticated.POST("/workspaces/:id/members", workspaceHandler.AddMember)
            authenticated.DELETE("/workspaces/:id/members/:userId", workspaceHandler.RemoveMember)
            
            // 环境管理
            envHandler := handler.NewEnvironmentHandler(envService)
            authenticated.GET("/environments", rbacMiddleware.RequirePermission("environments.read"), envHandler.ListEnvironments)
            authenticated.POST("/environments", rbacMiddleware.RequirePermission("environments.create"), envHandler.CreateEnvironment)
            authenticated.GET("/environments/:id", rbacMiddleware.RequirePermission("environments.read"), envHandler.GetEnvironment)
            authenticated.PUT("/environments/:id", rbacMiddleware.RequirePermission("environments.update"), envHandler.UpdateEnvironment)
            authenticated.DELETE("/environments/:id", rbacMiddleware.RequirePermission("environments.delete"), envHandler.DeleteEnvironment)
            
            // 集群管理
            clusterHandler := handler.NewClusterHandler(clusterService)
            authenticated.GET("/clusters", clusterHandler.ListClusters)
            authenticated.POST("/clusters", clusterHandler.CreateCluster)
            authenticated.GET("/clusters/:id", clusterHandler.GetClusterDetail)
            authenticated.PUT("/clusters/:id", clusterHandler.UpdateCluster)
            authenticated.DELETE("/clusters/:id", clusterHandler.DeleteCluster)
            authenticated.POST("/clusters/:id/test-etcd", clusterHandler.TestEtcdConnection)
            
            // 资源管理
            resourceHandler := handler.NewResourceHandler(resourceService)
            authenticated.GET("/resources", resourceHandler.ListResources)
            authenticated.POST("/resources", resourceHandler.CreateResource)
            authenticated.GET("/resources/:id", resourceHandler.GetResource)
            authenticated.PUT("/resources/:id", resourceHandler.UpdateResource)
            authenticated.DELETE("/resources/:id", resourceHandler.DeleteResource)
            authenticated.POST("/resources/test-connection", resourceHandler.TestConnection)
            authenticated.POST("/resources/:id/extract-tools", resourceHandler.ExtractTools)
            
            // 工具管理
            toolHandler := handler.NewToolHandler(toolService)
            authenticated.GET("/tools", toolHandler.ListTools)
            authenticated.POST("/tools", toolHandler.CreateTool)
            authenticated.GET("/tools/:id", toolHandler.GetTool)
            authenticated.PUT("/tools/:id", toolHandler.UpdateTool)
            authenticated.POST("/tools/test", toolHandler.TestTool)
            authenticated.POST("/tools/import", toolHandler.ImportTools)
            authenticated.POST("/tools/:id/publish", toolHandler.PublishTool)
            authenticated.POST("/tools/:id/deprecate", toolHandler.DeprecateTool)
            authenticated.GET("/tools/:id/versions", toolHandler.ListVersions)
            
            // 流程管理
            flowHandler := handler.NewFlowHandler(flowService)
            authenticated.GET("/flows", flowHandler.ListFlows)
            authenticated.POST("/flows", flowHandler.CreateFlow)
            authenticated.GET("/flows/:id", flowHandler.GetFlow)
            authenticated.PUT("/flows/:id", flowHandler.UpdateFlow)
            authenticated.DELETE("/flows/:id", flowHandler.DeleteFlow)
            authenticated.POST("/flows/test", flowHandler.TestFlow)
            authenticated.POST("/flows/:id/publish", flowHandler.PublishFlow)
            authenticated.GET("/flows/:id/versions", flowHandler.ListVersions)
            authenticated.POST("/flows/:id/rollback", flowHandler.RollbackVersion)
            
            // 触发器管理
            triggerHandler := handler.NewTriggerHandler(triggerService)
            authenticated.GET("/triggers", triggerHandler.ListTriggers)
            authenticated.POST("/triggers", triggerHandler.CreateTrigger)
            authenticated.GET("/triggers/:id", triggerHandler.GetTrigger)
            authenticated.PUT("/triggers/:id", triggerHandler.UpdateTrigger)
            authenticated.DELETE("/triggers/:id", triggerHandler.DeleteTrigger)
            authenticated.POST("/triggers/:id/enable", triggerHandler.EnableTrigger)
            authenticated.POST("/triggers/:id/disable", triggerHandler.DisableTrigger)
            authenticated.POST("/triggers/:id/publish", triggerHandler.PublishToCluster)
            
            // 运行记录
            runHandler := handler.NewRunHandler(runService)
            authenticated.GET("/runs", runHandler.ListRuns)
            authenticated.GET("/runs/:id", runHandler.GetRunDetail)
            authenticated.GET("/runs/:id/timeline", runHandler.GetExecutionTimeline)
            authenticated.POST("/runs/:id/retry", runHandler.RetryRun)
            authenticated.GET("/runs/stats", runHandler.GetRunStats)
        }
    }
    
    // 健康检查
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Swagger文档
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    return router
}
```

### 17.2 Swagger文档生成
**文件：`docs/swagger/docs.go`**
```go
// Package docs API文档
//
// @title API编排平台 API
// @version 1.0
// @description API编排平台的RESTful API文档
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
package docs
```
- [x] 为所有API添加Swagger注解
- [x] 生成swagger.json和swagger.yaml
- [x] 在 `/swagger` 路径提供文档界面

### 17.3 中间件实现
**文件：`internal/ports/http/middleware/logger.go`**
```go
func Logger() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

**文件：`internal/ports/http/middleware/cors.go`**
```go
func CORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

### 17.4 错误处理统一
**文件：`internal/ports/http/handler/error_handler.go`**
```go
func HandleError(c *gin.Context, err error) {
    switch e := err.(type) {
    case *domainerrors.ValidationError:
        c.JSON(400, gin.H{
            "code": 400,
            "msg":  e.Message,
            "data": nil,
        })
    case *domainerrors.NotFoundError:
        c.JSON(404, gin.H{
            "code": 404,
            "msg":  e.Message,
            "data": nil,
        })
    case *domainerrors.ForbiddenError:
        c.JSON(403, gin.H{
            "code": 403,
            "msg":  e.Message,
            "data": nil,
        })
    default:
        c.JSON(500, gin.H{
            "code": 500,
            "msg":  "internal server error",
            "data": nil,
        })
    }
}
```

### 17.5 服务入口
**文件：`cmd/server/main.go`**
```go
func main() {
    // 1. 加载配置
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // 2. 初始化数据库
    db := initDatabase(cfg.Database)
    
    // 3. 初始化etcd客户端
    etcdClient := initEtcd(cfg.Etcd)
    
    // 4. 初始化插件注册表
    pluginRegistry := plugin.GlobalRegistry()
    
    // 5. 初始化仓储层
    resourceRepo := mysql.NewResourceRepository(db)
    toolRepo := mysql.NewToolRepository(db)
    flowRepo := mysql.NewFlowRepository(db)
    triggerRepo := mysql.NewTriggerRepository(db)
    runRepo := mysql.NewRunRepository(db)
    clusterRepo := mysql.NewClusterRepository(db)
    userRepo := mysql.NewUserRepository(db)
    workspaceRepo := mysql.NewWorkspaceRepository(db)
    
    // 6. 初始化服务层
    resourceService := service.NewResourceService(resourceRepo, pluginRegistry)
    toolService := service.NewToolService(toolRepo, resourceRepo, pluginRegistry)
    flowService := service.NewFlowService(flowRepo, toolRepo, resourceRepo, pluginRegistry, runRepo, nodeLogRepo)
    triggerService := service.NewTriggerService(triggerRepo, flowRepo, clusterRepo, pluginRegistry, etcdClient)
    runService := service.NewRunService(runRepo, nodeLogRepo, triggerRepo, flowRepo, toolRepo)
    clusterService := service.NewClusterService(clusterRepo, etcdClient)
    envService := service.NewEnvironmentService(envRepo, clusterRepo)
    authService := service.NewAuthService(userRepo, jwtManager, passwordHasher)
    workspaceService := service.NewWorkspaceService(workspaceRepo, userRepo, memberRepo, roleRepo)
    
    // 7. 初始化HTTP服务器
    router := router.SetupRouter(
        authService,
        workspaceService,
        envService,
        clusterService,
        resourceService,
        toolService,
        flowService,
        triggerService,
        runService,
        jwtManager,
        pluginRegistry,
    )
    
    // 8. 启动服务器
    addr := fmt.Sprintf(":%d", cfg.Server.Port)
    log.Printf("Server starting on %s", addr)
    
    if err := router.Run(addr); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### 17.6 配置示例
**文件：`config.yaml.example`**
```yaml
server:
  port: 8080
  mode: debug  # debug|release

database:
  host: localhost
  port: 5432
  database: mengri_flow
  username: postgres
  password: postgres
  sslmode: disable

etcd:
  endpoints:
    - localhost:2379
  username: ""
  password: ""

jwt:
  secret: your-secret-key-here
  token_ttl: 24h
  refresh_token_ttl: 168h

logging:
  level: info  # debug|info|warn|error
  format: json  # json|text
```

### 17.7 Makefile构建配置
```makefile
.PHONY: build-server
build-server:
	@echo "Building server..."
	GOTOOLCHAIN=local go build -tags "$(BUILD_TAGS)" -o bin/server ./cmd/server/main.go

.PHONY: run
run:
	GOTOOLCHAIN=local go run -tags "$(BUILD_TAGS)" ./cmd/server/main.go --config=config.yaml

.PHONY: docker-build
docker-build:
	docker build -t mengri-flow:$(VERSION) \
		--build-arg BUILD_TAGS="$(BUILD_TAGS)" \
		-f Dockerfile .
```

## 验收标准
- [x] 所有路由正确注册
- [x] 中间件（日志、恢复、CORS、认证）正常工作
- [x] Swagger文档可访问且完整
- [x] 配置加载正确
- [x] 服务可正常启动
- [x] 错误处理统一
- [x] 构建脚本正确

## 参考文档
- `AGENTS.md` - Build & Run
- `docs/architecture-design.md` - Console API层

## 预估工时
3-4 天
