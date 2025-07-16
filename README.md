# CCFeedback - Claude Code Hooks Library

A high-performance Go library and CLI tool for handling Claude Code hooks with built-in linting capabilities.
Features automatic Go file formatting validation and test running. Built with
[go-json](https://github.com/goccy/go-json) for optimal JSON parsing performance.

## Features

- **Enhanced Go Linting**: golangci-lint integration with 30+ fast linters and intelligent fallback
- **Comprehensive Analysis**: Runs gosimple, ineffassign, gofmt, goimports, and many more linters
- **Fast Mode Optimization**: Uses golangci-lint's `--fast` flag for optimal individual file performance
- **Configuration Support**: Respects custom `.golangci.yml` configuration files
- **Module-Aware**: Correctly detects Go module roots and runs tests from proper directory
- **High Performance**: Uses go-json for 2-3x faster JSON parsing
- **Graceful Fallback**: Works even without golangci-lint installed
- **Fully Typed**: Strong typing for all hook message types
- **Extensible**: Pluggable linter and rule engine interface
- **Composable**: Chain multiple rule engines together
- **CLI Tool**: Ready-to-use command-line tool
- **Well Tested**: Comprehensive test coverage and benchmarks

## Installation

### Install with Homebrew (macOS/Linux)

```bash
brew tap jrossi/ccfeedback https://github.com/jrossi/ccfeedback
brew install jrossi/ccfeedback/ccfeedback
```

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/jrossi/ccfeedback/releases).

```bash
# Linux x86_64
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Linux_x86_64.tar.gz | tar xz

# macOS x86_64
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Darwin_x86_64.tar.gz | tar xz

# macOS arm64 (M1/M2)
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Darwin_arm64.tar.gz | tar xz

# Windows x86_64
# Download ccfeedback_Windows_x86_64.zip from releases page
```

### Install with Go

```bash
# Install the CLI tool
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest

# Use as a library
go get github.com/jrossi/ccfeedback
```

### Build from Source

```bash
git clone https://github.com/jrossi/ccfeedback.git
cd ccfeedback
make install
```

## Usage

### As a Library

```go
package main

import (
    "context"
    "github.com/jrossi/ccfeedback"
)

func main() {
    // Create API with default rule engine
    api := ccfeedback.New()

    // Or with a custom rule engine
    api := ccfeedback.NewWithRuleEngine(myRuleEngine)

    // Process stdin (for use as a hook)
    ctx := context.Background()
    if err := api.ProcessStdin(ctx); err != nil {
        // Handle error
    }
}
```

### Custom Rule Engine

```go
type MyRuleEngine struct{}

func (e *MyRuleEngine) EvaluatePreToolUse(ctx context.Context, msg *ccfeedback.PreToolUseMessage) (
    *ccfeedback.HookResponse, error) {
    // Block dangerous tools
    if msg.ToolName == "Bash" {
        return &ccfeedback.HookResponse{
            Decision: "block",
            Reason:   "Bash commands are not allowed",
        }, nil
    }
    return &ccfeedback.HookResponse{Decision: "approve"}, nil
}

// Implement other methods...
```

### Composite Rule Engines

```go
// Combine multiple rule engines
composite := ccfeedback.NewCompositeRuleEngine(
    securityEngine,
    loggingEngine,
    customEngine,
)

api := ccfeedback.NewWithRuleEngine(composite)
```

### Builder Pattern

```go
api := ccfeedback.NewBuilder().
    WithTimeout(30 * time.Second).
    WithRuleEngine(myEngine).
    RegisterHook(ccfeedback.HookConfig{
        Name:        "security-check",
        EventType:   ccfeedback.PreToolUseEvent,
        ToolPattern: "Write|Edit",
        Priority:    1,
        Timeout:     30 * time.Second,
    }).
    Build()
```

### CLI Tool

The CLI tool can be used as a hook processor or to analyze configuration:

#### Hook Processing Mode (Default)

Reads hook messages from stdin and writes responses to stdout:

```bash
# Basic usage
echo '{"session_id":"123","hook_event_name":"PreToolUse",...}' | ccfeedback

# With custom timeout
ccfeedback -timeout 30s

# Debug mode
ccfeedback -debug

# With custom configuration
ccfeedback -config my-config.json
```

#### Init Command

Set up ccfeedback in Claude Code settings:

```bash
# Initialize ccfeedback hooks in Claude Code settings
ccfeedback init

# Only update global settings (~/.claude/settings.json)
ccfeedback init --global

# Only update project settings (.claude/settings.json)
ccfeedback init --project

# Preview changes without applying them
ccfeedback init --dry-run

# Apply changes without confirmation prompt
ccfeedback init --force

# Configure for specific tools only (e.g., Write, Edit, Bash)
ccfeedback init --matcher "Write"
ccfeedback init --matcher "Bash"

# Empty matcher (default) matches all tools
ccfeedback init --matcher ""
```

The init command:
- Adds ccfeedback as a PostToolUse hook in Claude Code settings
- Shows proposed changes in diff format before applying
- Creates timestamped backups of existing settings
- Preserves all existing configuration and custom fields
- Detects when ccfeedback is already configured

#### Show Actions Command

Analyze which configuration rules would apply to specific files:

```bash
# Show configuration rules for a file
ccfeedback show-actions internal/api.go

# With custom configuration
ccfeedback -config team-config.json show-actions pkg/public/api.go

# Multiple files
ccfeedback show-actions internal/foo.go pkg/bar.go README.md

# Debug mode shows which patterns don't match
ccfeedback -debug show-actions internal/test.go
```

The show-actions command helps you understand:
- Which linters apply to each file type
- Base configuration for applicable linters
- Rule hierarchy showing pattern matching order
- Final merged configuration after all rules are applied
- Configuration file loading order (in debug mode)

### Go Linting Integration

CCFeedback provides comprehensive Go file linting with enhanced golangci-lint integration:

**Enhanced Linting with golangci-lint:**
- Automatically detects and uses golangci-lint for comprehensive analysis
- Runs golangci-lint in `--fast` mode for optimal performance on individual files
- Supports custom `.golangci.yml` configuration files
- Provides detailed issue reporting with line/column information
- Includes 30+ fast linters (gosimple, ineffassign, gofmt, goimports, etc.)

**Intelligent Fallback:**
- Gracefully falls back to basic `go/format` checking if golangci-lint is unavailable
- Maintains functionality even without golangci-lint installed
- Ensures consistent behavior across different development environments

**Pre-Write Validation:**
- Blocks writes of Go files with severe syntax errors
- Warns about linting issues (but allows the write)
- Skips generated files and testdata directories
- Module-aware operation for proper import resolution

**Performance Characteristics:**
- Enhanced linting: ~100ms per file (comprehensive analysis)
- Basic fallback: ~4μs per file (syntax/format only)
- Optimized for real-time development feedback

**Post-Write Actions:**
- Currently limited due to hook message structure - PostToolUse messages don't include file paths
- Test running is available during PreToolUse validation for immediate feedback
- All operations are module-aware and respect Go project structure

**Example Hook Configuration:**
```json
{
  "PreToolUse": [
    {
      "command": "/path/to/ccfeedback",
      "tool_patterns": ["Write", "Edit", "MultiEdit"]
    }
  ]
}
```

**Behavior Examples:**
```bash
# Clean Go code → Approved
{"decision": "approve"}

# Code with linting issues → Approved with detailed warnings
{
  "decision": "approve",
  "message": "Found 2 linting issues: Line 9: S1021 (gosimple), Line 13: needs gofmt"
}

# Syntax error → Blocked
{"decision": "block", "reason": "syntax: Go syntax error: missing ',' before newline"}

# golangci-lint unavailable → Basic linting fallback
{"decision": "approve", "message": "File test.go is not properly formatted. Consider running gofmt."}
```

## Hook Message Types

The library supports all Claude Code hook types:

- `PreToolUse`: Before tool execution
- `PostToolUse`: After tool execution
- `Notification`: System notifications
- `Stop`: Main agent completion
- `SubagentStop`: Subagent completion
- `PreCompact`: Before context compression

## Performance

Benchmarks show excellent performance across all components:

**Core System:**
- Message parsing: ~700ns per message
- Rule evaluation: <1ns for simple rules
- Full pipeline: ~22ns for handler processing

**Go Linting Performance:**
- Enhanced linting (golangci-lint --fast): ~100ms per file
- Basic fallback (go/format): ~4μs per file
- Performance optimized for real-time development feedback

## API Documentation

### Core Types

- `API`: Main interface for the library
- `RuleEngine`: Interface for custom rule implementations
- `Handler`: Processes hook messages
- `Parser`: High-performance JSON parser
- `Registry`: Manages hook configurations

### Response Format

Hook responses can use either exit codes or JSON:

**Exit Codes:**
- 0: Success (stdout shown)
- 2: Blocking error (stderr processed)
- Other: Non-blocking error

**JSON Response:**
```json
{
  "continue": false,
  "stopReason": "Security violation",
  "decision": "block",
  "reason": "Tool access denied"
}
```

## Examples

Working examples are included in the usage section above. For production use, ensure your hooks.json
configuration points to the installed ccfeedback binary location.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- Linting passes with no warnings
- Benchmarks show no performance regression

## License

[License to be determined]