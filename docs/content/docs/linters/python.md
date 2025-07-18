---
title: "Python Linting"
linkTitle: "Python"
weight: 20
description: >
  Python linting with UV/UVX and ruff integration
---

# Python Linting

Gismo provides modern Python linting through UV/UVX toolchain integration with ruff for fast,
comprehensive code analysis.

## Features

- **Modern Toolchain**: Uses UV/UVX for fast Python environment management
- **Ruff Integration**: Lightning-fast Python linter and formatter
- **Comprehensive Checks**: Syntax, style, imports, security, and complexity analysis
- **Fast Execution**: Rust-based tooling for optimal performance
- **Configurable Rules**: Extensive rule configuration options

## Basic Configuration

```json
{
  "linters": {
    "python": {
      "enabled": true,
      "config": {
        "tool": "uv",
        "timeout": "30s"
      }
    }
  }
}
```

## Advanced Configuration

### Custom Ruff Settings

```json
{
  "linters": {
    "python": {
      "enabled": true,
      "config": {
        "tool": "uv",
        "lintCommand": "ruff check --select=E,F,W,C,N",
        "formatCommand": "ruff format",
        "configFile": "pyproject.toml",
        "timeout": "60s"
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
      "pattern": "*_test.py",
      "linter": "python",
      "rules": {
        "timeout": "60s",
        "lintCommand": "ruff check --ignore=E501"
      }
    },
    {
      "pattern": "scripts/**",
      "linter": "python",
      "rules": {
        "lintCommand": "ruff check --ignore=T201"
      }
    }
  ]
}
```

## Ruff Rule Categories

### Error Prevention (E, F)

| Code | Description |
|------|-------------|
| `E9` | Syntax errors |
| `F` | Pyflakes (undefined names, imports) |
| `E` | PEP 8 style errors |

### Code Quality (W, C, N)

| Code | Description |
|------|-------------|
| `W` | PEP 8 style warnings |
| `C90` | McCabe complexity |
| `N` | PEP 8 naming conventions |

### Import Organization (I)

| Code | Description |
|------|-------------|
| `I001` | Import sorting |
| `I002` | Missing imports |

### Security (S)

| Code | Description |
|------|-------------|
| `S` | Bandit security rules |
| `S101` | Assert usage |
| `S608` | SQL injection |

## Tool Configuration

### UV/UVX Setup

Gismo automatically detects and uses UV/UVX when available:

```bash
# Install UV
curl -LsSf https://astral.sh/uv/install.sh | sh

# Verify installation
uv --version
```

### Ruff Configuration

Create `pyproject.toml` for project-specific settings:

```toml
[tool.ruff]
line-length = 88
target-version = "py39"

[tool.ruff.lint]
select = ["E", "F", "W", "C90", "I", "N"]
ignore = ["E501", "W503"]

[tool.ruff.lint.mccabe]
max-complexity = 10

[tool.ruff.lint.isort]
force-single-line = true
```

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check only changed Python files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.py$')
if [ -n "$changed_files" ]; then
    for file in $changed_files; do
        gismo --file "$file"
    done
fi
```

### Makefile Integration

```makefile
.PHONY: lint-python format-python test-python

lint-python:
	gismo --config .claude/gismo.json --filter="*.py"

format-python:
	uv run ruff format .
	uv run ruff check --fix .

test-python:
	uv run pytest tests/
```

### CI/CD Pipeline

```yaml
name: Python Quality
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Install UV
      run: curl -LsSf https://astral.sh/uv/install.sh | sh

    - name: Install gismo
      run: go install github.com/jrossi/gismo/cmd/gismo@latest

    - name: Lint Python code
      run: gismo --config .claude/gismo.json
```

## Common Rule Configurations

### Strict Mode

```json
{
  "config": {
    "lintCommand": "ruff check --select=ALL --ignore=COM,D"
  }
}
```

### Relaxed Mode

```json
{
  "config": {
    "lintCommand": "ruff check --select=E,F --ignore=E501,E203"
  }
}
```

### Security Focused

```json
{
  "config": {
    "lintCommand": "ruff check --select=E,F,S,B --ignore=S101"
  }
}
```

## Performance Tuning

### Fast Mode

```json
{
  "config": {
    "tool": "uv",
    "timeout": "15s",
    "lintCommand": "ruff check --select=E,F"
  }
}
```

### Comprehensive Analysis

```json
{
  "config": {
    "tool": "uv",
    "timeout": "120s",
    "lintCommand": "ruff check --select=ALL"
  }
}
```

## Troubleshooting

### Common Issues

#### UV Not Found

Install UV using the official installer:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
source $HOME/.cargo/env
```

#### Slow Performance

Reduce rule scope for faster execution:

```json
{
  "config": {
    "lintCommand": "ruff check --select=E,F",
    "timeout": "30s"
  }
}
```

#### Configuration Conflicts

Ensure pyproject.toml settings don't conflict:

```bash
uv run ruff check --show-settings
```

### Debug Mode

Enable verbose output:

```bash
gismo --config .claude/gismo.json --verbose
```

## Best Practices

### Configuration Strategy

1. **Start simple**: Begin with E and F rules
2. **Add gradually**: Introduce new rule categories incrementally
3. **Use ignores sparingly**: Fix issues rather than ignoring them
4. **Project-specific rules**: Use pyproject.toml for team standards

### Code Quality

1. **Fix syntax first**: Address E9 and F errors immediately
2. **Style consistency**: Enable E and W rules
3. **Import organization**: Use I rules for clean imports
4. **Security awareness**: Enable S rules for security checks

### Team Standards

```json
{
  "linters": {
    "python": {
      "enabled": true,
      "config": {
        "tool": "uv",
        "lintCommand": "ruff check --select=E,F,W,I,N",
        "formatCommand": "ruff format",
        "timeout": "60s"
      }
    }
  },
  "rules": [
    {
      "pattern": "tests/**",
      "linter": "python",
      "rules": {
        "lintCommand": "ruff check --select=E,F --ignore=E501"
      }
    },
    {
      "pattern": "migrations/**",
      "linter": "python",
      "rules": {
        "enabled": false
      }
    }
  ]
}
```

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [CLI Reference](/docs/cli/) - Command-line usage
- [Ruff Documentation](https://docs.astral.sh/ruff/) - Comprehensive ruff guide