---
title: "Go Linting"
linkTitle: "Go"
weight: 10
description: >
  Enhanced Go linting with golangci-lint integration
---

# Go Linting

Gismo provides comprehensive Go linting through golangci-lint integration with intelligent fallback
capabilities.

## Features

- **30+ Fast Linters**: bodyclose, errcheck, gofmt, goimports, gosimple, govet, ineffassign, misspell,
  staticcheck, typecheck, unconvert, unused, gosec
- **Module-Aware**: Automatically detects Go module roots and runs tests from proper directory
- **Fast Mode**: Uses `--fast` flag for optimal individual file performance
- **Graceful Fallback**: Works even without golangci-lint installed
- **Custom Configuration**: Respects `.golangci.yml` configuration files

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

## Module Detection

Gismo automatically detects Go modules and adjusts behavior:

1. **Module Root Detection**: Finds `go.mod` files to determine module boundaries
2. **Test Execution**: Runs tests from the module root directory
3. **Import Path Resolution**: Resolves import paths relative to module root
4. **Dependency Management**: Respects `go.mod` and `go.sum` files

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

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [CLI Reference](/docs/cli/) - Command-line usage
- [Library API](/docs/library/) - Go integration examples