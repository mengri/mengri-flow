package plugin

// SchemaBuilder JSON Schema构建器
type SchemaBuilder struct {
	schema JSONSchema
}

// NewSchemaBuilder 创建新的Schema构建器
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: make(JSONSchema),
	}
}

// Build 构建最终的Schema
func (b *SchemaBuilder) Build() JSONSchema {
	return b.schema
}

// AddStringField 添加字符串字段
func (b *SchemaBuilder) AddStringField(name, title, description string, required bool, format ...string) *SchemaBuilder {
	field := map[string]any{
		"type":        "string",
		"title":       title,
		"description": description,
	}
	if len(format) > 0 && format[0] != "" {
		field["format"] = format[0]
	}
	b.schema[name] = field
	if required {
		b.addRequired(name)
	}
	return b
}

// AddNumberField 添加数字字段
func (b *SchemaBuilder) AddNumberField(name, title, description string, required bool, min, max *float64) *SchemaBuilder {
	field := map[string]any{
		"type":        "number",
		"title":       title,
		"description": description,
	}
	if min != nil {
		field["minimum"] = *min
	}
	if max != nil {
		field["maximum"] = *max
	}
	b.schema[name] = field
	if required {
		b.addRequired(name)
	}
	return b
}

// AddBooleanField 添加布尔字段
func (b *SchemaBuilder) AddBooleanField(name, title, description string, required bool, defaultValue ...bool) *SchemaBuilder {
	field := map[string]any{
		"type":        "boolean",
		"title":       title,
		"description": description,
	}
	if len(defaultValue) > 0 {
		field["default"] = defaultValue[0]
	}
	b.schema[name] = field
	if required {
		b.addRequired(name)
	}
	return b
}

// AddObjectField 添加对象字段
func (b *SchemaBuilder) AddObjectField(name, title, description string, properties map[string]any, required []string) *SchemaBuilder {
	field := map[string]any{
		"type":        "object",
		"title":       title,
		"description": description,
		"properties":  properties,
	}
	if len(required) > 0 {
		field["required"] = required
	}
	b.schema[name] = field
	return b
}

// AddArrayField 添加数组字段
func (b *SchemaBuilder) AddArrayField(name, title, description string, items map[string]any) *SchemaBuilder {
	field := map[string]any{
		"type":        "array",
		"title":       title,
		"description": description,
		"items":       items,
	}
	b.schema[name] = field
	return b
}

// AddEnumField 添加枚举字段
func (b *SchemaBuilder) AddEnumField(name, title, description string, enum []any, defaultValue any, required bool) *SchemaBuilder {
	field := map[string]any{
		"type":        "string",
		"title":       title,
		"description": description,
		"enum":        enum,
	}
	if defaultValue != nil {
		field["default"] = defaultValue
	}
	b.schema[name] = field
	if required {
		b.addRequired(name)
	}
	return b
}

// addRequired 添加必填字段到required数组
func (b *SchemaBuilder) addRequired(field string) {
	if required, ok := b.schema["required"].([]string); ok {
		b.schema["required"] = append(required, field)
	} else {
		b.schema["required"] = []string{field}
	}
}

// BuildStringSchema 构建字符串类型Schema（辅助函数）
func BuildStringSchema(title, description string, required bool, format ...string) JSONSchema {
	schema := map[string]any{
		"type":        "string",
		"title":       title,
		"description": description,
	}
	if len(format) > 0 && format[0] != "" {
		schema["format"] = format[0]
	}
	return schema
}

// BuildNumberSchema 构建数字类型Schema（辅助函数）
func BuildNumberSchema(title, description string, required bool, min, max *float64) JSONSchema {
	schema := map[string]any{
		"type":        "number",
		"title":       title,
		"description": description,
	}
	if min != nil {
		schema["minimum"] = *min
	}
	if max != nil {
		schema["maximum"] = *max
	}
	return schema
}

// BuildObjectSchema 构建对象类型Schema（辅助函数）
func BuildObjectSchema(title, description string, properties map[string]JSONSchema, required []string) JSONSchema {
	schema := map[string]any{
		"type":        "object",
		"title":       title,
		"description": description,
		"properties":  properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// BuildArraySchema 构建数组类型Schema（辅助函数）
func BuildArraySchema(title, description string, items JSONSchema) JSONSchema {
	return map[string]any{
		"type":        "array",
		"title":       title,
		"description": description,
		"items":       items,
	}
}

// BuildEnumSchema 构建枚举类型Schema（辅助函数）
func BuildEnumSchema(title, description string, enum []any, defaultValue any) JSONSchema {
	schema := map[string]any{
		"type":        "string",
		"title":       title,
		"description": description,
		"enum":        enum,
	}
	if defaultValue != nil {
		schema["default"] = defaultValue
	}
	return schema
}
