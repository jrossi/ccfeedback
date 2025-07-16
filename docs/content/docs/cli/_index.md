---
title: "CLI Reference"
linkTitle: "CLI"
weight: 50
description: >
  Complete command-line interface documentation
---

# CLI Reference

CCFeedback provides a comprehensive command-line interface for processing Claude Code hooks and running
standalone linting operations.

## Installation

```bash
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
```

## Basic Usage

### Process Hook Messages

```bash
# Process from stdin
echo '{"type": "PreToolUse", "tool": "bash"}' | ccfeedback

# Process with configuration
ccfeedback --config .claude/ccfeedback.json

# Process specific file
ccfeedback --file src/main.go
```

### Configuration

```bash
# Use custom config file
ccfeedback --config path/to/config.json

# Validate configuration
ccfeedback --config .claude/ccfeedback.json --validate

# Show configuration
ccfeedback --show-config
```

## Command Options

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Configuration file path | Auto-detect |
| `--verbose` | Enable verbose output | false |
| `--quiet` | Suppress non-error output | false |
| `--version` | Show version information | - |
| `--help` | Show help message | - |

### Processing Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--file` | Process specific file | stdin |
| `--filter` | File pattern filter | all |
| `--dry-run` | Show what would be processed | false |
| `--fix` | Auto-fix issues when possible | false |

### Output Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output format (json, text, compact) | text |
| `--no-color` | Disable color output | false |
| `--profile` | Enable performance profiling | false |

## Configuration Commands

### Show Configuration

Display current configuration:

```bash
ccfeedback --show-config
```

Example output:
```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "fastMode": true,
        "testTimeout": "5m"
      }
    }
  },
  "parallel": {
    "maxWorkers": 4
  }
}
```

### Validate Configuration

Check configuration syntax and settings:

```bash
ccfeedback --config .claude/ccfeedback.json --validate
```

### Show Available Actions

Display configured rule hierarchy:

```bash
ccfeedback --show-actions
```

## File Processing

### Single File

Process a specific file:

```bash
ccfeedback --file src/main.go
ccfeedback --file docs/README.md
ccfeedback --file config.json
```

### File Patterns

Process files matching patterns:

```bash
ccfeedback --filter "*.go"
ccfeedback --filter "src/**/*.js"
ccfeedback --filter "docs/*.md"
```

### Directory Processing

Process all files in a directory:

```bash
find src -name "*.go" -exec ccfeedback --file {} \;
```

## Output Formats

### Text Format (Default)

Human-readable output with colors:

```bash
ccfeedback --format text
```

Example output:
```
✓ src/main.go: All checks passed
✗ src/config.go: 2 issues found
  - Line 15: unused variable 'x'
  - Line 23: missing error check
```

### JSON Format

Machine-readable JSON output:

```bash
ccfeedback --format json
```

Example output:
```json
{
  "files": [
    {
      "path": "src/main.go",
      "status": "passed",
      "issues": []
    },
    {
      "path": "src/config.go",
      "status": "failed",
      "issues": [
        {
          "line": 15,
          "message": "unused variable 'x'",
          "severity": "warning"
        }
      ]
    }
  ]
}
```

### Compact Format

Minimal output for CI/CD:

```bash
ccfeedback --format compact
```

Example output:
```
src/main.go: PASS
src/config.go: FAIL (2 issues)
```

## Hook Processing

### Standard Input

Process Claude Code hook messages from stdin:

```bash
echo '{"type": "PreToolUse", "tool": "bash", "command": "go build"}' | ccfeedback
```

### File Input

Process hook message from file:

```bash
ccfeedback < hook-message.json
```

### Interactive Mode

Process multiple hook messages interactively:

```bash
ccfeedback --interactive
```

## Advanced Usage

### Performance Profiling

Enable performance profiling:

```bash
ccfeedback --profile --config .claude/ccfeedback.json
```

### Debug Mode

Enable verbose debugging output:

```bash
ccfeedback --verbose --config .claude/ccfeedback.json
```

### Dry Run

See what would be processed without executing:

```bash
ccfeedback --dry-run --filter "*.go"
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success - all checks passed |
| 1 | Linting issues found |
| 2 | Configuration error |
| 3 | File not found |
| 4 | Permission error |
| 5 | Timeout error |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CCFEEDBACK_CONFIG` | Default config file path | auto-detect |
| `CCFEEDBACK_VERBOSE` | Enable verbose output | false |
| `CCFEEDBACK_NO_COLOR` | Disable color output | false |
| `CCFEEDBACK_TIMEOUT` | Default timeout | 5m |

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
exec ccfeedback --config .claude/ccfeedback.json
```

### CI/CD Pipeline

```yaml
- name: Run CCFeedback
  run: |
    ccfeedback --config .claude/ccfeedback.json --format json > results.json
    if [ $? -ne 0 ]; then
      cat results.json
      exit 1
    fi
```

### Make Target

```makefile
lint:
	ccfeedback --config .claude/ccfeedback.json

lint-fix:
	ccfeedback --config .claude/ccfeedback.json --fix

lint-ci:
	ccfeedback --config .claude/ccfeedback.json --format compact
```

## Troubleshooting

### Common Issues

#### Configuration Not Found

```bash
ccfeedback --show-config
# Shows current config file location and contents
```

#### Permission Errors

```bash
# Ensure ccfeedback has execute permissions
chmod +x $(which ccfeedback)
```

#### Timeout Issues

```bash
# Increase timeout for large projects
ccfeedback --config .claude/ccfeedback.json --timeout 10m
```

### Getting Help

```bash
# Show help for all commands
ccfeedback --help

# Show version information
ccfeedback --version

# Show current configuration
ccfeedback --show-config
```

## Related Documentation

- [Installation Guide](/docs/installation/) - Installation methods
- [Configuration Guide](/docs/configuration/) - Configuration options
- [Quick Start](/docs/quickstart/) - Getting started examples