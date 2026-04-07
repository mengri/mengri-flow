---

## 10. MQ触发器设计补充

### 10.1 MQ触发器的直接执行模式优势

基于你的正确理解，MQ触发器采用"消息到达时直接在当前线程执行，无需异步队列"的设计，具有以下优势：

1. **架构简化**：避免额外的消息队列和任务调度
2. **延迟降低**：消息处理后立即执行，减少中间环节
3. **资源节省**：不需要维护任务状态和任务队列
4. **错误处理简化**：消息确认与流程执行直接关联

### 10.2 消息确认策略

```go
// MQ消息确认模式
type MessageAckMode string

const (
    AckModeAuto   MessageAckMode = "auto"   // 自动确认（MQ触发器默认）
    AckModeManual MessageAckMode = "manual" // 手动确认（特殊场景）
)

// MQTriggerConfig MQ触发器配置
type MQTriggerConfig struct {
    MQType      string        `json:"mq_type"`      // rabbitmq, kafka
    Connection  MQConnection  `json:"connection"`   // 连接配置
    QueueName   string        `json:"queue_name"`   // RabbitMQ队列
    Topic       string        `json:"topic"`        // Kafka Topic
    GroupID     string        `json:"group_id"`     // 消费者组ID
    AckMode     MessageAckMode `json:"ack_mode"`    // 确认模式
    Concurrency int           `json:"concurrency"`  // 消费并发度
}

// 消息确认逻辑
func (c *MQConsumer) executeFlowFromMessage(ctx context.Context, msg amqp.Delivery) {
    // 开始时间，用于记录执行耗时
    startTime := time.Now()
    
    defer func() {
        // 记录执行指标
        duration := time.Since(startTime).Seconds()
        metrics.MQMessagesConsumed.WithLabelValues(
            c.triggerID, c.mqType, "processed",
        ).Inc()
        metrics.MQMessageProcessDuration.WithLabelValues(c.triggerID).Observe(duration)
    }()
    
    // 执行前确认或延迟确认
    if c.config.AckMode == AckModeAuto {
        // 自动确认：消息接收即确认，执行失败也不会重试
        if err := msg.Ack(false); err != nil {
            log.Errorf("Failed to ack message: %v", err)
        }
    }
    
    // 执行流程
    result, err := c.flowExecutor.Execute(ctx, c.flowID, c.buildInputs(msg))
    
    if err != nil {
        log.Errorf("MQ trigger execution failed: %s: %v", c.triggerID, err)
        metrics.MQMessagesConsumed.WithLabelValues(
            c.triggerID, c.mqType, "failed",
        ).Inc()
        
        // 如果是手动确认模式，并且在失败时不重新入队
        if c.config.AckMode == AckModeManual {
            // 可以记录到死信队列或直接拒绝
            if err := msg.Nack(false, false); err != nil {
                log.Errorf("Failed to nack message: %v", err)
            }
        }
    } else {
        log.Debugf("MQ trigger executed successfully: %s", c.triggerID)
        
        // 如果是手动确认模式，成功执行后确认
        if c.config.AckMode == AckModeManual {
            if err := msg.Ack(false); err != nil {
                log.Errorf("Failed to ack message: %v", err)
            }
        }
    }
}

// 构建流程输入
func (c *MQConsumer) buildInputs(msg amqp.Delivery) map[string]interface{} {
    inputs := map[string]interface{}{
        "mq_message": map[string]interface{}{
            "body":    extractMessageBody(msg),
            "headers": msg.Headers,
            "id":      msg.MessageId,
            "type":    msg.Type,
        },
        "trigger_info": map[string]interface{}{
            "trigger_id": c.triggerID,
            "mq_type":    c.mqType,
            "timestamp":  time.Now().Unix(),
        },
    }
    
    // 尝试解析JSON消息体
    var messageBody map[string]interface{}
    if err := json.Unmarshal(msg.Body, &messageBody); err == nil {
        // 如果是JSON格式，将字段平铺到inputs中
        for k, v := range messageBody {
            inputs[k] = v
        }
    } else {
        // 如果不是JSON，保持原始消息体
        inputs["raw_body"] = string(msg.Body)
    }
    
    return inputs
}

// 提取消息体
func extractMessageBody(msg amqp.Delivery) interface{} {
    // 根据Content-Type处理不同格式的消息体
    contentType := getHeader(msg.Headers, "content-type")
    
    switch {
    case strings.Contains(contentType, "application/json"):
        var data map[string]interface{}
        if err := json.Unmarshal(msg.Body, &data); err == nil {
            return data
        }
        return string(msg.Body)
        
    case strings.Contains(contentType, "text/plain"):
        return string(msg.Body)
        
    default:
        // 默认返回base64编码的二进制数据
        return base64.StdEncoding.EncodeToString(msg.Body)
    }
}
```

### 10.3 Kafka消费者设计补充

```go
// KafkaConsumerGroup Kafka消费者组
type KafkaConsumerGroup struct {
    triggerID    string
    flowID       string
    topic        string
    groupID      string
    flowExecutor FlowExecutor
    
    consumer sarama.ConsumerGroup
    done     chan struct{}
}

// Setup 消费者组初始化
func (cg *KafkaConsumerGroup) Setup(sarama.ConsumerGroupSession) error {
    log.Infof("Kafka consumer group setup for trigger %s", cg.triggerID)
    return nil
}

// Cleanup 消费者组清理
func (cg *KafkaConsumerGroup) Cleanup(sarama.ConsumerGroupSession) error {
    log.Infof("Kafka consumer group cleanup for trigger %s", cg.triggerID)
    return nil
}

// ConsumeClaim 消费消息
func (cg *KafkaConsumerGroup) ConsumeClaim(
    session sarama.ConsumerGroupSession,
    claim sarama.ConsumerGroupClaim) error {
    
    for message := range claim.Messages() {
        // 在当前协程中直接执行流程
        go func(msg *sarama.ConsumerMessage) {
            cg.executeFromKafkaMessage(session.Context(), msg)
            
            // 标记消息已处理
            session.MarkMessage(msg, "")
        }(message)
    }
    
    return nil
}

// executeFromKafkaMessage 执行Kafka消息对应的流程
func (cg *KafkaConsumerGroup) executeFromKafkaMessage(ctx context.Context, msg *sarama.ConsumerMessage) {
    log.Debugf("Processing Kafka message: topic=%s, partition=%d, offset=%d",
        msg.Topic, msg.Partition, msg.Offset)
    
    inputs := map[string]interface{}{
        "kafka_message": map[string]interface{}{
            "topic":     msg.Topic,
            "partition": msg.Partition,
            "offset":    msg.Offset,
            "key":       string(msg.Key),
            "value":     string(msg.Value),
            "headers":   extractKafkaHeaders(msg.Headers),
            "timestamp": msg.Timestamp,
        },
        "trigger_info": map[string]interface{}{
            "trigger_id": cg.triggerID,
            "mq_type":    "kafka",
            "topic":      msg.Topic,
        },
    }
    
    // 执行流程
    _, err := cg.flowExecutor.Execute(ctx, cg.flowID, inputs)
    if err != nil {
        log.Errorf("Kafka trigger execution failed: %s: %v", cg.triggerID, err)
        metrics.MQMessagesConsumed.WithLabelValues(
            cg.triggerID, "kafka", "failed",
        ).Inc()
    } else {
        log.Debugf("Kafka trigger executed successfully: %s", cg.triggerID)
        metrics.MQMessagesConsumed.WithLabelValues(
            cg.triggerID, "kafka", "processed",
        ).Inc()
    }
}

// 提取Kafka消息头
func extractKafkaHeaders(headers []*sarama.RecordHeader) map[string]string {
    result := make(map[string]string)
    for _, h := range headers {
        result[string(h.Key)] = string(h.Value)
    }
    return result
}
```

---

## 11. 执行器主程序更新

### 11.1 主程序初始化

```go
// NewExecutor 创建执行器实例
func NewExecutor(cfg ExecutorConfig) (*Executor, error) {
    // 1. 生成执行器ID
    executorID := cfg.ExecutorID
    if executorID == "" {
        executorID = fmt.Sprintf("executor-%s-%s",
            getHostname(),
            uuid.New().String()[:8])
    }
    
    // 2. 初始化etcd客户端
    etcdClient, err := initEtcdClient(cfg.EtcdEndpoints)
    if err != nil {
        return nil, fmt.Errorf("init etcd client failed: %w", err)
    }
    
    // 3. 初始化分布式锁
    distributedLock := &etcdDistributedLock{
        client: etcdClient,
    }
    
    // 4. 初始化配置管理器
    configManager := newConfigManager(etcdClient, cfg.ClusterID)
    
    // 5. 初始化流程执行器
    flowExecutor := newFlowExecutor()
    
    // 6. 初始化HTTP服务器（RESTful触发器需要）
    var httpServer *HTTPServer
    if cfg.HTTPPort > 0 {
        httpServer = &HTTPServer{
            port:         cfg.HTTPPort,
            address:      cfg.HTTPAddress,
            flowExecutor: flowExecutor,
            authChecker:  newAuthChecker(),
        }
    }
    
    // 7. 初始化定时调度器
    timerScheduler := &TimerScheduler{
        executorID:    executorID,
        clusterID:     cfg.ClusterID,
        etcdLock:      distributedLock,
        flowExecutor:  flowExecutor,
        scheduledJobs: make(map[string]*scheduledJob),
    }
    
    // 8. 创建执行器实例
    executor := &Executor{
        executorID:      executorID,
        clusterID:       cfg.ClusterID,
        hostname:        getHostname(),
        version:         getVersion(),
        configManager:   configManager,
        flowExecutor:    flowExecutor,
        httpServer:      httpServer,
        timerScheduler:  timerScheduler,
        distributedLock: distributedLock,
        mqConsumers:     make(map[string]MQConsumer),
        triggers:        make(map[string]*Trigger),
    }
    
    return executor, nil
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
    ClusterID         string   `yaml:"cluster_id"`
    ExecutorID        string   `yaml:"executor_id"`
    EtcdEndpoints     []string `yaml:"etcd_endpoints"`
    HTTPPort          int      `yaml:"http_port"`
    HTTPAddress       string   `yaml:"http_address"`
    RedisAddr         string   `yaml:"redis_addr"`
    RedisPassword     string   `yaml:"redis_password"`
    MaxConcurrent     int      `yaml:"max_concurrent"`
    LogLevel          string   `yaml:"log_level"`
    MetricsPort       int      `yaml:"metrics_port"`
    PProfPort         int      `yaml:"pprof_port"`
}
```

### 11.2 启动流程优化

```go
// StartExecutor 启动执行器
func StartExecutor(cfg ExecutorConfig) error {
    // 1. 初始化日志
    initLogger(cfg.LogLevel)
    
    // 2. 创建设置信号处理
    signalCtx, cancel := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    // 3. 创建执行器实例
    executor, err := NewExecutor(cfg)
    if err != nil {
        return fmt.Errorf("create executor failed: %w", err)
    }
    
    // 4. 启动执行器
    if err := executor.Start(signalCtx); err != nil {
        return fmt.Errorf("start executor failed: %w", err)
    }
    
    // 5. 启动监控服务
    if cfg.MetricsPort > 0 {
        go startMetricsServer(cfg.MetricsPort)
    }
    if cfg.PProfPort > 0 {
        go startPProfServer(cfg.PProfPort)
    }
    
    log.Infof(`Executor started successfully
    Executor ID: %s
    Cluster ID:  %s
    HTTP Server: %s:%d
    Metrics:     http://localhost:%d/metrics
    PProf:       http://localhost:%d/debug/pprof`,
        executor.executorID,
        executor.clusterID,
        cfg.HTTPAddress, cfg.HTTPPort,
        cfg.MetricsPort,
        cfg.PProfPort)
    
    // 6. 等待退出信号
    <-signalCtx.Done()
    log.Info("Received shutdown signal, stopping executor...")
    
    // 7. 停止执行器
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    // 启动优雅关闭
    go func() {
        if err := executor.Stop(); err != nil {
            log.Errorf("Executor stopped with error: %v", err)
        } else {
            log.Info("Executor stopped gracefully")
        }
        shutdownCancel()
    }()
    
    // 等待关闭完成或超时
    select {
    case <-shutdownCtx.Done():
        if shutdownCtx.Err() == context.DeadlineExceeded {
            log.Warn("Executor shutdown timed out")
        }
    }
    
    return nil
}
```

### 11.3 优雅关闭实现

```go
// Stop 优雅停止执行器
func (e *Executor) Stop() error {
    if !e.running {
        return nil
    }
    
    log.Info("Stopping executor gracefully...")
    e.running = false
    
    // 1. 取消上下文，通知所有组件开始关闭
    e.cancel()
    
    // 2. 按依赖顺序停止组件
    var wg sync.WaitGroup
    var errs []error
    
    // 停止MQ消费者
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := e.stopMQConsumers(); err != nil {
            errs = append(errs, fmt.Errorf("stop MQ consumers: %w", err))
        }
    }()
    
    // 停止Timer调度器
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := e.timerScheduler.Stop(); err != nil {
            errs = append(errs, fmt.Errorf("stop timer scheduler: %w", err))
        }
    }()
    
    // 停止HTTP服务器
    if e.httpServer != nil {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := e.httpServer.Shutdown(10*time.Second); err != nil {
                errs = append(errs, fmt.Errorf("stop http server: %w", err))
            }
        }()
    }
    
    // 停止触发器管理器
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := e.triggerManager.Stop(); err != nil {
            errs = append(errs, fmt.Errorf("stop trigger manager: %w", err))
        }
    }()
    
    // 等待所有组件停止完成
    wg.Wait()
    
    // 3. 最后停止状态管理器
    if err := e.stateManager.Stop(); err != nil {
        errs = append(errs, fmt.Errorf("stop state manager: %w", err))
    }
    
    // 4. 总结关闭结果
    if len(errs) > 0 {
        return fmt.Errorf("executor stopped with errors: %v", errs)
    }
    
    return nil
}

// stopMQConsumers 停止所有MQ消费者
func (e *Executor) stopMQConsumers() error {
    e.mu.RLock()
    consumers := make([]MQConsumer, 0, len(e.mqConsumers))
    for _, consumer := range e.mqConsumers {
        consumers = append(consumers, consumer)
    }
    e.mu.RUnlock()
    
    var errs []error
    for _, consumer := range consumers {
        if err := consumer.Stop(); err != nil {
            errs = append(errs, fmt.Errorf("stop consumer %s: %w", consumer.GetID(), err))
        }
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("stop MQ consumers failed: %v", errs)
    }
    
    return nil
}
```

---

## 12. 设计总结

### 12.1 关键设计点回顾

基于你的正确理解，本设计实现了以下关键点：

1. **RESTful触发器**：
   - 符合RESTful规范的Webhook
   - 执行器在指定端口创建API接口
   - 支持同步（立即返回）和异步（返回任务ID）两种模式
   - 核心目标：将流程开放为API供业务调用

2. **Timer触发器**：
   - 按时间触发（Cron表达式）
   - 通过etcd分布式锁保证集群级别唯一执行
   - 避免多个执行器实例重复执行
   - 锁粒度为分钟级别，优化性能

3. **MQ触发器**：
   - 监听MQ指定地址、Topic和消费策略
   - 每收到一条消息执行一次流程
   - 采用直接执行模式，无需额外异步队列
   - 消息确认与流程执行直接关联

### 12.2 架构优势

1. **清晰的职责分离**：
   - RESTful：HTTP Server + API路由
   - Timer：分布式调度 + 锁管理
   - MQ：消息消费 + 直接执行

2. **性能优化**：
   - RESTful同步模式：超时控制，避免长连接
   - RESTful异步模式：任务队列，支持长耗时任务
   - Timer触发器：分布式锁减少竞争
   - MQ触发器：直接执行，最低延迟

3. **可靠性保障**：
   - 心跳机制实时监控执行器状态
   - 配置动态更新，支持热部署
   - 完善的错误处理和重试机制
   - 分布式锁保证定时任务唯一性

4. **可观测性**：
   - 完整的Prometheus指标
   - 结构化日志记录
   - 健康检查端点
   - 性能剖析支持

### 12.3 扩展性考虑

1. **多MQ支持**：通过接口抽象支持不同MQ实现
2. **水平扩展**：无状态设计，支持任意数量执行器实例
3. **插件化**：资源类型和工具插件化扩展
4. **配置驱动**：运行时配置动态更新

### 12.4 部署灵活性

1. **二进制部署**：单一二进制，简化部署
2. **容器化部署**：完整的Docker和Kubernetes支持
3. **混合部署**：支持传统服务器和云原生部署
4. **自动化运维**：启动脚本、健康检查、指标收集

---

## 13. 下一步实施建议

### 13.1 实施优先级

1. **Phase 1: 核心架构**（2-3周）
   - [ ] 实现执行器主框架和etcd集成
   - [ ] 实现分布式锁（Timer触发器基础）
   - [ ] 实现状态管理和心跳机制
   - [ ] 实现配置监听和动态更新

2. **Phase 2: 触发器实现**（3-4周）
   - [ ] 实现RESTful触发器HTTP Server
   - [ ] 实现同步/异步接口范式
   - [ ] 实现Timer触发器调度器
   - [ ] 实现MQ触发器基础框架

3. **Phase 3: 增强功能**（2-3周）
   - [ ] 实现完整的错误处理和监控
   - [ ] 实现性能优化和资源管理
   - [ ] 实现部署和运维工具
   - [ ] 实现测试和验证框架

### 13.2 技术决策建议

1. **HTTP框架**：推荐使用Gin（轻量级，性能好）
2. **MQ客户端**：
   - RabbitMQ：使用amqp库
   - Kafka：使用sarama库
3. **分布式锁**：基于etcd实现，避免引入Redis锁的复杂性
4. **任务队列**：Redis Stream（MVP阶段），后续可扩展到RabbitMQ

### 13.3 风险控制

1. **分布式锁竞争**：设置合理的锁TTL和重试策略
2. **MQ消息积压**：监控队列长度，动态调整消费者数量
3. **HTTP服务超时**：合理设置同步和异步模式的超时时间
4. **配置更新冲突**：使用etcd的乐观锁机制

---

本设计文档完整地实现了基于你对触发器的正确理解，将流程开放为API供业务方调用，同时支持三种触发器的不同执行模式。设计考虑了性能、可靠性、可观测性和扩展性，为实际开发提供了完整的指导。