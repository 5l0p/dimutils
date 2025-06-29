package main

import (
	"context"
	"fmt"
	"os"

	"github.com/og-dim9/dimutils/pkg/apigen"
	makecmd "github.com/5l0p/go-make/pkg/cmd"
	"github.com/databricks/cli/cmd/root"
	"github.com/itchyny/gojq/cli"
	yqcmd "github.com/mikefarah/yq/v4/cmd"
	"github.com/og-dim9/dimutils/pkg/cbxxml2regex"
	"github.com/og-dim9/dimutils/pkg/config"
	"github.com/og-dim9/dimutils/pkg/datagen"
	"github.com/og-dim9/dimutils/pkg/ebcdic"
	"github.com/og-dim9/dimutils/pkg/eventdiff"
	"github.com/og-dim9/dimutils/pkg/gitaskop"
	"github.com/og-dim9/dimutils/pkg/mkgchat"
	"github.com/og-dim9/dimutils/pkg/regex2json"
	"github.com/og-dim9/dimutils/pkg/serve"
	"github.com/og-dim9/dimutils/pkg/shell"
	"github.com/og-dim9/dimutils/pkg/tandum"
	"github.com/og-dim9/dimutils/pkg/togchat"
	"github.com/og-dim9/dimutils/pkg/unexpect"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubectl/pkg/cmd"
)

// gitaskopCmd represents the gitaskop command
var gitaskopCmd = &cobra.Command{
	Use:                "gitaskop",
	Short:              "Git task scheduler and runner",
	Long:               `A git-based task scheduler that runs commands based on repository changes.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := gitaskop.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// eventdiffCmd represents the eventdiff command
var eventdiffCmd = &cobra.Command{
	Use:   "eventdiff",
	Short: "Event difference analyzer",
	Long:  `Analyze differences between events and data streams.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := eventdiff.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// unexpectCmd represents the unexpect command
var unexpectCmd = &cobra.Command{
	Use:   "unexpect",
	Short: "Test expectation framework",
	Long:  `A test framework for setting up expectations and validating outcomes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := unexpect.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP server utilities",
	Long:  `Simple HTTP server for development and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := serve.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// ebcdicCmd represents the ebcdic command
var ebcdicCmd = &cobra.Command{
	Use:   "ebcdic",
	Short: "EBCDIC encoding utilities",
	Long:  `Tools for working with EBCDIC encoded data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ebcdic.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// cbxxml2regexCmd represents the cbxxml2regex command
var cbxxml2regexCmd = &cobra.Command{
	Use:                "cbxxml2regex",
	Short:              "COBOL XML to regex converter",
	Long:               `Convert COBOL XML definitions to regular expressions.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cbxxml2regex.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// regex2jsonCmd represents the regex2json command
var regex2jsonCmd = &cobra.Command{
	Use:   "regex2json",
	Short: "Regex to JSON converter",
	Long:  `Convert regular expression patterns to JSON structures.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := regex2json.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// tandumCmd represents the tandum command
var tandumCmd = &cobra.Command{
	Use:   "tandum",
	Short: "Tandum data processing utility",
	Long:  `Process and transform tandum-format data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := tandum.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// mkgchatCmd represents the mkgchat command
var mkgchatCmd = &cobra.Command{
	Use:                "mkgchat",
	Short:              "Make Google Chat utility",
	Long:               `Utility for creating Google Chat messages and interactions.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := mkgchat.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// togchatCmd represents the togchat command
var togchatCmd = &cobra.Command{
	Use:                "togchat",
	Short:              "To Google Chat utility",
	Long:               `Send messages and data to Google Chat.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := togchat.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// jqCmd represents the jq command
var jqCmd = &cobra.Command{
	Use:                "jq",
	Short:              "JSON processor",
	Long:               `Command-line JSON processor using gojq implementation.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up args for gojq CLI
		oldArgs := os.Args
		os.Args = append([]string{"gojq"}, args...)
		defer func() {
			os.Args = oldArgs
		}()

		// Run gojq CLI
		exitCode := cli.Run()
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	},
}

// yqCmd represents the yq command
var yqCmd = &cobra.Command{
	Use:                "yq",
	Short:              "YAML processor",
	Long:               `Command-line YAML processor for querying and manipulating YAML data.`,
	DisableFlagParsing: true,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// Set up args for yq
		oldArgs := os.Args
		os.Args = append([]string{"yq"}, args...)
		defer func() {
			os.Args = oldArgs
		}()

		// Create and execute yq command
		yqCommand := yqcmd.New()
		yqCommand.SetArgs(args)
		if err := yqCommand.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// kubectlCmd represents the kubectl command
var kubectlCmd = &cobra.Command{
	Use:                "kubectl",
	Short:              "Kubernetes CLI",
	Long:               `Command-line tool for controlling Kubernetes clusters.`,
	DisableFlagParsing: true,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// Set up args for kubectl
		oldArgs := os.Args
		os.Args = append([]string{"kubectl"}, args...)
		defer func() {
			os.Args = oldArgs
		}()

		// Create kubectl command with factory
		kubectlCmd := kubectlcmd.NewDefaultKubectlCommand()
		kubectlCmd.SetArgs(args)
		if err := kubectlCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// databricksCmd represents the databricks command
var databricksCmd = &cobra.Command{
	Use:                "databricks",
	Short:              "Databricks CLI",
	Long:               `Command-line interface for Databricks.`,
	DisableFlagParsing: true,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// Set up args for databricks CLI
		oldArgs := os.Args
		os.Args = append([]string{"databricks"}, args...)
		defer func() {
			os.Args = oldArgs
		}()

		// Create and execute databricks command
		ctx := context.Background()
		databricksCmd := root.New(ctx)
		databricksCmd.SetArgs(args)
		if err := databricksCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:                "make",
	Short:              "Go-based make implementation",
	Long:               `A Go implementation of the make utility for building projects.`,
	DisableFlagParsing: true,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// Create go-make command with Makefile
		makeCommand, err := makecmd.New("Makefile")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating make command: %v\n", err)
			os.Exit(1)
		}

		// If no target specified, build default target
		if len(args) == 0 {
			if err := makeCommand.BuildDefault(); err != nil {
				fmt.Fprintf(os.Stderr, "Error building default target: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Build specified targets
			for _, target := range args {
				if err := makeCommand.Build(target); err != nil {
					fmt.Fprintf(os.Stderr, "Error building target '%s': %v\n", target, err)
					os.Exit(1)
				}
			}
		}
	},
}

// goshCmd represents the shell command
var goshCmd = &cobra.Command{
	Use:                "shell",
	Short:              "Shell interpreter",
	Long:               `An interactive shell interpreter with POSIX shell features.`,
	DisableFlagParsing: true,
	Run: func(cobraCmd *cobra.Command, args []string) {
		if err := shell.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// apigenCmd represents the apigen command
var apigenCmd = &cobra.Command{
	Use:   "apigen",
	Short: "API generator for read-only data APIs",
	Long:  `Generate REST APIs, HTML pages, and blob storage from data sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := apigen.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:                "config",
	Short:              "Interactive configuration management",
	Long:               `Create and manage configuration files, run command chains, and generate manifests.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// datagenCmd represents the datagen command
var datagenCmd = &cobra.Command{
	Use:   "datagen",
	Short: "Test data generation utility",
	Long:  `Generate realistic test data and shadow traffic for load testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := datagen.Run(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// runIndividualTool shows a placeholder message for now
func runIndividualTool(toolName string, args []string) {
	//fixme: we should fallback to the tools downloader if we need to
	cobra.CheckErr(fmt.Errorf("%s tool not yet integrated into multicall binary. Please use individual binary from src/%s/ or run 'make %s' to build it", toolName, toolName, toolName))
}

func init() {
	// Add all tool commands to root
	rootCmd.AddCommand(
		apigenCmd,
		datagenCmd,
		gitaskopCmd,
		eventdiffCmd,
		unexpectCmd,
		serveCmd,
		ebcdicCmd,
		cbxxml2regexCmd,
		regex2jsonCmd,
		tandumCmd,
		mkgchatCmd,
		togchatCmd,
		jqCmd,
		yqCmd,
		kubectlCmd,
		databricksCmd,
		makeCmd,
		goshCmd,
		configCmd,
	)
}