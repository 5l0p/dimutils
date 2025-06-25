package shell

import (
	"context"
	"fmt"
	"os"

	makecmd "github.com/5l0p/go-make/pkg/cmd"
	"github.com/databricks/cli/cmd/root"
	"github.com/itchyny/gojq/cli"
	yqcmd "github.com/mikefarah/yq/v4/cmd"
	"github.com/og-dim9/dimutils/pkg/cbxxml2regex"
	"github.com/og-dim9/dimutils/pkg/ebcdic"
	"github.com/og-dim9/dimutils/pkg/eventdiff"
	"github.com/og-dim9/dimutils/pkg/gitaskop"
	"github.com/og-dim9/dimutils/pkg/mkgchat"
	"github.com/og-dim9/dimutils/pkg/regex2json"
	"github.com/og-dim9/dimutils/pkg/serve"
	"github.com/og-dim9/dimutils/pkg/tandum"
	"github.com/og-dim9/dimutils/pkg/togchat"
	"github.com/og-dim9/dimutils/pkg/unexpect"
	"mvdan.cc/sh/v3/interp"
	kubectlcmd "k8s.io/kubectl/pkg/cmd"
)

// BuiltinFunc represents a builtin command function
type BuiltinFunc func(ctx context.Context, args []string) error

// builtins maps command names to their implementations
var builtins = map[string]BuiltinFunc{
	"gitaskop":     runGitaskop,
	"eventdiff":    runEventdiff,
	"unexpect":     runUnexpect,
	"serve":        runServe,
	"ebcdic":       runEbcdic,
	"cbxxml2regex": runCbxxml2regex,
	"regex2json":   runRegex2json,
	"tandum":       runTandum,
	"mkgchat":      runMkgchat,
	"togchat":      runTogchat,
	"jq":           runJq,
	"yq":           runYq,
	"kubectl":      runKubectl,
	"databricks":   runDatabricks,
	"make":         runMake,
}

// createExecHandler creates an exec handler that includes our builtins
func createExecHandler() interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		if len(args) == 0 {
			return nil
		}

		cmdName := args[0]
		
		// Check if it's one of our builtins
		if builtin, exists := builtins[cmdName]; exists {
			// Execute the builtin with remaining args
			return builtin(ctx, args[1:])
		}

		// Fall back to default behavior (execute external command)
		return interp.DefaultExecHandler(2*1024*1024)(ctx, args) // 2MB limit
	}
}

// Builtin command implementations

func runGitaskop(ctx context.Context, args []string) error {
	return gitaskop.Run(args)
}

func runEventdiff(ctx context.Context, args []string) error {
	return eventdiff.Run(args)
}

func runUnexpect(ctx context.Context, args []string) error {
	return unexpect.Run(args)
}

func runServe(ctx context.Context, args []string) error {
	return serve.Run(args)
}

func runEbcdic(ctx context.Context, args []string) error {
	return ebcdic.Run(args)
}

func runCbxxml2regex(ctx context.Context, args []string) error {
	return cbxxml2regex.Run(args)
}

func runRegex2json(ctx context.Context, args []string) error {
	return regex2json.Run(args)
}

func runTandum(ctx context.Context, args []string) error {
	return tandum.Run(args)
}

func runMkgchat(ctx context.Context, args []string) error {
	return mkgchat.Run(args)
}

func runTogchat(ctx context.Context, args []string) error {
	return togchat.Run(args)
}

func runJq(ctx context.Context, args []string) error {
	// Set up args for gojq CLI
	oldArgs := os.Args
	os.Args = append([]string{"gojq"}, args...)
	defer func() {
		os.Args = oldArgs
	}()

	// Run gojq CLI
	exitCode := cli.Run()
	if exitCode != 0 {
		return fmt.Errorf("jq exited with code %d", exitCode)
	}
	return nil
}

func runYq(ctx context.Context, args []string) error {
	// Set up args for yq
	oldArgs := os.Args
	os.Args = append([]string{"yq"}, args...)
	defer func() {
		os.Args = oldArgs
	}()

	// Create and execute yq command
	yqCommand := yqcmd.New()
	yqCommand.SetArgs(args)
	return yqCommand.Execute()
}

func runKubectl(ctx context.Context, args []string) error {
	// Set up args for kubectl
	oldArgs := os.Args
	os.Args = append([]string{"kubectl"}, args...)
	defer func() {
		os.Args = oldArgs
	}()

	// Create kubectl command with factory
	kubectlCmd := kubectlcmd.NewDefaultKubectlCommand()
	kubectlCmd.SetArgs(args)
	return kubectlCmd.Execute()
}

func runDatabricks(ctx context.Context, args []string) error {
	// Set up args for databricks CLI
	oldArgs := os.Args
	os.Args = append([]string{"databricks"}, args...)
	defer func() {
		os.Args = oldArgs
	}()

	// Create and execute databricks command
	databricksCmd := root.New(ctx)
	databricksCmd.SetArgs(args)
	return databricksCmd.Execute()
}

func runMake(ctx context.Context, args []string) error {
	// Create go-make command with Makefile
	makeCommand, err := makecmd.New("Makefile")
	if err != nil {
		return fmt.Errorf("error creating make command: %v", err)
	}

	// If no target specified, build default target
	if len(args) == 0 {
		return makeCommand.BuildDefault()
	}

	// Build specified targets
	for _, target := range args {
		if err := makeCommand.Build(target); err != nil {
			return fmt.Errorf("error building target '%s': %v", target, err)
		}
	}
	return nil
}