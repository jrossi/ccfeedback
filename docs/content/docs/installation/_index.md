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

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/jrossi/ccfeedback/releases).

```bash
# Linux x86_64
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Linux_x86_64.tar.gz | tar xz
sudo mv ccfeedback /usr/local/bin/

# macOS x86_64
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Darwin_x86_64.tar.gz | tar xz
sudo mv ccfeedback /usr/local/bin/

# macOS arm64 (M1/M2)
curl -L https://github.com/jrossi/ccfeedback/releases/latest/download/ccfeedback_Darwin_arm64.tar.gz | tar xz
sudo mv ccfeedback /usr/local/bin/

# Windows x86_64
# Download ccfeedback_Windows_x86_64.zip from releases page
```

### Using Go Install

```bash
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
```

### Build from Source

```bash
git clone https://github.com/jrossi/ccfeedback.git
cd ccfeedback
make install
```

## Go Library Integration

Add CCFeedback to your Go project:

```bash
go get github.com/jrossi/ccfeedback
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/jrossi/ccfeedback"
)

func main() {
    // Create API instance with default configuration
    api := ccfeedback.New()

    // Process hook message from stdin
    ctx := context.Background()
    if err := api.ProcessStdin(ctx); err != nil {
        log.Fatal(err)
    }
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