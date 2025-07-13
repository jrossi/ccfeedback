package ccfeedback

import (
	"context"
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
			// Add more linters here as they're implemented
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
	filePath, ok := msg.ToolInput["file_path"].(string)
	if !ok {
		return &HookResponse{Decision: "approve"}, nil
	}

	// For Edit/MultiEdit, we can't lint until after the edit is done
	if msg.ToolName == "Edit" || msg.ToolName == "MultiEdit" {
		return &HookResponse{Decision: "approve"}, nil
	}

	// For Write operations, check the content
	content, ok := msg.ToolInput["content"].(string)
	if !ok {
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

	// For now, we'll need to get the file path from the original input
	// In a real implementation, we might need to track this state
	// This is a limitation of the current message structure

	// Since we can't easily get the file path from PostToolUse message,
	// we'll return nil for now. In a production system, we'd need to
	// either enhance the message structure or maintain state between
	// pre and post hooks.
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
