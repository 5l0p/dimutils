package embed

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunOC runs OpenShift CLI commands
func (tm *ToolManager) RunOC(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running oc with args: %v\n", args)
	}

	// Check if oc is available in PATH
	if _, err := exec.LookPath("oc"); err != nil {
		return fmt.Errorf("oc command not found in PATH. Please install OpenShift CLI")
	}

	cmd := exec.Command("oc", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunDocker runs Docker CLI commands
func (tm *ToolManager) RunDocker(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running docker with args: %v\n", args)
	}

	// Check if docker is available in PATH
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker command not found in PATH. Please install Docker CLI")
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunHelm runs Helm commands
func (tm *ToolManager) RunHelm(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running helm with args: %v\n", args)
	}

	// Check if helm is available in PATH
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm command not found in PATH. Please install Helm")
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunKubectl runs kubectl commands
func (tm *ToolManager) RunKubectl(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running kubectl with args: %v\n", args)
	}

	// Check if kubectl is available in PATH, fallback to oc
	var cmdName string
	if _, err := exec.LookPath("kubectl"); err == nil {
		cmdName = "kubectl"
	} else if _, err := exec.LookPath("oc"); err == nil {
		cmdName = "oc"
	} else {
		return fmt.Errorf("neither kubectl nor oc found in PATH. Please install Kubernetes CLI")
	}

	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunGenericTool runs a generic tool command
func (tm *ToolManager) RunGenericTool(toolName string, args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running %s with args: %v\n", toolName, args)
	}

	// Check if tool is available in PATH
	if _, err := exec.LookPath(toolName); err != nil {
		return fmt.Errorf("%s command not found in PATH. Please install %s", toolName, toolName)
	}

	cmd := exec.Command(toolName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DiscoverTools discovers available tools in the system
func (tm *ToolManager) DiscoverTools() ([]ToolInfo, error) {
	commonTools := []string{
		"kubectl", "oc", "docker", "helm", "aws", "gcloud", "az",
		"psql", "mysql", "redis-cli", "mongo", "sqlite3",
		"curl", "wget", "git", "jq", "yq", "grep", "sed", "awk",
	}

	var discoveredTools []ToolInfo

	for _, tool := range commonTools {
		if path, err := exec.LookPath(tool); err == nil {
			info := ToolInfo{
				Name:        tool,
				Path:        path,
				Description: getToolDescription(tool),
				Commands:    []string{tool},
			}

			// Get version if possible
			if version := getToolVersion(tool); version != "" {
				info.Version = version
			}

			discoveredTools = append(discoveredTools, info)
		}
	}

	return discoveredTools, nil
}

// getToolDescription returns a description for common tools
func getToolDescription(tool string) string {
	descriptions := map[string]string{
		"kubectl":   "Kubernetes command-line tool",
		"oc":        "OpenShift CLI",
		"docker":    "Docker container platform CLI",
		"helm":      "Kubernetes package manager",
		"aws":       "AWS Command Line Interface",
		"gcloud":    "Google Cloud SDK CLI",
		"az":        "Azure CLI",
		"psql":      "PostgreSQL interactive terminal",
		"mysql":     "MySQL command-line client",
		"redis-cli": "Redis command-line interface",
		"mongo":     "MongoDB shell",
		"sqlite3":   "SQLite command-line interface",
		"curl":      "Command-line tool for transferring data",
		"wget":      "Network downloader",
		"git":       "Distributed version control system",
		"jq":        "Command-line JSON processor",
		"yq":        "Command-line YAML processor",
		"grep":      "Pattern searching utility",
		"sed":       "Stream editor",
		"awk":       "Pattern scanning and processing language",
	}

	if desc, exists := descriptions[tool]; exists {
		return desc
	}
	return fmt.Sprintf("%s command-line tool", tool)
}

// getToolVersion attempts to get the version of a tool
func getToolVersion(tool string) string {
	versionArgs := map[string][]string{
		"kubectl":   {"version", "--client", "--short"},
		"oc":        {"version", "--client"},
		"docker":    {"version", "--format", "{{.Client.Version}}"},
		"helm":      {"version", "--short", "--client"},
		"aws":       {"--version"},
		"gcloud":    {"version", "--format=value(version)"},
		"az":        {"--version"},
		"jq":        {"--version"},
		"git":       {"--version"},
		"curl":      {"--version"},
		"wget":      {"--version"},
	}

	if args, exists := versionArgs[tool]; exists {
		cmd := exec.Command(tool, args...)
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			// Clean up version output
			if strings.Contains(version, "\n") {
				version = strings.Split(version, "\n")[0]
			}
			return version
		}
	}

	return "unknown"
}

// InstallEmbeddedTool installs a tool into the embedded tools directory
func (tm *ToolManager) InstallEmbeddedTool(toolName, downloadURL string) error {
	if tm.config.Verbose {
		fmt.Printf("Installing embedded tool: %s from %s\n", toolName, downloadURL)
	}

	// Create tools directory
	if err := os.MkdirAll(tm.config.ToolsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// This is a placeholder for actual binary downloading and installation
	fmt.Printf("Binary installation not implemented. Would download %s to %s\n", 
		downloadURL, tm.config.ToolsDir)

	return nil
}