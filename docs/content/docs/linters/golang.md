---
title: "Go Linting"
linkTitle: "Go"
weight: 10
description: >
  Enhanced Go linting with golangci-lint integration
---

# Go Linting

Gismo provides comprehensive Go linting through golangci-lint integration with intelligent fallback
capabilities. The system operates through a sophisticated multi-phase architecture designed to provide
real-time feedback during development while maintaining optimal performance.

## Architecture & Execution Flow

### Hook-Based Architecture

Gismo integrates with Claude Code through two main execution phases:

#### **PreToolUse Hook** (Content Validation)
Triggered **BEFORE** content is written to disk during Write operations:

1. **Syntax Validation**: Parses Go AST to detect syntax errors
2. **Format Checking**: Runs `gofmt` to ensure proper formatting
3. **Basic Linting**: Performs lightweight checks using Go's built-in tools
4. **Blocking Behavior**: **BLOCKS** the write operation if critical errors are found

#### **PostToolUse Hook** (Comprehensive Analysis)
Triggered **AFTER** file modifications are written to disk:

1. **Module Detection**: Automatically discovers `go.mod` files and module boundaries
2. **Parallel Execution**: Processes multiple files concurrently using worker pools
3. **Enhanced Linting**: Runs full golangci-lint analysis with 30+ linters
4. **Test Execution**: Automatically runs relevant tests for test files
5. **Result Aggregation**: Collects and formats all issues for Claude Code feedback

### Three-Tier Fallback System

Gismo implements intelligent fallbacks to ensure consistent operation:

**Tier 1: Enhanced Analysis (golangci-lint)**
- **Primary Mode**: Uses golangci-lint with `--fast` flag for optimal performance
- **Module Context**: Automatically runs from Go module root for proper import resolution
- **Custom Configuration**: Supports `.golangci.yml` files for team-specific rules
- **Performance**: ~100-500ms per file with comprehensive analysis

**Tier 2: Basic Linting (Go Built-in Tools)**
- **Fallback Mode**: When golangci-lint is unavailable or fails
- **Core Checks**: Uses `go/format`, `go vet`, and `go/types` for essential validation
- **Performance**: ~4μs per file for syntax and basic formatting checks
- **Reliability**: Always available with any Go installation

**Tier 3: Graceful Degradation**
- **Safety Net**: Never blocks development workflow due to tooling issues
- **Minimal Validation**: Performs basic syntax checking only
- **Logging**: Records warnings about missing tools for later resolution

### Detailed Execution Flow

When Claude Code modifies a Go file, the following sequence occurs:

```text
┌─ Claude Code File Operation ─┐
│                              │
▼                              │
PreToolUse Hook                │
├─ Syntax validation           │
├─ Format checking (gofmt)     │
└─ BLOCK if errors found ──────┘
│
▼ (if no blocking errors)
File Written to Disk
│
▼
PostToolUse Hook
├─ Module detection & caching
├─ Parallel worker allocation
├─ Enhanced linting (golangci-lint)
├─ Test discovery & execution
└─ Result aggregation
│
▼
Results sent to Claude Code
├─ Exit Code 0: Success (logged)
└─ Exit Code 2: Issues found (shown to Claude)
```

### Performance Optimizations

**Intelligent Caching**
- Module root detection results are cached to avoid repeated filesystem walks
- golangci-lint binary location is cached after first discovery
- Test pattern generation uses AST parsing with fallback to filename patterns

**Parallel Processing**
- Worker pool size defaults to `runtime.NumCPU()` for optimal resource utilization
- Batching support groups multiple files for efficient linter execution
- Context cancellation ensures proper timeout handling

**Smart Filtering**
- Automatically skips generated files (`// Code generated` comments)
- Ignores test data directories (`/testdata/` paths)
- Excludes temporary files during test execution

## Features

- **30+ Fast Linters**: bodyclose, errcheck, gofmt, goimports, gosimple, govet, ineffassign, misspell,
  staticcheck, typecheck, unconvert, unused, gosec
- **Module-Aware**: Automatically detects Go module roots and runs tests from proper directory
- **Fast Mode**: Uses `--fast` flag for optimal individual file performance
- **Graceful Fallback**: Works even without golangci-lint installed
- **Custom Configuration**: Respects `.golangci.yml` configuration files
- **Test Integration**: Smart test discovery and execution with pattern optimization
- **Parallel Execution**: Concurrent processing with configurable worker pools
- **Context Awareness**: Proper timeout and cancellation handling

## Basic Configuration

```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "disabledChecks": [],
        "testTimeout": "5m",
        "fastMode": true
      }
    }
  }
}
```

## Advanced Configuration

### Custom golangci-lint Config

```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": ".golangci.yml",
        "disabledChecks": ["gofmt", "gosec"],
        "enabledChecks": ["bodyclose", "errcheck", "goimports"],
        "testTimeout": "10m",
        "fastMode": true
      }
    }
  }
}
```

### Pattern-Based Rules

```json
{
  "rules": [
    {
      "pattern": "*_test.go",
      "linter": "golang",
      "rules": {
        "testTimeout": "15m",
        "disabledChecks": ["dupl", "gocyclo"]
      }
    },
    {
      "pattern": "internal/**",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["gochecknoglobals"]
      }
    },
    {
      "pattern": "*.generated.go",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["all"]
      }
    }
  ]
}
```

## Supported Checks

### Enabled by Default

| Check | Description |
|-------|-------------|
| `bodyclose` | Checks HTTP response body closure |
| `errcheck` | Checks for unchecked errors |
| `gofmt` | Checks Go formatting |
| `goimports` | Checks import formatting |
| `gosimple` | Suggests code simplifications |
| `govet` | Go vet analysis |
| `ineffassign` | Detects ineffectual assignments |
| `misspell` | Finds misspelled words |
| `staticcheck` | Advanced static analysis |
| `typecheck` | Type checking |
| `unconvert` | Removes unnecessary conversions |
| `unused` | Finds unused code |
| `gosec` | Security analysis |

### Additional Available Checks

| Check | Description |
|-------|-------------|
| `dupl` | Code duplication detection |
| `gocyclo` | Cyclomatic complexity |
| `gochecknoglobals` | Global variable detection |
| `golint` | Go lint suggestions |
| `maligned` | Struct alignment optimization |
| `prealloc` | Slice preallocation suggestions |

## Performance Modes

### Fast Mode (Recommended)

```json
{
  "config": {
    "fastMode": true
  }
}
```

Fast mode uses golangci-lint's `--fast` flag for optimal performance when linting individual files.

### Enhanced Mode

```json
{
  "config": {
    "fastMode": false,
    "enabledChecks": ["bodyclose", "errcheck", "gosec", "staticcheck"]
  }
}
```

Enhanced mode runs full analysis with all configured checks.

## Linter Execution Order & Details

### Pre-Tool Phase (Content Validation)

When Claude Code prepares to write Go content, the following checks run in order:

1. **AST Parsing** (`go/parser`)
  - Validates Go syntax before file is written
  - Detects syntax errors that would break compilation
  - **Blocks write operation** if parsing fails

2. **Format Validation** (`gofmt`)
  - Checks if code meets Go formatting standards
  - Compares current formatting with gofmt output
  - **Blocks write operation** for critical formatting issues

### Post-Tool Phase (Comprehensive Analysis)

After the file is successfully written, comprehensive analysis begins:

1. **Module Discovery & Caching**
  - Walks directory tree upward to find `go.mod` files
  - Caches module root locations for performance
  - Determines proper execution context for linting and testing

2. **golangci-lint Execution** (Enhanced Mode)
  - **Discovery**: Locates golangci-lint binary (cached after first run)
  - **Configuration**: Loads `.golangci.yml` if present, otherwise uses defaults
  - **Execution**: Runs with `--fast` flag for individual file analysis
  - **Output Processing**: Parses JSON output for structured issue reporting

3. **Test Discovery & Execution** (for `*_test.go` files)
  - **AST Analysis**: Parses test file to extract actual test function names
  - **Pattern Generation**: Creates optimized test patterns:
    - Single test: `^TestSpecificFunction$`
    - Common prefix: `^TestCommon` (for `TestCommonFoo`, `TestCommonBar`)
    - Multiple tests: `^(TestFoo|TestBar|TestBaz)$`
  - **Execution**: Runs `go test` from module root with generated patterns
  - **Timeout**: Respects configured `testTimeout` (default: 5 minutes)

4. **Fallback Linting** (Basic Mode)
  - **Trigger**: Activates when golangci-lint is unavailable or fails
  - **Tools Used**: `go/format`, `go vet`, `go/types`
  - **Performance**: ~4μs per file vs ~100-500ms for enhanced mode
  - **Coverage**: Basic syntax, formatting, and type checking only

### Result Processing & Exit Codes

The system uses specific exit codes to communicate with Claude Code:

- **Exit Code 0**: Success - results logged to transcript, no Claude intervention
- **Exit Code 2**: Issues found - stderr content processed by Claude for action

## Module Detection

Gismo automatically detects Go modules and adjusts behavior:

1. **Module Root Detection**: Finds `go.mod` files to determine module boundaries
2. **Test Execution**: Runs tests from the module root directory
3. **Import Path Resolution**: Resolves import paths relative to module root
4. **Dependency Management**: Respects `go.mod` and `go.sum` files
5. **Caching Strategy**: Stores module information to avoid repeated filesystem operations

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check only changed Go files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')
if [ -n "$changed_files" ]; then
    for file in $changed_files; do
        gismo --file "$file"
    done
fi
```

### Makefile Integration

```makefile
.PHONY: lint lint-fix test-go

lint:
	gismo --config .claude/gismo.json

lint-fix:
	golangci-lint run --fix ./...
	gofmt -s -w .
	goimports -w .

test-go:
	go test -v -race ./...
```

### CI/CD Pipeline

```yaml
name: Go Quality
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'

    - name: Install golangci-lint
      run: |
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
        sh -s -- -b $(go env GOPATH)/bin v1.55.0

    - name: Install gismo
      run: go install github.com/jrossi/gismo/cmd/gismo@latest

    - name: Lint code
      run: gismo --config .claude/gismo.json
```

## Troubleshooting

### Common Issues

#### golangci-lint Not Found

If golangci-lint is not installed, Gismo falls back to basic Go tools:

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
sh -s -- -b $(go env GOPATH)/bin
```

#### Slow Performance

For large codebases, enable fast mode:

```json
{
  "config": {
    "fastMode": true,
    "testTimeout": "30s"
  }
}
```

#### Module Detection Issues

Ensure your project has a valid `go.mod` file:

```bash
go mod init your-module-name
go mod tidy
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
gismo --config .claude/gismo.json --verbose
```

## Best Practices

### Configuration

1. **Start with defaults**: Enable golang linter with minimal configuration
2. **Use fast mode**: Enable for individual file linting
3. **Disable selectively**: Disable specific checks rather than entire linter
4. **Pattern-based rules**: Use different rules for tests vs production code

### Code Quality

1. **Fix formatting first**: Address gofmt and goimports issues
2. **Handle errors**: Focus on errcheck violations
3. **Security review**: Pay attention to gosec findings
4. **Performance**: Consider ineffassign and unused code suggestions

### Team Standards

```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": ".golangci.yml",
        "fastMode": true,
        "testTimeout": "10m"
      }
    }
  },
  "rules": [
    {
      "pattern": "cmd/**",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["gochecknoglobals"]
      }
    },
    {
      "pattern": "*_test.go",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["dupl"]
      }
    }
  ]
}
```

## Advanced Configuration

### Hook Configuration Sources

Gismo loads configuration from multiple sources in order of precedence:

1. **Global**: `~/.claude/gismo.json` (user-wide settings)
2. **Project**: `<project>/.claude/gismo.json` (project-specific settings)
3. **Local**: `<project>/.claude/gismo.local.json` (local overrides, gitignored)
4. **Command-line**: `--config` flag (highest precedence)

### Performance Tuning

**For Large Codebases:**
```json
{
  "linters": {
    "golang": {
      "config": {
        "fastMode": true,
        "testTimeout": "30s",
        "disabledChecks": ["dupl", "gocyclo", "maligned"]
      }
    }
  }
}
```

**For Security-Critical Projects:**
```json
{
  "linters": {
    "golang": {
      "config": {
        "enabledChecks": ["gosec", "errcheck", "staticcheck"],
        "fastMode": false,
        "testTimeout": "15m"
      }
    }
  }
}
```

### Understanding Hook Output

When viewing `gismo show <file.go>`, you can see the complete execution flow:

- **Green sections**: Successfully configured and operational
- **Yellow sections**: Warnings or fallback modes active
- **Red sections**: Errors or missing dependencies

This visualization helps debug configuration issues and understand why certain linters
may not be running as expected.

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [CLI Reference](/docs/cli/) - Command-line usage
- [Library API](/docs/library/) - Go integration examples
- [golangci-lint Documentation](https://golangci-lint.run/) - External linter configuration