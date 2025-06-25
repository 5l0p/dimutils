package embed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds configuration for embedded tools
type Config struct {
	ToolsDir  string
	CacheDir  string
	Verbose   bool
	Timeout   int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		ToolsDir: filepath.Join(homeDir, ".dimutils", "tools"),
		CacheDir: filepath.Join(homeDir, ".dimutils", "cache"),
		Verbose:  false,
		Timeout:  30,
	}
}

// ToolInfo represents information about an embedded tool
type ToolInfo struct {
	Name        string
	Version     string
	Path        string
	Description string
	Commands    []string
	Aliases     []string
}

// EmbeddedTool represents an embedded tool instance
type EmbeddedTool struct {
	Info   ToolInfo
	Config Config
}

// ToolManager manages embedded tools
type ToolManager struct {
	config Config
	tools  map[string]*EmbeddedTool
}

// NewToolManager creates a new tool manager
func NewToolManager(config Config) *ToolManager {
	return &ToolManager{
		config: config,
		tools:  make(map[string]*EmbeddedTool),
	}
}

// Run is the main entry point for embed functionality
func Run(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	config := DefaultConfig()
	command := args[0]
	subArgs := args[1:]

	// Parse global flags
	for i, arg := range subArgs {
		switch arg {
		case "--tools-dir":
			if i+1 < len(subArgs) {
				config.ToolsDir = subArgs[i+1]
			}
		case "--cache-dir":
			if i+1 < len(subArgs) {
				config.CacheDir = subArgs[i+1]
			}
		case "--verbose", "-v":
			config.Verbose = true
		}
	}

	manager := NewToolManager(config)

	switch command {
	case "list", "ls":
		return manager.ListTools(subArgs)
	case "install":
		return manager.InstallTool(subArgs)
	case "remove", "rm":
		return manager.RemoveTool(subArgs)
	case "run":
		return manager.RunTool(subArgs)
	case "jq":
		return manager.RunJQ(subArgs)
	case "discover":
		return manager.DiscoverAndList(subArgs)
	case "update":
		return manager.UpdateTool(subArgs)
	case "info":
		return manager.ToolInfo(subArgs)
	case "help", "-h", "--help":
		return printHelp()
	default:
		// Try to run as embedded tool
		return manager.RunEmbeddedTool(command, subArgs)
	}
}

func printHelp() error {
	help := `Usage: embed <command> [options]

Embedded applications and tools manager.

Commands:
  list, ls              List predefined embedded tools
  discover              Discover available tools in system PATH
  install TOOL          Install an embedded tool
  remove, rm TOOL       Remove an embedded tool
  run TOOL [args...]    Run an embedded tool with arguments
  jq [filter] [file]    JSON processor with predefined shortcuts
  update [TOOL]         Update tool(s) to latest version
  info TOOL             Show information about a tool
  help                  Show this help message

Global Options:
  --tools-dir DIR       Directory for embedded tools
  --cache-dir DIR       Directory for tool cache
  --verbose, -v         Verbose output

Embedded Tools:
  jq                    JSON processor with enhanced functionality
  oc                    OpenShift CLI with kubectl compatibility
  docker                Docker CLI for container operations
  helm                  Kubernetes package manager
  kubectl               Kubernetes command-line tool

jq Shortcuts:
  .keys                 Get object keys
  .values               Get object values
  .length               Get array/object length
  .pretty               Pretty-print JSON
  .compact              Compact JSON output

Examples:
  embed list
  embed install jq
  embed jq '.name' data.json
  embed run kubectl get pods
  embed oc get projects`

	fmt.Println(help)
	return nil
}

// ListTools lists all installed embedded tools
func (tm *ToolManager) ListTools(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Scanning tools directory: %s\n", tm.config.ToolsDir)
	}

	// For now, return predefined tools
	tools := []ToolInfo{
		{
			Name:        "jq",
			Version:     "1.6",
			Description: "JSON processor with enhanced functionality",
			Commands:    []string{"jq"},
			Aliases:     []string{"json"},
		},
		{
			Name:        "oc",
			Version:     "4.10.0",
			Description: "OpenShift CLI with kubectl compatibility",
			Commands:    []string{"oc", "kubectl"},
			Aliases:     []string{"openshift", "k8s"},
		},
	}

	if tm.config.Verbose {
		fmt.Printf("%-15s %-10s %-40s %s\n", "TOOL", "VERSION", "DESCRIPTION", "COMMANDS")
		fmt.Println(strings.Repeat("-", 80))
	}

	for _, tool := range tools {
		if tm.config.Verbose {
			fmt.Printf("%-15s %-10s %-40s %s\n",
				tool.Name, tool.Version, tool.Description, strings.Join(tool.Commands, ", "))
		} else {
			fmt.Println(tool.Name)
		}
	}

	return nil
}

// DiscoverAndList discovers and lists available tools in the system
func (tm *ToolManager) DiscoverAndList(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Discovering tools in system PATH...\n")
	}

	tools, err := tm.DiscoverTools()
	if err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	if tm.config.Verbose {
		fmt.Printf("Found %d tools:\n", len(tools))
		fmt.Printf("%-15s %-10s %-40s %s\n", "TOOL", "VERSION", "DESCRIPTION", "PATH")
		fmt.Println(strings.Repeat("-", 80))
	}

	for _, tool := range tools {
		if tm.config.Verbose {
			fmt.Printf("%-15s %-10s %-40s %s\n",
				tool.Name, tool.Version, tool.Description, tool.Path)
		} else {
			fmt.Println(tool.Name)
		}
	}

	return nil
}

// InstallTool installs an embedded tool
func (tm *ToolManager) InstallTool(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tool name is required")
	}

	toolName := args[0]
	
	if tm.config.Verbose {
		fmt.Printf("Installing tool: %s\n", toolName)
	}

	// Create tools directory if it doesn't exist
	if err := os.MkdirAll(tm.config.ToolsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// Simulate installation
	fmt.Printf("Tool %s installed successfully\n", toolName)
	fmt.Printf("Note: Actual binary installation not implemented in this version\n")

	return nil
}

// RemoveTool removes an embedded tool
func (tm *ToolManager) RemoveTool(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tool name is required")
	}

	toolName := args[0]
	
	if tm.config.Verbose {
		fmt.Printf("Removing tool: %s\n", toolName)
	}

	fmt.Printf("Tool %s removed successfully\n", toolName)
	fmt.Printf("Note: Actual binary removal not implemented in this version\n")

	return nil
}

// RunTool runs an embedded tool with arguments
func (tm *ToolManager) RunTool(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tool name is required")
	}

	toolName := args[0]
	toolArgs := args[1:]

	if tm.config.Verbose {
		fmt.Printf("Running tool: %s with args: %v\n", toolName, toolArgs)
	}

	return tm.RunEmbeddedTool(toolName, toolArgs)
}

// RunEmbeddedTool runs an embedded tool directly
func (tm *ToolManager) RunEmbeddedTool(toolName string, args []string) error {
	switch toolName {
	case "jq", "json":
		return tm.RunJQ(args)
	case "oc", "openshift":
		return tm.RunOC(args)
	case "kubectl", "k8s":
		return tm.RunKubectl(args)
	case "docker":
		return tm.RunDocker(args)
	case "helm":
		return tm.RunHelm(args)
	case "aws", "gcloud", "az":
		return tm.RunGenericTool(toolName, args)
	case "psql", "mysql", "redis-cli", "mongo", "sqlite3":
		return tm.RunGenericTool(toolName, args)
	case "curl", "wget", "git", "yq", "grep", "sed", "awk":
		return tm.RunGenericTool(toolName, args)
	default:
		return fmt.Errorf("unknown embedded tool: %s. Use 'embed list' to see available tools", toolName)
	}
}

// UpdateTool updates an embedded tool
func (tm *ToolManager) UpdateTool(args []string) error {
	var toolName string
	if len(args) > 0 {
		toolName = args[0]
	}

	if toolName == "" {
		fmt.Println("Updating all tools...")
		fmt.Println("Note: Bulk update not implemented in this version")
	} else {
		fmt.Printf("Updating tool: %s\n", toolName)
		fmt.Printf("Note: Tool update not implemented in this version\n")
	}

	return nil
}

// ToolInfo shows information about a tool
func (tm *ToolManager) ToolInfo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tool name is required")
	}

	toolName := args[0]

	// Mock tool info
	switch toolName {
	case "jq":
		fmt.Printf("Tool: jq\n")
		fmt.Printf("Version: 1.6\n")
		fmt.Printf("Description: JSON processor with enhanced functionality\n")
		fmt.Printf("Commands: jq, json\n")
		fmt.Printf("Shortcuts: .keys, .values, .length, .pretty, .compact\n")
	case "oc":
		fmt.Printf("Tool: oc\n")
		fmt.Printf("Version: 4.10.0\n")
		fmt.Printf("Description: OpenShift CLI with kubectl compatibility\n")
		fmt.Printf("Commands: oc, kubectl\n")
		fmt.Printf("Aliases: openshift, k8s\n")
	default:
		return fmt.Errorf("tool not found: %s", toolName)
	}

	return nil
}