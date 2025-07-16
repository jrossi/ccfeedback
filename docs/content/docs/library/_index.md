---
title: "Library API"
linkTitle: "Library"
weight: 60
description: >
  Go library API documentation and integration examples
---

# Library API

CCFeedback provides a comprehensive Go library for integrating hook processing and linting capabilities
into your applications.

## Installation

```bash
go get github.com/jrossi/ccfeedback
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jrossi/ccfeedback"
)

func main() {
    // Create API with default configuration
    api := ccfeedback.NewAPI()

    // Process a hook message
    message := `{
        "type": "PreToolUse",
        "tool": "bash",
        "command": "go build"
    }`

    result, err := api.ProcessHookMessage(context.Background(), message)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Hook processed: %+v\n", result)
}
```

### With Custom Configuration

```go
package main

import (
    "context"
    "log"

    "github.com/jrossi/ccfeedback"
)

func main() {
    // Load configuration from file
    config, err := ccfeedback.LoadConfig(".claude/ccfeedback.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create API with custom configuration
    api := ccfeedback.NewAPIWithConfig(config)

    // Process hook message
    result, err := api.ProcessHookMessage(context.Background(), `{
        "type": "PreToolUse",
        "tool": "edit",
        "file": "main.go"
    }`)

    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Processing result: %+v", result)
}
```

## Core API

### API Creation

```go
// Create with default configuration
api := ccfeedback.NewAPI()

// Create with custom configuration
config := &ccfeedback.Config{
    Linters: map[string]ccfeedback.LinterConfig{
        "golang": {
            Enabled: true,
            Config: map[string]interface{}{
                "fastMode": true,
            },
        },
    },
}
api := ccfeedback.NewAPIWithConfig(config)

// Create with custom rule engine
engine := &MyCustomEngine{}
api := ccfeedback.NewAPIWithEngine(engine)
```

### Hook Processing

```go
// Process hook message from string
result, err := api.ProcessHookMessage(ctx, messageJSON)

// Process hook message from struct
msg := &ccfeedback.HookMessage{
    Type: "PreToolUse",
    Tool: "bash",
    Command: "go test",
}
result, err := api.ProcessHook(ctx, msg)

// Process file directly
result, err := api.ProcessFile(ctx, "src/main.go")
```

## Configuration API

### Loading Configuration

```go
// Load from file
config, err := ccfeedback.LoadConfig("config.json")

// Load with search paths
config, err := ccfeedback.LoadConfigWithPaths([]string{
    ".claude/ccfeedback.json",
    "~/.claude/ccfeedback.json",
})

// Default configuration
config := ccfeedback.DefaultConfig()
```

### Configuration Structure

```go
type Config struct {
    Linters  map[string]LinterConfig `json:"linters"`
    Rules    []Rule                  `json:"rules"`
    Parallel ParallelConfig          `json:"parallel"`
    Timeout  string                  `json:"timeout"`
}

type LinterConfig struct {
    Enabled bool                   `json:"enabled"`
    Config  map[string]interface{} `json:"config"`
}

type Rule struct {
    Pattern string                 `json:"pattern"`
    Linter  string                 `json:"linter"`
    Rules   map[string]interface{} `json:"rules"`
}
```

## Rule Engines

### Built-in Engines

```go
// Golang linting engine
engine := &ccfeedback.GolangEngine{
    Config: ccfeedback.GolangConfig{
        FastMode: true,
        TestTimeout: "5m",
    },
}

// Markdown linting engine
engine := &ccfeedback.MarkdownEngine{
    Config: ccfeedback.MarkdownConfig{
        MaxLineLength: 120,
        RequireFrontmatter: false,
    },
}

// Composite engine (multiple engines)
composite := ccfeedback.NewCompositeRuleEngine(
    &ccfeedback.GolangEngine{},
    &ccfeedback.MarkdownEngine{},
    &ccfeedback.JSONEngine{},
)
```

### Custom Rule Engine

```go
type MyRuleEngine struct {
    config MyConfig
}

func (e *MyRuleEngine) ShouldProcess(ctx context.Context, msg *ccfeedback.HookMessage) (bool, error) {
    // Determine if this engine should process the message
    return msg.Tool == "myTool", nil
}

func (e *MyRuleEngine) ProcessMessage(ctx context.Context, msg *ccfeedback.HookMessage) (*ccfeedback.Result, error) {
    // Custom processing logic
    return &ccfeedback.Result{
        Success: true,
        Message: "Custom processing completed",
    }, nil
}

// Use custom engine
api := ccfeedback.NewAPIWithEngine(&MyRuleEngine{})
```

## Message Types

### Hook Message Structure

```go
type HookMessage struct {
    Type      string                 `json:"type"`
    Tool      string                 `json:"tool"`
    Command   string                 `json:"command,omitempty"`
    File      string                 `json:"file,omitempty"`
    Content   string                 `json:"content,omitempty"`
    Arguments map[string]interface{} `json:"arguments,omitempty"`
}
```

### Supported Hook Types

```go
const (
    PreToolUse     = "PreToolUse"
    PostToolUse    = "PostToolUse"
    Notification   = "Notification"
    Stop           = "Stop"
    SubagentStop   = "SubagentStop"
    PreCompact     = "PreCompact"
)
```

### Result Structure

```go
type Result struct {
    Success   bool                   `json:"success"`
    Message   string                 `json:"message"`
    Issues    []Issue                `json:"issues,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Duration  time.Duration          `json:"duration"`
}

type Issue struct {
    File     string `json:"file"`
    Line     int    `json:"line,omitempty"`
    Column   int    `json:"column,omitempty"`
    Message  string `json:"message"`
    Severity string `json:"severity"`
    Rule     string `json:"rule,omitempty"`
}
```

## Advanced Usage

### Parallel Processing

```go
// Configure parallel execution
config := &ccfeedback.Config{
    Parallel: ccfeedback.ParallelConfig{
        MaxWorkers: 8,
        DisableParallel: false,
    },
}

api := ccfeedback.NewAPIWithConfig(config)
```

### Context and Cancellation

```go
// Process with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := api.ProcessHookMessage(ctx, message)

// Process with cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(10 * time.Second)
    cancel() // Cancel after 10 seconds
}()

result, err := api.ProcessHookMessage(ctx, message)
```

### Error Handling

```go
result, err := api.ProcessHookMessage(ctx, message)
if err != nil {
    switch {
    case errors.Is(err, ccfeedback.ErrConfigNotFound):
        log.Println("Configuration file not found")
    case errors.Is(err, ccfeedback.ErrInvalidMessage):
        log.Println("Invalid hook message format")
    case errors.Is(err, ccfeedback.ErrTimeout):
        log.Println("Processing timed out")
    default:
        log.Printf("Unexpected error: %v", err)
    }
}

// Check result for issues
if !result.Success {
    for _, issue := range result.Issues {
        fmt.Printf("%s:%d: %s (%s)\n",
            issue.File, issue.Line, issue.Message, issue.Severity)
    }
}
```

## Performance Considerations

### Memory Usage

```go
// For large files, consider streaming
config := &ccfeedback.Config{
    Linters: map[string]ccfeedback.LinterConfig{
        "json": {
            Config: map[string]interface{}{
                "streamingMode": true,
                "maxFileSize": "100MB",
            },
        },
    },
}
```

### Caching

```go
// Enable rule engine caching
api := ccfeedback.NewAPI()
api.EnableCaching(true)

// Custom cache implementation
cache := &MyCustomCache{}
api.SetCache(cache)
```

## Integration Examples

### HTTP Server

```go
package main

import (
    "encoding/json"
    "net/http"
    "log"

    "github.com/jrossi/ccfeedback"
)

func main() {
    api := ccfeedback.NewAPI()

    http.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
        var msg ccfeedback.HookMessage
        if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        result, err := api.ProcessHook(r.Context(), &msg)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    })

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### CLI Tool

```go
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"

    "github.com/jrossi/ccfeedback"
)

func main() {
    var (
        configFile = flag.String("config", "", "Configuration file")
        file       = flag.String("file", "", "File to process")
    )
    flag.Parse()

    var api *ccfeedback.API
    if *configFile != "" {
        config, err := ccfeedback.LoadConfig(*configFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
            os.Exit(1)
        }
        api = ccfeedback.NewAPIWithConfig(config)
    } else {
        api = ccfeedback.NewAPI()
    }

    var result *ccfeedback.Result
    var err error

    if *file != "" {
        result, err = api.ProcessFile(context.Background(), *file)
    } else {
        // Read from stdin
        var msg ccfeedback.HookMessage
        if err := json.NewDecoder(os.Stdin).Decode(&msg); err != nil {
            fmt.Fprintf(os.Stderr, "Error reading message: %v\n", err)
            os.Exit(1)
        }
        result, err = api.ProcessHook(context.Background(), &msg)
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if !result.Success {
        os.Exit(1)
    }
}
```

## Testing

### Unit Tests

```go
package main

import (
    "context"
    "testing"

    "github.com/jrossi/ccfeedback"
)

func TestAPIProcessing(t *testing.T) {
    api := ccfeedback.NewAPI()

    msg := &ccfeedback.HookMessage{
        Type: "PreToolUse",
        Tool: "bash",
        Command: "echo test",
    }

    result, err := api.ProcessHook(context.Background(), msg)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    if !result.Success {
        t.Errorf("Expected success, got failure: %s", result.Message)
    }
}
```

### Benchmarks

```go
func BenchmarkAPIProcessing(b *testing.B) {
    api := ccfeedback.NewAPI()
    msg := &ccfeedback.HookMessage{
        Type: "PreToolUse",
        Tool: "bash",
        Command: "echo test",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := api.ProcessHook(context.Background(), msg)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Related Documentation

- [Configuration Guide](/docs/configuration/) - Configuration options
- [CLI Reference](/docs/cli/) - Command-line interface
- [Linter Documentation](/docs/linters/) - Language-specific linting