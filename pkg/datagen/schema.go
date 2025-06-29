package datagen

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// JSONSchema represents a JSON Schema definition
type JSONSchema struct {
	Type                 string                    `json:"type,omitempty"`
	Properties           map[string]*JSONSchema    `json:"properties,omitempty"`
	Items                *JSONSchema               `json:"items,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Enum                 []interface{}             `json:"enum,omitempty"`
	Minimum              *float64                  `json:"minimum,omitempty"`
	Maximum              *float64                  `json:"maximum,omitempty"`
	MinLength            *int                      `json:"minLength,omitempty"`
	MaxLength            *int                      `json:"maxLength,omitempty"`
	Pattern              string                    `json:"pattern,omitempty"`
	Format               string                    `json:"format,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Default              interface{}               `json:"default,omitempty"`
	ExclusiveMinimum     bool                      `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool                      `json:"exclusiveMaximum,omitempty"`
	MultipleOf           *float64                  `json:"multipleOf,omitempty"`
	MinItems             *int                      `json:"minItems,omitempty"`
	MaxItems             *int                      `json:"maxItems,omitempty"`
	UniqueItems          bool                      `json:"uniqueItems,omitempty"`
	Title                string                    `json:"title,omitempty"`
	Examples             []interface{}             `json:"examples,omitempty"`
	AdditionalProperties interface{}               `json:"additionalProperties,omitempty"`
	Definitions          map[string]*JSONSchema    `json:"definitions,omitempty"`
	Ref                  string                    `json:"$ref,omitempty"`
	AllOf                []*JSONSchema             `json:"allOf,omitempty"`
	AnyOf                []*JSONSchema             `json:"anyOf,omitempty"`
	OneOf                []*JSONSchema             `json:"oneOf,omitempty"`
	Not                  *JSONSchema               `json:"not,omitempty"`
}

// AvroSchema represents an Avro schema definition
type AvroSchema struct {
	Type      interface{}          `json:"type"`
	Name      string               `json:"name,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
	Fields    []AvroField          `json:"fields,omitempty"`
	Items     interface{}          `json:"items,omitempty"`
	Values    interface{}          `json:"values,omitempty"`
	Symbols   []string             `json:"symbols,omitempty"`
	Size      int                  `json:"size,omitempty"`
	LogicalType string             `json:"logicalType,omitempty"`
	Precision int                  `json:"precision,omitempty"`
	Scale     int                  `json:"scale,omitempty"`
}

// AvroField represents a field in an Avro record schema
type AvroField struct {
	Name    string      `json:"name"`
	Type    interface{} `json:"type"`
	Doc     string      `json:"doc,omitempty"`
	Default interface{} `json:"default,omitempty"`
	Order   string      `json:"order,omitempty"`
	Aliases []string    `json:"aliases,omitempty"`
}

// SchemaConverter converts schemas to DataTemplate
type SchemaConverter struct {
	Config Config
}

// NewSchemaConverter creates a new schema converter
func NewSchemaConverter(config Config) *SchemaConverter {
	return &SchemaConverter{Config: config}
}

// ConvertJSONSchema converts a JSON Schema to DataTemplate
func (sc *SchemaConverter) ConvertJSONSchema(schemaPath string) (*DataTemplate, error) {
	schema, err := sc.loadJSONSchema(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON schema: %w", err)
	}

	template := &DataTemplate{
		Name:        schema.Title,
		Description: schema.Description,
		Fields:      make(map[string]FieldConfig),
	}

	if template.Name == "" {
		template.Name = "generated_from_json_schema"
	}

	if schema.Type == "object" && schema.Properties != nil {
		for propName, propSchema := range schema.Properties {
			fieldConfig := sc.convertJSONSchemaProperty(propSchema)
			
			// Check if field is required
			for _, required := range schema.Required {
				if required == propName {
					fieldConfig.Nullable = false
					break
				}
			}

			template.Fields[propName] = fieldConfig
		}
	}

	return template, nil
}

func (sc *SchemaConverter) loadJSONSchema(schemaPath string) (*JSONSchema, error) {
	file, err := os.Open(schemaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var schema JSONSchema
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (sc *SchemaConverter) convertJSONSchemaProperty(schema *JSONSchema) FieldConfig {
	config := FieldConfig{
		Nullable:   true,
		NullChance: 0.1,
	}

	// Handle enum values
	if len(schema.Enum) > 0 {
		config.Type = "string"
		config.Values = make([]string, len(schema.Enum))
		for i, val := range schema.Enum {
			config.Values[i] = fmt.Sprintf("%v", val)
		}
		return config
	}

	// Handle different types
	switch schema.Type {
	case "string":
		config.Type = sc.mapStringFormat(schema.Format)
		if schema.Pattern != "" {
			config.Pattern = schema.Pattern
		}
		if schema.MinLength != nil {
			config.Min = float64(*schema.MinLength)
		}
		if schema.MaxLength != nil {
			config.Max = float64(*schema.MaxLength)
		}

	case "integer":
		config.Type = "int"
		if schema.Minimum != nil {
			config.Min = *schema.Minimum
		}
		if schema.Maximum != nil {
			config.Max = *schema.Maximum
		}

	case "number":
		config.Type = "float"
		if schema.Minimum != nil {
			config.Min = *schema.Minimum
		}
		if schema.Maximum != nil {
			config.Max = *schema.Maximum
		}

	case "boolean":
		config.Type = "bool"

	case "array":
		config.Type = "array"
		// For now, generate a simple string array
		config.Min = 1.0
		config.Max = 5.0

	case "object":
		config.Type = "object"

	default:
		config.Type = "string"
	}

	// Set default value if provided
	if schema.Default != nil {
		config.Values = []string{fmt.Sprintf("%v", schema.Default)}
	}

	return config
}

func (sc *SchemaConverter) mapStringFormat(format string) string {
	switch format {
	case "date":
		return "date"
	case "date-time":
		return "timestamp"
	case "time":
		return "time"
	case "email":
		return "email"
	case "uri", "url":
		return "url"
	case "uuid":
		return "uuid"
	case "ipv4", "ipv6":
		return "ip"
	case "hostname":
		return "string"
	case "byte":
		return "string"
	case "binary":
		return "string"
	case "password":
		return "string"
	default:
		return "string"
	}
}

// ConvertAvroSchema converts an Avro schema to DataTemplate
func (sc *SchemaConverter) ConvertAvroSchema(schemaPath string) (*DataTemplate, error) {
	schema, err := sc.loadAvroSchema(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load Avro schema: %w", err)
	}

	template := &DataTemplate{
		Name:        schema.Name,
		Description: fmt.Sprintf("Generated from Avro schema: %s", schema.Name),
		Fields:      make(map[string]FieldConfig),
	}

	if template.Name == "" {
		template.Name = "generated_from_avro_schema"
	}

	// Handle record type
	for _, field := range schema.Fields {
		fieldConfig := sc.convertAvroField(field)
		template.Fields[field.Name] = fieldConfig
	}

	return template, nil
}

func (sc *SchemaConverter) loadAvroSchema(schemaPath string) (*AvroSchema, error) {
	file, err := os.Open(schemaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var schema AvroSchema
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (sc *SchemaConverter) convertAvroField(field AvroField) FieldConfig {
	config := FieldConfig{
		Nullable:   false,
		NullChance: 0.0,
	}

	// Handle union types (e.g., ["null", "string"])
	if fieldType := sc.extractAvroType(field.Type); fieldType != "" {
		config.Type = sc.mapAvroType(fieldType)
		
		// Check if null is allowed in union
		if sc.isNullableAvroType(field.Type) {
			config.Nullable = true
			config.NullChance = 0.1
		}
	}

	// Set default value if provided
	if field.Default != nil {
		config.Values = []string{fmt.Sprintf("%v", field.Default)}
	}

	return config
}

func (sc *SchemaConverter) extractAvroType(typeField interface{}) string {
	switch t := typeField.(type) {
	case string:
		return t
	case []interface{}:
		// Union type, find the non-null type
		for _, unionType := range t {
			if str, ok := unionType.(string); ok && str != "null" {
				return str
			}
		}
	case map[string]interface{}:
		// Complex type
		if typeStr, ok := t["type"].(string); ok {
			return typeStr
		}
	}
	return "string"
}

func (sc *SchemaConverter) isNullableAvroType(typeField interface{}) bool {
	if unionTypes, ok := typeField.([]interface{}); ok {
		for _, unionType := range unionTypes {
			if str, ok := unionType.(string); ok && str == "null" {
				return true
			}
		}
	}
	return false
}

func (sc *SchemaConverter) mapAvroType(avroType string) string {
	switch avroType {
	case "null":
		return "string"
	case "boolean":
		return "bool"
	case "int", "long":
		return "int"
	case "float", "double":
		return "float"
	case "bytes", "string":
		return "string"
	case "array":
		return "array"
	case "map", "record":
		return "object"
	case "enum":
		return "string"
	case "fixed":
		return "string"
	default:
		return "string"
	}
}

// InferSchemaFromData creates a DataTemplate by analyzing sample data
func (sc *SchemaConverter) InferSchemaFromData(data interface{}) (*DataTemplate, error) {
	template := &DataTemplate{
		Name:        "inferred_schema",
		Description: "Schema inferred from sample data",
		Fields:      make(map[string]FieldConfig),
	}

	switch d := data.(type) {
	case map[string]interface{}:
		for key, value := range d {
			template.Fields[key] = sc.inferFieldFromValue(value)
		}
	case []interface{}:
		if len(d) > 0 {
			if firstObj, ok := d[0].(map[string]interface{}); ok {
				for key, value := range firstObj {
					template.Fields[key] = sc.inferFieldFromValue(value)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported data type for schema inference: %T", data)
	}

	return template, nil
}

func (sc *SchemaConverter) inferFieldFromValue(value interface{}) FieldConfig {
	config := FieldConfig{
		Nullable:   true,
		NullChance: 0.1,
	}

	if value == nil {
		config.Type = "string"
		config.Nullable = true
		config.NullChance = 0.3
		return config
	}

	switch v := value.(type) {
	case bool:
		config.Type = "bool"
	case int, int32, int64:
		config.Type = "int"
		config.Min = 0.0
		config.Max = 1000.0
	case float32, float64:
		config.Type = "float"
		config.Min = 0.0
		config.Max = 1000.0
	case string:
		config.Type = sc.inferStringType(v)
		config.Min = 5.0
		config.Max = 50.0
	case []interface{}:
		config.Type = "array"
		config.Min = 1.0
		config.Max = 5.0
	case map[string]interface{}:
		config.Type = "object"
	default:
		config.Type = "string"
	}

	return config
}

func (sc *SchemaConverter) inferStringType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))

	// Email pattern
	if strings.Contains(value, "@") && strings.Contains(value, ".") {
		return "email"
	}

	// URL pattern
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return "url"
	}

	// UUID pattern (simplified)
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		return "uuid"
	}

	// Date patterns
	if strings.Contains(value, "-") && len(value) >= 8 {
		// Could be date
		return "date"
	}

	// Phone pattern (simplified)
	if strings.Contains(value, "(") && strings.Contains(value, ")") {
		return "phone"
	}

	// IP address pattern (simplified)
	if strings.Count(value, ".") == 3 {
		return "ip"
	}

	return "string"
}

// GenerateSchemaFromTemplate creates a JSON Schema from a DataTemplate
func (sc *SchemaConverter) GenerateSchemaFromTemplate(template *DataTemplate) (*JSONSchema, error) {
	schema := &JSONSchema{
		Type:        "object",
		Title:       template.Name,
		Description: template.Description,
		Properties:  make(map[string]*JSONSchema),
		Required:    []string{},
	}

	for fieldName, fieldConfig := range template.Fields {
		propSchema := sc.convertFieldToJSONSchema(fieldConfig)
		schema.Properties[fieldName] = propSchema

		// Add to required if not nullable
		if !fieldConfig.Nullable {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	return schema, nil
}

func (sc *SchemaConverter) convertFieldToJSONSchema(config FieldConfig) *JSONSchema {
	schema := &JSONSchema{}

	switch config.Type {
	case "string", "name", "email", "phone", "address", "company", "url", "lorem", "uuid":
		schema.Type = "string"
		if config.Pattern != "" {
			schema.Pattern = config.Pattern
		}
		if config.Min != nil {
			minLen := int(config.Min.(float64))
			schema.MinLength = &minLen
		}
		if config.Max != nil {
			maxLen := int(config.Max.(float64))
			schema.MaxLength = &maxLen
		}
		if len(config.Values) > 0 {
			schema.Enum = make([]interface{}, len(config.Values))
			for i, v := range config.Values {
				schema.Enum[i] = v
			}
		}

	case "int":
		schema.Type = "integer"
		if config.Min != nil {
			min := config.Min.(float64)
			schema.Minimum = &min
		}
		if config.Max != nil {
			max := config.Max.(float64)
			schema.Maximum = &max
		}

	case "float":
		schema.Type = "number"
		if config.Min != nil {
			min := config.Min.(float64)
			schema.Minimum = &min
		}
		if config.Max != nil {
			max := config.Max.(float64)
			schema.Maximum = &max
		}

	case "bool":
		schema.Type = "boolean"

	case "date":
		schema.Type = "string"
		schema.Format = "date"

	case "time":
		schema.Type = "string"
		schema.Format = "time"

	case "timestamp":
		schema.Type = "string"
		schema.Format = "date-time"

	case "array":
		schema.Type = "array"
		schema.Items = &JSONSchema{Type: "string"}

	case "object":
		schema.Type = "object"

	default:
		schema.Type = "string"
	}

	return schema
}

// ValidateDataAgainstSchema validates generated data against a JSON schema
func (sc *SchemaConverter) ValidateDataAgainstSchema(data interface{}, schema *JSONSchema) error {
	// Basic validation implementation
	return sc.validateValue(data, schema)
}

func (sc *SchemaConverter) validateValue(value interface{}, schema *JSONSchema) error {
	if value == nil {
		// Check if null is allowed
		return nil
	}

	valueType := reflect.TypeOf(value).Kind()

	switch schema.Type {
	case "string":
		if valueType != reflect.String {
			return fmt.Errorf("expected string, got %v", valueType)
		}
		str := value.(string)
		if schema.MinLength != nil && len(str) < *schema.MinLength {
			return fmt.Errorf("string too short: %d < %d", len(str), *schema.MinLength)
		}
		if schema.MaxLength != nil && len(str) > *schema.MaxLength {
			return fmt.Errorf("string too long: %d > %d", len(str), *schema.MaxLength)
		}

	case "integer":
		if valueType != reflect.Int && valueType != reflect.Int32 && valueType != reflect.Int64 {
			return fmt.Errorf("expected integer, got %v", valueType)
		}

	case "number":
		if valueType != reflect.Float32 && valueType != reflect.Float64 {
			return fmt.Errorf("expected number, got %v", valueType)
		}

	case "boolean":
		if valueType != reflect.Bool {
			return fmt.Errorf("expected boolean, got %v", valueType)
		}

	case "array":
		if valueType != reflect.Slice {
			return fmt.Errorf("expected array, got %v", valueType)
		}

	case "object":
		if valueType != reflect.Map {
			return fmt.Errorf("expected object, got %v", valueType)
		}
	}

	return nil
}