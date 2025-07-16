---
title: "CLI Reference"
linkTitle: "CLI"
weight: 50
description: >
  Complete command-line interface documentation
---

# CLI Reference

CCFeedback provides a command-line interface for processing Claude Code hooks and analyzing linting configurations.

## Installation

```bash
go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
```

## Commands

### Default Mode (Hook Processing)

Process Claude Code hook messages from stdin:

```bash
# Basic usage - reads from stdin
ccfeedback

# With custom timeout
ccfeedback -timeout 30s

# With custom configuration
ccfeedback -config my-config.json

# Debug mode
ccfeedback -debug
```

### init Command

Set up ccfeedback in Claude Code settings:

```bash
# Initialize ccfeedback hooks (updates both global and project settings)
ccfeedback init

# Only update global settings (~/.claude/settings.json)
ccfeedback init --global

# Only update project settings (.claude/settings.json)
ccfeedback init --project

# Preview changes without applying them
ccfeedback init --dry-run

# Apply changes without confirmation prompt
ccfeedback init --force

# Configure for specific tools only
ccfeedback init --matcher "Write"    # Only for Write tool
ccfeedback init --matcher "Edit"     # Only for Edit tool
ccfeedback init --matcher "Bash"     # Only for Bash tool
ccfeedback init --matcher ""         # All tools (default)
```

The `init` command:
- Adds ccfeedback as a PostToolUse hook in Claude Code settings
- Shows proposed changes in diff format before applying
- Creates timestamped backups of existing settings
- Preserves all existing configuration and custom fields
- Detects when ccfeedback is already configured

### show-actions Command

Analyze which configuration rules would apply to specific files:

```bash
# Show configuration for a single file
ccfeedback show-actions internal/api.go

# Multiple files
ccfeedback show-actions src/main.go pkg/util.go README.md

# With custom configuration
ccfeedback -config team-config.json show-actions pkg/api.go

# Debug mode shows pattern matching details
ccfeedback -debug show-actions internal/test.go
```

The `show-actions` command displays:
- Which linters apply to each file type
- Base configuration for applicable linters
- Rule hierarchy showing pattern matching order
- Final merged configuration after all rules are applied
- Configuration file loading order (in debug mode)

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-config` | Path to configuration file | Auto-detect |
| `-debug` | Enable debug output | false |
| `-timeout` | Hook execution timeout | 60s |
| `-version` | Show version information | - |

## Exit Codes

| Code | Description | Usage |
|------|-------------|-------|
| 0 | Success | Output shown in transcript (stdout) |
| 1 | Non-blocking error | General errors |
| 2 | Blocking error | Feedback processed by Claude (stderr) |

## Configuration

CCFeedback looks for configuration files in the following order:
1. `~/.claude/ccfeedback.json` (user global)
2. `.claude/ccfeedback.json` (project-specific)
3. `.claude/ccfeedback.local.json` (local overrides, git-ignored)
4. File specified with `-config` flag

## Hook Processing

### Processing Flow

1. Read hook message from stdin
2. Parse message type (PreToolUse, PostToolUse, etc.)
3. Apply configured rules based on tool and file patterns
4. Run applicable linters for file operations
5. Return response with decision and feedback

### Example Hook Message

```json
{
  "session_id": "123",
  "hook_event_name": "PostToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "src/main.go",
    "content": "package main\n\nfunc main() {\n\t// code\n}"
  }
}
```

### Example Response

```json
{
  "decision": "approve",
  "message": "✅ Style clean. Continue with your task."
}
```

## Integration Examples

### Claude Code Settings

After running `ccfeedback init`, your Claude Code settings will include:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "ccfeedback",
            "timeout": 60000,
            "continueOnError": false
          }
        ]
      }
    ]
  }
}
```

### Project Configuration

Example `.claude/ccfeedback.json`:

```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": ".golangci.yml"
      }
    },
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120
      }
    }
  },
  "rules": [
    {
      "pattern": "**/*_test.go",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["line-length"]
      }
    }
  ]
}
```

## Troubleshooting

### Common Issues

#### ccfeedback not found

Ensure ccfeedback is in your PATH:

```bash
# Check installation
which ccfeedback

# Add Go bin to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Configuration not loaded

Check which configuration files are being loaded:

```bash
ccfeedback -debug show-actions test.go
```

#### Hook not triggering

Verify Claude Code settings:

```bash
# Check if hook is configured
cat ~/.claude/settings.json | jq '.hooks.PostToolUse'

# Re-run init if needed
ccfeedback init
```

### Getting Help

```bash
# Show usage
ccfeedback -help

# Show version
ccfeedback -version
```

## Related Documentation

- [Installation Guide](/docs/installation/) - Detailed installation instructions
- [Configuration Guide](/docs/configuration/) - Configuration options and examples
- [Quick Start](/docs/quickstart/) - Getting started guide
- [Linters](/docs/linters/) - Available linters and their options