package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the main configuration structure
type Config struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Commands    []Command              `yaml:"commands,omitempty"`
	Variables   map[string]string      `yaml:"variables,omitempty"`
	Templates   map[string]string      `yaml:"templates,omitempty"`
	Generators  map[string]interface{} `yaml:"generators,omitempty"`
}

// Command represents a command to execute
type Command struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	WorkDir     string            `yaml:"workdir,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
}

// GeneratorType represents different generator types
type GeneratorType string

const (
	GeneratorDocker     GeneratorType = "docker"
	GeneratorKubernetes GeneratorType = "kubernetes"
	GeneratorHelm       GeneratorType = "helm"
	GeneratorTerraform  GeneratorType = "terraform"
	GeneratorMakefile   GeneratorType = "makefile"
)

// Run is the main entry point for config functionality
func Run(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	command := args[0]
	switch command {
	case "init":
		return initConfig(args[1:])
	case "run":
		return runConfig(args[1:])
	case "generate":
		return generateFromConfig(args[1:])
	case "validate":
		return validateConfig(args[1:])
	case "list":
		return listCommands(args[1:])
	case "-h", "--help":
		return printHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printHelp() error {
	help := `Usage: config <command> [options]

Interactive configuration management and generator tool.

Commands:
  init [name]                 Create a new config file interactively
  run <config-file> [command] Run commands from config file
  generate <type> <config>    Generate manifests from config
  validate <config-file>      Validate configuration file
  list <config-file>          List available commands in config

Generate Types:
  docker                      Generate Dockerfile and docker-compose.yml
  kubernetes                  Generate Kubernetes manifests
  helm                        Generate Helm chart
  terraform                   Generate Terraform configuration
  makefile                    Generate Makefile with recipes

Options:
  -h, --help                  Show this help message

Examples:
  config init myapp
  config run myapp.yaml build
  config generate docker myapp.yaml
  config generate kubernetes myapp.yaml
  config validate myapp.yaml`

	fmt.Println(help)
	return nil
}

func initConfig(args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		name = "dimutils-config"
	}

	fmt.Printf("Creating new configuration: %s\n", name)
	
	config := Config{
		Name:      name,
		Version:   "1.0.0",
		Variables: make(map[string]string),
		Templates: make(map[string]string),
		Generators: make(map[string]interface{}),
	}

	// Interactive prompts
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Print("Description: ")
	if scanner.Scan() {
		config.Description = scanner.Text()
	}

	fmt.Print("Add commands? (y/n): ")
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		config.Commands = []Command{}
		
		for {
			fmt.Print("Command name (or 'done' to finish): ")
			if !scanner.Scan() {
				break
			}
			cmdName := strings.TrimSpace(scanner.Text())
			if cmdName == "done" || cmdName == "" {
				break
			}

			cmd := Command{Name: cmdName}
			
			fmt.Print("Command description: ")
			if scanner.Scan() {
				cmd.Description = scanner.Text()
			}
			
			fmt.Print("Command to execute: ")
			if scanner.Scan() {
				cmd.Command = scanner.Text()
			}
			
			fmt.Print("Working directory (optional): ")
			if scanner.Scan() {
				workdir := strings.TrimSpace(scanner.Text())
				if workdir != "" {
					cmd.WorkDir = workdir
				}
			}

			config.Commands = append(config.Commands, cmd)
		}
	}

	// Add default variables
	config.Variables["timestamp"] = time.Now().Format("2006-01-02T15:04:05Z")
	config.Variables["user"] = os.Getenv("USER")
	if config.Variables["user"] == "" {
		config.Variables["user"] = "unknown"
	}

	// Save to file
	filename := name + ".yaml"
	return saveConfig(&config, filename)
}

func runConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config file required")
	}

	configFile := args[0]
	var targetCommand string
	if len(args) > 1 {
		targetCommand = args[1]
	}

	config, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if targetCommand == "" {
		// Run all commands
		return executeCommands(config.Commands, config.Variables)
	}

	// Run specific command
	for _, cmd := range config.Commands {
		if cmd.Name == targetCommand {
			return executeCommand(&cmd, config.Variables)
		}
	}

	return fmt.Errorf("command '%s' not found in config", targetCommand)
}

func generateFromConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("generator type and config file required")
	}

	genType := GeneratorType(args[0])
	configFile := args[1]

	config, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch genType {
	case GeneratorDocker:
		return generateDocker(config)
	case GeneratorKubernetes:
		return generateKubernetes(config)
	case GeneratorHelm:
		return generateHelm(config)
	case GeneratorTerraform:
		return generateTerraform(config)
	case GeneratorMakefile:
		return generateMakefile(config)
	default:
		return fmt.Errorf("unknown generator type: %s", genType)
	}
}

func validateConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config file required")
	}

	_, err := loadConfig(args[0])
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("Configuration is valid")
	return nil
}

func listCommands(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config file required")
	}

	config, err := loadConfig(args[0])
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Commands in %s:\n", config.Name)
	for _, cmd := range config.Commands {
		fmt.Printf("  %s - %s\n", cmd.Name, cmd.Description)
	}

	return nil
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Configuration saved to %s\n", filename)
	return nil
}

func executeCommands(commands []Command, variables map[string]string) error {
	for _, cmd := range commands {
		if err := executeCommand(&cmd, variables); err != nil {
			return fmt.Errorf("command '%s' failed: %w", cmd.Name, err)
		}
	}
	return nil
}

func executeCommand(cmd *Command, variables map[string]string) error {
	fmt.Printf("Executing: %s\n", cmd.Name)

	// Substitute variables in command
	command := substituteVariables(cmd.Command, variables)
	
	var args []string
	for _, arg := range cmd.Args {
		args = append(args, substituteVariables(arg, variables))
	}

	execCmd := exec.Command(command, args...)
	
	// Set working directory
	if cmd.WorkDir != "" {
		execCmd.Dir = substituteVariables(cmd.WorkDir, variables)
	}

	// Set environment variables
	execCmd.Env = os.Environ()
	for key, value := range cmd.Env {
		execCmd.Env = append(execCmd.Env, key+"="+substituteVariables(value, variables))
	}

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func substituteVariables(text string, variables map[string]string) string {
	for key, value := range variables {
		text = strings.ReplaceAll(text, "{{."+key+"}}", value)
		text = strings.ReplaceAll(text, "${"+key+"}", value)
	}
	return text
}

func generateDocker(config *Config) error {
	dockerfileTemplate := `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o {{.Name}} ./cmd/{{.Name}}

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/{{.Name}} .
CMD ["./{{.Name}}"]
`

	composeTemplate := `version: '3.8'
services:
  {{.Name}}:
    build: .
    container_name: {{.Name}}
    environment:
{{range $key, $value := .Variables}}      - {{$key}}={{$value}}
{{end}}
`

	// Generate Dockerfile
	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return err
	}

	dockerfile, err := os.Create("Dockerfile")
	if err != nil {
		return err
	}
	defer dockerfile.Close()

	if err := tmpl.Execute(dockerfile, config); err != nil {
		return err
	}

	// Generate docker-compose.yml
	tmpl, err = template.New("compose").Parse(composeTemplate)
	if err != nil {
		return err
	}

	composefile, err := os.Create("docker-compose.yml")
	if err != nil {
		return err
	}
	defer composefile.Close()

	if err := tmpl.Execute(composefile, config); err != nil {
		return err
	}

	fmt.Println("Generated Dockerfile and docker-compose.yml")
	return nil
}

func generateKubernetes(config *Config) error {
	deploymentTemplate := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
      - name: {{.Name}}
        image: {{.Name}}:latest
        ports:
        - containerPort: 8080
        env:
{{range $key, $value := .Variables}}        - name: {{$key}}
          value: "{{$value}}"
{{end}}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-service
spec:
  selector:
    app: {{.Name}}
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
`

	if err := os.MkdirAll("k8s", 0755); err != nil {
		return err
	}

	tmpl, err := template.New("k8s").Parse(deploymentTemplate)
	if err != nil {
		return err
	}

	k8sFile, err := os.Create(filepath.Join("k8s", "deployment.yaml"))
	if err != nil {
		return err
	}
	defer k8sFile.Close()

	if err := tmpl.Execute(k8sFile, config); err != nil {
		return err
	}

	fmt.Println("Generated Kubernetes manifests in k8s/")
	return nil
}

func generateHelm(config *Config) error {
	chartTemplate := `apiVersion: v2
name: {{.Name}}
description: {{.Description}}
version: {{.Version}}
appVersion: "{{.Version}}"
`

	valuesTemplate := `replicaCount: 1

image:
  repository: {{.Name}}
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80

env:
{{range $key, $value := .Variables}}  {{$key}}: "{{$value}}"
{{end}}
`

	// Create helm chart structure
	chartDir := fmt.Sprintf("helm-%s", config.Name)
	dirs := []string{
		chartDir,
		filepath.Join(chartDir, "templates"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Generate Chart.yaml
	tmpl, err := template.New("chart").Parse(chartTemplate)
	if err != nil {
		return err
	}

	chartFile, err := os.Create(filepath.Join(chartDir, "Chart.yaml"))
	if err != nil {
		return err
	}
	defer chartFile.Close()

	if err := tmpl.Execute(chartFile, config); err != nil {
		return err
	}

	// Generate values.yaml
	tmpl, err = template.New("values").Parse(valuesTemplate)
	if err != nil {
		return err
	}

	valuesFile, err := os.Create(filepath.Join(chartDir, "values.yaml"))
	if err != nil {
		return err
	}
	defer valuesFile.Close()

	if err := tmpl.Execute(valuesFile, config); err != nil {
		return err
	}

	fmt.Printf("Generated Helm chart in %s/\n", chartDir)
	return nil
}

func generateTerraform(config *Config) error {
	terraformTemplate := `terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "{{.Name}}"
}

# Example ECS service configuration
resource "aws_ecs_cluster" "main" {
  name = var.app_name
}

resource "aws_ecs_service" "app" {
  name            = var.app_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
}

resource "aws_ecs_task_definition" "app" {
  family                   = var.app_name
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = 256
  memory                  = 512
  
  container_definitions = jsonencode([
    {
      name  = var.app_name
      image = "{{.Name}}:latest"
      
      environment = [
{{range $key, $value := .Variables}}        {
          name  = "{{$key}}"
          value = "{{$value}}"
        },
{{end}}      ]
    }
  ])
}

output "cluster_name" {
  value = aws_ecs_cluster.main.name
}
`

	if err := os.MkdirAll("terraform", 0755); err != nil {
		return err
	}

	tmpl, err := template.New("terraform").Parse(terraformTemplate)
	if err != nil {
		return err
	}

	tfFile, err := os.Create(filepath.Join("terraform", "main.tf"))
	if err != nil {
		return err
	}
	defer tfFile.Close()

	if err := tmpl.Execute(tfFile, config); err != nil {
		return err
	}

	fmt.Println("Generated Terraform configuration in terraform/")
	return nil
}

func generateMakefile(config *Config) error {
	makefileTemplate := `# Generated Makefile for {{.Name}}
# Description: {{.Description}}

.PHONY: help{{range .Commands}} {{.Name}}{{end}}

help:
	@echo "Available targets:"
{{range .Commands}}	@echo "  {{.Name}}{{if .Description}} - {{.Description}}{{end}}"
{{end}}

{{range .Commands}}{{.Name}}:
{{if .WorkDir}}	cd {{.WorkDir}} && {{.Command}}{{if .Args}} {{join .Args " "}}{{end}}
{{else}}	{{.Command}}{{if .Args}} {{join .Args " "}}{{end}}
{{end}}

{{end}}`

	// Add join function to template
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("makefile").Funcs(funcMap).Parse(makefileTemplate)
	if err != nil {
		return err
	}

	makeFile, err := os.Create("Makefile.generated")
	if err != nil {
		return err
	}
	defer makeFile.Close()

	if err := tmpl.Execute(makeFile, config); err != nil {
		return err
	}

	fmt.Println("Generated Makefile.generated")
	return nil
}