# 双角色部署指南

根据架构设计要求，控制台(Console)和执行器(Executor)是同一个二进制程序，通过CLI参数`--role`来区分角色。

## 启动方式

### 1. 启动控制台（Console）角色

控制台提供Web UI和API服务，用于管理资源、工具、流程和触发器定义。

```bash
# 基本启动
./mengri-flow.exe --role=console --config=config.yaml

# 或简写（console是默认角色）
./mengri-flow.exe --config=config.yaml

# Docker启动
docker run -d \
  --name mengri-flow-console \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  mengri-flow:latest \
  --role=console \
  --config=/app/config.yaml
```

**控制台职责：**
- 提供Web UI和RESTful API
- 管理资源、工具、流程、触发器定义
- 向etcd发布配置
- 数据存储：PostgreSQL（资源、工具、流程、用户、权限等）
- 运行记录存储和查询

**配置文件示例（config.yaml）：**

```yaml
server:
  port: 8080
  mode: debug

database:
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/mengri_flow?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

log:
  level: info
  format: json

plugins:
  enabled:
    - http
    - grpc
    - timer
```

### 2. 启动执行器（Executor）角色

执行器连接etcd监听配置变更，执行触发器和流程。

```bash
# 基本启动
./mengri-flow.exe --role=executor \
  --etcd-endpoints=etcd-1:2379,etcd-2:2379,etcd-3:2379 \
  --cluster-id=cluster-prod-001 \
  --node-id=executor-node-1 \
  --log-level=info

# 自动生成node-id（如果不指定）
./mengri-flow.exe --role=executor \
  --etcd-endpoints=localhost:2379 \
  --cluster-id=cluster-dev-001

# Docker启动
docker run -d \
  --name mengri-flow-executor-1 \
  mengri-flow:latest \
  --role=executor \
  --etcd-endpoints=etcd-1:2379,etcd-2:2379 \
  --cluster-id=cluster-prod-001 \
  --node-id=executor-$(hostname)
```

**执行器职责：**
- 连接etcd监听配置变更
- 执行触发器（RESTful/Timer/MQ）
- 执行流程编排
- 上报运行状态和心跳
- 无本地状态存储（状态存储在etcd）

**执行器参数说明：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--role` | 是 | 必须设置为`executor` |
| `--etcd-endpoints` | 是 | etcd集群地址，多个用逗号分隔 |
| `--cluster-id` | 是 | 集群ID，对应控制台中创建的集群 |
| `--node-id` | 否 | 执行器节点ID，不指定则自动生成 |
| `--executor-port` | 否 | RESTful触发器HTTP服务端口 |
| `--log-level` | 否 | 日志级别：debug/info/warn/error，默认info |

## 架构说明

### 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Console (控制台)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Web UI    │  │   API 服务   │  │      发布服务            │  │
│  │  (Vue 3)    │  │   (REST)    │  │  (同步配置到 etcd)       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                          │                                      │
│                          ▼                                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              数据库 (PostgreSQL/MySQL)                   │   │
│  │  - 资源/工具/流程/触发器定义                              │   │
│  │  - 工作空间/用户/权限                                    │   │
│  │  - 运行记录/审计日志                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼ 发布配置
┌─────────────────────────────────────────────────────────────────┐
│                        etcd 配置中心                             │
│  /clusters/{cluster-id}/                                        │
│    ├── /flows/{flow-id}          # 流程定义                      │
│    ├── /triggers/{trigger-id}    # 触发器配置                    │
│    ├── /resources/{res-id}       # 资源配置                      │
│    └── /tools/{tool-id}          # 工具定义                      │
└─────────────────────────────────────────────────────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                 ▼
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│  Executor 1     │   │  Executor 2     │   │  Executor N     │
│  (同一二进制)    │   │  (同一二进制)    │   │  (同一二进制)    │
│                 │   │                 │   │                 │
│  --role=executor│   │  --role=executor│   │  --role=executor│
└─────────────────┘   └─────────────────┘   └─────────────────┘
```

### 为什么使用同一二进制？

1. **简化部署**：只需构建和维护一个二进制文件
2. **减少镜像大小**：Docker镜像只需包含一个程序
3. **统一版本管理**：Console和Executor版本保持一致
4. **共享代码**：共享配置解析、日志、插件系统等基础组件
5. **灵活的部署模式**：同一二进制可在不同环境灵活部署不同角色

### 角色对比

| 特性 | Console | Executor |
|------|---------|----------|
| **启动参数** | `--role=console` | `--role=executor` |
| **依赖服务** | PostgreSQL, Redis | etcd |
| **对外端口** | 8080 (Web + API) | 动态（RESTful触发器） |
| **是否有状态** | 有（数据库） | 无（状态在etcd） |
| **部署数量** | 通常1-2个 | 可水平扩展多个实例 |
| **资源需求** | 中等（管理用） | 较高（执行流程） |

## 部署示例

### Docker Compose部署

```yaml
version: '3.8'

services:
  # 数据库
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: mengri_flow
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Redis
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # etcd
  etcd:
    image: quay.io/coreos/etcd:v3.5.9
    command: >
      etcd
      --advertise-client-urls http://0.0.0.0:2379
      --listen-client-urls http://0.0.0.0:2379
    ports:
      - "2379:2379"

  # Console
  console:
    image: mengri-flow:latest
    command:
      - --role=console
      - --config=/app/config.yaml
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
    depends_on:
      - postgres
      - redis

  # Executor实例1
  executor-1:
    image: mengri-flow:latest
    command:
      - --role=executor
      - --etcd-endpoints=etcd:2379
      - --cluster-id=cluster-prod-001
      - --node-id=executor-1
    depends_on:
      - etcd
    deploy:
      replicas: 2  # 可以扩展多个副本

volumes:
  postgres_data:
```

### Kubernetes部署

```yaml
# Console Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mengri-flow-console
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mengri-flow-console
  template:
    metadata:
      labels:
        app: mengri-flow-console
    spec:
      containers:
      - name: console
        image: mengri-flow:latest
        args:
        - --role=console
        - --config=/app/config.yaml
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
      volumes:
      - name: config
        configMap:
          name: mengri-flow-config
---
# Executor Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mengri-flow-executor
spec:
  replicas: 3  # 3个执行器实例
  selector:
    matchLabels:
      app: mengri-flow-executor
  template:
    metadata:
      labels:
        app: mengri-flow-executor
    spec:
      containers:
      - name: executor
        image: mengri-flow:latest
        args:
        - --role=executor
        - --etcd-endpoints=etcd:2379
        - --cluster-id=cluster-prod-001
        - --node-id=$(POD_NAME)
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
```

## 验证部署

### 验证Console

```bash
# 检查Console是否运行
curl http://localhost:8080/api/v1/greet

# 应该返回:
{"message":"Welcome to Mengri Flow"}
```

### 验证Executor

```bash
# 查看Executor日志
docker logs mengri-flow-executor-1

# 应该看到:
# "Executor executor-node-1 started successfully for cluster cluster-prod-001"

# 检查etcd中的执行器状态
etcdctl get /clusters/cluster-prod-001/executors/ --prefix
```

## 故障排查

### Console启动失败

1. **检查数据库连接**
   ```bash
   # 查看日志
   docker logs mengri-flow-console
   ```

2. **检查配置**
   ```bash
   # 验证config.yaml格式
   ./mengri-flow.exe --role=console --config=config.yaml
   ```

### Executor无法连接etcd

1. **检查etcd地址**
   ```bash
   # 测试etcd连通性
   etcdctl --endpoints=etcd:2379 endpoint health
   ```

2. **检查cluster-id**
   ```bash
   # 确认cluster-id在控制台中已创建
   # 查看etcd中的集群配置
   etcdctl get /clusters/cluster-prod-001 --prefix
   ```

3. **查看Executor日志**
   ```bash
   docker logs mengri-flow-executor-1
   ```

## 最佳实践

1. **生产环境部署**
   - Console：至少2个实例（高可用）
   - Executor：根据负载动态扩展（3-10个实例）
   - etcd：3节点或5节点集群
   - PostgreSQL：主从复制

2. **网络隔离**
   - Console部署在管理网络
   - Executor部署在执行网络
   - 通过etcd进行跨网络通信

3. **监控告警**
   - 监控Console的API响应时间
   - 监控Executor的心跳和运行状态
   - 监控etcd的存储和性能
   - 配置告警规则

4. **日志收集**
   - 统一收集Console和Executor的日志
   - 使用结构化日志（JSON格式）
   - 配置日志级别（生产环境用info或warn）

5. **版本管理**
   - Console和Executor使用相同版本
   - 滚动升级，先升级Console再升级Executor
   - 验证兼容性后再全量升级
