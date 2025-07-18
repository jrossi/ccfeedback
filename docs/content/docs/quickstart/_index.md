---
title: "Quick Start"
linkTitle: "Quick Start"
weight: 20
description: >
  Get started with Gismo in minutes
---

# Quick Start

Get Gismo running in your environment quickly with these examples.

## Prerequisites

- Go 1.23+ (for library usage)
- golangci-lint (recommended for Go linting)

## CLI Quick Start

### 1. Install Gismo

```bash
# Install with Homebrew (macOS/Linux)
brew tap jrossi/gismo https://github.com/jrossi/gismo
brew install jrossi/gismo/gismo

# Or download pre-built binary (Linux x86_64)
curl -L https://github.com/jrossi/gismo/releases/latest/download/gismo_Linux_x86_64.tar.gz | tar xz
sudo mv gismo /usr/local/bin/

# Or install with Go
go install github.com/jrossi/gismo/cmd/gismo@latest
```

### 2. Set Up Claude Code Integration

Initialize gismo as a Claude Code hook:

```bash
# Set up gismo in Claude Code settings
gismo init

# Preview changes without applying
gismo init --dry-run
```

This will configure gismo as a PostToolUse hook in your Claude Code settings.

### 3. Basic Usage

Process a Claude Code hook message:

```bash
echo '{"hook_event_name": "PostToolUse", "tool_name": "Write", "tool_input": {"file_path": "test.go"}}' | gismo
```

### 4. With Configuration

Create a basic configuration:

```bash
mkdir -p .claude
cat > .claude/gismo.json << 'EOF'
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
gismo --config .claude/gismo.json
```

## Library Quick Start

### 1. Add to Your Project

```bash
go mod init your-project
go get github.com/jrossi/gismo
```

### 2. Basic Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jrossi/gismo"
)

func main() {
    // Create API with default configuration
    api := gismo.NewAPI()

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

    "github.com/jrossi/gismo"
)

func main() {
    // Load configuration from file
    config, err := gismo.LoadConfig(".claude/gismo.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create API with custom configuration
    api := gismo.NewAPIWithConfig(config)

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

Use Gismo to validate code before commits:

```bash
# Check Go files
find . -name "*.go" -exec gismo --file {} \;

# Check markdown documentation
find docs -name "*.md" -exec gismo --file {} \;
```

### CI/CD Integration

Add to your GitHub Actions:

```yaml
- name: Run Gismo
  run: |
    go install github.com/jrossi/gismo/cmd/gismo@latest
    gismo --config .claude/gismo.json
```

### Claude Code Hooks

After running `gismo init`, your Claude Code settings will include:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "gismo",
            "timeout": 60000
          }
        ]
      }
    ]
  }
}
```

## What's Next?

- [Configuration Guide](../configuration/) - Detailed configuration options
- [CLI Reference](../cli-reference/) - Complete command documentation
- [Library API](../library/) - Full Go API reference
- [Linter Documentation](../linters/) - Language-specific linting guides