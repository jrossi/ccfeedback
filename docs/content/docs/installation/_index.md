---
title: "Installation"
linkTitle: "Installation"
weight: 10
description: >
  Install Gismo CLI and integrate the Go library
---

# Installation

Gismo provides multiple installation options to fit your workflow.

## CLI Installation

### Install with Homebrew (Recommended for macOS/Linux)

```bash
brew tap jrossi/gismo https://github.com/jrossi/gismo
brew install jrossi/gismo/gismo
```

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/jrossi/gismo/releases).

```bash
# Linux x86_64
curl -L https://github.com/jrossi/gismo/releases/latest/download/gismo_Linux_x86_64.tar.gz | tar xz
sudo mv gismo /usr/local/bin/

# macOS x86_64
curl -L https://github.com/jrossi/gismo/releases/latest/download/gismo_Darwin_x86_64.tar.gz | tar xz
sudo mv gismo /usr/local/bin/

# macOS arm64 (M1/M2)
curl -L https://github.com/jrossi/gismo/releases/latest/download/gismo_Darwin_arm64.tar.gz | tar xz
sudo mv gismo /usr/local/bin/

# Windows x86_64
# Download gismo_Windows_x86_64.zip from releases page
```

### Using Go Install

```bash
go install github.com/jrossi/gismo/cmd/gismo@latest
```

### Build from Source

```bash
git clone https://github.com/jrossi/gismo.git
cd gismo
make install
```

## Go Library Integration

Add Gismo to your Go project:

```bash
go get github.com/jrossi/gismo
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/jrossi/gismo"
)

func main() {
    // Create API instance with default configuration
    api := gismo.New()

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
gismo --version
```

## Next Steps

- [Quick Start Guide](../quickstart/) - Get started with basic usage
- [Configuration](../configuration/) - Set up your linting rules
- [CLI Reference](../cli-reference/) - Complete command documentation