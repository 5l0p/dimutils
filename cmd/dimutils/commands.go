package main

import (
	"fmt"
	"os"

	"github.com/og-dim9/dimutils/pkg/apigen"
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
	"github.com/spf13/cobra"
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

// runIndividualTool shows a placeholder message for now
func runIndividualTool(toolName string, args []string) {
	cobra.CheckErr(fmt.Errorf("%s tool not yet integrated into multicall binary. Please use individual binary from src/%s/ or run 'make %s' to build it", toolName, toolName, toolName))
}

func init() {
	// Add all tool commands to root
	rootCmd.AddCommand(
		apigenCmd,
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
	)
}