package datagen

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Config holds configuration for data generation
type Config struct {
	Count       int
	Format      string
	Output      string
	Schema      string
	Template    string
	Rate        int
	Duration    time.Duration
	ShadowMode  bool
	Seed        int64
	Concurrent  int
}

// Generator manages data generation
type Generator struct {
	Config Config
	Rand   *rand.Rand
}

// DataTemplate defines the structure for generating data
type DataTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Fields      map[string]FieldConfig `json:"fields"`
	Relations   []Relation             `json:"relations,omitempty"`
}

// FieldConfig defines how to generate a field
type FieldConfig struct {
	Type        string      `json:"type"`
	Min         interface{} `json:"min,omitempty"`
	Max         interface{} `json:"max,omitempty"`
	Values      []string    `json:"values,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Format      string      `json:"format,omitempty"`
	Nullable    bool        `json:"nullable,omitempty"`
	NullChance  float64     `json:"null_chance,omitempty"`
	Reference   string      `json:"reference,omitempty"`
	Distribution string     `json:"distribution,omitempty"`
}

// Relation defines relationships between data entities
type Relation struct {
	Type       string `json:"type"`
	Target     string `json:"target"`
	Field      string `json:"field"`
	Cardinality string `json:"cardinality"`
}

// DefaultConfig returns default data generation configuration
func DefaultConfig() Config {
	return Config{
		Count:      100,
		Format:     "json",
		Output:     "-",
		Rate:       10,
		Duration:   1 * time.Minute,
		ShadowMode: false,
		Seed:       time.Now().UnixNano(),
		Concurrent: 1,
	}
}

// Run executes the data generator
func Run(args []string) error {
	config := DefaultConfig()
	
	// Parse arguments
	for i, arg := range args {
		switch arg {
		case "--count", "-c":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					config.Count = count
				}
			}
		case "--format", "-f":
			if i+1 < len(args) {
				config.Format = args[i+1]
			}
		case "--output", "-o":
			if i+1 < len(args) {
				config.Output = args[i+1]
			}
		case "--schema", "-s":
			if i+1 < len(args) {
				config.Schema = args[i+1]
			}
		case "--template", "-t":
			if i+1 < len(args) {
				config.Template = args[i+1]
			}
		case "--rate", "-r":
			if i+1 < len(args) {
				if rate, err := strconv.Atoi(args[i+1]); err == nil {
					config.Rate = rate
				}
			}
		case "--duration", "-d":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					config.Duration = duration
				}
			}
		case "--shadow":
			config.ShadowMode = true
		case "--seed":
			if i+1 < len(args) {
				if seed, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					config.Seed = seed
				}
			}
		case "--concurrent", "-j":
			if i+1 < len(args) {
				if concurrent, err := strconv.Atoi(args[i+1]); err == nil {
					config.Concurrent = concurrent
				}
			}
		case "--help", "-h":
			return showHelp()
		}
	}

	generator := NewGenerator(config)
	return generator.Generate()
}

func showHelp() error {
	fmt.Printf(`datagen - Test data generation utility

Usage: datagen [options]

Options:
  -c, --count       Number of records to generate (default: 100)
  -f, --format      Output format (json, csv, sql, kafka) (default: json)
  -o, --output      Output file (default: stdout)
  -s, --schema      JSON schema file for validation
  -t, --template    Data template file for structure
  -r, --rate        Records per second for streaming (default: 10)
  -d, --duration    Duration for streaming mode (default: 1m)
      --shadow      Enable shadow traffic mode
      --seed        Random seed for reproducible data
  -j, --concurrent  Number of concurrent generators (default: 1)
  -h, --help        Show this help message

Data Types:
  - string, int, float, bool, date, time, timestamp
  - email, phone, name, address, company, url
  - uuid, guid, hash, ip, mac, credit_card
  - lorem, paragraph, sentence, word

Examples:
  datagen -c 1000 -f json -o testdata.json
  datagen --shadow -r 50 -d 5m
  datagen -t user_template.json -c 500
  echo '{"name": "string", "age": "int"}' | datagen -t -
`)
	return nil
}

// NewGenerator creates a new data generator instance
func NewGenerator(config Config) *Generator {
	return &Generator{
		Config: config,
		Rand:   rand.New(rand.NewSource(config.Seed)),
	}
}

// Generate creates the specified data
func (g *Generator) Generate() error {
	var template *DataTemplate
	var err error

	// Load template if specified
	if g.Config.Template != "" {
		template, err = g.loadTemplate()
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}
	} else {
		// Use default template
		template = g.getDefaultTemplate()
	}

	if g.Config.ShadowMode {
		return g.generateShadowTraffic(template)
	}

	return g.generateBatch(template)
}

func (g *Generator) loadTemplate() (*DataTemplate, error) {
	var reader *os.File
	var err error

	if g.Config.Template == "-" {
		reader = os.Stdin
	} else {
		reader, err = os.Open(g.Config.Template)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	}

	var template DataTemplate
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&template); err != nil {
		return nil, err
	}

	return &template, nil
}

func (g *Generator) getDefaultTemplate() *DataTemplate {
	return &DataTemplate{
		Name:        "default",
		Description: "Default data template",
		Fields: map[string]FieldConfig{
			"id":         {Type: "uuid"},
			"name":       {Type: "name"},
			"email":      {Type: "email"},
			"age":        {Type: "int", Min: 18, Max: 80},
			"created_at": {Type: "timestamp"},
			"active":     {Type: "bool"},
		},
	}
}

func (g *Generator) generateBatch(template *DataTemplate) error {
	var output *os.File
	var err error

	if g.Config.Output == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(g.Config.Output)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	switch g.Config.Format {
	case "json":
		return g.generateJSONBatch(template, output)
	case "csv":
		return g.generateCSVBatch(template, output)
	case "sql":
		return g.generateSQLBatch(template, output)
	default:
		return fmt.Errorf("unsupported format: %s", g.Config.Format)
	}
}

func (g *Generator) generateJSONBatch(template *DataTemplate, output *os.File) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	var records []map[string]interface{}
	
	for i := 0; i < g.Config.Count; i++ {
		record := g.generateRecord(template)
		records = append(records, record)
	}

	return encoder.Encode(records)
}

func (g *Generator) generateCSVBatch(template *DataTemplate, output *os.File) error {
	// Get field names in consistent order
	fieldNames := make([]string, 0, len(template.Fields))
	for name := range template.Fields {
		fieldNames = append(fieldNames, name)
	}

	// Write header
	for i, name := range fieldNames {
		if i > 0 {
			output.WriteString(",")
		}
		output.WriteString(name)
	}
	output.WriteString("\n")

	// Write records
	for i := 0; i < g.Config.Count; i++ {
		record := g.generateRecord(template)
		
		for j, name := range fieldNames {
			if j > 0 {
				output.WriteString(",")
			}
			
			value := record[name]
			if value != nil {
				output.WriteString(fmt.Sprintf("%v", value))
			}
		}
		output.WriteString("\n")
	}

	return nil
}

func (g *Generator) generateSQLBatch(template *DataTemplate, output *os.File) error {
	tableName := template.Name
	if tableName == "" {
		tableName = "generated_data"
	}

	// Write table creation
	output.WriteString(fmt.Sprintf("-- Generated data for table: %s\n", tableName))
	output.WriteString(fmt.Sprintf("-- Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	// Generate INSERT statements
	for i := 0; i < g.Config.Count; i++ {
		record := g.generateRecord(template)
		
		output.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
		
		// Write column names
		fieldNames := make([]string, 0, len(record))
		for name := range record {
			fieldNames = append(fieldNames, name)
		}
		
		for j, name := range fieldNames {
			if j > 0 {
				output.WriteString(", ")
			}
			output.WriteString(name)
		}
		
		output.WriteString(") VALUES (")
		
		// Write values
		for j, name := range fieldNames {
			if j > 0 {
				output.WriteString(", ")
			}
			
			value := record[name]
			if value == nil {
				output.WriteString("NULL")
			} else if str, ok := value.(string); ok {
				output.WriteString(fmt.Sprintf("'%s'", str))
			} else {
				output.WriteString(fmt.Sprintf("%v", value))
			}
		}
		
		output.WriteString(");\n")
	}

	return nil
}

func (g *Generator) generateRecord(template *DataTemplate) map[string]interface{} {
	record := make(map[string]interface{})
	
	for fieldName, fieldConfig := range template.Fields {
		value := g.generateFieldValue(fieldConfig)
		record[fieldName] = value
	}
	
	return record
}

func (g *Generator) generateFieldValue(config FieldConfig) interface{} {
	// Handle null values
	if config.Nullable && g.Rand.Float64() < config.NullChance {
		return nil
	}
	
	switch config.Type {
	case "string":
		return g.generateString(config)
	case "int":
		return g.generateInt(config)
	case "float":
		return g.generateFloat(config)
	case "bool":
		return g.Rand.Intn(2) == 1
	case "uuid":
		return g.generateUUID()
	case "name":
		return g.generateName()
	case "email":
		return g.generateEmail()
	case "phone":
		return g.generatePhone()
	case "address":
		return g.generateAddress()
	case "company":
		return g.generateCompany()
	case "url":
		return g.generateURL()
	case "timestamp":
		return g.generateTimestamp()
	case "date":
		return g.generateDate()
	case "time":
		return g.generateTime()
	case "lorem":
		return g.generateLorem(config)
	case "ip":
		return g.generateIP()
	case "mac":
		return g.generateMAC()
	default:
		return g.generateString(config)
	}
}

func (g *Generator) generateShadowTraffic(template *DataTemplate) error {
	fmt.Printf("Starting shadow traffic generation: %d records/sec for %v\n", 
		g.Config.Rate, g.Config.Duration)
	
	ticker := time.NewTicker(time.Second / time.Duration(g.Config.Rate))
	defer ticker.Stop()
	
	timeout := time.After(g.Config.Duration)
	count := 0
	
	for {
		select {
		case <-ticker.C:
			record := g.generateRecord(template)
			
			// Output the record (could be sent to Kafka, HTTP, etc.)
			if data, err := json.Marshal(record); err == nil {
				fmt.Printf("[%d] %s\n", count, string(data))
			}
			
			count++
			
		case <-timeout:
			fmt.Printf("Shadow traffic generation completed. Generated %d records.\n", count)
			return nil
		}
	}
}