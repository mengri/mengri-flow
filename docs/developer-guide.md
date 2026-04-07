# API 编排平台开发者实施指南

## 产品概述

API 编排平台是一个**私有化部署的接口编排内部工具**，旨在帮助企业快速构建、管理和自动化 API 工作流。

### 核心数据模型

平台采用四层数据模型，层层递进：

```
资源(Resource) → 工具(Tool) → 流程(Flow) → 触发器(Trigger)
```

| 层级 | 描述 | 示例 |
|------|------|------|
| **资源 (Resource)** | 外部服务的连接配置 | HTTP 服务、gRPC 服务、数据库、消息队列 |
| **工具 (Tool)** | 具体可调用的接口单元 | HTTP 接口、gRPC 方法、SQL 操作 |
| **流程 (Flow)** | 工具的编排组合 | 多个工具的串联、条件分支、循环 |
| **触发器 (Trigger)** | 流程的启动方式 | RESTful 调用、定时任务、消息队列事件 |

### 多租户工作空间

平台支持多租户架构，以**工作空间 (Workspace)** 为单位进行资源隔离和协作：

- 每个工作空间独立管理自己的资源、工具、流程和触发器
- 工作空间成员可分配不同角色（管理员、开发者、只读成员）
- 跨工作空间调用可通过资源授权机制实现

### 账号体系

平台提供完善的账号和权限管理：

| 认证方式 | 说明 |
|----------|------|
| 本地账号 | 用户名密码注册登录 |
| SSO | 支持 CAS、SAML、OAuth 2.0 |
| LDAP/AD | 企业目录服务集成 |
| 第三方登录 | 企业微信、钉钉、飞书 |
| MFA | 多因素认证（TOTP、短信） |

权限模型采用 **RBAC (Role-Based Access Control)**：

- **平台管理员**：系统配置、用户管理、工作空间审批
- **工作空间负责人**：工作空间设置、成员管理、角色分配
- **开发者**：创建和编辑资源、工具、流程
- **只读成员**：查看和执行流程

---

## 技术选型建议

### 前端技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 框架 | React 18+ with TypeScript | 主流企业级前端框架 |
| 构建工具 | Vite | 快速开发启动和热更新 |
| 样式方案 | Tailwind CSS + CSS Modules | Tailwind 用于全局样式，CSS Modules 用于组件级样式隔离 |
| UI 组件库 | shadcn/ui | 基于 Radix UI 的可定制组件库 |
| 状态管理 | Zustand + TanStack Query | Zustand 管理客户端状态，TanStack Query 管理服务端状态和缓存 |
| 路由 | React Router v6 | 声明式路由 |
| 画布编排 | React Flow (@xyflow/react) | 专业流程编排画布库 |
| 代码编辑器 | Monaco Editor | VS Code 同款编辑器，用于 JSON/SQL 编辑 |
| 表单处理 | React Hook Form + Zod | 高性能表单 + Schema 验证 |
| HTTP 客户端 | Axios + TanStack Query | 请求库 + 数据缓存 |
| 图标 | Lucide React | 简洁一致的图标库 |
| 国际化 | i18next | 预留国际化能力 |
| 主题管理 | next-themes | 深色/浅色主题切换 |

### 后端技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 开发语言 | **Go** (推荐) | 高性能、良好的并发模型，适合微服务和插件系统 |
| 备选语言 | Node.js (TypeScript) | 团队熟悉度较高时可选 |
| API 风格 | RESTful + gRPC | RESTful 用于外部 API，gRPC 用于内部服务通信 |
| 主数据库 | PostgreSQL | 强大的关系型数据库，支持 JSON 类型 |
| 缓存/会话 | Redis | 高速缓存、会话存储、消息队列 |
| 消息队列 | RabbitMQ / Kafka / Redis Streams | 通过插件系统支持多种 MQ |
| 认证 | JWT + Refresh Token | 无状态认证，支持令牌刷新 |
| 外部认证集成 | SAML / OAuth / CAS / LDAP | 支持多种企业认证协议 |
| 插件系统 | Go Plugin 模式 / RPC 插件接口 | 动态加载扩展功能 |

### DevOps 技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 容器化 | Docker + Docker Compose | 私有部署标准化 |
| CI/CD | GitLab CI / GitHub Actions | 自动化构建和部署 |
| 监控 | Prometheus + Grafana | 指标采集和可视化 |
| 日志 | 结构化 JSON 日志 | 支持 ELK 可选方案 |

---

## 项目目录结构

```
api-orchestrator/
├── frontend/                    # 前端项目
│   ├── src/
│   │   ├── app/                 # 页面路由
│   │   │   ├── (auth)/          # 认证相关页面 (login, register, mfa)
│   │   │   │   ├── login/page.tsx
│   │   │   │   └── mfa/page.tsx
│   │   │   ├── (main)/          # 主布局页面
│   │   │   │   ├── layout.tsx           # 顶部导航 + 工作空间切换器
│   │   │   │   ├── page.tsx             # Dashboard
│   │   │   │   ├── resources/
│   │   │   │   │   ├── page.tsx         # 资源列表
│   │   │   │   │   ├── new/page.tsx     # 创建资源
│   │   │   │   │   └── [id]/page.tsx    # 资源详情
│   │   │   │   ├── tools/
│   │   │   │   │   ├── page.tsx         # 工具列表
│   │   │   │   │   ├── new/page.tsx
│   │   │   │   │   ├── import/page.tsx  # 批量导入
│   │   │   │   │   └── [id]/page.tsx    # 工具详情/编辑
│   │   │   │   ├── flows/
│   │   │   │   │   ├── page.tsx         # 流程列表
│   │   │   │   │   ├── new/page.tsx
│   │   │   │   └── [id]/
│   │   │   │       ├── page.tsx         # 编排画布 ⭐
│   │   │   │       └── versions/page.tsx
│   │   │   │   ├── triggers/
│   │   │   │   │   ├── page.tsx
│   │   │   │   │   ├── new/page.tsx
│   │   │   │   │   └── [id]/page.tsx
│   │   │   │   ├── runs/
│   │   │   │   │   ├── page.tsx          # 运行记录列表
│   │   │   │   │   └── [id]/page.tsx     # 运行详情
│   │   │   │   ├── profile/page.tsx      # 个人中心
│   │   │   │   └── admin/                # 管理员页面
│   │   │   │       ├── users/page.tsx
│   │   │   │       └── roles/page.tsx
│   │   ├── components/          # 共享组件
│   │   │   ├── ui/               # shadcn/ui 基础组件
│   │   │   ├── layout/           # 布局组件 (NavBar, WorkspaceSwitcher, UserMenu)
│   │   │   ├── resource/         # 资源相关组件
│   │   │   ├── tool/             # 工具相关组件
│   │   │   ├── flow/             # 流程编排组件
│   │   │   │   ├── canvas/       # 画布组件 (FlowCanvas, FlowNode, FlowEdge)
│   │   │   │   ├── panel/        # 左侧工具面板
│   │   │   │   ├── drawer/       # 右侧配置抽屉
│   │   │   │   └── test/         # 底部测试面板
│   │   │   ├── trigger/          # 触发器相关组件
│   │   │   └── auth/             # 认证相关组件
│   │   ├── hooks/                # 自定义 Hooks
│   │   ├── stores/               # Zustand 状态管理
│   │   ├── lib/                  # 工具函数 + API 客户端
│   │   ├── types/                # TypeScript 类型定义
│   │   └── styles/               # 全局样式 + Design Tokens
│   ├── public/
│   └── package.json
├── backend/                     # 后端项目
│   ├── cmd/                     # 入口
│   ├── internal/
│   │   ├── api/                 # HTTP handlers
│   │   ├── domain/              # 业务逻辑
│   │   │   ├── resource/
│   │   │   ├── tool/
│   │   │   ├── flow/
│   │   │   ├── trigger/
│   │   │   ├── auth/
│   │   │   └── workspace/
│   │   ├── plugin/              # 插件系统
│   │   ├── repo/                # 数据访问层
│   │   └── pkg/                 # 公共包
│   ├── plugins/                 # 内置插件 (http, grpc, mysql, timer, mq)
│   └── go.mod
├── deploy/                      # 部署配置
│   ├── docker-compose.yml
│   └── Dockerfile
└── docs/                        # 文档
```

---

## 核心组件拆分策略

### 编排画布 (Flow Canvas) — 最核心组件

编排画布是平台的核心交互界面，负责可视化地串联各个工具节点。

**技术实现：**

- 使用 **React Flow** (@xyflow/react) 作为画布基础
- 支持节点拖拽、连线、缩放、平移等交互

**自定义节点类型：**

| 节点类型 | 说明 | 特殊属性 |
|----------|------|----------|
| StartNode | 流程起始节点 | 触发参数定义 |
| EndNode | 流程结束节点 | 返回值配置 |
| ToolNode | 工具执行节点 | 工具ID、版本号、参数映射 |
| ConditionNode | 条件分支节点 | 条件表达式 |

**自定义边类型：**

| 边类型 | 说明 |
|--------|------|
| 顺序流 (SequenceEdge) | 默认连线，表示执行顺序 |
| 条件分支 (ConditionalEdge) | 带条件判断的分支连线 |

**节点数据结构：**

```typescript
interface FlowNodeData {
  toolId?: string;           // 关联的工具ID
  toolVersion?: number;      // 锁定的工具版本
  config: {
    inputMapping: Mapping[];   // 输入参数映射
    outputMapping: Mapping[]; // 输出参数映射
    timeout?: number;          // 超时时间(ms)
    retry?: number;            // 重试次数
  };
  status: 'idle' | 'running' | 'success' | 'error';
  hasUpdate?: boolean;       // 是否有新版本
}
```

**版本升级提示：**
当关联的工具发布新版本时，节点右上角显示升级图标，点击可查看变更内容并选择是否升级。

### 配置抽屉 (Config Drawer)

右侧滑出的配置面板，用于编辑选中节点的详细配置。

**组件结构：**

```
ConfigDrawer
├── Tabs
│   ├── 参数映射 (Input Mapping)
│   ├── 条件配置 (Condition)
│   └── 返回值 (Output)
├── MappingTable
│   ├── Source (来源)
│   │   ├── 节点选择
│   │   ├── 输出字段
│   │   └── 转换函数
│   └── Target (目标)
│       ├── 参数名称
│       └── 必填/可选
└── Actions
    ├── 保存
    └── 取消
```

**变量引用语法：**
支持在参数中使用 `{{node_id.output.field}}` 语法引用前序节点的输出。

### 工具面板 (Tool Panel)

左侧面板，展示当前工作空间可用的工具列表。

**功能特性：**

- 搜索框：按名称搜索工具
- 筛选器：按资源类型筛选
- 拖拽支持：拖拽工具到画布创建节点
- 状态指示：显示工具版本、发布状态
- 升级提示：工具新版本角标

**数据结构：**

```typescript
interface ToolListItem {
  id: string;
  name: string;
  resourceId: string;
  resourceName: string;
  resourceType: 'http' | 'grpc' | 'mysql' | ...;
  version: number;
  status: 'draft' | 'published' | 'deprecated';
  description?: string;
}
```

---

## API 设计规范

### RESTful API 约定

| 规范 | 说明 |
|------|------|
| 版本前缀 | `/api/v1/` |
| 响应格式 | `{ code: number, data: T, message: string }` |
| 认证方式 | `Authorization: Bearer {token}` |
| 分页参数 | `?page=1&page_size=20` |
| 错误码 | 2xx 成功，4xx 客户端错误，5xx 服务端错误 |

**响应格式示例：**

```json
// 成功响应
{
  "code": 0,
  "data": {
    "id": "123",
    "name": "示例工具"
  },
  "message": "success"
}

// 分页响应
{
  "code": 0,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  },
  "message": "success"
}

// 错误响应
{
  "code": 40001,
  "data": null,
  "message": "资源不存在"
}
```

### 核心 API 端点

#### 资源管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/resources` | 获取资源列表 |
| POST | `/api/v1/resources` | 创建资源 |
| GET | `/api/v1/resources/:id` | 获取资源详情 |
| PUT | `/api/v1/resources/:id` | 更新资源 |
| DELETE | `/api/v1/resources/:id` | 删除资源 |
| POST | `/api/v1/resources/:id/test` | 测试资源连接 |

#### 工具管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/tools` | 获取工具列表 |
| POST | `/api/v1/tools` | 创建工具 |
| GET | `/api/v1/tools/:id` | 获取工具详情 |
| PUT | `/api/v1/tools/:id` | 更新工具 |
| DELETE | `/api/v1/tools/:id` | 删除工具 |
| POST | `/api/v1/tools/:id/publish` | 发布工具 |
| POST | `/api/v1/tools/:id/unpublish` | 取消发布 |
| POST | `/api/v1/tools/import` | 批量导入工具 |
| GET | `/api/v1/tools/:id/versions` | 获取版本历史 |
| POST | `/api/v1/tools/:id/test` | 测试工具调用 |
| GET | `/api/v1/tools/:id/impact-analysis` | 获取影响分析 |

#### 流程管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/flows` | 获取流程列表 |
| POST | `/api/v1/flows` | 创建流程 |
| GET | `/api/v1/flows/:id` | 获取流程详情 |
| PUT | `/api/v1/flows/:id` | 更新流程 |
| DELETE | `/api/v1/flows/:id` | 删除流程 |
| POST | `/api/v1/flows/:id/publish` | 发布流程 |
| POST | `/api/v1/flows/:id/run` | 执行流程 |
| GET | `/api/v1/flows/:id/versions` | 获取版本历史 |

#### 触发器管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/triggers` | 获取触发器列表 |
| POST | `/api/v1/triggers` | 创建触发器 |
| GET | `/api/v1/triggers/:id` | 获取触发器详情 |
| PUT | `/api/v1/triggers/:id` | 更新触发器 |
| DELETE | `/api/v1/triggers/:id` | 删除触发器 |
| PUT | `/api/v1/triggers/:id/status` | 启用/禁用触发器 |

#### 运行记录

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/runs` | 获取运行记录列表 |
| GET | `/api/v1/runs/:id` | 获取运行详情 |

#### 认证

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| POST | `/api/v1/auth/refresh` | 刷新令牌 |
| GET | `/api/v1/auth/profile` | 获取当前用户信息 |

---

## 数据库核心表设计

### users - 用户表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| username | VARCHAR(50) | 用户名，唯一 |
| email | VARCHAR(255) | 邮箱，唯一 |
| password_hash | VARCHAR(255) | 密码哈希 |
| status | VARCHAR(20) | 状态: active, disabled, deleted |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### workspaces - 工作空间表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 工作空间名称 |
| description | TEXT | 描述 |
| created_by | UUID | 创建者ID |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### workspace_members - 工作空间成员表

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | UUID | 用户ID |
| workspace_id | UUID | 工作空间ID |
| role | VARCHAR(20) | 角色: owner, admin, developer, viewer |
| joined_at | TIMESTAMP | 加入时间 |

### resources - 资源表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 资源名称 |
| type | VARCHAR(50) | 资源类型: http, grpc, mysql, postgres, rabbitmq, kafka |
| config_json | JSONB | 连接配置 |
| workspace_id | UUID | 所属工作空间 |
| status | VARCHAR(20) | 状态: active, inactive |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### tools - 工具表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 工具名称 |
| resource_id | UUID | 关联资源ID |
| config_json | JSONB | 工具配置 |
| version | INTEGER | 当前版本号 |
| status | VARCHAR(20) | 状态: draft, published, deprecated |
| workspace_id | UUID | 所属工作空间 |
| published_at | TIMESTAMP | 发布时间 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### tool_versions - 工具版本表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| tool_id | UUID | 工具ID |
| version | INTEGER | 版本号 |
| config_snapshot | JSONB | 配置快照 |
| published_at | TIMESTAMP | 发布时间 |
| changelog | TEXT | 版本变更说明 |

### flows - 流程表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 流程名称 |
| config_json | JSONB | 流程配置（画布数据） |
| version | INTEGER | 当前版本号 |
| status | VARCHAR(20) | 状态: draft, published |
| workspace_id | UUID | 所属工作空间 |
| published_at | TIMESTAMP | 发布时间 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### flow_versions - 流程版本表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| flow_id | UUID | 流程ID |
| version | INTEGER | 版本号 |
| config_snapshot | JSONB | 配置快照 |
| published_at | TIMESTAMP | 发布时间 |

### flow_tool_refs - 流程工具引用表

| 字段 | 类型 | 说明 |
|------|------|------|
| flow_id | UUID | 流程ID |
| tool_id | UUID | 工具ID |
| tool_version | INTEGER | 锁定的工具版本 |
| node_id | VARCHAR(50) | 画布节点ID |

### triggers - 触发器表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 触发器名称 |
| type | VARCHAR(50) | 类型: restful, timer, mq |
| config_json | JSONB | 触发器配置 |
| flow_id | UUID | 关联流程ID |
| flow_version | INTEGER | 锁定的流程版本 |
| status | VARCHAR(20) | 状态: active, paused |
| workspace_id | UUID | 所属工作空间 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### runs - 运行记录表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| trigger_id | UUID | 触发器ID |
| flow_id | UUID | 流程ID |
| flow_version | INTEGER | 执行的流程版本 |
| status | VARCHAR(20) | 状态: pending, running, success, failed |
| input_json | JSONB | 输入数据 |
| output_json | JSONB | 输出数据 |
| started_at | TIMESTAMP | 开始时间 |
| finished_at | TIMESTAMP | 结束时间 |

### run_node_logs - 节点执行日志表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| run_id | UUID | 运行记录ID |
| node_id | VARCHAR(50) | 节点ID |
| tool_id | UUID | 工具ID |
| status | VARCHAR(20) | 状态 |
| input_json | JSONB | 输入数据 |
| output_json | JSONB | 输出数据 |
| duration_ms | INTEGER | 执行时长(毫秒) |
| created_at | TIMESTAMP | 创建时间 |

---

## 插件系统设计

插件系统是平台扩展能力的核心，支持动态添加新的资源类型和触发器类型。

### 插件接口定义

```go
// ResourceType 资源插件接口
type ResourceType interface {
    // PluginName 插件名称
    PluginName() string
    
    // ConfigSchema 返回连接配置的 JSON Schema
    ConfigSchema() map[string]interface{}
    
    // TestConnection 测试连接
    TestConnection(config map[string]interface{}) error
    
    // ExtractTools 从资源中提取工具列表
    ExtractTools(config map[string]interface{}) ([]ToolDefinition, error)
}

// TriggerType 触发器插件接口
type TriggerType interface {
    // PluginName 插件名称
    PluginName() string
    
    // ConfigSchema 返回触发器配置的 JSON Schema
    ConfigSchema() map[string]interface{}
    
    // InputSchema 输入数据 schema
    InputSchema() map[string]interface{}
    
    // OutputSchema 输出数据 schema
    OutputSchema() map[string]interface{}
    
    // Start 启动触发器
    Start(config map[string]interface{}, handler TriggerHandler) error
    
    // Stop 停止触发器
    Stop() error
}
```

### 资源插件

资源插件负责与外部服务建立连接并提取可调用工具。

**配置 Schema 示例：**

```json
{
  "type": "object",
  "required": ["baseUrl"],
  "properties": {
    "baseUrl": {
      "type": "string",
      "title": "服务地址"
    },
    "auth": {
      "type": "object",
      "title": "认证配置",
      "properties": {
        "type": {"type": "string", "enum": ["none", "basic", "bearer", "apiKey"]},
        "username": {"type": "string"},
        "password": {"type": "string"},
        "token": {"type": "string"},
        "apiKey": {"type": "string"}
      }
    },
    "timeout": {
      "type": "integer",
      "default": 30000,
      "title": "超时时间(ms)"
    }
  }
}
```

### 触发器插件

触发器插件负责监听外部事件并触发流程执行。

**支持的触发类型：**

| 类型 | 说明 | 配置示例 |
|------|------|----------|
| restful | RESTful API 调用 | 监听路径、HTTP 方法 |
| timer | 定时任务 | Cron 表达式 |
| mq | 消息队列 | 队列名称、交换机 |

### 内置插件

| 资源插件 | 说明 |
|----------|------|
| http | HTTP/HTTPS 服务 |
| grpc | gRPC 服务 |
| mysql | MySQL 数据库 |
| postgres | PostgreSQL 数据库 |

| 触发器插件 | 说明 |
|------------|------|
| timer | 定时任务触发 |
| rabbitmq | RabbitMQ 消息触发 |
| kafka | Kafka 消息触发 |

### 插件加载机制

1. 启动时扫描 `plugins/` 目录
2. 加载插件动态库（Go Plugin）或配置文件
3. 注册到插件管理器
4. 前端动态渲染配置表单

---

## 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Nginx / 反向代理                        │
│                    (SSL 终止、负载均衡)                       │
└─────────────────────────────────────────────────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
   ┌─────┴─────┐        ┌─────┴─────┐        ┌─────┴─────┐
   │  Frontend │        │  Backend  │        │  Plugin   │
   │  (React)  │        │   (Go)    │        │  Services │
   │   :3000   │◄──────►│   :8080   │◄──────►│  (独立进程)│
   └───────────┘        └─────┬─────┘        └───────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
        ┌─────┴─────┐    ┌─────┴─────┐    ┌─────┴─────┐
        │  Postgres │    │   Redis   │    │    MQ     │
        │   :5432   │    │   :6379   │    │ 5672/9092 │
        └───────────┘    └───────────┘    └───────────┘
```

### 组件说明

| 组件 | 说明 | 端口 |
|------|------|------|
| Nginx | 反向代理、SSL 终止、静态资源服务 | 80/443 |
| Frontend | React 单页应用 | 3000 |
| Backend | Go RESTful API 服务 | 8080 |
| Plugin Services | 插件服务（可选独立部署） | 动态分配 |
| PostgreSQL | 主数据库 | 5432 |
| Redis | 缓存、会话、消息队列 | 6379 |
| MQ | 消息队列 (RabbitMQ/Kafka) | 5672/9092 |

### Docker Compose 部署

```yaml
version: '3.8'
services:
  nginx:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - frontend
      - backend

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/db
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: api_orchestrator
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine

volumes:
  pgdata:
```

---

## 设计系统落地指南

### CSS Design Tokens

设计 Tokens 已定义在 `design-system/styles/design-tokens.css` 中，包括：

- **色板**：主色、辅助色、功能色、语义色
- **字体**：字号、字重、字体族
- **间距**：间距 scale
- **圆角**：圆角 scale
- **阴影**：阴影层级

### Tailwind CSS 配置

将 Design Tokens 映射到 `tailwind.config.ts`：

```typescript
import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: 'var(--color-primary-50)',
          // ... 其他层级
          500: 'var(--color-primary-500)',
          600: 'var(--color-primary-600)',
          // ...
        },
      },
      fontSize: {
        xs: 'var(--font-size-xs)',
        sm: 'var(--font-size-sm)',
        base: 'var(--font-size-base)',
        // ...
      },
      spacing: {
        '0': 'var(--spacing-0)',
        '1': 'var(--spacing-1)',
        // ...
      },
      borderRadius: {
        none: 'var(--radius-none)',
        sm: 'var(--radius-sm)',
        // ...
      },
    },
  },
  plugins: [],
} satisfies Config
```

### shadcn/ui 主题配置

修改 `components.json` 中的主题色配置：

```json
{
  "style": "default",
  "tailwind": {
    "baseColor": "slate",
    "cssVariables": true
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils"
  }
}
```

然后使用 Design Tokens 覆盖 CSS 变量：

```css
:root {
  --primary: var(--color-primary-500);
  --primary-foreground: var(--color-primary-50);
  --background: var(--color-bg-primary);
  --foreground: var(--color-text-primary);
  /* ... 其他变量 */
}
```

### 组件展示页参考

- **Design Tokens 展示页**: `design-system/tokens.html`
- **组件规范展示页**: `design-system/components.html`

---

## 总结

本文档提供了 API 编排平台的完整技术实施指南，涵盖：

1. **产品定义**：核心数据模型、多租户架构、账号体系
2. **技术选型**：前端 React 技术栈、后端 Go 语言、DevOps 工具
3. **项目结构**：前后端目录组织、核心组件拆分
4. **API 设计**：RESTful 规范、完整端点列表
5. **数据模型**：核心数据库表设计
6. **插件系统**：接口定义、加载机制
7. **部署架构**：容器化部署方案
8. **设计系统**：Tokens 落地指南

开发团队可基于本指南快速启动项目开发，实现一个功能完善、性能优异的私有化 API 编排平台。
