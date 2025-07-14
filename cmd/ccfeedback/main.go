package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jrossi/ccfeedback"
)

func main() {
	var (
		timeout     = flag.Duration("timeout", 60*time.Second, "Hook execution timeout")
		showVersion = flag.Bool("version", false, "Show version information")
		debug       = flag.Bool("debug", false, "Enable debug output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CCFeedback - Claude Code Hooks Feedback System\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nThe tool reads hook messages from stdin and writes responses to stdout.\n")
		fmt.Fprintf(os.Stderr, "Exit codes:\n")
		fmt.Fprintf(os.Stderr, "  0 - Success (stdout shown in transcript)\n")
		fmt.Fprintf(os.Stderr, "  2 - Blocking error (stderr processed by Claude)\n")
		fmt.Fprintf(os.Stderr, "  Other - Non-blocking error\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println("ccfeedback version 0.1.0")
		os.Exit(0)
	}

	// Create rule engine with linting capabilities
	ruleEngine := ccfeedback.NewLintingRuleEngine()

	// Create executor
	executor := ccfeedback.NewExecutor(ruleEngine)
	executor.SetTimeout(*timeout)

	// Create context
	ctx := context.Background()

	// Execute
	exitCode, err := executor.ExecuteWithExitCode(ctx)

	// Always flush both stdout and stderr before exiting
	os.Stdout.Sync()
	os.Stderr.Sync()

	if err != nil {
		// Errors are non-blocking (exit 1) and shown on stderr
		fmt.Fprintf(os.Stderr, "\n> Hook execution error:\n")
		fmt.Fprintf(os.Stderr, "  - [ccfeedback]: ❌ %v\n", err)
		if *debug {
			fmt.Fprintf(os.Stderr, "  - Debug: Full error: %v\n", err)
		}
		// Default to non-blocking error
		os.Exit(1)
	}

	// Show status for successful exit codes in debug mode
	if exitCode == 0 && *debug {
		// Success messages go to stdout for exit code 0
		fmt.Fprintf(os.Stdout, "\n> Hook execution completed:\n")
		fmt.Fprintf(os.Stdout, "  - [ccfeedback]: ✅ Success (exit code 0)\n")
	}

	// Exit with the proper code
	os.Exit(exitCode)
}
