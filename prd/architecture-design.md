# API 编排平台 — 架构设计文档

> **版本**：v1.4  
> **日期**：2026-04-07  
> **状态**：Draft  
> **作者**：软件架构师  

---

## 文档修订记录

| 版本 | 日期 | 修订内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-04-01 | 初稿，基于 PRD v1.0 和开发者实施指南编写 | 软件架构师 |
| v1.1 | 2026-04-01 | 插件架构从运行时改为编译时集成（Build-time Plugin），通过 Go build tags 按客户需求选择性编译 | 软件架构师 |
| v1.2 | 2026-04-01 | 重大架构变更：引入控制台(Console)+执行器(Executor)双角色架构，etcd作为配置中心，新增环境与集群概念 | 软件架构师 |
| v1.3 | 2026-04-01 | 补充执行器状态上报机制和部署面板设计 | 软件架构师 |
| v1.4 | 2026-04-07 | 评审修订：树形流程、分支/迭代控制节点、自动布局(ElkJS)、3段版本号、执行器心跳5s、Timer锁粒度 | 软件架构师 |


## 1. 架构愿景与设计原则

### 1.1 架构愿景

构建一个**控制台(Console) + 执行器(Executor)** 双角色架构的 API 编排平台，具备以下核心特征：

- **双角色二进制**：同一二进制通过 CLI 参数区分控制台（管理）和执行器（执行）角色
- **etcd 配置中心**：Console 将配置发布到 etcd，Executor 从 etcd 监听并热加载
- **多集群支持**：一套 Console 可管理多个集群，每个集群绑定独立的 etcd 实例
- **环境标签**：集群按环境（dev/staging/prod）分类，便于管理
- **私有化部署优先**：Docker Compose 一键启动，无外部 SaaS 依赖
- **插件驱动扩展**：资源类型和触发器类型通过插件接口扩展，编译时按需打包
- **版本一致性保证**：工具→流程→触发器的版本锁定机制，确保运行时可复现
- **多租户数据隔离**：以工作空间为单位的数据和权限隔离

### 1.2 设计原则

| 原则 | 说明 | 权衡 |
|------|------|------|
| **双角色架构** | 控制台（管理）和执行器（执行）编译到同一二进制，通过 CLI 区分角色 | 简化部署和版本管理，但二进制体积略大 |
| **配置中心解耦** | Console 和 Executor 通过 etcd 解耦，Executor 无状态设计 | 引入 etcd 运维复杂度，但实现配置热更新和水平扩容 |
| **模块化单体优先** | 通过 Go internal package 实现模块边界，单进程内按领域划分 | 牺牲独立扩缩容能力，换取更简单的开发、调试和部署 |
| **领域驱动分层** | 领域层不依赖基础设施层，依赖倒置通过接口实现 | 增加代码结构复杂度，但提高可测试性和可替换性 |
| **插件化核心能力** | 资源类型和触发器类型通过插件接口扩展，编译时按客户需求选择性集成 | 新增能力需要重新编译，不适合客户自助扩展 |
| **版本锁定而非版本跟随** | 流程运行时使用发布时锁定的工具版本 | 升级需要显式操作，但保证运行时一致性 |
| **异步执行 + 同步反馈** | 流程执行异步调度，通过轮询或 WebSocket 提供实时状态 | 增加系统复杂度，但避免长时间 HTTP 连接阻塞 |
| **数据加密存储** | 敏感配置（密码、API Key）加密落库 | 增加性能开销，但满足企业安全合规 |

### 1.3 架构风险与缓解

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| etcd 单点故障 | 中 | 高 | etcd 集群部署（3/5 节点），自动选主 |
| 执行器故障转移 | 中 | 高 | 多执行器实例 + 心跳检测 + 触发器自动重平衡 |
| 配置同步延迟 | 低 | 中 | etcd Watch 机制 + 本地缓存 + 降级读取 |
| 流程引擎性能瓶颈 | 中 | 高 | 异步执行 + 限流 + 超时控制 |
| 插件集成构建复杂度 | 低 | 中 | CI/CD 流水线支持多配置编译，plugins.yaml 声明式管理 |
| 画布状态同步冲突 | 中 | 中 | 乐观锁 + 最后写入胜出 + 冲突提示 |
| 大规模并发流程 | 低 | 高 | 执行器水平扩容，无状态设计支持多实例 |

---

## 2. 系统上下文

### 2.1 C4 — 系统上下文图

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               企业内网环境                                        │
│                                                                                 │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐      │
│  │ 开发者    │   │ 管理员    │   │ 第三方系统 │   │ 消息队列  │   │  外部服务  │      │
│  │ (浏览器)  │   │ (浏览器)  │   │ (HTTP)   │   │ (MQ)    │   │(HTTP/DB) │      │
│  └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘   └────▲─────┘      │
│       │              │              │              │               │            │
│       │    ┌─────────┴──────────────┴──────────────┘               │            │
│       │    │                                                       │            │
│       ▼    ▼                                                       │            │
│  ┌────────────────────────────────────────────────────────────┐    │            │
│  │                      Console (控制台)                       │    │            │
│  │                                                            │    │            │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────────────┐ │    │            │
│  │  │ Web 前端  │  │ Go API   │  │      发布服务             │ │    │            │
│  │  │ (React)  │  │ 服务     │  │  (写入 etcd)              │ │    │            │
│  │  └──────────┘  └────┬─────┘  └──────────────────────────┘ │    │            │
│  └───────────────────────┼────────────────────────────────────┘    │            │
│                          │                                         │            │
│              ┌───────────┼───────────┐                             │            │
│              ▼           ▼           ▼                             │            │
│        ┌──────────┐ ┌──────────┐ ┌──────────┐                     │            │
│        │PostgreSQL│ │  Redis   │ │   etcd   │◄────────────────────┘            │
│        │ (数据存储)│ │(缓存/会话)│ │(配置中心)│                                  │
│        └──────────┘ └──────────┘ └────┬─────┘                                  │
│                                       │                                        │
│                              ┌────────┴────────┐                               │
│                              ▼                 ▼                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                        Executor (执行器集群)                             │  │
│  │                                                                         │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │  │
│  │  │ 配置监听服务  │  │ 触发器管理器  │  │ 流程执行引擎  │  │ 插件运行时    │ │  │
│  │  │ (Watch etcd) │  │(REST/Cron/MQ)│  │ (异步调度)   │  │              │ │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘ │  │
│  │                                                                         │  │
│  │  实例 x N（无状态，水平扩容）                                             │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 参与者与交互

| 参与者 | 角色 | 交互方式 |
|--------|------|---------|
| 接口开发者 | 创建资源、定义工具、编排流程、配置触发器 | Web 浏览器 → Console 前端 → Console API |
| 系统管理员 | 管理用户、工作空间、集群、插件、系统配置 | Web 浏览器 → Console 前端 → Console API |
| 第三方系统 | 调用触发器触发的 API 端点 | HTTP → Executor 触发器 |
| 外部服务 | 被工具调用的 HTTP/gRPC 服务、数据库等 | Executor → 插件 → 外部服务 |
| 消息队列 | 触发 MQ 类型触发器的消息源 | MQ Consumer → Executor |
| etcd | 配置中心，存储集群配置 | Console 写入 ↔ Executor 监听 |

---

## 3. 技术决策记录 (ADR)

### ADR-001: 模块化单体架构（非微服务）

**状态**：已接受

**上下文**：我们需要决定 MVP 阶段的系统架构风格。选项包括微服务、模块化单体、事件驱动架构。

**决策**：采用**模块化单体**架构。在单体进程中通过 Go 的 internal package 划分清晰的模块边界，每个模块有独立的领域层、数据访问层。

**后果**：
- ✅ 开发部署简单，Docker Compose 一个服务搞定
- ✅ 调试方便，事务一致性天然保证
- ✅ 模块边界清晰，未来可拆微服务
- ❌ 无法独立扩缩容某个模块
- ❌ 单点故障风险（通过多实例水平扩容缓解）

---

### ADR-002: 后端语言选型 — Go

**状态**：已接受

**上下文**：后端需要选择开发语言。候选方案：Go、Java、Node.js、Python。

**决策**：选择 **Go** 作为后端开发语言。

**理由**：
- 高性能 + 低内存占用，适合私有化部署场景
- 原生并发模型（goroutine），适合流程执行引擎
- 单二进制部署，无需运行时环境
- 编译型语言，类型安全，适合编译时插件集成
- 交叉编译支持，方便为不同客户构建定制版本

**权衡**：
- 相比 Node.js，前端团队上手成本略高
- 相比 Java，企业级生态（ORM、监控）不够成熟

---

### ADR-003: 前端技术栈 — React + TypeScript + Vite

**状态**：已接受

**上下文**：前端需要选择框架和工具链。候选方案：React、Vue、Angular。

**决策**：选择 **React 18+ TypeScript + Vite + Tailwind CSS + shadcn/ui**。

**理由**：
- React 生态最成熟，React Flow 提供专业的流程编排画布
- TypeScript 提供类型安全，shadcn/ui 基于 Radix UI 可定制性强
- Tailwind CSS 配合 Design Tokens 实现一致的视觉系统
- Zustand + TanStack Query 管理状态，轻量高效

---

### ADR-004: 数据库 — PostgreSQL

**状态**：已接受

**上下文**：主数据库选型。候选方案：PostgreSQL、MySQL、MongoDB。

**决策**：选择 **PostgreSQL 15+**。

**理由**：
- JSONB 类型支持，天然适配工具配置、画布数据等半结构化数据
- 强大的查询能力，支持 CTE、窗口函数
- 行级安全策略（RLS），可辅助多租户数据隔离
- 成熟稳定，社区活跃

---

### ADR-005: 插件系统 — 编译时集成（Build-time Plugin）

**状态**：已接受

**上下文**：资源类型和触发器类型需要可扩展。需要根据不同客户的合同需求，交付包含不同能力组合的平台版本。Go 的 `plugin` 包存在严格的版本兼容性限制（同一 Go 版本编译才能加载），不适合生产使用。

**决策**：采用**编译时插件（Build-time Plugin）**模式——所有插件在编译阶段通过 Go build tags 选择性打包进二进制，不同客户编译出包含不同插件组合的定制版本。

- **Build Tags 选择**：通过 `//go:build` 标签控制哪些插件参与编译
- **插件注册表**：每个插件的 `init()` 函数在编译时自动注册到全局 PluginRegistry
- **配置文件驱动**：通过 `plugins.yaml` 声明当前构建包含的插件列表
- **Feature Flag API**：运行时通过 `PluginRegistry` 查询已编译的插件能力，前端据此动态渲染 UI

**构建流程**：
```
1. 根据客户合同确定需要的插件组合
2. 生成 plugins.yaml（声明启用哪些插件）
3. go build -tags "http,grpc,mysql,timer,restful,rabbitmq" -o api-orchestrator
4. 产物是单一二进制文件，包含选定插件的全部能力
```

**后果**：
- ✅ 零运行时插件兼容性问题——所有插件编译期类型检查
- ✅ 零动态加载开销——进程内直接调用，性能最优
- ✅ 部署简单——单一二进制，无需额外插件文件
- ✅ 不同客户版本明确隔离，不会误加载不需要的插件
- ❌ 新增/修改插件需要重新编译部署
- ❌ 客户无法自助添加新插件，需要厂商支持
- ❌ 需要维护构建流水线（CI/CD）支持多版本编译

---

### ADR-006: 流程执行 — 异步调度 + 实时状态推送

**状态**：已接受

**上下文**：流程执行可能耗时数秒到数分钟（取决于工具调用延迟）。需要决定执行模式。

**决策**：流程执行采用**异步调度**模式：
1. 触发器收到请求 → 创建 Run 记录（状态 pending）→ 立即返回 Run ID
2. 执行引擎异步消费 Run → 按节点顺序执行 → 更新状态
3. 前端通过 **轮询**（MVP）或 **WebSocket**（迭代）获取实时状态

**后果**：
- ✅ 避免长时间 HTTP 连接阻塞
- ✅ 支持并发执行多个流程
- ✅ 执行失败可自动重试
- ❌ 前端需要额外轮询逻辑
- ❌ 不是立即返回结果的同步体验（对于 RESTful 触发器场景，可在短超时内等待结果）

---

### ADR-007: 双角色架构 — Console + Executor

**状态**：已接受

**上下文**：需要决定系统的部署架构。选项包括：
1. 单一服务（管理 + 执行合一）
2. 微服务拆分（管理服务、执行服务分离）
3. **双角色二进制**（同一二进制，CLI 区分角色）

**决策**：采用**双角色二进制**架构：
- 同一二进制包含 Console（管理）和 Executor（执行）两种角色
- 通过 `--role=console` 或 `--role=executor` 启动参数区分
- Console 提供 Web UI 和 API，管理资源/工具/流程/触发器定义
- Executor 连接 etcd，监听配置变更，执行触发器和流程

**后果**：
- ✅ 简化部署，一套二进制即可满足两种场景
- ✅ 版本一致性，Console 和 Executor 永远版本匹配
- ✅ 灵活部署，单机开发可双角色同时启动，生产环境可分离部署
- ✅ 执行器无状态，支持水平扩容
- ❌ 二进制体积略大（包含两种角色的代码）
- ❌ 需要 etcd 作为配置中心，增加基础设施依赖

---

### ADR-008: 配置中心选型 — etcd

**状态**：已接受

**上下文**：Console 和 Executor 需要解耦，Executor 需要实时获取配置变更。候选方案：
1. **etcd** — 原生 Watch 机制，Kubernetes 同款
2. **Consul** — 服务发现强，但 Watch 机制不如 etcd
3. **Redis** — 熟悉度高，但无原生 Watch，需轮询
4. **数据库轮询** — 简单，但延迟高、效率低

**决策**：选择 **etcd** 作为配置中心。

**理由**：
- 原生 Watch API，配置变更实时推送
- 强一致性，基于 Raft 共识算法
- 键值存储模型简单，适合配置存储
- 云原生生态成熟，运维工具丰富
- 支持租约（Lease）机制，可用于执行器心跳

**etcd 存储结构**：
```
/clusters/{cluster-id}/
  ├── /flows/{flow-id}          # 流程定义 JSON
  ├── /triggers/{trigger-id}    # 触发器配置 JSON
  ├── /resources/{res-id}       # 资源配置 JSON
  ├── /tools/{tool-id}          # 工具定义 JSON
  └── /executors/{executor-id}  # 执行器心跳信息
```

**后果**：
- ✅ 配置变更实时同步（< 1s）
- ✅ 执行器无状态，故障后可快速恢复
- ✅ 支持多集群，每个集群独立 etcd 实例
- ❌ 引入新的基础设施组件，需要运维 etcd 集群
- ❌ 需要处理 etcd 连接异常、重连等边界情况

---

### ADR-009: 集群与环境模型

**状态**：已接受

**上下文**：需要设计多集群部署模型，支持不同环境（dev/staging/prod）的隔离。

**决策**：
1. **环境（Environment）** 是集群的标签/分类，用于逻辑分组
2. **集群（Cluster）** 是一组运行相同配置的执行器实例
3. 一个集群绑定**唯一一套 etcd**（一对多）
4. 一套 etcd 可服务**多个集群**（通过不同根节点路径区分）

**模型关系**：
```
Environment (标签) 1 ─── N Cluster (执行器组)
Cluster 1 ─── 1 etcd 实例 (通过 cluster-id 区分根节点)
etcd 实例 1 ─── N Cluster (多对多)
```

**后果**：
- ✅ 清晰的部署边界，每个集群独立扩缩容
- ✅ 环境标签便于管理和筛选
- ✅ 支持多地域部署（不同集群使用不同 etcd）
- ❌ 流程/触发器需要指定目标集群，增加配置复杂度
- ❌ 跨集群流程调用需要额外设计（MVP 暂不支持）

---

## 4. 领域驱动设计 — 限界上下文

### 4.1 上下文映射

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            平台限界上下文映射                                  │
│                                                                             │
│  ┌────────────────┐     ┌────────────────┐     ┌────────────────┐          │
│  │  身份认证上下文   │◄───►│  工作空间上下文   │◄───►│  集群管理上下文   │          │
│  │  Identity      │     │  Workspace     │     │  Cluster       │          │
│  │  (核心域)       │     │  (核心域)       │     │  (核心域)       │          │
│  └───────┬────────┘     └───────┬────────┘     └───────┬────────┘          │
│          │ 下游                  │ 下游                  │ 下游              │
│  ┌───────┴────────┐             │                      │                  │
│  │  资源管理上下文   │             │                      │                  │
│  │  Resource      │─────────────┘                      │                  │
│  │  (核心域)       │                                  │                  │
│  └───────┬────────┘                                  │                  │
│          │                                          │                  │
│  ┌───────┴────────┐     ┌────────────────┐          │                  │
│  │  工具管理上下文   │────►│  流程编排上下文   │◄─────────┘                  │
│  │  Tool          │     │  Flow          │     (触发器绑定集群)              │
│  │  (核心域)       │     │  (核心域)       │                             │
│  └────────────────┘     └───────┬────────┘                             │
│                                 │                                      │
│                         ┌───────┴────────┐                             │
│                         │  触发器上下文     │                             │
│                         │  Trigger       │                             │
│                         │  (核心域)       │                             │
│                         └───────┬────────┘                             │
│                                 │                                      │
│                         ┌───────┴────────┐                             │
│                         │  运行记录上下文   │                             │
│                         │  Run           │                             │
│                         │  (支撑域)       │                             │
│                         └────────────────┘                             │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────┐         │
│  │ 跨切面关注点（通用域）                                      │         │
│  │  ● 审计日志 (AuditLog)                                   │         │
│  │  ● 插件管理 (PluginRegistry)                             │         │
│  │  ● 系统设置 (SystemConfig)                               │         │
│  │  ● 配置发布 (ConfigPublisher) ──► etcd                   │         │
│  └──────────────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 上下文详细描述

#### 4.2.1 身份认证上下文 (Identity Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | User |
| **职责** | 用户注册/登录、MFA、SSO 集成、JWT 令牌管理 |
| **上游** | 无 |
| **下游** | Workspace（提供用户身份验证）、所有上下文（权限校验） |
| **防腐层** | SSO Provider Adapter（CAS/SAML/OAuth）、LDAP Adapter、第三方登录 Adapter |

#### 4.2.2 工作空间上下文 (Workspace Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Workspace |
| **职责** | 工作空间创建/管理、成员邀请/移除、角色分配、资源隔离 |
| **上游** | Identity（获取用户信息） |
| **下游** | Resource、Flow、Trigger（提供工作空间上下文）、Cluster（发布配置到集群） |
| **防腐层** | 无 |

#### 4.2.3 集群管理上下文 (Cluster Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Cluster |
| **职责** | 集群创建/管理、etcd 连接配置、执行器心跳监控、配置发布 |
| **上游** | Workspace（获取工作空间信息） |
| **下游** | Trigger（触发器绑定集群）、etcd（配置存储） |
| **防腐层** | etcd Client（封装 etcd 操作）、Executor Health Monitor（执行器健康检查） |

**关键概念：**
- **Environment**：环境的标签/分类，如 dev/staging/prod
- **Cluster**：执行器集群，绑定唯一的 etcd 实例
- **ConfigPublisher**：将配置发布到指定集群的 etcd

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Workspace、WorkspaceMember |
| **职责** | 工作空间 CRUD、成员管理、角色分配、数据隔离 |
| **上游** | Identity（验证用户身份） |
| **下游** | Resource、Tool、Flow、Trigger（提供 workspace_id 隔离） |
| **防腐层** | 无 |

#### 4.2.4 资源管理上下文 (Resource Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Resource |
| **职责** | 资源 CRUD、连接测试、状态探活、敏感信息加密 |
| **上游** | Workspace（数据隔离）、PluginRegistry（插件 Schema） |
| **下游** | Tool（资源下挂工具） |
| **防腐层** | 插件调用层（将资源配置映射为插件调用参数） |

#### 4.2.5 工具管理上下文 (Tool Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Tool、ToolVersion |
| **职责** | 工具 CRUD、版本管理、发布/下线、批量导入、影响分析 |
| **版本号** | 采用 3 段版本号（x.y.z），如 v1.0.0，每次发布自动递增 |
| **上游** | Resource（所属资源）、PluginRegistry（执行工具） |
| **下游** | Flow（已发布工具供流程编排） |
| **防腐层** | OpenAPI Parser、Proto Parser、SQLc Parser |

#### 4.2.6 流程编排上下文 (Flow Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Flow、FlowVersion、FlowToolRef |
| **职责** | 画布数据 CRUD、版本管理、发布、工具引用锁定 |
| **版本号** | 采用 3 段版本号（x.y.z），如 v1.0.0，每次发布自动递增 |
| **上游** | Tool（引用已发布工具）、Workspace（数据隔离） |
| **下游** | Trigger（绑定已发布流程）、Run（提供执行配置） |

#### 4.2.7 触发器上下文 (Trigger Context)

| 维度 | 说明 |
|------|------|
| **类型** | 核心域 |
| **聚合根** | Trigger |
| **职责** | 触发器 CRUD、启停控制、输入输出映射、错误处理配置 |
| **上游** | Flow（绑定已发布流程）、PluginRegistry（触发器插件） |
| **下游** | Run（创建运行记录） |

#### 4.2.8 运行记录上下文 (Run Context)

| 维度 | 说明 |
|------|------|
| **类型** | 支撑域 |
| **聚合根** | Run、RunNodeLog |
| **职责** | 运行记录查询、节点执行日志、数据流追踪 |
| **上游** | Trigger（触发创建）、Flow（提供执行配置） |
| **下游** | 无（纯查询上下文） |

---

## 5. 领域模型

### 5.1 聚合与不变量

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                             领域模型关系图                                     │
│                                                                             │
│  ┌──────────────┐                                                           │
│  │ Environment  │     1:N                                                    │
│  │  ─────────── │◄──────────┐                                                │
│  │  id          │           │                                                │
│  │  key         │           │                                                │
│  │  name        │           │                                                │
│  └──────────────┘           │                                                │
│                             │                                                │
│  ┌──────────────┐      1:N  │      1:N      ┌──────────────┐               │
│  │   Workspace   │◄──────────┼──────────────│   Resource    │               │
│  │  ─────────── │           │               │  ─────────── │               │
│  │  id          │           │               │  id          │               │
│  │  name        │           │               │  name        │               │
│  │  members[]   │           │               │  type        │               │
│  └──────┬───────┘           │               │  config      │               │
│         │ 1:N               │               │  status      │               │
│  ┌──────┴───────┐           │               └──────┬───────┘               │
│  │   Cluster    │◄──────────┘      1:N             │                       │
│  │  ─────────── │◄───────────────────────────────│                       │
│  │  id          │                                  │                       │
│  │  name        │      1:N      ┌──────────────┐  │                       │
│  │  etcd_config │◄─────────────│   Tool       │◄─┘                       │
│  │  status      │               │  ─────────── │                          │
│  └──────────────┘               │  id          │                          │
│                                 │  name        │                          │
│                                 │  version     │                          │
│                                 │  status      │                          │
│                                 │  versions[]  │                          │
│                                 └──────┬───────┘                          │
│                                        │ N:M (FlowToolRef)                 │
│                                 ┌──────┴───────┐                          │
│                                 │    Flow      │                          │
│                                 │  ─────────── │                          │
│                                 │  id          │                          │
│                                 │  name        │                          │
│                                 │  canvas_data │                          │
│                                 │  version     │                          │
│                                 │  versions[]  │                          │
│                                 │  tool_refs[] │                          │
│                                 └──────┬───────┘                          │
│                                        │ 1:N                               │
│                                 ┌──────┴───────┐      N:M (TriggerCluster)│
│                                 │   Trigger    │◄─────────────────────────┤
│                                 │  ─────────── │                          │
│                                 │  id          │                          │
│                                 │  name        │                          │
│                                 │  type        │                          │
│                                 │  flow_id     │                          │
│                                 │  flow_version│                          │
│                                 │  cluster_ids │                          │
│                                 └──────┬───────┘                          │
│                                        │ 1:N                               │
│                                 ┌──────┴───────┐                          │
│                                 │    Run       │                          │
│                                 │  ─────────── │                          │
│                                 │  id          │                          │
│                                 │  cluster_id  │                          │
│  │  status      │                      │                            │
│  │  node_logs[] │                      │                            │
│  └──────────────┘                      │                            │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 核心不变量（Invariants）

| 聚合 | 不变量 | 校验时机 |
|------|--------|---------|
| **Resource** | 同一工作空间内名称唯一 | 创建/更新时 |
| **Resource** | config 必须符合对应资源类型的 JSON Schema | 创建/更新时 |
| **Tool** | 同一资源下名称唯一 | 创建/更新时 |
| **Tool** | 只有 status=draft 的工具可以编辑 | 更新操作前 |
| **Tool** | 有 Flow 引用（任何版本）的 Tool 不可下线 | 下线操作前 |
| **Flow** | 只能引用已发布的 Tool | 创建/更新画布时 |
| **Flow** | 只有 status=draft 的流程可以编辑 | 更新操作前 |
| **Flow** | 有 Trigger 绑定的 Flow 不可删除 | 删除操作前 |
| **Trigger** | 只能绑定已发布的 Flow | 创建/更新时 |
| **Trigger** | 同一工作空间内 RESTful 类型的 path+method 唯一 | 创建时 |
| **WorkspaceMember** | 每个用户在每个工作空间只有一条成员记录 | 邀请时 |
| **Run** | Run 的 flow_version 必须等于 Trigger 锁定的版本 | 创建时 |

### 5.3 领域事件

| 事件 | 触发时机 | 消费者 | 说明 |
|------|---------|--------|------|
| `resource.created` | 资源创建后 | 审计日志 | |
| `resource.updated` | 资源更新后 | 审计日志 | |
| `resource.deleted` | 资源删除后 | 审计日志 | |
| `resource.health_changed` | 连接状态变更 | Dashboard 推送 | 定时探活触发 |
| `tool.published` | 工具发布后 | 影响分析服务 | 检测下游流程并标记升级 |
| `tool.deprecated` | 工具下线后 | 通知服务 | |
| `flow.published` | 流程发布后 | — | |
| `trigger.created` | 触发器创建后 | 触发器调度器 | 注册到调度器 |
| `trigger.enabled` | 触发器启用后 | 触发器调度器 | |
| `trigger.disabled` | 触发器停用后 | 触发器调度器 | 从调度器注销 |
| `run.created` | 运行记录创建 | 执行引擎 | 入队执行 |
| `run.completed` | 运行完成 | Dashboard 推送、审计日志 | |
| `run.failed` | 运行失败 | 审计日志、告警（可配） | |

---

## 6. 分层架构与模块结构

### 6.1 后端分层架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        后端分层架构                                   │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   API Layer (接口层)                          │  │
│  │  HTTP Handlers  │  Request/Response DTOs  │  Middleware       │  │
│  │  (Gin/Echo)     │  Validation (Zod-like)  │  (Auth/CORS/...) │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │                                       │
│  ┌──────────────────────────┴───────────────────────────────────┐  │
│  │                Application Layer (应用层)                      │  │
│  │  Use Cases / Services  │  Command Handlers  │  Event Handlers │  │
│  │  (Orchestration)       │  (CQRS Commands)  │  (Async)        │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │                                       │
│  ┌──────────────────────────┴───────────────────────────────────┐  │
│  │                  Domain Layer (领域层)                         │  │
│  │  Entities  │  Value Objects  │  Aggregates  │  Domain Events  │  │
│  │  (Pure Go, No Dependencies)                               │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │ 依赖倒置 (接口)                        │
│  ┌──────────────────────────┴───────────────────────────────────┐  │
│  │              Infrastructure Layer (基础设施层)                  │  │
│  │  Repository Impl  │  Plugin Client  │  Cache (Redis)          │  │
│  │  Database (pgx)   │  MQ Producer    │  External Service       │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 项目目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 入口：初始化依赖、启动 HTTP Server
│
├── internal/                       # 内部包（模块边界）
│   ├── api/                        # 接口层
│   │   ├── middleware/             # HTTP 中间件
│   │   │   ├── auth.go            # JWT 认证中间件
│   │   │   ├── cors.go            # CORS 中间件
│   │   │   ├── rbac.go            # RBAC 权限校验
│   │   │   ├── requestid.go       # 请求追踪 ID
│   │   │   └── ratelimit.go       # 限流
│   │   ├── handler/               # HTTP Handler（按上下文分组）
│   │   │   ├── auth.go
│   │   │   ├── resource.go
│   │   │   ├── tool.go
│   │   │   ├── flow.go
│   │   │   ├── trigger.go
│   │   │   ├── run.go
│   │   │   ├── workspace.go
│   │   │   ├── user.go
│   │   │   └── admin.go
│   │   ├── router.go              # 路由注册
│   │   └── dto/                   # Request/Response DTO
│   │       ├── auth.go
│   │       ├── resource.go
│   │       ├── tool.go
│   │       ├── flow.go
│   │       ├── trigger.go
│   │       └── run.go
│   │
│   ├── domain/                     # 领域层（纯 Go，无外部依赖）
│   │   ├── common/                 # 共享值对象
│   │   │   ├── id.go              # UUID 值对象
│   │   │   ├── version.go         # 版本号值对象
│   │   │   └── event.go           # 领域事件基类
│   │   ├── identity/               # 身份认证上下文
│   │   │   ├── user.go            # User 聚合根
│   │   │   ├── auth_token.go      # JWT 令牌值对象
│   │   │   ├── repository.go      # Repository 接口
│   │   │   └── service.go         # 领域服务
│   │   ├── workspace/              # 工作空间上下文
│   │   │   ├── workspace.go       # Workspace 聚合根
│   │   │   ├── member.go          # Member 实体
│   │   │   ├── role.go            # Role 值对象
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── resource/               # 资源管理上下文
│   │   │   ├── resource.go
│   │   │   ├── resource_config.go # 资源配置值对象
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── tool/                   # 工具管理上下文
│   │   │   ├── tool.go
│   │   │   ├── tool_version.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── flow/                   # 流程编排上下文
│   │   │   ├── flow.go
│   │   │   ├── flow_version.go
│   │   │   ├── canvas.go          # 画布数据值对象
│   │   │   ├── flow_tool_ref.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── trigger/                # 触发器上下文
│   │   │   ├── trigger.go
│   │   │   ├── input_output_mapping.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── run/                    # 运行记录上下文
│   │       ├── run.go
│   │       ├── run_node_log.go
│   │       ├── repository.go
│   │       └── service.go
│   │
│   ├── app/                        # 应用层（用例编排）
│   │   ├── auth/
│   │   │   ├── login_usecase.go
│   │   │   └── refresh_usecase.go
│   │   ├── resource/
│   │   │   ├── create_resource.go
│   │   │   ├── test_connection.go
│   │   │   └── health_check.go
│   │   ├── tool/
│   │   │   ├── create_tool.go
│   │   │   ├── import_tools.go
│   │   │   ├── publish_tool.go
│   │   │   └── test_tool.go
│   │   ├── flow/
│   │   │   ├── create_flow.go
│   │   │   ├── save_canvas.go
│   │   │   ├── publish_flow.go
│   │   │   └── test_run_flow.go
│   │   ├── trigger/
│   │   │   ├── create_trigger.go
│   │   │   └── invoke_trigger.go
│   │   └── run/
│   │       ├── execute_flow.go    # 流程执行用例
│   │       └── query_run.go
│   │
│   ├── infra/                      # 基础设施层
│   │   ├── database/               # 数据库
│   │   │   ├── postgres.go        # 连接管理
│   │   │   ├── migrations/        # SQL 迁移文件
│   │   │   └── repo/              # Repository 实现
│   │   │       ├── user_repo.go
│   │   │       ├── workspace_repo.go
│   │   │       ├── resource_repo.go
│   │   │       ├── tool_repo.go
│   │   │       ├── flow_repo.go
│   │   │       ├── trigger_repo.go
│   │   │       └── run_repo.go
│   │   ├── cache/                  # Redis 缓存
│   │   │   ├── redis.go
│   │   │   └── cache_provider.go
│   │   ├── plugin/                 # 插件基础设施
│   │   │   ├── registry.go        # 插件注册表（运行时查询已编译插件）
│   │   │   └── types.go           # 插件接口定义（ResourcePlugin / TriggerPlugin）
│   │   ├── mq/                     # 消息队列
│   │   │   ├── producer.go
│   │   │   └── consumer.go
│   │   ├── crypto/                 # 加密服务
│   │   │   └── encryptor.go       # AES-256 加密敏感配置
│   │   └── eventbus/               # 领域事件总线
│   │       ├── bus.go
│   │       └── handler.go
│   │
│   └── pkg/                        # 公共工具包
│       ├── logger/                 # 结构化日志
│       ├── errors/                 # 统一错误处理
│       ├── pagination/             # 分页工具
│       └── validator/              # 参数校验
│
├── plugins/                        # 编译时插件（通过 build tags 选择性编译）
│   ├── plugins.yaml                # 声明当前构建包含的插件列表（CI 用）
│   ├── resource/                   # 资源插件
│   │   ├── http/
│   │   │   ├── plugin.go           //go:build http
│   │   │   ├── schema.json
│   │   │   ├── executor.go
│   │   │   └── openapi_parser.go
│   │   ├── grpc/
│   │   │   ├── plugin.go           //go:build grpc
│   │   │   ├── schema.json
│   │   │   └── executor.go
│   │   ├── mysql/
│   │   │   ├── plugin.go           //go:build mysql
│   │   │   └── schema.json
│   │   └── postgres/
│   │       ├── plugin.go           //go:build postgres
│   │       └── schema.json
│   └── trigger/                    # 触发器插件
│       ├── restful/
│       │   ├── plugin.go           //go:build restful
│       │   └── schema.json
│       ├── timer/
│       │   ├── plugin.go           //go:build timer
│       │   ├── schema.json
│       │   └── scheduler.go
│       ├── rabbitmq/
│       │   ├── plugin.go           //go:build rabbitmq
│       │   └── schema.json
│       └── kafka/
│           ├── plugin.go           //go:build kafka
│           └── schema.json
│
├── config/
│   ├── config.go                   # 配置结构体
│   └── default.yaml                # 默认配置
│
├── migrations/                     # 数据库迁移（SQL 文件）
│   ├── 001_init.sql
│   ├── 002_add_indexes.sql
│   └── ...
│
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

---

## 7. 核心流程引擎设计

### 7.1 执行架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                      流程执行引擎架构                                  │
│                                                                     │
│  ┌───────────┐     ┌──────────────┐     ┌──────────────────────┐  │
│  │  触发器     │────►│  Run 创建器   │────►│  执行队列 (Redis)     │  │
│  │  Handler   │     │  (状态pending) │     │  List: run:pending  │  │
│  └───────────┘     └──────────────┘     └──────────┬───────────┘  │
│                                                   │                 │
│                                                   │ BRPOP          │
│                                                   ▼                 │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    执行引擎 (Worker Pool)                      │  │
│  │                                                              │  │
│  │  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐ │  │
│  │  │   Worker 1   │     │   Worker 2   │     │   Worker N   │ │  │
│  │  │  (goroutine) │     │  (goroutine) │     │  (goroutine) │ │  │
│  │  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘ │  │
│  │         │                    │                    │         │  │
│  │         ▼                    ▼                    ▼         │  │
│  │  ┌──────────────────────────────────────────────────────┐   │  │
│  │  │              节点执行器 (Node Executor)                 │   │  │
│  │  │                                                      │   │  │
│  │  │  1. 解析画布数据 → 构建执行 DAG                        │   │  │
│  │  │  2. 拓扑排序 → 确定节点执行顺序                         │   │  │
│  │  │  3. 按序执行节点:                                      │   │  │
│  │  │     a. 解析参数映射 ({{node_id.output.field}})          │   │  │
│  │  │     b. 调用插件执行工具                                  │   │  │
│  │  │     c. 记录节点日志 (input/output/status/duration)       │   │  │
│  │  │     d. 失败时根据策略处理 (终止/重试/跳过)                │   │  │
│  │  │  4. 更新 Run 状态 (success/failed)                      │   │  │
│  │  │  5. 发布 run.completed / run.failed 事件                │   │  │
│  │  └──────────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                 参数映射引擎 (Mapping Engine)                   │  │
│  │                                                              │  │
│  │  {{node_1.output.userId}}  ──►  从节点1输出中提取 userId      │  │
│  │  {{flow.username}}         ──►  从流程输入参数中获取 username  │  │
│  │  {{trigger.body.orderId}}  ──►  从触发器请求体中获取 orderId   │  │
│  │  {{$now}}                  ──►  当前时间戳（系统变量）          │  │
│  │  {{$uuid}}                 ──►  生成新 UUID（系统变量）        │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.2 执行流程时序

```
触发器请求  │  后端服务  │  Redis  │  执行引擎  │  外部服务
    │          │          │         │           │
    ├─POST────►│          │         │           │
    │          │─创建Run──►│         │           │
    │◄─202 +ID─│          │         │           │
    │          │─LPUSH────►│         │           │
    │          │          ├─BRPOP──►│           │
    │          │          │         │─更新:running
    │          │          │         │─执行Node1──►│
    │          │          │         │◄──200 response
    │          │          │         │─记录NodeLog │
    │          │          │         │─执行Node2──►│
    │          │          │         │◄──200 response
    │          │          │         │─记录NodeLog │
    │          │          │         │─更新:success│
    │          │          │         │           │
    │  GET status───►     │         │           │
    │◄─{status:success}──│         │           │
```

### 7.3 画布数据 DAG 解析

```go
// CanvasDAG 画布有向无环图
type CanvasDAG struct {
    Nodes    map[string]*CanvasNode  // node_id -> node
    Edges    []*CanvasEdge           // 连线
    AdjList  map[string][]string     // node_id -> 下游 node_ids
    InDegree map[string]int          // node_id -> 入度
}

// TopologicalSort 拓扑排序，返回执行顺序
func (dag *CanvasDAG) TopologicalSort() ([]string, error) {
    // Kahn 算法
    // 1. 找到入度为 0 的节点（Start 节点）
    // 2. 依次取出，更新邻接节点入度
    // 3. 如果存在环则返回错误
}

// ResolveMapping 解析参数映射
// 将 {{node_1.output.userId}} 替换为实际的值
func ResolveMapping(template string, context *ExecutionContext) (interface{}, error)
```

### 7.4 节点执行策略

| 策略 | 说明 | 配置 |
|------|------|------|
| **失败终止** | 任一节点失败则整个流程终止 | 默认策略 |
| **重试** | 节点失败后自动重试 N 次 | 可配置次数和间隔 |
| **超时** | 节点执行超过指定时间则中断 | 可配置超时时间 |
| **跳过** | 条件不满足时跳过节点 | 条件分支场景（MVP 后） |

### 7.5 控制节点 — 分支与迭代

树形流程支持两种控制节点：**分支（Branch）** 和 **迭代（Loop）**。

#### 7.5.1 分支节点

**功能**：根据条件判断，将数据路由到不同的分支执行。

**数据结构**：
```go
// BranchNodeConfig 分支节点配置
type BranchNodeConfig struct {
    Branches    []*Branch      `json:"branches"`     // 分支列表
    MultiMatch  bool           `json:"multiMatch"`   // 是否多命中
    Parallel    bool           `json:"parallel"`     // 多命中时是否并行执行
}

// Branch 分支定义
type Branch struct {
    ID       string              `json:"id"`        // 分支 ID
    Name     string              `json:"name"`      // 分支名称
    Condition *ConditionExpr     `json:"condition"` // 条件表达式
    NodeID   string              `json:"nodeId"`    // 分支执行的第一个节点 ID
}

// ConditionExpr 条件表达式
type ConditionExpr struct {
    Field    string `json:"field"`     // 引用字段，如 {{node_1.output.status}}
    Operator string `json:"operator"`  // 操作符: ==, !=, >, <, >=, <=, contains, in
    Value    any    `json:"value"`    // 比较值
}
```

**执行逻辑**：

```
┌────────────────────────────────────────────────────────────┐
│                    分支节点执行流程                         │
│                                                            │
│  1. 解析条件字段值                                          │
│  2. 按顺序遍历分支，判断条件                                 │
│  3. 收集命中的分支列表                                       │
│     ├── 单次命中: 执行第一个命中分支，忽略后续               │
│     └── 多命中:                                             │
│          ├── 禁用并行: 按顺序依次执行                        │
│          └── 启用并行: 并发执行所有命中分支                   │
│  4. 收集所有分支执行结果                                     │
└────────────────────────────────────────────────────────────┘
```

**示例配置**：
```json
{
  "type": "branch",
  "multiMatch": true,
  "parallel": true,
  "branches": [
    { "id": "b1", "name": "VIP用户", "condition": { "field": "{{input.userLevel}}", "operator": "==", "value": "VIP" }, "nodeId": "vip_process" },
    { "id": "b2", "name": "普通用户", "condition": { "field": "{{input.userLevel}}", "operator": "==", "value": "normal" }, "nodeId": "normal_process" },
    { "id": "b3", "name": "异常处理", "condition": { "field": "{{input.status}}", "operator": "==", "value": "error" }, "nodeId": "error_handler" }
  ]
}
```

**执行引擎实现**：
```go
func (e *TreeExecutor) executeBranchNode(ctx context.Context, node *TreeNode, input map[string]interface{}) (*NodeResult, error) {
    config := node.Config.(BranchNodeConfig)
    
    // 1. 按顺序检查条件，收集命中分支
    var matchedBranches []*Branch
    for _, branch := range config.Branches {
        if e.evalCondition(branch.Condition, input) {
            matchedBranches = append(matchedBranches, branch)
            if !config.MultiMatch {
                break // 单次命中，退出
            }
        }
    }
    
    if len(matchedBranches) == 0 {
        return &NodeResult{Status: Success, Output: input}, nil // 无命中，原值传递
    }
    
    // 2. 执行命中的分支
    var results []*NodeResult
    if config.Parallel && config.MultiMatch && len(matchedBranches) > 1 {
        // 并行执行
        var wg sync.WaitGroup
        mu := sync.Mutex{}
        for _, branch := range matchedBranches {
            wg.Add(1)
            go func(b *Branch) {
                result, _ := e.executeBranch(ctx, b, input)
                mu.Lock()
                results = append(results, result)
                mu.Unlock()
                wg.Done()
            }(branch)
        }
        wg.Wait()
    } else {
        // 顺序执行
        for _, branch := range matchedBranches {
            result, err := e.executeBranch(ctx, branch, input)
            results = append(results, result)
            if err != nil && config.BranchFailStrategy == "abort" {
                return result, err
            }
        }
    }
    
    // 3. 合并结果
    return mergeBranchResults(results), nil
}
```

**分支结束点**：
- 每个分支节点的底部有一个隐式**结束点**
- 所有分支执行完成后，结果汇入结束点
- 分支无命中时：输入直接透传到结束点

```
┌─────────────────────────────────────────────────────────┐
│                   分支节点执行结果汇聚                   │
│                                                         │
│   分支A ──┐                                             │
│   分支B ──┼──► [分支结束点] ──► 后续节点                 │
│   分支C ──┘                                             │
│                                                         │
│   无命中 ──────────────► [分支结束点] ──► 后续节点       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**结果合并规则**：
```go
func mergeBranchResults(results []*NodeResult) *NodeResult {
    if len(results) == 0 {
        // 无命中，透传输入
        return &NodeResult{Status: Success, Output: nil}
    }
    
    // 多分支结果合并
    var mergedOutput []map[string]interface{}
    for _, r := range results {
        if r.Output != nil {
            if arr, ok := r.Output.([]map[string]interface{}); ok {
                mergedOutput = append(mergedOutput, arr...)
            } else {
                mergedOutput = append(mergedOutput, map[string]interface{}{
                    "branchOutput": r.Output,
                })
            }
        }
    }
    
    return &NodeResult{
        Status: Success,
        Output: map[string]interface{}{
            "results": mergedOutput,
        },
    }
}
```

#### 7.5.2 迭代节点

**功能**：对数组或字典的每个元素执行子流程。

**数据结构**：
```go
// LoopNodeConfig 迭代节点配置
type LoopNodeConfig struct {
    Source      string  `json:"source"`       // 迭代源字段，如 {{node_1.output.items}}
    Concurrency int     `json:"concurrency"`  // 并发数，0 为顺序执行
    MaxQueue   int     `json:"maxQueue"`     // 排队阈值，超过时排队等待
    SubFlowID   string  `json:"subFlowId"`   // 子流程 ID（每个元素的执行逻辑）
    InputMap    map[string]string `json:"inputMap"`  // 输入映射：element -> {{.item}}
    OutputMap   string  `json:"outputMap"`  // 输出映射：结果收集字段
}

// 迭代上下文
type LoopContext struct {
    Index   int                     // 当前索引
    Key     string                 // 字典 key（字典迭代时）
    Value   any                    // 当前元素值
    Total   int                   // 总元素数
}
```

**执行逻辑**：

```
┌────────────────────────────────────────────────────────────┐
│                    迭代节点执行流程                         │
│                                                            │
│  1. 获取迭代源数据（数组/字典）                              │
│  2. 遍历元素，准备执行上下文                                │
│     ├── 顺序执行: 依次执行每个元素                          │
│     └── 并发执行:                                           │
│          ├── 并发数 <= 阈值: 直接并发                        │
│          └── 并发数 > 阈值: 信号量控制，排队执行              │
│  3. 对每个元素调用子流程                                    │
│  4. 收集所有结果                                            │
│     ├── 数组: [result1, result2, ...]                      │
│     └── 字典: {key1: result1, key2: result2, ...}          │
└────────────────────────────────────────────────────────────┘
```

**示例配置**：
```json
{
  "type": "loop",
  "source": "{{node_1.output.users}}",
  "concurrency": 5,
  "maxQueue": 10,
  "subFlowId": "process_user_flow",
  "inputMap": {
    "userId": "{{.item.id}}",
    "userName": "{{.item.name}}"
  },
  "outputMap": "results"
}
```

**执行引擎实现**：
```go
func (e *TreeExecutor) executeLoopNode(ctx context.Context, node *TreeNode, input map[string]interface{}) (*NodeResult, error) {
    config := node.Config.(LoopNodeConfig)
    
    // 1. 获取迭代源
    source, err := ResolveMapping(config.Source, input)
    if err != nil {
        return nil, fmt.Errorf("resolve loop source: %w", err)
    }
    
    // 2. 转换为可迭代切片
    items, err := normalizeToSlice(source)
    if err != nil {
        return nil, fmt.Errorf("normalize loop source: %w", err)
    }
    
    // 3. 执行迭代
    var results []map[string]interface{}
    
    if config.Concurrency <= 1 {
        // 顺序执行
        for i, item := range items {
            ctx := &LoopContext{Index: i, Total: len(items), Value: item}
            result, err := e.executeLoopItem(ctx, node, input, item, config)
            results = append(results, result)
            if err != nil {
                return nil, err
            }
        }
    } else {
        // 并发执行（带信号量控制）
        semaphore := semaphore.NewWeighted(int64(config.Concurrency))
        queue := make(chan struct{}, config.MaxQueue)
        
        var wg sync.WaitGroup
        mu := sync.Mutex{}
        
        for i, item := range items {
            // 排队等待
            if config.MaxQueue > 0 {
                queue <- struct{}{}
                defer func() { <-queue }()
            }
            
            semaphore.Acquire(ctx, 1)
            wg.Add(1)
            
            go func(i int, item any) {
                defer semaphore.Release(1)
                defer wg.Done()
                
                ctx := &LoopContext{Index: i, Total: len(items), Value: item}
                result, _ := e.executeLoopItem(ctx, node, input, item, config)
                
                mu.Lock()
                results = append(results, result)
                mu.Unlock()
            }(i, item)
        }
        
        wg.Wait()
    }
    
    // 4. 返回结果
    output := map[string]interface{}{
        config.OutputMap: results,
    }
    return &NodeResult{Status: Success, Output: output}, nil
}
```

**并发控制细节**：
- `concurrency=0`：顺序执行
- `concurrency=1`：禁用并发，等同顺序
- `concurrency>1`：启用并发
- `maxQueue`：当待执行元素超过此值时，新元素需等待

**空数组处理**：
```
┌─────────────────────────────────────────────────────────┐
│                 迭代节点空数组处理                        │
│                                                         │
│  获取迭代源数据后:                                       │
│     ├── 有数据: 遍历执行每个元素                          │
│     └── 空数组:                                          │
│          └── 输出空数组，直接进入后续节点                 │
│                                                         │
│  实现:                                                  │
│  if len(items) == 0 {                                   │
│      return &NodeResult{                                 │
│          Status: Success,                                │
│          Output: map[string]interface{}{                 │
│              config.OutputMap: []interface{}{},         │
│          }, nil                                          │
│      }                                                  │
│  }                                                      │
└─────────────────────────────────────────────────────────┘
```

#### 7.5.3 树形流程完整示例

```
                    [Start]
                       │
                       ▼
                  [获取用户列表]
                       │
                       ▼
                  [迭代节点]
                   /    \
        ┌──────────┐    └──────────┐
        ▼                         ▼
   [分支节点]                [分支节点]
   (用户类型A)               (用户类型B)
        │                         │
        ▼                         ▼
   [处理A流程]             [处理B流程]
        │                         │
        └──────────┬──────────────┘
                   ▼
              [汇总结果]
```

### 7.6 RESTful 触发器的同步等待优化

对于 RESTful 触发器场景，调用者通常期望同步获取结果。采用**混合模式**：

```
┌──────────────────────────────────────────┐
│  RESTful 触发器请求处理流程                  │
│                                          │
│  1. 收到请求 → 创建 Run (pending)         │
│  2. 同步入队执行（不走异步 Worker）         │
│  3. 等待执行完成（带超时，默认 30s）        │
│  4. 超时内完成 → 直接返回结果              │
│  5. 超时 → 返回 202 + Run ID（异步模式）    │
└──────────────────────────────────────────┘
```

---

## 8. 数据存储设计

### 8.1 存储架构概览

平台采用**双存储架构**：
- **PostgreSQL**：持久化存储资源、工具、流程、触发器定义，以及运行记录、审计日志
- **etcd**：配置中心，存储发布到集群的运行时配置，Executor 通过 Watch 机制实时同步

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              存储架构                                         │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         Console (控制台)                             │   │
│  │                                                                     │   │
│  │  ┌──────────────────┐          ┌──────────────────┐                │   │
│  │  │   PostgreSQL     │          │      etcd        │                │   │
│  │  │  (定义存储)       │          │  (配置中心)       │                │   │
│  │  │                  │          │                  │                │   │
│  │  │  • 资源定义       │  发布    │  • 流程配置       │                │   │
│  │  │  • 工具定义       │ ───────► │  • 触发器配置     │                │   │
│  │  │  • 流程定义       │          │  • 资源配置       │                │   │
│  │  │  • 触发器定义     │          │  • 工具配置       │                │   │
│  │  │  • 运行记录       │          │  • 执行器心跳     │                │   │
│  │  │  • 审计日志       │          │                  │                │   │
│  │  └──────────────────┘          └────────┬─────────┘                │   │
│  └─────────────────────────────────────────┼──────────────────────────┘   │
│                                            │                               │
│                              Watch / Get   │                               │
│                                            ▼                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        Executor (执行器)                             │   │
│  │                                                                     │   │
│  │  ┌──────────────────┐          ┌──────────────────┐                │   │
│  │  │   本地内存缓存     │          │   流程执行引擎    │                │   │
│  │  │                  │          │                  │                │   │
│  │  │  • 流程定义       │◄────────►│  • 触发器管理    │                │   │
│  │  │  • 触发器配置     │          │  • 节点调度      │                │   │
│  │  │  • 资源配置       │          │  • 工具调用      │                │   │
│  │  └──────────────────┘          └──────────────────┘                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 8.2 数据库 ER 关系

```
┌──────────┐     ┌──────────────┐     ┌──────────┐
│  users   │────►│workspace_    │◄────│workspace_│
│          │     │members       │     │          │
└──────────┘     └──────┬───────┘     └────┬─────┘
                        │                  │
              ┌─────────┴──────────────────┘
              │ workspace_id
              │
     ┌────────┴────────┬─────────────┬─────────────┐
     ▼                 ▼             ▼             ▼
┌──────────┐    ┌──────────┐   ┌──────────┐  ┌──────────┐
│resources │    │  tools   │   │  flows   │  │ clusters │
│          │───►│          │◄──►│          │  │          │
└──────────┘    └────┬─────┘   └────┬─────┘  └────┬─────┘
                     │              │             │
                ┌────┴────┐    ┌────┴────┐   ┌────┴────┐
                ▼         │    ▼         │   ▼         │
          ┌──────────┐    │  ┌──────────┐│  ┌──────────┐
          │tool_     │    │  │flow_     ││  │cluster_  │
          │versions  │    │  │versions  ││  │triggers  │
          └──────────┘    │  └──────────┘│  └──────────┘
                         │              │
                ┌────────┴──────────────┘
                │ flow_tool_refs (N:M)
                ▼
          ┌──────────┐    ┌──────────┐
          │triggers  │───►│  runs    │
          └──────────┘    └────┬─────┘
                             │ 1:N
                             ▼
                       ┌──────────┐
                       │run_node_ │
                       │  logs    │
                       └──────────┘
```

### 8.2 核心表设计

> 以下表结构在开发者指南已有详细定义，此处补充索引策略和约束。

#### 索引策略

```sql
-- 工作空间隔离的核心索引
CREATE INDEX idx_resources_workspace ON resources(workspace_id);
CREATE INDEX idx_tools_workspace ON tools(workspace_id);
CREATE INDEX idx_flows_workspace ON flows(workspace_id);
CREATE INDEX idx_triggers_workspace ON triggers(workspace_id);

-- 关联查询索引
CREATE INDEX idx_tools_resource ON tools(resource_id);
CREATE INDEX idx_triggers_flow ON triggers(flow_id);
CREATE INDEX idx_runs_trigger ON runs(trigger_id);
CREATE INDEX idx_runs_flow ON runs(flow_id);
CREATE INDEX idx_run_logs_run ON run_node_logs(run_id);

-- 状态筛选索引
CREATE INDEX idx_tools_status ON tools(status) WHERE status != 'deprecated';
CREATE INDEX idx_triggers_status ON triggers(status);
CREATE INDEX idx_runs_status ON runs(status, created_at DESC);

-- 唯一性约束
CREATE UNIQUE INDEX ux_users_username ON users(username);
CREATE UNIQUE INDEX ux_users_email ON users(email);
CREATE UNIQUE INDEX ux_resources_name_workspace ON resources(name, workspace_id);
CREATE UNIQUE INDEX ux_tools_name_resource ON tools(name, resource_id);

-- 全文搜索（PostgreSQL GIN 索引）
CREATE INDEX idx_resources_name_search ON resources USING gin(to_tsvector('simple', name));
CREATE INDEX idx_tools_name_search ON tools USING gin(to_tsvector('simple', name));
CREATE INDEX idx_flows_name_search ON flows USING gin(to_tsvector('simple', name));
```

#### 数据加密

```go
// 敏感字段加密存储（AES-256-GCM）
// 适用于：resources.config 中的密码、API Key、Token
// 加密密钥通过环境变量注入，不落库

type Encryptor struct {
    key []byte // 从 CONFIG_ENCRYPTION_KEY 环境变量加载
}

func (e *Encryptor) Encrypt(plaintext string) (string, error)
func (e *Encryptor) Decrypt(ciphertext string) (string, error)
```

### 8.3 etcd 存储设计

**存储路径规范**：

```
/clusters/{cluster-id}/                    # 集群根节点
├── /metadata                              # 集群元数据
│   └── version: "1.0"                     # 配置格式版本
│
├── /flows/{flow-id}                       # 流程定义
│   └── data: {flow definition JSON}
│
├── /flows/{flow-id}/versions/{version}    # 流程历史版本
│   └── data: {flow definition JSON}
│
├── /triggers/{trigger-id}                 # 触发器配置
│   └── data: {trigger config JSON}
│
├── /resources/{resource-id}               # 资源配置
│   └── data: {resource config JSON}
│
├── /tools/{tool-id}                       # 工具定义
│   └── data: {tool definition JSON}
│
├── /tools/{tool-id}/versions/{version}    # 工具历史版本
│   └── data: {tool definition JSON}
│
└── /executors/{executor-id}               # 执行器心跳
    ├── /heartbeat: {timestamp}
    ├── /status: "active"
    └── /metadata: {ip, hostname, version}
```

**数据格式示例**：

```json
// /clusters/prod-001/flows/flow-123
data: {
  "id": "flow-123",
  "name": "用户注册流程",
  "version": 3,
  "canvas_data": { ... },
  "input_schema": { ... },
  "output_schema": { ... },
  "tool_refs": [
    {"tool_id": "tool-456", "version": 2}
  ],
  "published_at": "2026-04-01T10:00:00Z",
  "published_by": "user-789"
}

// /clusters/prod-001/triggers/trigger-456
data: {
  "id": "trigger-456",
  "name": "用户注册API",
  "type": "restful",
  "flow_id": "flow-123",
  "flow_version": 3,
  "config": {
    "method": "POST",
    "path": "/api/v1/users"
  },
  "input_mapping": [ ... ],
  "output_mapping": [ ... ]
}
```

**Watch 机制**：

```go
// Executor 监听配置变更
func (e *Executor) watchConfig(ctx context.Context, clusterID string) {
    prefix := fmt.Sprintf("/clusters/%s/", clusterID)
    watchChan := e.etcdClient.Watch(ctx, prefix, clientv3.WithPrefix())
    
    for watchResp := range watchChan {
        for _, event := range watchResp.Events {
            key := string(event.Kv.Key)
            switch event.Type {
            case clientv3.EventTypePut:
                e.handleConfigUpdate(key, event.Kv.Value)
            case clientv3.EventTypeDelete:
                e.handleConfigDelete(key)
            }
        }
    }
}
```

**租约机制（心跳）**：

```go
// 执行器启动时注册心跳
func (e *Executor) registerHeartbeat(ctx context.Context) {
    // 创建 10 秒租约（心跳频率 5s，超时阈值 10s）
    lease, _ := e.etcdClient.Grant(ctx, 10)
    
    // 写入执行器信息，绑定租约
    key := fmt.Sprintf("/clusters/%s/executors/%s", e.clusterID, e.executorID)
    e.etcdClient.Put(ctx, key, executorInfo, clientv3.WithLease(lease.ID))
    
    // 保持租约
    keepAliveChan, _ := e.etcdClient.KeepAlive(ctx, lease.ID)
    go func() {
        for range keepAliveChan {
            // 租约续期成功
        }
    }()
}
```

**执行器状态上报**：

```go
// ExecutorStatus 执行器完整状态信息
type ExecutorStatus struct {
    // 基础信息
    ExecutorID    string    `json:"executor_id"`     // 执行器唯一标识 (UUID)
    Hostname      string    `json:"hostname"`        // 主机名
    IP            string    `json:"ip"`              // IP 地址
    Version       string    `json:"version"`         // 软件版本
    BuildTags     string    `json:"build_tags"`      // 编译标签（包含的插件）
    
    // 运行时状态
    Status        string    `json:"status"`          // active / draining / offline
    StartedAt     time.Time `json:"started_at"`      // 启动时间
    LastHeartbeat time.Time `json:"last_heartbeat"`  // 最后心跳时间
    UptimeSeconds int64     `json:"uptime_seconds"`  // 运行时长
    
    // 资源使用
    CPUUsage      float64   `json:"cpu_usage"`       // CPU 使用率 (%)
    MemoryUsage   int64     `json:"memory_usage"`    // 内存使用 (bytes)
    MemoryTotal   int64     `json:"memory_total"`    // 总内存 (bytes)
    Goroutines    int       `json:"goroutines"`      // Goroutine 数量
    
    // 负载统计
    ActiveRuns    int       `json:"active_runs"`     // 当前执行中的流程数
    QueuedRuns    int       `json:"queued_runs"`     // 队列等待数
    TotalRuns     int64     `json:"total_runs"`      // 累计执行数
    
    // 配置统计
    LoadedFlows     int `json:"loaded_flows"`      // 已加载流程数
    LoadedTriggers  int `json:"loaded_triggers"`   // 已加载触发器数
    LoadedResources int `json:"loaded_resources"`  // 已加载资源数
    LoadedTools     int `json:"loaded_tools"`      // 已加载工具数
}

// 状态上报流程
func (e *Executor) startStatusReporter(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)  // 每 5 秒上报一次
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            status := e.collectStatus()  // 收集当前状态
            e.reportStatus(ctx, status)  // 写入 etcd
        }
    }
}

// 写入 etcd 的状态路径
// /clusters/{cluster-id}/executors/{executor-id}/status
```

**etcd 执行器状态存储结构**：

```
/clusters/{cluster-id}/
└── /executors/
    └── /{executor-id}/
        ├── /info           # 基础信息（静态，启动时写入）
        ├── /status         # 运行时状态（动态，定时更新）
        ├── /heartbeat      # 心跳时间戳（租约绑定）
        └── /metrics        # 指标数据（可选，用于监控）
```

### 8.4 数据库迁移策略

| 策略 | 说明 |
|------|------|
| **迁移工具** | golang-migrate |
| **迁移方向** | 仅 forward，不回滚（生产安全） |
| **命名规则** | `NNN_description.sql`（数字前缀 + 下划线描述） |
| **执行时机** | 服务启动时自动检查并执行 |
| **锁机制** | 使用 `advisory_lock` 防止多实例重复执行 |

---

## 9. 缓存策略

### 9.1 Redis 使用场景

| 用途 | Key 模式 | TTL | 说明 |
|------|---------|-----|------|
| **JWT 黑名单** | `auth:token:blacklist:{jti}` | Token 剩余有效期 | 登出时将令牌加入黑名单 |
| **运行状态缓存** | `run:status:{run_id}` | 1h | 流程执行中的实时状态，减少 DB 查询 |
| **执行队列** | `run:queue:pending` | — | Redis List，BRPOP 消费 |
| **资源健康状态** | `resource:health:{resource_id}` | 30s | 定时探活结果缓存 |
| **插件 Schema 缓存** | `plugin:schema:{plugin_name}` | 24h | 插件配置 Schema |
| **工作空间成员缓存** | `ws:members:{workspace_id}` | 10min | RBAC 权限校验加速 |
| **分布式锁** | `lock:trigger:timer:{trigger_id}:{execution_time}` | TTL=单次触发时长 | Timer 触发器防重复执行，锁粒度为一次触发 |

### 9.2 缓存一致性

```
┌────────────────────────────────────────────────────────────────┐
│                    缓存一致性策略                                │
│                                                                │
│  写操作流程：                                                   │
│  1. 更新数据库                                                  │
│  2. 删除相关缓存（Cache-Aside 模式）                             │
│  3. 发布领域事件                                                │
│                                                                │
│  读操作流程：                                                   │
│  1. 查询缓存                                                    │
│  2. 缓存命中 → 返回                                            │
│  3. 缓存未命中 → 查询数据库 → 写入缓存 → 返回                   │
│                                                                │
│  一致性保证：                                                   │
│  - 缓存 TTL 设置合理的过期时间作为兜底                            │
│  - 关键数据（权限、配置）使用短 TTL + 主动失效                    │
│  - 非关键数据（统计、健康状态）接受短暂不一致                     │
└────────────────────────────────────────────────────────────────┘
```

---

---

## 10. API 设计

### 11.1 API 约定

| 规范 | 说明 |
|------|------|
| **版本前缀** | `/api/v1/` |
| **响应格式** | `{ code: number, data: T, message: string }` |
| **认证方式** | `Authorization: Bearer {jwt_access_token}` |
| **分页参数** | `?page=1&page_size=20` |
| **排序** | `?sort_by=created_at&sort_order=desc` |
| **过滤** | `?status=active&type=http` |
| **搜索** | `?q=keyword` |
| **错误码** | 业务错误码 5 位数（如 40001 = 资源不存在） |
| **HTTP 状态码** | 200 成功 / 201 创建 / 400 参数错误 / 401 未认证 / 403 无权限 / 404 不存在 / 500 内部错误 |

### 11.2 部署面板 API

**执行器状态查询**：

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/clusters/:id/executors` | 获取集群所有执行器状态列表 | 平台管理员 |
| GET | `/api/v1/clusters/:id/executors/:executorId` | 获取单个执行器详情 | 平台管理员 |
| GET | `/api/v1/clusters/:id/dashboard` | 获取集群部署面板汇总数据 | 平台管理员 |

**部署面板数据结构**：

```go
// DeploymentDashboard 部署面板数据
type DeploymentDashboard struct {
    // 集群基本信息
    ClusterID       string    `json:"cluster_id"`
    ClusterName     string    `json:"cluster_name"`
    Environment     string    `json:"environment"`
    Status          string    `json:"status"`          // active / inactive / error
    
    // 执行器汇总
    ExecutorSummary struct {
        Total       int `json:"total"`       // 总执行器数
        Active      int `json:"active"`      // 在线数
        Draining    int `json:"draining"`    // 排空中的数
        Offline     int `json:"offline"`     // 离线数
    } `json:"executor_summary"`
    
    // 实时负载
    LoadSummary struct {
        ActiveRuns  int   `json:"active_runs"`   // 当前执行中流程数
        QueuedRuns  int   `json:"queued_runs"`   // 队列等待数
        TotalRuns   int64 `json:"total_runs"`    // 累计执行数
        AvgCPU      float64 `json:"avg_cpu"`     // 平均 CPU 使用率
        AvgMemory   float64 `json:"avg_memory"`  // 平均内存使用率
    } `json:"load_summary"`
    
    // 配置统计
    ConfigSummary struct {
        LoadedFlows     int `json:"loaded_flows"`
        LoadedTriggers  int `json:"loaded_triggers"`
        LoadedResources int `json:"loaded_resources"`
        LoadedTools     int `json:"loaded_tools"`
    } `json:"config_summary"`
    
    // 执行器列表
    Executors []ExecutorStatus `json:"executors"`
    
    // 最近事件
    RecentEvents []DeploymentEvent `json:"recent_events"`
}

// DeploymentEvent 部署事件
type DeploymentEvent struct {
    Timestamp   time.Time `json:"timestamp"`
    Type        string    `json:"type"`        // executor_joined / executor_left / config_updated / run_failed
    ExecutorID  string    `json:"executor_id,omitempty"`
    Message     string    `json:"message"`
}
```

**Console 读取 etcd 执行器状态**：

```go
// Console 通过 etcd 查询执行器状态
func (c *ClusterService) GetExecutors(ctx context.Context, clusterID string) ([]ExecutorStatus, error) {
    prefix := fmt.Sprintf("/clusters/%s/executors/", clusterID)
    
    // 获取所有执行器键
    resp, err := c.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }
    
    var executors []ExecutorStatus
    for _, kv := range resp.Kvs {
        // 解析路径提取 executor-id
        // 解析 value 获取状态数据
        var status ExecutorStatus
        if err := json.Unmarshal(kv.Value, &status); err != nil {
            continue
        }
        executors = append(executors, status)
    }
    
    return executors, nil
}
```

### 11.3 核心 API 端点

> 详细端点已在开发者指南中定义，此处按上下文分组整理。

#### 身份认证

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | 登录 | 公开 |
| POST | `/api/v1/auth/logout` | 登出 | 已认证 |
| POST | `/api/v1/auth/refresh` | 刷新 Token | 已认证 |
| GET | `/api/v1/auth/profile` | 当前用户信息 | 已认证 |
| PUT | `/api/v1/auth/profile` | 更新个人信息 | 已认证 |
| PUT | `/api/v1/auth/password` | 修改密码 | 已认证 |
| POST | `/api/v1/auth/mfa/enable` | 启用 MFA | 已认证 |
| POST | `/api/v1/auth/mfa/verify` | MFA 验证 | 公开（登录流程中） |

#### 工作空间

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces` | 我的工作空间列表 | 已认证 |
| POST | `/api/v1/workspaces` | 创建工作空间 | 已认证 |
| GET | `/api/v1/workspaces/:id` | 工作空间详情 | 成员 |
| PUT | `/api/v1/workspaces/:id` | 更新工作空间 | 负责人/管理员 |
| GET | `/api/v1/workspaces/:id/members` | 成员列表 | 成员 |
| POST | `/api/v1/workspaces/:id/members` | 邀请成员 | 负责人/管理员 |
| DELETE | `/api/v1/workspaces/:id/members/:userId` | 移除成员 | 负责人/管理员 |
| PUT | `/api/v1/workspaces/:id/members/:userId/role` | 修改角色 | 负责人/管理员 |

#### 资源管理

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces/:wsId/resources` | 资源列表 | 成员 |
| POST | `/api/v1/workspaces/:wsId/resources` | 创建资源 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/resources/:id` | 资源详情 | 成员 |
| PUT | `/api/v1/workspaces/:wsId/resources/:id` | 更新资源 | 开发者+ |
| DELETE | `/api/v1/workspaces/:wsId/resources/:id` | 删除资源 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/resources/:id/test` | 测试连接 | 开发者+ |

#### 工具管理

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces/:wsId/tools` | 工具列表 | 成员 |
| POST | `/api/v1/workspaces/:wsId/tools` | 创建工具 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/tools/:id` | 工具详情 | 成员 |
| PUT | `/api/v1/workspaces/:wsId/tools/:id` | 更新工具 | 开发者+ |
| DELETE | `/api/v1/workspaces/:wsId/tools/:id` | 删除工具 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/tools/:id/publish` | 发布工具 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/tools/:id/unpublish` | 取消发布 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/tools/:id/test` | 测试工具 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/tools/import` | 批量导入 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/tools/:id/versions` | 版本历史 | 成员 |
| GET | `/api/v1/workspaces/:wsId/tools/:id/impact` | 影响分析 | 成员 |

#### 流程管理

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces/:wsId/flows` | 流程列表 | 成员 |
| POST | `/api/v1/workspaces/:wsId/flows` | 创建流程 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/flows/:id` | 流程详情 | 成员 |
| PUT | `/api/v1/workspaces/:wsId/flows/:id` | 更新流程 | 开发者+ |
| DELETE | `/api/v1/workspaces/:wsId/flows/:id` | 删除流程 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/flows/:id/publish` | 发布流程 | 开发者+ |
| POST | `/api/v1/workspaces/:wsId/flows/:id/test` | 测试运行 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/flows/:id/versions` | 版本历史 | 成员 |

#### 触发器管理

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces/:wsId/triggers` | 触发器列表 | 成员 |
| POST | `/api/v1/workspaces/:wsId/triggers` | 创建触发器 | 开发者+ |
| GET | `/api/v1/workspaces/:wsId/triggers/:id` | 触发器详情 | 成员 |
| PUT | `/api/v1/workspaces/:wsId/triggers/:id` | 更新触发器 | 开发者+ |
| DELETE | `/api/v1/workspaces/:wsId/triggers/:id` | 删除触发器 | 开发者+ |
| PUT | `/api/v1/workspaces/:wsId/triggers/:id/status` | 启停触发器 | 开发者+ |

#### 运行记录

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/runs` | 运行记录列表 | 成员 |
| GET | `/api/v1/runs/:id` | 运行详情 | 成员 |
| POST | `/api/v1/runs/:id/retry` | 重新运行 | 开发者+ |

#### 平台管理

| 方法 | 端点 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/admin/users` | 用户列表 | 平台管理员 |
| POST | `/api/v1/admin/users/invite` | 邀请用户 | 平台管理员 |
| PUT | `/api/v1/admin/users/:id/status` | 启禁用用户 | 平台管理员 |
| GET | `/api/v1/admin/roles` | 角色列表 | 平台管理员 |
| POST | `/api/v1/admin/roles` | 创建角色 | 平台管理员 |
| GET | `/api/v1/admin/plugins` | 插件列表 | 平台管理员 |
| GET | `/api/v1/admin/system/health` | 系统健康 | 平台管理员 |

### 11.3 RESTful 触发器生成的端点

当创建 RESTful 类型触发器时，自动生成对应的调用端点：

```
POST /api/v1/trigger/{trigger_id}/invoke
```

该端点由触发器插件动态注册到路由器，处理输入映射、流程执行和输出映射。

### 10.5 WebSocket 端点（迭代阶段）

```
WS /api/v1/ws?token={jwt}

消息格式：
→ 订阅运行状态: { "type": "subscribe:run", "runId": "xxx" }
← 状态更新: { "type": "run:status", "runId": "xxx", "status": "running" }
← 节点更新: { "type": "run:node", "runId": "xxx", "nodeId": "node_1", "status": "success" }
```

---

## 11. 认证与授权架构

### 12.1 JWT 认证流程

```
┌────────────────────────────────────────────────────────────────┐
│                    JWT 认证架构                                   │
│                                                                │
│  登录流程：                                                     │
│  1. 用户提交 username + password                                │
│  2. 服务端验证 → 生成 JWT Access Token (15min) + Refresh Token (7d) │
│  3. Refresh Token 存储到 Redis: refresh:{user_id}:{jti}        │
│  4. 返回两个 Token 给前端                                       │
│                                                                │
│  请求认证：                                                     │
│  1. 前端在 Authorization Header 携带 Access Token               │
│  2. Auth Middleware 解析 JWT → 验证签名 → 检查黑名单             │
│  3. 注入用户信息到 Context（user_id, roles）                    │
│  4. RBAC Middleware 根据用户角色 + 工作空间校验权限               │
│                                                                │
│  Token 刷新：                                                   │
│  1. Access Token 过期 → 前端用 Refresh Token 调用 /auth/refresh  │
│  2. 服务端验证 Refresh Token (Redis 中存在 + 未过期)             │
│  3. 生成新的 Access Token                                       │
│  4. 可选：轮换 Refresh Token（提升安全性）                       │
│                                                                │
│  登出：                                                         │
│  1. 将当前 Access Token 的 jti 加入 Redis 黑名单                │
│  2. 删除 Redis 中的 Refresh Token                               │
│  3. 前端清除本地存储                                             │
└────────────────────────────────────────────────────────────────┘
```

### 12.2 JWT Payload

```json
{
  "sub": "user-uuid",
  "username": "zhangsan",
  "email": "zhangsan@company.com",
  "roles": {
    "global": "platform_admin",
    "workspaces": {
      "ws-uuid-1": "owner",
      "ws-uuid-2": "developer"
    }
  },
  "jti": "token-uuid",
  "iat": 1712000000,
  "exp": 1712000900
}
```

> 注意：JWT 中携带角色信息以避免每次请求都查库。角色变更时通过短 Token 有效期（15min）保证时效性。

### 12.3 RBAC 权限校验流程

```go
// RBAC Middleware 伪代码
func RBACMiddleware(requiredPermission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 JWT Context 获取用户信息
        user := c.Get("user")

        // 2. 获取当前工作空间 ID（从 URL path 或 header）
        wsId := c.Param("wsId")

        // 3. 查询用户在该工作空间的角色（优先用 JWT 中的缓存，降级查 Redis/DB）
        role := getUserRole(user.ID, wsId)

        // 4. 校验角色是否具有所需权限
        if !role.HasPermission(requiredPermission) {
            c.AbortWithStatusJSON(403, "无权限")
            return
        }

        c.Next()
    }
}
```

### 11.4 权限矩阵（数据模型）

```go
// Permission 权限定义
type Permission struct {
    Module  string // resource / tool / flow / trigger / run / workspace / user / admin
    Action  string // view / create / edit / delete / publish / manage
    Scope   string // own_workspace / all
}

// RolePermissionMap 预设角色权限映射
var RolePermissionMap = map[string][]Permission{
    "viewer": {
        {Module: "resource", Action: "view", Scope: "own_workspace"},
        {Module: "tool", Action: "view", Scope: "own_workspace"},
        {Module: "flow", Action: "view", Scope: "own_workspace"},
        {Module: "trigger", Action: "view", Scope: "own_workspace"},
        {Module: "run", Action: "view", Scope: "own_workspace"},
    },
    "developer": {
        // 继承 viewer 所有权限
        {Module: "resource", Action: "create", Scope: "own_workspace"},
        {Module: "resource", Action: "edit", Scope: "own_workspace"},
        {Module: "resource", Action: "delete", Scope: "own_workspace"},
        {Module: "tool", Action: "create", Scope: "own_workspace"},
        {Module: "tool", Action: "edit", Scope: "own_workspace"},
        {Module: "tool", Action: "delete", Scope: "own_workspace"},
        {Module: "tool", Action: "publish", Scope: "own_workspace"},
        {Module: "flow", Action: "create", Scope: "own_workspace"},
        {Module: "flow", Action: "edit", Scope: "own_workspace"},
        {Module: "flow", Action: "delete", Scope: "own_workspace"},
        {Module: "flow", Action: "publish", Scope: "own_workspace"},
        {Module: "trigger", Action: "create", Scope: "own_workspace"},
        {Module: "trigger", Action: "edit", Scope: "own_workspace"},
        {Module: "trigger", Action: "delete", Scope: "own_workspace"},
        {Module: "trigger", Action: "manage", Scope: "own_workspace"}, // 启停
        {Module: "run", Action: "retry", Scope: "own_workspace"},
    },
    "owner": {
        // 继承 developer 所有权限
        {Module: "workspace", Action: "manage", Scope: "own_workspace"}, // 成员管理
    },
    "platform_admin": {
        // 所有权限，Scope = all
        {Module: "user", Action: "manage", Scope: "all"},
        {Module: "admin", Action: "manage", Scope: "all"},
    },
}
```

---

## 12. 插件架构

### 13.1 设计理念

本平台的插件采用**编译时集成（Build-time Plugin）**模式，而非运行时动态加载。所有插件源码位于仓库 `plugins/` 目录下，通过 Go build tags 在编译阶段选择性打包进二进制。不同客户的合同需求不同，交付的版本包含不同的插件组合。

**为什么不用 Go Plugin（`plugin` 包）？**
- Go Plugin 要求插件和主程序使用**完全相同的 Go 版本和依赖版本**编译，否则 panic
- 跨平台支持差（Windows 上 Go Plugin 不可用）
- CGO 依赖复杂，容器化部署受限
- 调试困难，运行时崩溃难以追踪

**编译时插件的优势：**
- 编译期类型检查，零运行时兼容问题
- 进程内直接调用，零加载开销
- 单一二进制部署，无需额外文件
- 交叉编译简单，`GOOS=linux GOARCH=amd64 go build` 即可

### 13.2 插件接口定义

```go
// infra/plugin/types.go

// PluginType 插件类型
type PluginType string

const (
    PluginTypeResource PluginType = "resource"
    PluginTypeTrigger  PluginType = "trigger"
)

// ResourcePlugin 资源插件接口
// 资源插件负责管理外部服务的连接和工具执行
type ResourcePlugin interface {
    PluginMeta() PluginMeta
    ConfigSchema() JSONSchema       // 资源配置表单 Schema（前端动态渲染）
    TestConnection(ctx context.Context, config map[string]interface{}) error
    ExecuteTool(ctx context.Context, config, toolConfig map[string]interface{}, input interface{}) (*ToolResult, error)
    ExtractTools(ctx context.Context, config map[string]interface{}) ([]ToolDefinition, error) // 可选：批量导入工具
}

// TriggerPlugin 触发器插件接口
// 触发器插件负责监听外部事件并触发流程执行
type TriggerPlugin interface {
    PluginMeta() PluginMeta
    ConfigSchema() JSONSchema       // 触发器配置表单 Schema
    InputSchema() JSONSchema        // 输入参数 Schema
    OutputSchema() JSONSchema       // 输出结果 Schema
    Start(ctx context.Context, config map[string]interface{}, handler TriggerHandler) error
    Stop() error
}

// TriggerHandler 触发器回调（当触发器收到事件时调用）
type TriggerHandler func(ctx context.Context, input map[string]interface{}) (*TriggerResult, error)

// PluginMeta 插件元数据
type PluginMeta struct {
    Name        string    `json:"name"`        // 唯一标识，如 "http", "grpc", "timer"
    Type        PluginType `json:"type"`        // "resource" or "trigger"
    Version     string    `json:"version"`     // 语义化版本
    Description string    `json:"description"` // 中文描述
    Author      string    `json:"author"`      // 开发者/团队
    BuildTag    string    `json:"buildTag"`    // 对应的 Go build tag
}

// JSONSchema 配置 Schema（用于前端动态渲染表单）
type JSONSchema = map[string]interface{}

// ToolDefinition 从资源提取的工具定义（批量导入用）
type ToolDefinition struct {
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`         // http_method / grpc_method / sql_query
    Method       string                 `json:"method"`
    Path         string                 `json:"path"`
    InputSchema  JSONSchema             `json:"inputSchema"`
    OutputSchema JSONSchema             `json:"outputSchema"`
    Description  string                 `json:"description"`
    Config       map[string]interface{} `json:"config"`
}

// ToolResult 工具执行结果
type ToolResult struct {
    StatusCode int               `json:"statusCode"`
    Data       interface{}       `json:"data"`
    Headers    map[string]string `json:"headers,omitempty"`
    Duration   time.Duration     `json:"duration"`
    Error      string            `json:"error,omitempty"`
}

// TriggerResult 触发器处理结果
type TriggerResult struct {
    Success bool                   `json:"success"`
    Data    map[string]interface{} `json:"data,omitempty"`
    Error   string                 `json:"error,omitempty"`
}
```

### 13.3 插件注册表

```go
// infra/plugin/registry.go

// PluginRegistry 插件注册表
// 运行时查询已编译插件的能力，不负责加载
type PluginRegistry struct {
    resources map[string]ResourcePlugin  // name -> ResourcePlugin
    triggers  map[string]TriggerPlugin   // name -> TriggerPlugin
    mu        sync.RWMutex
}

func NewPluginRegistry() *PluginRegistry

// RegisterResource 注册资源插件（由插件 init() 调用）
func (r *PluginRegistry) RegisterResource(plugin ResourcePlugin) error

// RegisterTrigger 注册触发器插件（由插件 init() 调用）
func (r *PluginRegistry) RegisterTrigger(plugin TriggerPlugin) error

// GetResource 获取资源插件
func (r *PluginRegistry) GetResource(name string) (ResourcePlugin, error)

// GetTrigger 获取触发器插件
func (r *PluginRegistry) GetTrigger(name string) (TriggerPlugin, error)

// ListResources 列出所有已编译的资源插件
func (r *PluginRegistry) ListResources() []PluginMeta

// ListTriggers 列出所有已编译的触发器插件
func (r *PluginRegistry) ListTriggers() []PluginMeta

// GetAllSchemas 获取所有插件的 Schema（前端渲染配置表单用）
func (r *PluginRegistry) GetAllSchemas() map[string]PluginSchemaInfo

// HasPlugin 检查指定插件是否已编译
func (r *PluginRegistry) HasPlugin(pluginType PluginType, name string) bool
```

### 12.4 插件注册机制（Build Tag + init()）

每个插件通过 Go build tag + `init()` 函数实现自动注册：

```go
// plugins/resource/http/plugin.go
//go:build http

package http

import "backend/internal/infra/plugin"

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterResource(&HTTPPlugin{})
}

type HTTPPlugin struct{}

func (p *HTTPPlugin) PluginMeta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "http",
        Type:        plugin.PluginTypeResource,
        Version:     "1.0.0",
        Description: "HTTP/HTTPS 服务资源",
        BuildTag:    "http",
    }
}

// ... ConfigSchema(), TestConnection(), ExecuteTool() 等实现
```

```go
// plugins/trigger/timer/plugin.go
//go:build timer

package timer

import "backend/internal/infra/plugin"

func init() {
    registry := plugin.GlobalRegistry()
    registry.RegisterTrigger(&TimerPlugin{})
}
```

**关键点**：`init()` 在 Go 程序启动时自动执行。只有被 build tag 选中的文件会参与编译，未被选中的插件文件完全不会出现在二进制中。

### 13.5 构建配置与 CI/CD

#### plugins.yaml 声明式构建配置

```yaml
# plugins/plugins.yaml
# 声明当前构建包含的插件列表
# CI/CD 流水线读取此文件生成 go build -tags 参数

version: "1.0"
build_tags:
  - http        # HTTP 资源插件
  - grpc        # gRPC 资源插件
  - postgres    # PostgreSQL 资源插件
  - timer       # 定时器触发器
  - restful     # RESTful 触发器

# 以下插件未包含在此构建中：
# - mysql       # MySQL 资源插件（需要时添加）
# - rabbitmq    # RabbitMQ 触发器（需要时添加）
# - kafka       # Kafka 触发器（需要时添加）
```

#### Makefile 构建目标

```makefile
# Makefile

# 从 plugins.yaml 提取 build tags
BUILD_TAGS := $(shell yq '.build_tags | join(",")' plugins/plugins.yaml)
VERSION := $(shell cat version.txt)

.PHONY: build
build:
	go build -tags "$(BUILD_TAGS)" \
		-ldflags "-X main.version=$(VERSION) -X main.buildTags=$(BUILD_TAGS)" \
		-o bin/api-orchestrator ./cmd/server/

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -tags "$(BUILD_TAGS)" \
		-ldflags "-X main.version=$(VERSION)" \
		-o bin/api-orchestrator-linux-amd64 ./cmd/server/

# 为特定客户构建定制版本
.PHONY: build-customer
build-customer:
	@echo "Building for customer: $(CUSTOMER)"
	@cp configs/customers/$(CUSTOMER)/plugins.yaml plugins/plugins.yaml
	$(MAKE) build
```

#### 客户配置目录

```
configs/
└── customers/
    ├── acme-corp/
    │   └── plugins.yaml       # ACME 公司：HTTP + gRPC + Timer + RESTful
    ├── big-bank/
    │   └── plugins.yaml       # 大银行：HTTP + PostgreSQL + RabbitMQ
    └── startup-io/
        └── plugins.yaml       # 初创公司：HTTP + Timer + RESTful（精简版）
```

### 13.6 服务启动时的插件发现

```
服务启动
    │
    ▼
1. main.go 初始化 PluginRegistry（空）
    │
    ▼
2. import 被选中的插件包 → 触发各插件的 init()
    │  （未被 build tag 选中的插件不会被 import，自然不会注册）
    │
    ▼
3. PluginRegistry.ListAll() → 打印已注册插件清单到日志
    │  例: [INFO] Registered plugins: resource=[http,grpc,postgres] trigger=[timer,restful]
    │
    ▼
4. 暴露 /api/v1/admin/plugins 端点
    │  → 返回已注册插件列表 + 各插件的 ConfigSchema
    │
    ▼
5. 前端获取插件 Schema → 动态渲染资源配置表单、触发器配置表单
    │  例: 只有 http 和 grpc 插件 → 资源创建页面只显示这两种类型
    │
    ▼
6. 启动触发器插件（TriggerPlugin.Start()）
    │  遍历数据库中 status=active 的触发器 → 按 type 查找已注册的 TriggerPlugin
    │  未注册的插件类型 → 日志告警 + 标记触发器为 error 状态
```

### 13.7 前端适配

前端通过 `/api/v1/admin/plugins` 获取当前版本的插件能力列表，据此动态控制 UI：

```typescript
// 前端根据插件能力动态控制 UI

// 1. 资源创建 — 只显示已编译的资源类型
const resourcePlugins = await api.get('/admin/plugins', { type: 'resource' });
// 响应: { plugins: [{ name: 'http', schema: {...} }, { name: 'grpc', schema: {...} }] }
// → 资源创建表单的「类型」下拉框只显示 http 和 grpc

// 2. 触发器创建 — 只显示已编译的触发器类型
const triggerPlugins = await api.get('/admin/plugins', { type: 'trigger' });
// → 触发器创建表单的「类型」下拉框只显示 timer 和 restful

// 3. 工具批量导入 — 根据资源类型匹配可用插件
// 例: 资源类型为 http → 显示「从 OpenAPI 导入」按钮
//     资源类型为 grpc → 显示「从 Proto 导入」按钮

// 4. 系统信息页 — 显示当前版本的插件能力
// 「当前版本包含：HTTP 资源、gRPC 资源、PostgreSQL 资源、定时器触发器、RESTful 触发器」
```

### 13.8 新增插件开发流程

```
1. 在 plugins/resource/ 或 plugins/trigger/ 下创建新目录
    │
    ▼
2. 实现 ResourcePlugin 或 TriggerPlugin 接口
    │  ├── plugin.go（//go:build tag + init() 注册）
    │  ├── schema.json（配置 Schema）
    │  └── executor.go / scheduler.go（核心逻辑）
    │
    ▼
3. 在客户的 plugins.yaml 中添加对应的 build tag
    │
    ▼
4. 编译验证 → 测试 → 交付
```

---

## 13. 前端架构

### 13.1 技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| 框架 | React 18+ TypeScript | 企业级 SPA |
| 构建 | Vite 5+ | 快速 HMR |
| 样式 | Tailwind CSS + CSS Modules | 全局 + 组件级隔离 |
| UI 组件 | shadcn/ui (Radix UI) | 可定制组件库 |
| 状态管理 | Zustand (客户端) + TanStack Query (服务端) | 轻量高效 |
| 路由 | React Router v6 | 声明式路由 |
| 画布 | @xyflow/react (React Flow) | 流程编排画布 |
| 代码编辑 | Monaco Editor | JSON/SQL 编辑器 |
| 表单 | React Hook Form + Zod | 表单 + Schema 校验 |
| HTTP | Axios + TanStack Query | 请求 + 缓存 |
| 图标 | Lucide React | SVG 图标库 |
| 国际化 | i18next (预留) | MVP 仅中文 |
| 主题 | next-themes | 深色/浅色切换 |

### 14.2 前端分层

```
┌────────────────────────────────────────────────────────────────┐
│                      前端分层架构                                 │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │                    Pages (页面层)                          │ │
│  │  路由页面组件，负责组合 Layout + Feature Components         │ │
│  └──────────────────────┬───────────────────────────────────┘ │
│                         │                                      │
│  ┌──────────────────────┴───────────────────────────────────┐ │
│  │              Feature Components (功能组件层)                │ │
│  │  resource/  tool/  flow/  trigger/  run/  auth/          │ │
│  │  面向业务场景的组件组合                                     │ │
│  └──────────────────────┬───────────────────────────────────┘ │
│                         │                                      │
│  ┌──────────────────────┴───────────────────────────────────┐ │
│  │                  Shared Components (共享组件层)             │ │
│  │  ui/ (shadcn/ui)  layout/  common/                       │ │
│  │  与业务无关的可复用组件                                      │ │
│  └──────────────────────┬───────────────────────────────────┘ │
│                         │                                      │
│  ┌──────────────────────┴───────────────────────────────────┐ │
│  │                    Hooks + Stores                         │ │
│  │  useAuth  useResource  useTool  useFlow  useRun          │ │
│  │  Zustand stores: workspaceStore, flowEditorStore          │ │
│  └──────────────────────┬───────────────────────────────────┘ │
│                         │                                      │
│  ┌──────────────────────┴───────────────────────────────────┐ │
│  │                      lib / API                            │ │
│  │  apiClient (Axios instance)  api/ (按模块分组)            │ │
│  │  utils/  constants/  types/                              │ │
│  └──────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
```

### 14.3 核心页面状态管理

```
┌────────────────────────────────────────────────────────────────┐
│                    状态管理策略                                   │
│                                                                │
│  TanStack Query (服务端状态):                                    │
│  ├── useQuery: 资源列表、工具列表、流程列表、运行记录...           │
│  ├── useMutation: 创建/更新/删除操作                             │
│  ├── 自动缓存、后台刷新、乐观更新                                │
│  └── queryClient.invalidateQueries() 联动刷新                   │
│                                                                │
│  Zustand (客户端状态):                                           │
│  ├── workspaceStore: 当前工作空间、用户角色                      │
│  ├── flowEditorStore: 画布节点、连线、选中状态、Undo/Redo       │
│  ├── triggerFormStore: 多步骤表单状态                           │
│  └── uiStore: 侧边栏、面板折叠、主题                             │
│                                                                │
│  画布状态特殊处理:                                               │
│  ├── React Flow 内部管理节点位置、连线                            │
│  ├── Zustand 管理选中、配置、保存状态                             │
│  ├── 自动保存: dirty 变更 2s 后 debounce 保存                    │
│  └── Undo/Redo: Zustand temporal middleware                    │
└────────────────────────────────────────────────────────────────┘
```

### 14.4 编排画布组件架构

```
FlowCanvasPage
├── FlowToolbar (顶部工具栏)
│   ├── 返回按钮
│   ├── 流程标题 (可编辑)
│   ├── 版本标签
│   ├── 发布状态
│   ├── 保存状态指示
│   ├── 保存按钮 (Ctrl+S)
│   ├── 运行按钮 (Ctrl+Enter)
│   └── 发布按钮
│
├── MainContent (Flex 布局)
│   ├── ToolPanel (左侧, 240px, 可折叠)
│   │   ├── 搜索框
│   │   ├── 资源类型筛选
│   │   └── 工具列表 (DragSource)
│   │
│   ├── ReactFlowCanvas (中央)
│   │   ├── Background (点阵网格)
│   │   ├── Controls (缩放控制)
│   │   ├── MiniMap
│   │   ├── Custom Nodes:
│   │   │   ├── StartNode
│   │   │   ├── EndNode
│   │   │   ├── ToolNode (含状态指示、版本标记)
│   │   │   └── ConditionNode (迭代阶段)
│   │   └── Custom Edges:
│   │       ├── SequenceEdge (流动动画)
│   │       └── ConditionalEdge
│   │
│   └── ConfigDrawer (右侧, 320px, 选中时滑出)
│       ├── NodeInfoHeader
│       ├── TabBar (参数映射 / 条件 / 设置)
│       ├── MappingTable (参数映射)
│       ├── ConditionEditor (条件配置)
│       └── AdvancedSettings (重试/超时)
│
└── TestPanel (底部, 可收起, 参考 Coze 调试体验)
    ├── DebugPanel (调试面板)
    │   ├── InputArea (输入参数)
    │   │   ├── JSON 编辑器 (支持语法高亮)
    │   │   └── 快捷输入模板 (常用参数预设)
    │   │
    │   ├── RunButton (运行按钮)
    │   │   ├── 调试运行 (开发中流程)
    │   │   └── 试运行 (针对单个节点)
    │   │
    │   └── OutputArea (输出结果)
    │       └── 执行状态、返回数据、耗时
    │
    ├── RunLogPanel (运行日志)
    │   ├── NodeList (节点列表)
    │   │   ├── 节点名称、状态图标
    │   │   ├── 执行状态: pending / running / success / failed
    │   │   └── 时长显示
    │   │
    │   └── NodeDetail (节点详情)
    │       ├── Input (输入参数)
    │       ├── Output (输出结果)
    │       ├── Error (错误信息, 失败时)
    │       └── Duration (执行时长)
    │
    └── HistoryPanel (历史记录)
        ├── 最近 10 次调试记录
        └── 点击可回放历史输入和输出
```

**调试交互流程**（参考 Coze）：

```
┌─────────────────────────────────────────────────────────────────┐
│                        调试流程                                   │
│                                                                 │
│  1. 用户在 InputArea 输入测试参数                                │
│  2. 点击「调试运行」                                             │
│     ├── 创建调试 Run (状态: debugging)                          │
│     ├── 实时推送执行进度                                        │
│     └── 节点执行时高亮当前节点                                    │
│  3. 执行完成后                                                  │
│     ├── OutputArea 展示最终结果                                  │
│     ├── RunLogPanel 展示完整执行链路                             │
│     └── 失败节点可点击查看详情                                   │
│  4. 支持「试运行」单节点                                         │
│     ├── 右键节点 → 试运行                                        │
│     └── 仅执行该节点，查看输入输出                                │
└─────────────────────────────────────────────────────────────────┘
```

**调试运行 vs 试运行**：
| 类型 | 范围 | 用途 |
|------|------|------|
| 调试运行 | 整个流程 | 测试完整流程 |
| 试运行 | 单个节点 | 调试特定节点配置 |

#### 14.4.1 画布交互设计

**节点拖拽**：

```
ToolPanel ──Drag──► ReactFlowCanvas
                              │
                              ▼
                        添加节点到画布
                              │
                              ▼
                    自动对齐到网格 (Snap to Grid)
```

- 左侧工具栏节点拖拽到画布 → 创建新节点（自动布局）
- **用户不可手动拖拽节点位置**
- **画布采用自动布局算法**，节点位置由系统计算生成

**自动布局规则**：
```
┌─────────────────────────────────────────────────────────────────┐
│                    自动布局规则                                  │
│                                                                 │
│  1. 树形结构：从上到下布局                                       │
│     └── 根节点在顶部，子节点依次向下排列                         │
│                                                                 │
│  2. 同级节点：从左到右水平排列                                   │
│     └── 间距自动计算，避免重叠                                   │
│                                                                 │
│  3. 线框包裹：                                                  │
│     └── 含有子域的节点（如迭代、分支、子流程）                  │
│         用虚线边框完整包裹，内部可展开/折叠                      │
│                                                                 │
│  4. 布局触发时机：                                             │
│     └── 节点添加/删除/连线变化时自动重排                        │
│                                                                 │
│  5. 技术选型：                                                 │
│     └── 使用 ElkJS 布局引擎（支持 1000+ 节点，性能优于 Dagre）   │
└─────────────────────────────────────────────────────────────────┘
```

**线框包裹示例**：
```
                    [Start]
                       │
                       ▼
         ┌─────────────────────────────┐
         │   🔄 迭代节点 (可折叠)        │
         │  ┌───┐   ┌───┐   ┌───┐      │
         │  │ A │ → │ B │ → │ C │      │
         │  └───┘   └───┘   └───┘      │
         └─────────────────────────────┘
                       │
                       ▼
                    [End]
```

**节点类型与线框**：
| 节点类型 | 是否有线框 | 说明 |
|---------|-----------|------|
| Start/End | 无 | 边界节点 |
| Tool | 无 | 工具节点 |
| Branch | 有 | 分支节点，线框内包裹多个分支 |
| Loop | 有 | 迭代节点，线框内包裹循环体内容 |

> **注意**：子流程仅作为迭代节点的子内容存在，不作为独立节点类型。

**连线规则**：
```typescript
// 连线校验规则
const validateConnection = (connection: Connection): boolean | string => {
  // 1. 不能连接自己
  if (connection.source === connection.target) {
    return '不能连接自己'
  }
  
  // 2. 树形模式：目标节点只能有一个上游
  const existingInputs = edges.filter(e => e.target === connection.target)
  if (isTreeMode && existingInputs.length > 0) {
    return '树形模式：每个节点最多一个上游'
  }
  
  // 3. 检查循环（树形模式下可跳过）
  if (wouldCreateCycle(connection)) {
    return '不能创建循环'
  }
  
  // 4. 起始节点只能作为源
  if (targetNode.type === 'start') {
    return 'Start 节点不能作为目标'
  }
  
  // 5. 结束节点只能作为目标
  if (sourceNode.type === 'end') {
    return 'End 节点不能作为源'
  }
  
  return true
}
```

**交互行为**：
| 操作 | 行为 |
|------|------|
| 单击节点 | 选中，右侧 drawer 打开 |
| 双击节点 | 进入节点编辑模式 |
| 拖拽节点 | 移动位置 |
| 点击连线 | 选中，可删除 |
| 右键节点 | 上下文菜单（复制/删除/注释） |
| 滚轮 | 缩放画布 |
| 空格+拖拽 | 平移画布 |

**键盘快捷键**：
| 快捷键 | 功能 |
|--------|------|
| Ctrl+S | 保存 |
| Ctrl+Z | 撤销 |
| Ctrl+Shift+Z | 重做 |
| Delete | 删除选中 |
| Ctrl+A | 全选 |
| Escape | 取消选中 |

---

## 14. 部署架构

### 15.1 容器化部署（MVP）

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         Docker Compose 部署架构                                   │
│                                                                                 │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │                          Docker Network                                    │ │
│  │                                                                           │ │
│  │  ┌───────────┐                                                            │ │
│  │  │   Nginx   │ :80/:443                                                   │ │
│  │  │ (反向代理)  │                                                           │ │
│  │  │  - SSL    │                                                            │ │
│  │  │  - 静态资源 │                                                            │ │
│  │  │  - /api→Console│                                                        │ │
│  │  └─────┬─────┘                                                            │ │
│  │        │                                                                   │ │
│  │  ┌─────┴──────────┐     ┌─────────────────────────────────────────────┐   │ │
│  │  │  Frontend      │     │  Console (Go)                               │   │ │
│  │  │  (React)       │     │  ─────────────────────────────────────────  │   │ │
│  │  │  :3000         │     │  --role=console                             │   │ │
│  │  └────────────────┘     │  :8080                                       │   │ │
│  │                         └─────┬───────────────────────────────────────┘   │ │
│  │                               │                                           │ │
│  │           ┌───────────────────┼───────────────────┐                       │ │
│  │           ▼                   ▼                   ▼                       │ │
│  │     ┌───────────┐      ┌───────────┐      ┌───────────┐                  │ │
│  │     │PostgreSQL │      │   Redis   │      │   etcd    │                  │ │
│  │     │  :5432    │      │  :6379    │      │  :2379    │                  │ │
│  │     │           │      │           │      │           │                  │ │
│  │     └───────────┘      └───────────┘      └─────┬─────┘                  │ │
│  │                                                 │                         │ │
│  │                               ┌─────────────────┘                         │ │
│  │                               │ Watch/Get                                  │ │
│  │                               ▼                                            │ │
│  │  ┌─────────────────────────────────────────────────────────────────────┐  │ │
│  │  │  Executor (Go)                                                      │  │ │
│  │  │  ─────────────────────────────────────────────────────────────────  │  │ │
│  │  │  --role=executor --etcd-endpoints=etcd:2379 --cluster-id=default   │  │ │
│  │  │  :8081 (RESTful 触发器) / :8082 (gRPC 内部)                         │  │ │
│  │  └─────────────────────────────────────────────────────────────────────┘  │ │
│  │                                                                           │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**启动命令示例**：

```yaml
# docker-compose.yml
services:
  console:
    image: api-orchestrator:latest
    command: ["--role=console", "--http-port=8080"]
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - ETCD_ENDPOINTS=etcd:2379
    ports:
      - "8080:8080"

  executor:
    image: api-orchestrator:latest
    command: ["--role=executor", "--etcd-endpoints=etcd:2379", "--cluster-id=default"]
    environment:
      - ETCD_ENDPOINTS=etcd:2379
    ports:
      - "8081:8081"  # RESTful 触发器入口
    depends_on:
      - etcd

  etcd:
    image: bitnami/etcd:latest
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    ports:
      - "2379:2379"
```

### 15.2 多集群部署架构

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              生产多集群部署架构                                   │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                           Console 集群                                   │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │   │
│  │  │  Console-1  │    │  Console-2  │    │  Console-3  │                │   │
│  │  │  (LB 后)     │    │  (LB 后)     │    │  (LB 后)     │                │   │
│  │  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                │   │
│  │         └───────────────────┼───────────────────┘                       │   │
│  │                             │                                           │   │
│  │         ┌───────────────────┴───────────────────┐                       │   │
│  │         ▼                   ▼                   ▼                       │   │
│  │   ┌───────────┐       ┌───────────┐       ┌───────────┐                │   │
│  │   │PostgreSQL │       │   Redis   │       │   etcd    │                │   │
│  │   │ (主从)     │       │  (集群)    │       │(Console用)│                │   │
│  │   └───────────┘       └───────────┘       └───────────┘                │   │
│  │                                                                         │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                            │
│                          发布配置   │                                            │
│                                    ▼                                            │
│  ┌────────────────────────────┐  ┌────────────────────────────┐                │
│  │      Dev Cluster           │  │      Prod Cluster          │                │
│  │  ┌──────────────────────┐  │  │  ┌──────────────────────┐  │                │
│  │  │  etcd (dev)          │  │  │  │  etcd (prod)         │  │                │
│  │  │  :2379               │◄─┘  │  │  :2379               │◄─┘                │
│  │  └──────────┬───────────┘      │  └──────────┬───────────┘                  │
│  │             │ Watch             │             │ Watch                        │
│  │  ┌──────────┴───────────┐      │  ┌──────────┴───────────┐                  │
│  │  │  Executor Pool       │      │  │  Executor Pool       │                  │
│  │  │  ┌────┐ ┌────┐ ┌────┐│      │  │  ┌────┐ ┌────┐ ┌────┐│                  │
│  │  │  │Ex-1│ │Ex-2│ │Ex-3││      │  │  │Ex-1│ │Ex-2│ │Ex-3││                  │
│  │  │  └────┘ └────┘ └────┘│      │  │  └────┘ └────┘ └────┘│                  │
│  │  └──────────────────────┘      │  └──────────────────────┘                  │
│  │        (env: dev)              │        (env: prod)                          │
│  └────────────────────────────────┘  └────────────────────────────────          │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**集群配置示例**：

```yaml
# clusters.yaml
clusters:
  - id: dev-cluster
    name: "开发集群"
    environment: dev
    etcd_endpoints:
      - "etcd-dev-1:2379"
      - "etcd-dev-2:2379"
      - "etcd-dev-3:2379"
    
  - id: prod-cluster-beijing
    name: "生产集群-北京"
    environment: prod
    etcd_endpoints:
      - "etcd-prod-bj-1:2379"
      - "etcd-prod-bj-2:2379"
      - "etcd-prod-bj-3:2379"
    
  - id: prod-cluster-shanghai
    name: "生产集群-上海"
    environment: prod
    etcd_endpoints:
      - "etcd-prod-sh-1:2379"
      - "etcd-prod-sh-2:2379"
      - "etcd-prod-sh-3:2379"
```

### 15.3 执行器水平扩容

```
┌────────────────────────────────────────────────────────────────┐
│                    水平扩容架构                                    │
│                                                                │
│  ┌───────────┐                                                 │
│  │   Nginx   │  upstream backend {                              │
│  │ (LB)      │    server backend_1:8080;                        │
│  │           │    server backend_2:8080;                        │
│  │           │    server backend_3:8080;                        │
│  │           │  }                                               │
│  └─────┬─────┘                                                 │
│        │                                                       │
│   ┌────┼────────────┬────────────┐                              │
│   ▼    ▼            ▼            ▼                              │
│ ┌─────┐┌─────┐  ┌─────┐    ┌─────┐                            │
│ │BE 1 ││BE 2 │  │BE 3 │... │BE N │  (无状态，Session 在 Redis)  │
│ └──┬──┘└──┬──┘  └──┬──┘    └──┬──┘                            │
│    └──────┴────────┴──────────┘                                │
│                    │                                            │
│              ┌─────┴─────┐                                     │
│              │  Redis    │  (Session + Cache + 执行队列)          │
│              └───────────┘                                     │
│                                                                │
│  注意事项：                                                     │
│  - 后端服务必须无状态（JWT + Redis Session）                      │
│  - 执行队列通过 Redis Stream 实现跨实例任务分配                    │
│  - 资源探活需要分布式锁避免重复执行                                │
│  - 幂等性：Run 创建需要检查重复                                    │
└────────────────────────────────────────────────────────────────┘
```

### 15.3 Docker Compose 配置要点

```yaml
# 关键配置说明
services:
  backend:
    environment:
      - APP_ENV=production
      - DATABASE_URL=postgres://user:pass@postgres:5432/api_orchestrator?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=${JWT_SECRET}           # 必须通过环境变量注入
      - CONFIG_ENCRYPTION_KEY=${ENCRYPTION_KEY} # 数据加密密钥
      - LOG_LEVEL=info
    deploy:
      replicas: 2                          # 水平扩容
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  postgres:
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=api_orchestrator
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $DB_USER"]
      interval: 10s
```

### 15.4 环境配置

| 环境 | 说明 |
|------|------|
| **开发环境** | 本地运行，Vite Dev Server + Go 本地运行，PostgreSQL + Redis 通过 Docker |
| **测试环境** | Docker Compose 全栈部署，模拟生产配置 |
| **生产环境** | Docker Compose 或 K8s 部署，环境变量注入敏感配置，持久化数据卷 |

---

## 15. 可观测性与监控

### 15.1 日志

```
┌────────────────────────────────────────────────────────────────┐
│                      日志规范                                    │
│                                                                │
│  格式: 结构化 JSON                                              │
│  级别: DEBUG < INFO < WARN < ERROR                             │
│                                                                │
│  字段规范:                                                      │
│  {                                                             │
│    "level": "info",                                            │
│    "ts": "2026-04-01T15:30:00.000Z",                           │
│    "msg": "流程执行完成",                                         │
│    "request_id": "req-xxx",     // 请求追踪 ID                   │
│    "user_id": "user-xxx",      // 操作用户                      │
│    "workspace_id": "ws-xxx",   // 工作空间                      │
│    "trace_id": "trace-xxx",    // 分布式追踪 ID                  │
│    "span_id": "span-xxx",      // 追踪段 ID                     │
│    "duration_ms": 1250,        // 耗时                          │
│    "run_id": "run-xxx",        // 运行记录 ID                   │
│    "flow_id": "flow-xxx",      // 流程 ID                      │
│    "error": {                                               │
│      "code": "TOOL_EXECUTION_FAILED",                        │
│      "message": "连接超时",                                    │
│      "stack": "..."                                        │
│    }                                                        │
│  }                                                          │
│                                                                │
│  审计日志（单独存储）:                                           │
│  - 用户登录/登出                                                │
│  - 权限变更                                                    │
│  - 资源/工具/流程的创建/更新/删除                                │
│  - 发布/下线操作                                               │
│  - 触发器启停                                                  │
└────────────────────────────────────────────────────────────────┘
```

### 15.2 指标

```
┌────────────────────────────────────────────────────────────────┐
│                      关键指标                                    │
│                                                                │
│  系统指标 (Prometheus):                                         │
│  ├── api_request_total{method, path, status_code}             │
│  ├── api_request_duration_seconds{method, path} (histogram)    │
│  ├── api_active_requests                                       │
│  ├── flow_execution_total{status}                              │
│  ├── flow_execution_duration_seconds (histogram)               │
│  ├── flow_active_executions                                    │
│  ├── tool_invocation_total{plugin, status}                     │
│  ├── tool_invocation_duration_seconds (histogram)              │
│  ├── resource_health_status{type, status}                      │
│  ├── redis_operations_total{command}                           │
│  ├── db_query_duration_seconds (histogram)                     │
│  └── go_goroutines / go_memstats                              │
│                                                                │
│  业务指标 (Dashboard 展示):                                      │
│  ├── 今日运行总数 / 成功率 / 平均耗时                             │
│  ├── 资源健康状态分布 (active/slow/error)                       │
│  ├── 工具使用排行                                               │
│  ├── 触发器调用次数 Top 10                                      │
│  └── 工作空间活跃度排名                                          │
└────────────────────────────────────────────────────────────────┘
```

### 16.3 健康检查

```go
// GET /health
{
    "status": "ok",
    "version": "1.0.0",
    "uptime": "72h15m",
    "components": {
        "database": "ok",       // PostgreSQL 连接检查
        "redis": "ok",          // Redis PING
        "plugins": ["http", "grpc", "postgres", "timer", "restful"]  // 已编译插件
    },
    "build_tags": "http,grpc,postgres,timer,restful",
    "go_version": "go1.22"
}
```

---

## 16. 非功能约束与质量属性

### 16.1 质量属性矩阵

| 属性 | 目标 | 策略 |
|------|------|------|
| **性能** | API P95 < 300ms | 数据库索引优化、Redis 缓存、连接池 |
| **可扩展性** | 50+ 并发流程执行 | 无状态后端 + 多实例水平扩容 + Redis 队列 |
| **可靠性** | 流程执行不丢失 | 持久化队列 + 重试机制 + 死信队列 |
| **安全性** | 企业级 | JWT + RBAC + 数据加密 + HTTPS + 审计日志 |
| **可维护性** | 模块边界清晰 | DDD 分层 + 模块化单体 + 结构化日志 |
| **可部署性** | 一键启动 | Docker Compose + 健康检查 + 数据库迁移 |
| **可观测性** | 全链路追踪 | 结构化日志 + Prometheus 指标 + Request ID |

### 16.2 并发与限流

| 场景 | 策略 |
|------|------|
| API 限流 | 全局 1000 req/s，单用户 100 req/s（令牌桶） |
| 流程并发 | 单实例 50 并发 goroutine，超出排队等待 |
| 资源探活 | 最多 10 并发探活 goroutine |
| 数据库连接池 | max_open=50, max_idle=10, max_lifetime=30min |
| Redis 连接池 | pool_size=20 |

### 16.3 错误处理策略

```
┌────────────────────────────────────────────────────────────────┐
│                      错误处理分层                                 │
│                                                                │
│  1. 工具执行错误:                                                │
│     - 记录到 RunNodeLog (error_message + input/output)           │
│     - 根据 Flow 配置的失败策略 (终止/重试)                        │
│     - Run 状态 → failed                                         │
│                                                                │
│  2. 触发器调用错误:                                              │
│     - 输入参数校验失败 → 400 + 错误详情                           │
│     - 绑定流程不存在 → 503                                       │
│     - 执行超时 → 504 + Run ID (可查后续结果)                     │
│                                                                │
│  3. 系统级错误:                                                  │
│     - 数据库连接失败 → 503 + 自动重连                             │
│     - Redis 不可用 → 降级到直接查库 (性能下降但不中断)              │
│     - 插件执行崩溃 → 捕获 panic → 节点状态 → error                │
│                                                                │
│  4. 统一错误响应格式:                                             │
│     {                                                          │
│       "code": 40001,                                          │
│       "message": "资源不存在",                                   │
│       "details": [                                             │
│         { "field": "id", "message": "无效的 UUID" }            │
│       ]                                                      │
│     }                                                        │
└────────────────────────────────────────────────────────────────┘
```

### 16.4 数据备份

| 策略 | 说明 |
|------|------|
| **PostgreSQL** | `pg_dump` 定时备份，保留最近 7 天 |
| **Redis** | RDB 快照每小时一次 + AOF 持久化 |
| **备份存储** | 本地挂载卷 + 可选远程 S3 存储 |
| **恢复测试** | 每月一次恢复演练 |

---

## 17. 演进路线图

### 17.1 架构演进路径

```
Phase 1: 双角色架构 (MVP, 当前)
├── Console + Executor 双角色二进制
├── PostgreSQL (定义存储) + etcd (配置中心)
├── Redis (缓存/会话)
├── 内置插件
└── Docker Compose 部署

Phase 2: 能力增强 (迭代 1, +4-6 周)
├── WebSocket 实时状态推送
├── MQ 触发器 (RabbitMQ/Kafka 插件)
├── 版本对比和回滚
├── 影响分析和下线保护完善
├── 多集群管理完善
├── 执行器自动扩缩容
└── 审计日志 UI

Phase 3: 规模化 (迭代 2, +6-8 周)
├── 条件分支/循环节点
├── 新增插件类型扩展（根据客户需求按编译时模式添加）
├── 国际化 (i18n)
├── API 文档门户
├── Prometheus + Grafana 监控
├── 跨集群调用（流程 A 调用集群 B 的流程）
└── 接口 Mock 服务

Phase 4: 可选拆分 (远期)
├── 如需独立扩缩容，可将 Executor 拆为独立部署单元
├── 如需高可用事件处理，引入 Kafka
├── 如需多区域部署，考虑服务网格 (Istio)
└── 拆分原则：按业务能力拆，不按技术层拆
```

### 17.2 拆分预判

如果未来需要从双角色架构进一步拆分，优先拆分顺序：

1. **Executor 独立部署** — 已经是独立角色，可进一步分离为独立二进制
2. **触发器调度服务** — 需要与外部 MQ 深度集成
3. **认证服务** — 无状态，可独立部署多实例

**不拆分的模块**（保持单体）：
- 资源管理、工具管理、流程管理、运行记录 — 数据强关联，拆分收益低
- Console 保持单体，通过多实例水平扩容

### 17.3 技术债务跟踪

| 债务 | 优先级 | 计划 |
|------|--------|------|
| RESTful 触发器同步等待优化 | P1 | MVP 后立即优化 |
| WebSocket 实时推送 | P1 | 迭代 1 |
| 前端画布性能优化（50+ 节点） | P2 | 迭代 2 |
| 插件热加载能力（可选） | P3 | 远期评估，如客户有自助扩展需求再考虑 |
| 条件分支节点支持 | P2 | 迭代 2 |
| 接口 Mock 服务 | P3 | 迭代 3 |
| OpenAPI 文档自动生成 | P3 | 迭代 3 |

---

## 附录

### A. 术语表

| 术语 | 定义 |
|------|------|
| 双角色架构 | Console（管理）和 Executor（执行）编译到同一二进制，通过 CLI 区分的架构 |
| 模块化单体 | 在单体应用内部通过代码组织实现模块边界的架构风格 |
| 限界上下文 | DDD 中划分领域边界的模式 |
| 聚合根 | 聚合的入口实体，负责维护一致性边界 |
| 领域事件 | 领域内发生的重要业务事件 |
| 依赖倒置 | 高层模块不依赖低层模块，二者都依赖抽象接口 |
| 版本锁定 | 流程/触发器运行时使用发布时锁定的具体版本 |
| 环境 | 对集群的分类标签，如 dev/staging/prod |
| 集群 | 一组运行相同配置的执行器实例，绑定唯一的 etcd |
| etcd | 配置中心，存储发布到集群的运行时配置 |
| Console | 控制台，提供 Web UI 和 API，管理资源/工具/流程/触发器定义 |
| Executor | 执行器，连接 etcd 监听配置变更，执行触发器和流程 |

### B. 参考文档

| 文档 | 路径 |
|------|------|
| 产品需求文档 | `docs/PRD.md` |
| 开发者实施指南 | `docs/developer-guide.md` |
| 信息架构图 | `docs/information-architecture.md` |
| 用户旅程地图 | `docs/user-journey-map.md` |
| 页面流程图 | `docs/page-flow.md` |
| CSS Design Tokens | `design-system/styles/design-tokens.css` |
| HTML 原型 | `prototype/` |

### C. 架构决策索引

| ADR | 标题 | 状态 |
|-----|------|------|
| ADR-001 | 模块化单体架构 | 已接受 |
| ADR-002 | 后端语言 Go | 已接受 |
| ADR-003 | 前端 React + TypeScript + Vite | 已接受 |
| ADR-004 | 数据库 PostgreSQL | 已接受 |
| ADR-005 | 编译时插件（Build-time Plugin） | 已接受 |
| ADR-006 | 异步调度 + 实时状态推送 | 已接受 |
| ADR-007 | 双角色架构 — Console + Executor | 已接受 |
| ADR-008 | 配置中心 — etcd | 已接受 |
| ADR-009 | 集群与环境模型 | 已接受 |
