package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	prompt = "dim > "
)
// Run executes the shell command with the provided arguments
func Run(args []string) error {
	parser := syntax.NewParser()
	runner, err := interp.New(
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
		interp.ExecHandler(createExecHandler()),
	)
	if err != nil {
		return fmt.Errorf("error creating shell interpreter: %v", err)
	}

	if len(args) == 0 {
		// Check if stdin is a terminal (interactive) or pipe/redirect (script mode)
		if term.IsTerminal(int(os.Stdin.Fd())) {
			// Interactive mode - stdin is a terminal
			return runInteractive(parser, runner)
		} else {
			// Script mode - stdin is piped/redirected
			return runPipedScript(parser, runner)
		}
	}

	if len(args) == 2 && args[0] == "-c" {
		// Execute command string
		return runCommand(parser, runner, args[1])
	}

	// Execute script file
	return runScript(parser, runner, args[0])
}

func runInteractive(parser *syntax.Parser, runner *interp.Runner) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print(prompt)
	var src strings.Builder
	for scanner.Scan() {
		src.WriteString(scanner.Text())
		src.WriteByte('\n')

		prog, err := parser.Parse(strings.NewReader(src.String()), "")
		if err != nil {
			if syntax.IsIncomplete(err) {
				fmt.Print("> ")
				continue
			}
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			src.Reset()
			fmt.Print(prompt)
			continue
		}

		src.Reset()
		if err := runner.Run(context.Background(), prog); err != nil {
			if status, ok := interp.IsExitStatus(err); ok {
				os.Exit(int(status))
			}
			fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		}
		fmt.Print(prompt)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %v", err)
	}
	return nil
}

func runPipedScript(parser *syntax.Parser, runner *interp.Runner) error {
	// Read all input from stdin and execute as a script
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading piped input: %v", err)
	}

	if len(input) == 0 {
		return nil // Empty input, nothing to do
	}

	prog, err := parser.Parse(strings.NewReader(string(input)), "")
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	if err := runner.Run(context.Background(), prog); err != nil {
		if status, ok := interp.IsExitStatus(err); ok {
			os.Exit(int(status))
		}
		return fmt.Errorf("runtime error: %v", err)
	}
	return nil
}

func runCommand(parser *syntax.Parser, runner *interp.Runner, command string) error {
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	if err := runner.Run(context.Background(), prog); err != nil {
		if status, ok := interp.IsExitStatus(err); ok {
			os.Exit(int(status))
		}
		return fmt.Errorf("runtime error: %v", err)
	}
	return nil
}

func runScript(parser *syntax.Parser, runner *interp.Runner, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening script file: %v", err)
	}
	defer file.Close()

	prog, err := parser.Parse(file, filename)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	if err := runner.Run(context.Background(), prog); err != nil {
		if status, ok := interp.IsExitStatus(err); ok {
			os.Exit(int(status))
		}
		return fmt.Errorf("runtime error: %v", err)
	}
	return nil
}