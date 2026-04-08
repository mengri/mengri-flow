package common

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// SQLParam 表示SQL查询参数
type SQLParam struct {
	Name     string
	Type     string
	Position int
}

// SQLQuery 表示一个SQLc命名查询
type SQLQuery struct {
	Name        string
	SQL         string
	Params      []SQLParam
	Description string
}

// ParseSQLcFile 解析SQLc文件，提取命名查询
func ParseSQLcFile(content string) ([]SQLQuery, error) {
	var queries []SQLQuery
	var currentQuery *SQLQuery

	scanner := bufio.NewScanner(strings.NewReader(content))
	var sqlLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 检查是否是命名查询的开始
		if strings.HasPrefix(line, "-- name:") {
			// 如果之前有查询，先保存
			if currentQuery != nil && len(sqlLines) > 0 {
				currentQuery.SQL = strings.TrimSpace(strings.Join(sqlLines, " "))
				queries = append(queries, *currentQuery)
				sqlLines = []string{}
			}

			// 解析新的查询定义
			name, params, desc := parseNameLine(line)
			currentQuery = &SQLQuery{
				Name:        name,
				Description: desc,
				Params:      params,
			}
			continue
		}

		// 如果是SQL行且当前有活跃查询
		if currentQuery != nil && !strings.HasPrefix(line, "--") && line != "" {
			// 跳过注释行
			if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "*/") {
				continue
			}
			sqlLines = append(sqlLines, line)
		}
	}

	// 处理最后一个查询
	if currentQuery != nil && len(sqlLines) > 0 {
		currentQuery.SQL = strings.TrimSpace(strings.Join(sqlLines, " "))
		queries = append(queries, *currentQuery)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan SQLc file: %w", err)
	}

	// 如果没有解析到任何查询，返回错误
	if len(queries) == 0 {
		return nil, fmt.Errorf("no SQL queries found in the file")
	}

	return queries, nil
}

// parseNameLine 解析 -- name: 行
// 格式: -- name: GetUserByID :one
// 或: -- name: GetUserByID :many
// 或: -- name: GetUserByID
func parseNameLine(line string) (string, []SQLParam, string) {
	// 移除 -- name: 前缀
	line = strings.TrimPrefix(line, "-- name:")
	line = strings.TrimSpace(line)

	// 按空格分割
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", nil, ""
	}

	// 第一个部分是查询名称
	name := parts[0]

	// 解析参数（如果存在:one/:many标签）
	var params []SQLParam
	for i := 1; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], ":") {
			// 返回类型标签，忽略
			continue
		}
	}

	// 尝试从名称中提取参数（如GetUserByIDAndName -> user_id, name）
	params = extractParamsFromName(name)

	return name, params, ""
}

// extractParamsFromName 从查询名称提取参数
func extractParamsFromName(name string) []SQLParam {
	var params []SQLParam

	// 简单的启发式方法：查找By、And、Or后面的部分
	patterns := []string{"By", "And", "Or"}
	nameUpper := strings.ToUpper(name)

	for _, pattern := range patterns {
		idx := strings.Index(nameUpper, pattern)
		if idx >= 0 && idx+len(pattern) < len(name) {
			remaining := name[idx+len(pattern):]

			// 按大写字母分割
			re := regexp.MustCompile(`([A-Z][^A-Z]*)`)
			matches := re.FindAllString(remaining, -1)

			for i, match := range matches {
				paramName := strings.ToLower(match)
				params = append(params, SQLParam{
					Name:     paramName,
					Type:     "string", // 默认类型
					Position: i,
				})
			}
		}
	}

	return params
}

// ExtractSQLParams 从SQL语句中提取参数（:param格式）
func ExtractSQLParams(sql string) []SQLParam {
	var params []SQLParam

	// 正则表达式匹配 :param_name
	re := regexp.MustCompile(`:[a-zA-Z_][a-zA-Z0-9_]*`)
	matches := re.FindAllString(sql, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			paramName := strings.TrimPrefix(match, ":")
			params = append(params, SQLParam{
				Name:     paramName,
				Type:     "string", // 默认类型，实际使用时需要根据上下文推断
				Position: len(params),
			})
		}
	}

	return params
}

// ConvertToToolDefinition 将SQLQuery转换为ToolDefinition
func ConvertToToolDefinition(query SQLQuery) interface{} {
	// 返回一个可以转换为plugin.ToolDefinition的map
	return map[string]interface{}{
		"name":        query.Name,
		"description": query.Description,
		"sql":         query.SQL,
		"params":      query.Params,
	}
}

// ValidateSQLQuery 验证SQL查询
func ValidateSQLQuery(query string) error {
	// 只允许SELECT查询
	upperQuery := strings.TrimSpace(strings.ToUpper(query))

	if !strings.HasPrefix(upperQuery, "SELECT") {
		return fmt.Errorf("only SELECT queries are allowed for security reasons")
	}

	// 检查危险关键字
	dangerousKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "TRUNCATE", "ALTER",
		"CREATE", "GRANT", "REVOKE", "EXECUTE", "CALL",
	}

	for _, keyword := range dangerousKeywords {
		// 检查是否有这些关键字（不在字符串中的）
		if containsKeyword(upperQuery, keyword) {
			return fmt.Errorf("dangerous SQL keyword '%s' is not allowed", keyword)
		}
	}

	return nil
}

// containsKeyword 检查SQL中是否包含关键字（简单实现）
func containsKeyword(sql, keyword string) bool {
	// 移除字符串字面量，避免误匹配
	cleanSQL := removeStringLiterals(sql)
	return strings.Contains(cleanSQL, keyword)
}

// removeStringLiterals 从SQL中移除字符串字面量
func removeStringLiterals(sql string) string {
	// 简单的字符串替换，实际使用可能需要更复杂的解析
	re := regexp.MustCompile(`'[^']*'`)
	return re.ReplaceAllString(sql, "''")
}

// ExampleSQLcFile 返回一个示例SQLc文件内容，用于测试
func ExampleSQLcFile() string {
	return `-- name: GetUserByID :one
-- Get user by ID
SELECT id, name, email, created_at FROM users WHERE id = ? LIMIT 1;

-- name: GetUsersByAge :many
-- Get users by age range
SELECT id, name, email, age FROM users WHERE age >= ? AND age <= ? ORDER BY age;

-- name: CountUsers :one
-- Count total users
SELECT COUNT(*) as total FROM users;
`
}
