package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mengri-flow/internal/executor"
	"mengri-flow/internal/infra/config"
)

func main() {
	// 1. 解析命令行参数
	var (
		etcdEndpoints = flag.String("etcd-endpoints", "localhost:2379", "etcd endpoints")
		etcdUsername  = flag.String("etcd-username", "", "etcd username")
		etcdPassword  = flag.String("etcd-password", "", "etcd password")
		clusterID     = flag.String("cluster-id", "", "cluster ID")
		nodeID        = flag.String("node-id", "", "executor node ID (auto-generated if empty)")
		logLevel      = flag.String("log-level", "info", "log level")
	)
	flag.Parse()

	if *clusterID == "" {
		log.Fatal("cluster-id is required")
	}

	// 2. 生成Node ID（如果不指定）
	if *nodeID == "" {
		*nodeID = generateNodeID()
		log.Printf("Auto-generated node ID: %s", *nodeID)
	}

	// 3. 加载配置
	cfg := &config.ExecutorConfig{
		EtcdEndpoints: *etcdEndpoints,
		EtcdUsername:  *etcdUsername,
		EtcdPassword:  *etcdPassword,
		ClusterID:     *clusterID,
		NodeID:        *nodeID,
		LogLevel:      *logLevel,
	}

	// 4. 初始化Executor
	exec := executor.NewExecutor(cfg)

	// 5. 启动
	ctx := context.Background()
	if err := exec.Start(ctx); err != nil {
		log.Fatal("Failed to start executor:", err)
	}

	log.Printf("Executor %s started successfully for cluster %s", *nodeID, *clusterID)

	// 6. 等待信号
	waitForShutdownSignal()

	// 7. 优雅关闭
	log.Printf("Shutting down executor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := exec.Stop(shutdownCtx); err != nil {
		log.Fatal("Failed to stop executor:", err)
	}

	log.Printf("Executor stopped gracefully")
}

func generateNodeID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

func waitForShutdownSignal() {
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Printf("Received shutdown signal")
}
