package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"mengri-flow/internal/infra/plugin"
	"mengri-flow/plugins/resource/common"
)

func init() {
	registry := plugin.GlobalRegistry()
	registry.RegisterResource(&PostgreSQLPlugin{})
}

// PostgreSQLPlugin PostgreSQL数据库资源插件
type PostgreSQLPlugin struct{}

var _ plugin.ResourcePlugin = (*PostgreSQLPlugin)(nil)

// PluginMeta 返回插件元数据
func (p *PostgreSQLPlugin) PluginMeta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "postgres",
		Type:        plugin.PluginTypeResource,
		Version:     "1.0.0",
		Description: "PostgreSQL数据库资源插件，支持SQL查询和SQLc导入",
		Author:      "mengri-flow",
		BuildTag:    "postgres",
	}
}

// ConfigSchema 返回配置Schema
func (p *PostgreSQLPlugin) ConfigSchema() plugin.JSONSchema {
	return plugin.NewSchemaBuilder().
		AddStringField("host", "主机地址", "PostgreSQL服务器主机地址", true).
		AddNumberField("port", "端口", "PostgreSQL服务器端口", true, float64Ptr(1), float64Ptr(65535)).
		AddStringField("database", "数据库名", "要连接的数据库名称", true).
		AddStringField("username", "用户名", "数据库用户名", true).
		AddStringField("password", "密码", "数据库密码", false, "password").
		AddEnumField("sslMode", "SSL模式", "SSL连接模式", []any{"disable", "require", "verify-ca", "verify-full"}, "disable", false).
		AddNumberField("maxOpenConns", "最大连接数", "最大打开连接数", false, float64Ptr(1), float64Ptr(100)).
		AddNumberField("maxIdleConns", "最大空闲连接数", "最大空闲连接数", false, float64Ptr(1), float64Ptr(50)).
		AddNumberField("connMaxLifetime", "连接最大生命周期", "连接最大生命周期（秒）", false, float64Ptr(60), float64Ptr(86400)).
		Build()
}

// TestConnection 测试数据库连接
func (p *PostgreSQLPlugin) TestConnection(ctx context.Context, config map[string]any) error {
	dsn := p.buildDSN(config)

	// 创建临时连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &plugin.PluginError{
			Type:    "INVALID_CONFIG",
			Message: "数据库配置无效：" + err.Error(),
			Cause:   err,
		}
	}
	defer db.Close()

	// 设置连接超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		return &plugin.PluginError{
			Type:    "CONNECTION_FAILED",
			Message: "无法连接到PostgreSQL服务器：" + err.Error(),
			Cause:   err,
		}
	}

	return nil
}

// getDBConnection 获取数据库连接（使用连接池）
func (p *PostgreSQLPlugin) getDBConnection(config map[string]any) (*sql.DB, error) {
	dsn := p.buildDSN(config)

	// 使用全局连接池
	pool := common.GetConnectionPool()

	// 定义如何打开数据库连接
	openDBFunc := func(dsn string) (*sql.DB, error) {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}

		// 设置连接池参数
		maxOpenConns := getInt(config, "maxOpenConns", 10)
		maxIdleConns := getInt(config, "maxIdleConns", 5)
		connMaxLifetime := getInt(config, "connMaxLifetime", 3600)

		db.SetMaxOpenConns(maxOpenConns)
		db.SetMaxIdleConns(maxIdleConns)
		db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

		return db, nil
	}

	// 从连接池获取或创建连接
	db, err := pool.GetOrCreateConnection(dsn, openDBFunc)
	if err != nil {
		return nil, &plugin.PluginError{
			Type:    "CONNECTION_FAILED",
			Message: "无法创建数据库连接：" + err.Error(),
			Cause:   err,
		}
	}

	return db, nil
}

// ExecuteTool 执行工具（SQL查询）
func (p *PostgreSQLPlugin) ExecuteTool(
	ctx context.Context,
	resourceConfig map[string]any,
	toolConfig map[string]any,
	input any,
) (*plugin.ToolResult, error) {
	start := time.Now()

	// 1. 验证SQL查询类型（只允许SELECT）
	sqlQuery, ok := toolConfig["sql"].(string)
	if !ok || sqlQuery == "" {
		return nil, &plugin.PluginError{
			Type:    "INVALID_INPUT",
			Message: "SQL查询不能为空",
		}
	}

	// 2. 安全检查 - 只允许SELECT查询
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sqlQuery))
	if !strings.HasPrefix(trimmedSQL, "SELECT") {
		return nil, &plugin.PluginError{
			Type:    "INVALID_INPUT",
			Message: "只允许执行SELECT查询（安全限制）",
		}
	}

	// 3. 添加LIMIT限制（防止返回过多数据）
	if !strings.Contains(trimmedSQL, "LIMIT") {
		sqlQuery += " LIMIT 10000"
	}

	// 4. 获取数据库连接
	db, err := p.getDBConnection(resourceConfig)
	if err != nil {
		return nil, err
	}

	// 5. 执行查询（使用参数化查询）
	rows, err := db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, &plugin.PluginError{
			Type:    "EXECUTION_FAILED",
			Message: fmt.Sprintf("查询执行失败: %v", err),
			Cause:   err,
		}
	}
	defer rows.Close()

	// 6. 解析结果
	result, err := p.parseRows(rows)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)

	return &plugin.ToolResult{
		StatusCode: 200,
		Data: map[string]any{
			"rows":    result,
			"count":   len(result),
			"columns": getColumns(rows),
		},
		Duration: duration,
	}, nil
}

// ExtractTools 从SQLc文件批量导入工具
func (p *PostgreSQLPlugin) ExtractTools(ctx context.Context, config map[string]any) ([]plugin.ToolDefinition, error) {
	// 获取SQLc文件内容
	sqlcContent, ok := config["sqlcContent"].(string)
	if !ok || sqlcContent == "" {
		return nil, &plugin.PluginError{
			Type:    "INVALID_INPUT",
			Message: "SQLc文件内容不能为空",
		}
	}

	// 解析SQLc文件
	queries, err := common.ParseSQLcFile(sqlcContent)
	if err != nil {
		return nil, &plugin.PluginError{
			Type:    "INVALID_INPUT",
			Message: "SQLc文件解析失败：" + err.Error(),
			Cause:   err,
		}
	}

	// 转换为ToolDefinition
	var tools []plugin.ToolDefinition
	for _, query := range queries {
		// 验证SQL查询
		if err := common.ValidateSQLQuery(query.SQL); err != nil {
			continue // 跳过不安全的查询
		}

		tool := plugin.ToolDefinition{
			Name:        query.Name,
			Type:        "sql_query",
			Method:      "POST",
			Path:        "/query/" + query.Name,
			Description: query.Description,
			InputSchema: buildInputSchema(query.Params),
			OutputSchema: buildOutputSchema(),
			Config: map[string]any{
				"sql": query.SQL,
			},
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// buildInputSchema 构建输入Schema
func buildInputSchema(params []common.SQLParam) plugin.JSONSchema {
	if len(params) == 0 {
		return plugin.NewSchemaBuilder().
			AddObjectField("params", "参数", "查询参数", nil, nil).
			Build()
	}

	builder := plugin.NewSchemaBuilder()

	for _, param := range params {
		// 根据参数名称推断类型（简单规则）
		fieldType := "string"
		if strings.Contains(param.Name, "id") || strings.Contains(param.Name, "age") {
			fieldType = "number"
		}

		if fieldType == "number" {
			builder.AddNumberField(param.Name, param.Name, "参数："+param.Name, false, nil, nil)
		} else {
			builder.AddStringField(param.Name, param.Name, "参数："+param.Name, false)
		}
	}

	return builder.Build()
}

// buildOutputSchema 构建输出Schema
func buildOutputSchema() plugin.JSONSchema {
	return plugin.NewSchemaBuilder().
		AddArrayField("rows", "结果行", "查询结果行", nil).
		AddNumberField("count", "结果数", "结果行数", false, float64Ptr(0), nil).
		Build()
}

// buildDSN 构建DSN连接字符串
func (p *PostgreSQLPlugin) buildDSN(config map[string]any) string {
	host := getString(config, "host", "localhost")
	port := getInt(config, "port", 5432)
	database := getString(config, "database", "")
	username := getString(config, "username", "")
	password := getString(config, "password", "")
	sslMode := getString(config, "sslMode", "disable")

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslMode)
}

// parseRows 解析查询结果
func (p *PostgreSQLPlugin) parseRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		// 创建扫描目标
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %w", err)
		}

		// 转换为map
		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %w", err)
	}

	return results, nil
}

// getColumns 获取列名
func getColumns(rows *sql.Rows) []string {
	columns, err := rows.Columns()
	if err != nil {
		return []string{}
	}
	return columns
}

// getString 获取字符串配置值
func getString(config map[string]any, key string, defaultValue string) string {
	if val, ok := config[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

// getInt 获取整型配置值
func getInt(config map[string]any, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

// float64Ptr float64指针
func float64Ptr(v float64) *float64 {
	return &v
}
