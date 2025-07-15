---
title: "Installation"
linkTitle: "Installation"
weight: 10
description: >
  Install CCFeedback CLI and integrate the Go library
---

# Installation

CCFeedback provides multiple installation options to fit your workflow.

## CLI Installation

### Using Go Install (Recommended)

```bash
go install github.com/jrossi-claude/ccfeedback/cmd/ccfeedback@latest
```

### Using Homebrew

```bash
brew tap jrossi-claude/ccfeedback
brew install ccfeedback
```

### Download Binary

Download the latest release from GitHub:

```bash
curl -L https://github.com/jrossi-claude/ccfeedback/releases/latest/download/ccfeedback-$(uname -s)-$(uname -m) \
  -o ccfeedback
chmod +x ccfeedback
sudo mv ccfeedback /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/jrossi-claude/ccfeedback.git
cd ccfeedback
make build
```

## Go Library Integration

Add CCFeedback to your Go project:

```bash
go get github.com/jrossi-claude/ccfeedback
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/jrossi-claude/ccfeedback"
)

func main() {
    // Create API instance with default configuration
    api := ccfeedback.NewAPI()

    // Process hook message
    message := `{"type": "PreToolUse", "tool": "bash", "command": "ls"}`
    result, err := api.ProcessHookMessage(context.Background(), message)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %+v", result)
}
```

## Verification

Verify your installation:

```bash
ccfeedback --version
```

## Next Steps

- [Quick Start Guide](/docs/quickstart/) - Get started with basic usage
- [Configuration](/docs/configuration/) - Set up your linting rules
- [CLI Reference](/docs/cli/) - Complete command documentation