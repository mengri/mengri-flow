package grpc

import (
	"context"

	"mengri-flow/internal/infra/plugin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterResource(&GRPCPlugin{})
}

// GRPCPlugin gRPC资源插件
type GRPCPlugin struct{}

var _ plugin.ResourcePlugin = (*GRPCPlugin)(nil)

// PluginMeta 返回插件元数据
func (p *GRPCPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "grpc",
		Type:        plugin.PluginTypeResource,
		Version:     "1.0.0",
		Description: "gRPC资源插件",
		Author:      "mengri-flow",
		BuildTag:    "resource_grpc",
	}
}

// ConfigSchema 返回配置Schema
func (p *GRPCPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.NewSchemaBuilder().
		AddStringField("endpoint", "服务端点", "gRPC服务地址，如 grpc.example.com:443", true).
		AddBooleanField("tls", "启用TLS", "是否启用TLS加密", false, true).
		AddStringField("tlsCert", "TLS证书", "TLS证书内容（PEM格式）", false).
		Build()
}

// TestConnection 测试连接
func (p *GRPCPlugin) TestConnection(ctx context.Context, config map[string]any) error {
	conn, err := p.createConnection(ctx, config)
	if err != nil {
		return plugin.NewPluginError("connection_failed", "连接失败", err)
	}
	defer conn.Close()

	// 测试连接状态
	if conn.GetState() == connectivity.Connecting {
		return plugin.ErrConnectionFailed
	}

	return nil
}

// ExecuteTool 执行gRPC工具
func (p *GRPCPlugin) ExecuteTool(
	ctx context.Context,
	resourceConfig map[string]any,
	toolConfig map[string]any,
	input any,
) (*plugin.ToolResult, error) {
	conn, err := p.createConnection(ctx, resourceConfig)
	if err != nil {
		return nil, plugin.NewPluginError("connection_failed", "创建连接失败", err)
	}
	defer conn.Close()

	// TODO: 实现gRPC调用逻辑
	// 这需要使用gRPC反射或预生成的客户端代码

	return &plugin.ToolResult{
		StatusCode: 200,
		Data:       map[string]any{"result": "success"},
		Headers:    map[string]string{},
		Duration:   100,
		Error:      "",
	}, nil
}

// ExtractTools 提取工具定义
func (p *GRPCPlugin) ExtractTools(ctx context.Context, config map[string]any) ([]plugin.ToolDefinition, error) {
	// gRPC插件不自动提取工具，需要手动配置或使用反射
	return []plugin.ToolDefinition{}, nil
}

// createConnection 创建gRPC连接
func (p *GRPCPlugin) createConnection(ctx context.Context, config map[string]any) (*grpc.ClientConn, error) {
	enabledTLS, _ := config["tls"].(bool)
	endpoint, _ := config["endpoint"].(string)

	var opts []grpc.DialOption

	if enabledTLS {
		certData, _ := config["tlsCert"].(string)
		if certData != "" {
			// 使用自定义证书
			creds, err := credentials.NewServerTLSFromFile(certData, "")
			if err != nil {
				return nil, err
			}
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			// 使用系统证书
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.DialContext(ctx, endpoint, opts...)
}
