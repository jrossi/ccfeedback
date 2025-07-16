---
title: "Quick Start"
linkTitle: "Quick Start"
weight: 20
description: >
  Get started with CCFeedback in minutes
---

# Quick Start

Get CCFeedback running in your environment quickly with these examples.

## Prerequisites

- Go 1.23+ (for library usage)
- golangci-lint (recommended for Go linting)

## CLI Quick Start

### 1. Install CCFeedback

```bash
# Install with Homebrew (macOS/Linux)
brew tap jrossi/ccfeedback https://github.com/jrossi/ccfeedback
brew install jrossi/ccfeedback/ccfeedback

# Or download pre-built binary (Linux x86_64)
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Linux_x86_64.tar.gz | tar xz
sudo mv ccfeedback /usr/local/bin/

# Or install with Go
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
```

### 2. Basic Usage

Process a Claude Code hook message:

```bash
echo '{"type": "PreToolUse", "tool": "bash", "command": "ls"}' | ccfeedback
```

### 3. With Configuration

Create a basic configuration:

```bash
mkdir -p .claude
cat > .claude/ccfeedback.json << 'EOF'
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "disabledChecks": [],
        "testTimeout": "5m"
      }
    },
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120,
        "requireFrontmatter": false
      }
    }
  }
}
EOF
```

Run with configuration:

```bash
ccfeedback --config .claude/ccfeedback.json
```

## Library Quick Start

### 1. Add to Your Project

```bash
go mod init your-project
go get github.com/jrossi/ccfeedback
```

### 2. Basic Library Usage

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
        "command": "go fmt ./..."
    }`

    result, err := api.ProcessHookMessage(context.Background(), message)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Hook result: %+v\n", result)
}
```

### 3. Custom Configuration

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

    // Use the API
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

## Common Use Cases

### Code Quality Checks

Use CCFeedback to validate code before commits:

```bash
# Check Go files
find . -name "*.go" -exec ccfeedback --file {} \;

# Check markdown documentation
find docs -name "*.md" -exec ccfeedback --file {} \;
```

### CI/CD Integration

Add to your GitHub Actions:

```yaml
- name: Run CCFeedback
  run: |
    go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
    ccfeedback --config .claude/ccfeedback.json
```

### Claude Code Hooks

Set up as a Claude Code hook processor:

```json
{
  "hooks": {
    "pre_tool_use": "ccfeedback"
  }
}
```

## What's Next?

- [Configuration Guide](../configuration/) - Detailed configuration options
- [CLI Reference](../cli-reference/) - Complete command documentation
- [Library API](../library/) - Full Go API reference
- [Linter Documentation](../linters/) - Language-specific linting guides