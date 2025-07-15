package ccfeedback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jrossi/ccfeedback/linters"
	"github.com/jrossi/ccfeedback/linters/golang"
	jsonlinter "github.com/jrossi/ccfeedback/linters/json"
	"github.com/jrossi/ccfeedback/linters/markdown"
	"github.com/jrossi/ccfeedback/linters/python"
)

// LintingRuleEngine implements RuleEngine to provide linting functionality
type LintingRuleEngine struct {
	linters  []linters.Linter
	executor *linters.ParallelExecutor
	config   *AppConfig
}

// LintingConfig provides configuration options for the linting engine
type LintingConfig struct {
	// MaxWorkers sets the maximum number of concurrent workers
	// If 0 or negative, defaults to runtime.NumCPU()
	MaxWorkers int
	// DisableParallel disables parallel execution for debugging
	DisableParallel bool
}

// NewLintingRuleEngine creates a new linting rule engine with default linters
func NewLintingRuleEngine() *LintingRuleEngine {
	return NewLintingRuleEngineWithConfig(LintingConfig{})
}

// NewLintingRuleEngineWithConfig creates a new linting rule engine with custom configuration
func NewLintingRuleEngineWithConfig(config LintingConfig) *LintingRuleEngine {
	maxWorkers := config.MaxWorkers
	if config.DisableParallel {
		maxWorkers = 1
	}

	engine := &LintingRuleEngine{
		linters:  []linters.Linter{},
		executor: linters.NewParallelExecutor(maxWorkers),
		config:   NewAppConfig(),
	}

	// Initialize linters with empty configs for now
	// We'll update them when SetAppConfig is called
	engine.linters = append(engine.linters, golang.NewGoLinter())
	engine.linters = append(engine.linters, jsonlinter.NewJSONLinter())
	engine.linters = append(engine.linters, markdown.NewMarkdownLinter())
	engine.linters = append(engine.linters, python.NewPythonLinter())

	return engine
}

// AddLinter adds a custom linter to the engine
func (e *LintingRuleEngine) AddLinter(linter linters.Linter) {
	e.linters = append(e.linters, linter)
}

// SetAppConfig sets the application configuration
func (e *LintingRuleEngine) SetAppConfig(config *AppConfig) {
	e.config = config

	// Update linter configurations
	if config != nil {
		for _, linter := range e.linters {
			// Check if this linter is disabled
			if !config.IsLinterEnabled(linter.Name()) {
				continue
			}

			// Get linter-specific configuration
			if linterConfig, ok := config.GetLinterConfig(linter.Name()); ok {
				// Try to cast to configurable linter
				if configurable, ok := linter.(ConfigurableLinter); ok {
					if err := configurable.SetConfig(linterConfig); err != nil {
						// Log error but continue
						fmt.Fprintf(os.Stderr, "Warning: Failed to configure %s linter: %v\n", linter.Name(), err)
					}
				}
			}
		}
	}
}

// GetAppConfig returns the application configuration
func (e *LintingRuleEngine) GetAppConfig() *AppConfig {
	return e.config
}

// ConfigurableLinter is an interface for linters that support runtime configuration
type ConfigurableLinter interface {
	linters.Linter
	SetConfig(config json.RawMessage) error
}

// applyRuleOverrides applies any rule overrides for the given file path
func (e *LintingRuleEngine) applyRuleOverrides(filePath string) {
	if e.config == nil {
		return
	}

	// Apply overrides for each linter
	for _, linter := range e.linters {
		// Get any rule overrides for this file and linter
		overrides := e.config.GetRuleOverrides(filePath, linter.Name())
		if len(overrides) == 0 {
			continue
		}

		// Try to cast to configurable linter
		if configurable, ok := linter.(ConfigurableLinter); ok {
			// Merge all overrides into a single config
			mergedConfig := make(map[string]interface{})

			for _, override := range overrides {
				var overrideMap map[string]interface{}
				if err := json.Unmarshal(override, &overrideMap); err != nil {
					continue
				}

				// Merge this override into the merged config
				for k, v := range overrideMap {
					mergedConfig[k] = v
				}
			}

			// Convert back to JSON and apply
			if configData, err := json.Marshal(mergedConfig); err == nil {
				if err := configurable.SetConfig(configData); err != nil {
					// Log error but continue
					fmt.Fprintf(os.Stderr, "Warning: Failed to apply rule override for %s linter: %v\n", linter.Name(), err)
				}
			}
		}
	}
}

// EvaluatePreToolUse checks files before they're written
func (e *LintingRuleEngine) EvaluatePreToolUse(ctx context.Context, msg *PreToolUseMessage) (*HookResponse, error) {
	// Only check Write and Edit operations
	if msg.ToolName != "Write" && msg.ToolName != "Edit" && msg.ToolName != "MultiEdit" {
		return &HookResponse{Decision: "approve"}, nil
	}

	// Extract file path from tool input
	filePathRaw, exists := msg.ToolInput["file_path"]
	if !exists {
		return &HookResponse{Decision: "approve"}, nil
	}
	var filePath string
	if err := json.Unmarshal(filePathRaw, &filePath); err != nil {
		return &HookResponse{Decision: "approve"}, nil
	}

	// For Edit/MultiEdit, we can't lint until after the edit is done
	if msg.ToolName == "Edit" || msg.ToolName == "MultiEdit" {
		return &HookResponse{Decision: "approve"}, nil
	}

	// For Write operations, check the content
	contentRaw, exists := msg.ToolInput["content"]
	if !exists {
		return &HookResponse{Decision: "approve"}, nil
	}
	var content string
	if err := json.Unmarshal(contentRaw, &content); err != nil {
		return &HookResponse{Decision: "approve"}, nil
	}

	// Apply rule overrides for this file
	e.applyRuleOverrides(filePath)

	// Run all applicable linters in parallel
	results := e.executor.ExecuteLinters(ctx, e.linters, filePath, []byte(content))

	// Aggregate results
	aggregatedResult, errs := linters.AggregateResults(results)

	// Handle any linting errors
	if len(errs) > 0 {
		return &HookResponse{
			Decision: "block",
			Reason:   fmt.Sprintf("Linting error: %v", errs[0]),
		}, nil
	}

	// Check for issues and format detailed output
	var errorIssues, warningIssues []linters.Issue
	for _, issue := range aggregatedResult.Issues {
		if issue.Severity == "error" {
			errorIssues = append(errorIssues, issue)
		} else {
			warningIssues = append(warningIssues, issue)
		}
	}

	// If there are syntax errors, block the write
	if len(errorIssues) > 0 {
		output := e.formatLintOutput(filePath, errorIssues, true)
		// Write detailed output to stderr for user visibility
		fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
		return &HookResponse{
			Decision: "block",
			Reason:   fmt.Sprintf("Found %d error(s) in %s", len(errorIssues), filePath),
		}, nil
	}

	// If formatting is needed, inform but don't block
	if len(warningIssues) > 0 {
		output := e.formatLintOutput(filePath, warningIssues, false)
		// Write detailed output to stderr for user visibility
		fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
		return &HookResponse{
			Decision: "approve",
			Message:  fmt.Sprintf("Found %d warning(s) in %s", len(warningIssues), filePath),
		}, nil
	}

	// Write success message to stderr (matching smart-lint.sh behavior)
	fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: ✅ Style clean. Continue with your task.\n")
	return &HookResponse{Decision: "approve"}, nil
}

// EvaluatePostToolUse runs linters and tests after file operations
func (e *LintingRuleEngine) EvaluatePostToolUse(ctx context.Context, msg *PostToolUseMessage) (*HookResponse, error) {
	// Only check Write and Edit operations
	if msg.ToolName != "Write" && msg.ToolName != "Edit" && msg.ToolName != "MultiEdit" {
		// Show status for non-file operations on stderr (matching smart-lint.sh behavior)
		fmt.Fprintf(os.Stderr, "\n> Tool execution feedback:\n  - [ccfeedback]: ℹ️  %s operation completed (no linting required)\n", msg.ToolName)
		return nil, nil
	}

	// Skip if there was an error
	if msg.ToolError != "" {
		// Tool errors trigger exit code 1, shown on stderr
		fmt.Fprintf(os.Stderr, "\n> Tool execution feedback:\n  - [ccfeedback]: ⚠️  Tool error: %s (skipping linting)\n", msg.ToolError)
		return nil, nil
	}

	// Extract file path from tool input
	filePathRaw, exists := msg.ToolInput["file_path"]
	if !exists {
		return nil, nil
	}
	var filePath string
	if err := json.Unmarshal(filePathRaw, &filePath); err != nil || filePath == "" {
		return nil, nil
	}

	// Read the actual file from disk
	content, err := os.ReadFile(filePath)
	if err != nil {
		// File errors shown on stderr (matching smart-lint.sh behavior)
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: ⚠️  File not found: %s\n", filePath)
		} else {
			fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: ⚠️  Cannot read file: %v\n", err)
		}
		return nil, nil
	}

	// Apply rule overrides for this file
	e.applyRuleOverrides(filePath)

	// Run all applicable linters in parallel
	results := e.executor.ExecuteLinters(ctx, e.linters, filePath, content)

	// Aggregate results
	aggregatedResult, errs := linters.AggregateResults(results)

	// Handle any linting errors
	for _, err := range errs {
		// Linting errors trigger exit code 1, shown on stderr
		fmt.Fprintf(os.Stderr, "\n> Linting error for %s: %v\n", filePath, err)
	}

	// Check for issues and format detailed output
	var errorIssues, warningIssues []linters.Issue
	for _, issue := range aggregatedResult.Issues {
		if issue.Severity == "error" {
			errorIssues = append(errorIssues, issue)
		} else {
			warningIssues = append(warningIssues, issue)
		}
	}

	// Issues trigger exit code 1, shown on stderr
	if len(errorIssues) > 0 {
		output := e.formatLintOutput(filePath, errorIssues, true)
		fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
	} else if len(warningIssues) > 0 {
		output := e.formatLintOutput(filePath, warningIssues, false)
		fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
	} else if len(errs) == 0 {
		// Success shown on stderr (matching smart-lint.sh behavior)
		fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: ✅ Style clean. Continue with your task.\n")
	}

	// Check for associated test files if it's a Go file
	if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
		e.checkTestFile(ctx, filePath)
	}

	// Always return nil for PostToolUse to avoid JSON output interfering with stderr
	// The exit code is controlled by executor.go based on IsPostToolUseHook()
	return nil, nil
}

// EvaluateNotification handles system notifications
func (e *LintingRuleEngine) EvaluateNotification(ctx context.Context, msg *NotificationMessage) (*HookResponse, error) {
	return nil, nil
}

// EvaluateStop handles main agent completion
func (e *LintingRuleEngine) EvaluateStop(ctx context.Context, msg *StopMessage) (*HookResponse, error) {
	return nil, nil
}

// EvaluateSubagentStop handles subagent completion
func (e *LintingRuleEngine) EvaluateSubagentStop(ctx context.Context, msg *SubagentStopMessage) (*HookResponse, error) {
	return nil, nil
}

// EvaluatePreCompact handles pre-compact events
func (e *LintingRuleEngine) EvaluatePreCompact(ctx context.Context, msg *PreCompactMessage) (*HookResponse, error) {
	return nil, nil
}

// formatLintOutput formats linting issues in a style similar to smart-lint.sh
func (e *LintingRuleEngine) formatLintOutput(filePath string, issues []linters.Issue, isBlocking bool) string {
	var output strings.Builder

	// Header similar to smart-lint.sh
	output.WriteString(fmt.Sprintf("- [ccfeedback:%s]: ", filePath))

	// Add details for each issue
	for i, issue := range issues {
		if i > 0 {
			output.WriteString("\n  ")
		}

		// Format: file:line:column: message
		if issue.Line > 0 && issue.Column > 0 {
			output.WriteString(fmt.Sprintf("%s:%d:%d: %s",
				strings.TrimPrefix(filePath, "/Users/jrossi/src/ccfeedback/"),
				issue.Line, issue.Column, issue.Message))
		} else {
			output.WriteString(issue.Message)
		}

		if issue.Rule != "" {
			output.WriteString(fmt.Sprintf(" (%s)", issue.Rule))
		}
	}

	output.WriteString("\n")

	// Add footer similar to smart-lint.sh
	if isBlocking {
		issueCount := len(issues)
		output.WriteString(fmt.Sprintf("\n❌ Found %d blocking issue(s) - fix all above\n", issueCount))
		output.WriteString("⛔ BLOCKING: Must fix ALL errors above before continuing")
	} else {
		issueCount := len(issues)
		output.WriteString(fmt.Sprintf("\n⚠️  Found %d warning(s) - consider fixing\n", issueCount))
		output.WriteString("📝 NON-BLOCKING: Issues detected but you can continue")
	}

	return output.String()
}

// checkTestFile checks for an associated _test.go file and runs linting on it
func (e *LintingRuleEngine) checkTestFile(ctx context.Context, filePath string) {
	// Construct test file path
	base := strings.TrimSuffix(filePath, ".go")
	testPath := base + "_test.go"

	// Check if test file exists
	content, err := os.ReadFile(testPath)
	if err != nil {
		// No test file, that's ok
		return
	}

	// Run all applicable linters on test file in parallel
	results := e.executor.ExecuteLinters(ctx, e.linters, testPath, content)

	// Aggregate results
	aggregatedResult, errs := linters.AggregateResults(results)

	// Handle any linting errors
	for _, err := range errs {
		// Test file linting errors trigger exit code 1, shown on stderr
		fmt.Fprintf(os.Stderr, "\n> Test file linting error for %s: %v\n", testPath, err)
	}

	// Report any issues found in test file
	if len(aggregatedResult.Issues) > 0 {
		var errorIssues, warningIssues []linters.Issue
		for _, issue := range aggregatedResult.Issues {
			if issue.Severity == "error" {
				errorIssues = append(errorIssues, issue)
			} else {
				warningIssues = append(warningIssues, issue)
			}
		}

		// Test file issues trigger exit code 1, shown on stderr
		if len(errorIssues) > 0 {
			output := e.formatLintOutput(testPath, errorIssues, true)
			fmt.Fprintf(os.Stderr, "\n> Test file feedback:\n%s\n", output)
		} else if len(warningIssues) > 0 {
			output := e.formatLintOutput(testPath, warningIssues, false)
			fmt.Fprintf(os.Stderr, "\n> Test file feedback:\n%s\n", output)
		}
	}
}
