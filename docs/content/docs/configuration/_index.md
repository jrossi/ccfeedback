---
title: "Configuration"
linkTitle: "Configuration"
weight: 30
description: >
  Configure Gismo linters and rules for your project
---

# Configuration

Gismo uses a flexible configuration system that supports hierarchical loading and pattern-based rule overrides.

## Configuration Loading Order

Gismo loads configuration files in this order (later files override earlier ones):

1. `~/.claude/gismo.json` - User's global configuration
2. `PROJECT_DIR/.claude/gismo.json` - Project-specific configuration
3. `PROJECT_DIR/.claude/gismo.local.json` - Local overrides (git-ignored)

You can also specify a custom configuration file:

```bash
gismo --config path/to/config.json
```

## Basic Configuration

### Simple Setup

```json
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
```

### Global Settings

```json
{
  "parallel": {
    "maxWorkers": 4,
    "disableParallel": false
  },
  "timeout": "5m"
}
```

## Linter-Specific Configuration

### Go Linting

```json
{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": "path/to/.golangci.yml",
        "disabledChecks": ["gofmt", "gosec"],
        "testTimeout": "10m",
        "enabledChecks": ["bodyclose", "errcheck", "goimports"],
        "fastMode": true
      }
    }
  }
}
```

### Markdown Linting

```json
{
  "linters": {
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120,
        "requireFrontmatter": false,
        "maxBlankLines": 2,
        "listIndentSize": 2,
        "disabledRules": ["line-length"],
        "frontmatterSchema": "path/to/schema.json"
      }
    }
  }
}
```

### Python Linting

```json
{
  "linters": {
    "python": {
      "enabled": true,
      "config": {
        "tool": "uv",
        "lintCommand": "ruff check",
        "formatCommand": "ruff format",
        "timeout": "30s"
      }
    }
  }
}
```

### JavaScript Linting

```json
{
  "linters": {
    "javascript": {
      "enabled": true,
      "config": {
        "eslintConfig": ".eslintrc.js",
        "disabledRules": ["no-console"],
        "timeout": "30s"
      }
    }
  }
}
```

### JSON Validation

```json
{
  "linters": {
    "json": {
      "enabled": true,
      "config": {
        "validateSchema": true,
        "schemaPath": "schema.json",
        "allowComments": false
      }
    }
  }
}
```

## Pattern-Based Rule Overrides

Use pattern-based rules to apply different configurations to specific files:

### Disable Linting for Generated Files

```json
{
  "rules": [
    {
      "pattern": "*.generated.go",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["all"]
      }
    },
    {
      "pattern": "vendor/**",
      "linter": "*",
      "rules": {
        "enabled": false
      }
    }
  ]
}
```

### Different Rules for Tests

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
      "pattern": "testdata/**",
      "linter": "*",
      "rules": {
        "enabled": false
      }
    }
  ]
}
```

### Documentation Standards

```json
{
  "rules": [
    {
      "pattern": "docs/*.md",
      "linter": "markdown",
      "rules": {
        "requireFrontmatter": true,
        "maxLineLength": 80,
        "frontmatterSchema": "docs/schema.json"
      }
    },
    {
      "pattern": "README.md",
      "linter": "markdown",
      "rules": {
        "maxLineLength": 100,
        "requireFrontmatter": false
      }
    }
  ]
}
```

## Advanced Configuration

### Team Configuration Example

```json
{
  "parallel": {
    "maxWorkers": 8
  },
  "timeout": "10m",
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": ".golangci.yml",
        "testTimeout": "15m"
      }
    },
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120,
        "requireFrontmatter": true
      }
    },
    "python": {
      "enabled": true,
      "config": {
        "tool": "uv"
      }
    }
  },
  "rules": [
    {
      "pattern": "internal/**",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["gochecknoglobals"]
      }
    },
    {
      "pattern": "*.pb.go",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["all"]
      }
    },
    {
      "pattern": "CHANGELOG.md",
      "linter": "markdown",
      "rules": {
        "maxLineLength": 200,
        "requireFrontmatter": false
      }
    }
  ]
}
```

## Configuration Tips

### Best Practices

1. **Start Simple**: Begin with basic configuration and add complexity as needed
2. **Use Local Overrides**: Use `.claude/gismo.local.json` for personal preferences
3. **Pattern Specificity**: Use specific patterns to handle special cases
4. **Disable vs Configure**: Prefer disabling specific checks over entire linters

### Common Patterns

```json
{
  "rules": [
    {
      "pattern": "cmd/**",
      "linter": "golang",
      "rules": {
        "disabledChecks": ["gochecknoglobals"]
      }
    },
    {
      "pattern": "examples/**",
      "linter": "*",
      "rules": {
        "enabled": false
      }
    }
  ]
}
```

### Environment-Specific Configuration

Use different configurations for different environments:

```bash
# Development
gismo --config .claude/gismo.dev.json

# CI/CD
gismo --config .claude/gismo.ci.json

# Production
gismo --config .claude/gismo.prod.json
```

## Validation

Test your configuration:

```bash
# Validate configuration syntax
gismo --config .claude/gismo.json --validate

# Dry run to see what would be checked
gismo --config .claude/gismo.json --dry-run

# Verbose output for debugging
gismo --config .claude/gismo.json --verbose
```

## Next Steps

- [Linter Documentation](/docs/linters/) - Language-specific linter guides
- [CLI Reference](/docs/cli/) - Complete command documentation
- [Library API](/docs/library/) - Integration examples