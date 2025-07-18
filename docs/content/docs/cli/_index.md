---
title: "CLI Reference"
linkTitle: "CLI"
weight: 50
description: >
  Complete command-line interface documentation
---

# CLI Reference

Gismo provides a command-line interface for processing Claude Code hooks and analyzing linting configurations.

## Installation

```bash
go install github.com/jrossi/gismo/cmd/gismo@latest
```

## Commands

### Default Mode (Hook Processing)

Process Claude Code hook messages from stdin:

```bash
# Basic usage - reads from stdin
gismo

# With custom timeout
gismo -timeout 30s

# With custom configuration
gismo -config my-config.json

# Debug mode
gismo -debug
```

### init Command

Set up gismo in Claude Code settings:

```bash
# Initialize gismo hooks (updates both global and project settings)
gismo init

# Only update global settings (~/.claude/settings.json)
gismo init --global

# Only update project settings (.claude/settings.json)
gismo init --project

# Preview changes without applying them
gismo init --dry-run

# Apply changes without confirmation prompt
gismo init --force

# Configure for specific tools only
gismo init --matcher "Write"    # Only for Write tool
gismo init --matcher "Edit"     # Only for Edit tool
gismo init --matcher "Bash"     # Only for Bash tool
gismo init --matcher ""         # All tools (default)
```

The `init` command:
- Adds gismo as a PostToolUse hook in Claude Code settings
- Shows proposed changes in diff format before applying
- Creates timestamped backups of existing settings
- Preserves all existing configuration and custom fields
- Detects when gismo is already configured

### show Command

The `show` command provides comprehensive visibility into gismo's configuration and behavior.
It includes several subcommands:

#### show config

Display the current merged configuration:

```bash
# Show current configuration
gismo show config

# With custom configuration file
gismo show --config team-config.json config

# Debug mode shows configuration sources
gismo show --debug config
```

#### show filter

Analyze which rules and linters apply to specific files:

```bash
# Show configuration for a single file
gismo show filter internal/api.go

# With custom configuration
gismo show --config team-config.json filter pkg/api.go

# Debug mode shows pattern matching details
gismo show --debug filter internal/test.go
```

#### show setup

Check gismo setup status:

```bash
# Show setup status
gismo show setup

# Debug mode includes environment details
gismo show --debug setup
```

#### show linters

List all available linters and their status:

```bash
# Show all linters
gismo show linters

# With custom configuration
gismo show --config team-config.json linters
```

#### Backward Compatibility

The old `show-actions` command still works and maps to `show filter`:

```bash
# These are equivalent:
gismo show-actions internal/api.go
gismo show filter internal/api.go
```

#### Show Command Features

- **`show config`**: Displays the complete merged configuration in JSON format
- **`show filter <file>`**: Shows which linters and rules apply to a specific file
- **`show setup`**: Checks binary availability, config files, and Claude integration
- **`show linters`**: Lists all linters with their supported files and tool requirements

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

Gismo looks for configuration files in the following order:
1. `~/.claude/gismo.json` (user global)
2. `.claude/gismo.json` (project-specific)
3. `.claude/gismo.local.json` (local overrides, git-ignored)
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

Example `.claude/gismo.json`:

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

#### gismo not found

Ensure gismo is in your PATH:

```bash
# Check installation
which gismo

# Add Go bin to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Configuration not loaded

Check which configuration files are being loaded:

```bash
gismo -debug show-actions test.go
```

#### Hook not triggering

Verify Claude Code settings:

```bash
# Check if hook is configured
cat ~/.claude/settings.json | jq '.hooks.PostToolUse'

# Re-run init if needed
gismo init
```

### Getting Help

```bash
# Show usage
gismo -help

# Show version
gismo -version
```

## Related Documentation

- [Installation Guide](/docs/installation/) - Detailed installation instructions
- [Configuration Guide](/docs/configuration/) - Configuration options and examples
- [Quick Start](/docs/quickstart/) - Getting started guide
- [Linters](/docs/linters/) - Available linters and their options