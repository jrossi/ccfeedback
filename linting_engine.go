package ccfeedback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jrossi/ccfeedback/linters"
)

// LintingRuleEngine implements RuleEngine to provide linting functionality
type LintingRuleEngine struct {
	linters []linters.Linter
}

// NewLintingRuleEngine creates a new linting rule engine with default linters
func NewLintingRuleEngine() *LintingRuleEngine {
	return &LintingRuleEngine{
		linters: []linters.Linter{
			linters.NewGoLinter(),
			linters.NewMarkdownLinter(),
		},
	}
}

// AddLinter adds a custom linter to the engine
func (e *LintingRuleEngine) AddLinter(linter linters.Linter) {
	e.linters = append(e.linters, linter)
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

	// Find appropriate linter
	for _, linter := range e.linters {
		if linter.CanHandle(filePath) {
			result, err := linter.Lint(ctx, filePath, []byte(content))
			if err != nil {
				return &HookResponse{
					Decision: "block",
					Reason:   fmt.Sprintf("Linting error: %v", err),
				}, nil
			}

			// Check for issues and format detailed output
			var errorIssues, warningIssues []linters.Issue
			for _, issue := range result.Issues {
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
		}
	}

	// Write success message to stderr for user visibility
	fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: 👉 Style clean. Continue with your task.\n")
	return &HookResponse{Decision: "approve"}, nil
}

// EvaluatePostToolUse runs linters and tests after file operations
func (e *LintingRuleEngine) EvaluatePostToolUse(ctx context.Context, msg *PostToolUseMessage) (*HookResponse, error) {
	// Only check Write and Edit operations
	if msg.ToolName != "Write" && msg.ToolName != "Edit" && msg.ToolName != "MultiEdit" {
		return nil, nil
	}

	// Skip if there was an error
	if msg.ToolError != "" {
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
		// File might have been deleted or moved, that's ok
		return nil, nil
	}

	// Check each linter
	for _, linter := range e.linters {
		if linter.CanHandle(filePath) {
			result, err := linter.Lint(ctx, filePath, content)
			if err != nil {
				// Log error but don't block
				fmt.Fprintf(os.Stderr, "\n> Linting error for %s: %v\n", filePath, err)
				continue
			}

			// Check for issues and format detailed output
			var errorIssues, warningIssues []linters.Issue
			for _, issue := range result.Issues {
				if issue.Severity == "error" {
					errorIssues = append(errorIssues, issue)
				} else {
					warningIssues = append(warningIssues, issue)
				}
			}

			// Track if we found any issues
			hasIssues := false

			// Always provide feedback via stderr
			if len(errorIssues) > 0 {
				output := e.formatLintOutput(filePath, errorIssues, true)
				fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
				hasIssues = true
			} else if len(warningIssues) > 0 {
				output := e.formatLintOutput(filePath, warningIssues, false)
				fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n%s\n", output)
				hasIssues = true
			} else {
				fmt.Fprintf(os.Stderr, "\n> Write operation feedback:\n  - [ccfeedback]: 👉 Style clean. Continue with your task.\n")
			}

			// Check for associated test files if it's a Go file
			if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
				e.checkTestFile(ctx, filePath)
			}

			// Return a response that will trigger exit code 1 if there are issues
			// This makes the output visible while still being non-blocking
			if hasIssues {
				return &HookResponse{
					Decision: "approve",
					Message:  "Linting issues found - see stderr for details",
				}, nil
			}
		}
	}

	// Return nil for clean files
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
		output.WriteString("\n⚠️  Found formatting issues - consider running gofmt")
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

	// Run linters on test file
	for _, linter := range e.linters {
		if linter.CanHandle(testPath) {
			result, err := linter.Lint(ctx, testPath, content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n> Test file linting error for %s: %v\n", testPath, err)
				continue
			}

			// Report any issues found in test file
			if len(result.Issues) > 0 {
				var errorIssues, warningIssues []linters.Issue
				for _, issue := range result.Issues {
					if issue.Severity == "error" {
						errorIssues = append(errorIssues, issue)
					} else {
						warningIssues = append(warningIssues, issue)
					}
				}

				if len(errorIssues) > 0 {
					output := e.formatLintOutput(testPath, errorIssues, true)
					fmt.Fprintf(os.Stderr, "\n> Test file feedback:\n%s\n", output)
				} else if len(warningIssues) > 0 {
					output := e.formatLintOutput(testPath, warningIssues, false)
					fmt.Fprintf(os.Stderr, "\n> Test file feedback:\n%s\n", output)
				}
			}
		}
	}
}
