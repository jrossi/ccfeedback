# CCFeedback - Claude Code Hooks Library

A high-performance Go library and CLI tool for handling Claude Code hooks with customizable rule engines. Built with [go-json](https://github.com/goccy/go-json) for optimal JSON parsing performance.

## Features

- **High Performance**: Uses go-json for 2-3x faster JSON parsing
- **Fully Typed**: Strong typing for all hook message types
- **Extensible**: Pluggable rule engine interface
- **Composable**: Chain multiple rule engines together
- **CLI Tool**: Ready-to-use command-line tool
- **Well Tested**: Comprehensive test coverage and benchmarks

## Installation

```bash
go get github.com/jrossi/ccfeedback
```

### CLI Tool

```bash
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
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

func (e *MyRuleEngine) EvaluatePreToolUse(ctx context.Context, msg *ccfeedback.PreToolUseMessage) (*ccfeedback.HookResponse, error) {
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
        Name:      "security-check",
        EventType: ccfeedback.PreToolUseEvent,
        Priority:  1,
    }).
    Build()
```

### CLI Tool

The CLI tool reads hook messages from stdin and writes responses to stdout:

```bash
# Basic usage
echo '{"session_id":"123","hook_event_name":"PreToolUse",...}' | ccfeedback

# With custom timeout
ccfeedback -timeout 30s

# Debug mode
ccfeedback -debug
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

Benchmarks show excellent performance:

- Message parsing: ~700ns per message
- Rule evaluation: <1ns for simple rules
- Full pipeline: ~22ns for handler processing

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

See the `examples/` directory for more complete examples (to be added).

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- Linting passes with no warnings
- Benchmarks show no performance regression

## License

[License to be determined]