package embed

import (
	"fmt"
	"os"

	"github.com/itchyny/gojq/cli"
)

// JQProcessor handles jq-like JSON processing
type JQProcessor struct {
	config Config
}

// NewJQProcessor creates a new jq processor
func NewJQProcessor(config Config) *JQProcessor {
	return &JQProcessor{config: config}
}

// RunJQ runs jq processing using gojq/cli
func (tm *ToolManager) RunJQ(args []string) error {
	if tm.config.Verbose {
		fmt.Printf("Running jq with args: %v\n", args)
	}

	// Handle predefined shortcuts
	if len(args) > 0 {
		switch args[0] {
		case ".keys":
			args[0] = "keys"
		case ".values":
			args[0] = "values" 
		case ".length":
			args[0] = "length"
		case ".pretty":
			args[0] = "."
		case ".compact":
			args[0] = "."
		case ".type":
			args[0] = "type"
		case ".reverse":
			args[0] = "reverse"
		case ".sort":
			args[0] = "sort"
		case ".unique":
			args[0] = "unique"
		case ".flatten":
			args[0] = "flatten"
		case ".min":
			args[0] = "min"
		case ".max":
			args[0] = "max"
		case ".sum":
			args[0] = "add"
		case ".avg":
			args[0] = "add/length"
		}
	}

	// Set up args for the CLI
	oldArgs := os.Args
	os.Args = append([]string{"gojq"}, args...)
	defer func() {
		os.Args = oldArgs
	}()

	// Use gojq CLI directly
	exitCode := cli.Run()
	if exitCode != 0 {
		return fmt.Errorf("jq processing failed with exit code: %d", exitCode)
	}

	return nil
}

