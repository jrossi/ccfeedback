---
title: "Protocol Buffers"
linkTitle: "Protobuf"
weight: 70
description: >
  Protocol Buffer linting with buf, protolint, and protoc
---

# Protocol Buffer Linting

CCFeedback provides comprehensive Protocol Buffer linting support using industry-standard tools:
`buf` for modern linting, `protolint` as an alternative, and `protoc` for basic syntax validation.

## Prerequisites

At least one of the following tools should be installed:
- **buf** (recommended) - Modern protobuf linter with excellent error messages
- **protolint** - Alternative linter with different rule sets
- **protoc** - Protocol buffer compiler (for basic syntax validation)

## Features

- **Multi-Tool Support**: Automatically detects and uses available tools
- **buf Integration**: Full support for buf.yaml and buf.work.yaml configurations
- **Style Enforcement**: Checks naming conventions, package structure, and more
- **Syntax Validation**: Ensures proto files compile correctly
- **Workspace Support**: Handles buf workspaces and complex project structures
- **Graceful Degradation**: Falls back to simpler tools when advanced ones aren't available

## Configuration

### Basic Configuration

```json
{
  "linters": {
    "protobuf": {
      "enabled": true
    }
  }
}
```

### Advanced Configuration

```json
{
  "linters": {
    "protobuf": {
      "enabled": true,
      "config": {
        "preferredTools": ["buf", "protolint"],
        "bufConfigPath": "./buf.yaml",
        "disabledChecks": ["PACKAGE_VERSION_SUFFIX"],
        "categories": ["STANDARD", "COMMENTS"],
        "testTimeout": "2m",
        "verbose": false
      }
    }
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `preferredTools` | string[] | `["buf", "protolint", "protoc"]` | Tools to try in order of preference |
| `forceTool` | string | - | Force a specific tool to be used |
| `bufPath` | string | - | Path to buf binary |
| `protocPath` | string | - | Path to protoc binary |
| `protolintPath` | string | - | Path to protolint binary |
| `bufConfigPath` | string | - | Path to buf.yaml configuration |
| `bufWorkPath` | string | - | Path to buf.work.yaml for workspaces |
| `disabledChecks` | string[] | `[]` | buf lint checks to disable |
| `categories` | string[] | - | buf lint categories to check |
| `protolintConfig` | string | - | Path to .protolint.yaml configuration |
| `maxFileSize` | number | `10485760` | Maximum file size in bytes (10MB) |
| `testTimeout` | string | `"2m"` | Timeout for linting operations |
| `verbose` | boolean | `false` | Enable verbose output |

## Tool-Specific Features

### buf

buf is the recommended tool for protobuf linting. It provides:

- Comprehensive lint rules following industry best practices
- Breaking change detection (with additional configuration)
- Clear, actionable error messages
- Support for workspaces and modules

#### Example buf.yaml

```yaml
version: v1
lint:
  use:
    - DEFAULT
  except:
    - PACKAGE_VERSION_SUFFIX
  ignore:
    - vendor/
```

### protolint

protolint offers different lint rules and can be used alongside or instead of buf:

- Customizable rule sets
- Support for auto-fixing some issues
- Different naming convention checks

#### Example .protolint.yaml

```yaml
lint:
  rules:
    no_default: false
    add:
      - ENUM_FIELD_NAMES_PREFIX
      - ENUM_FIELD_NAMES_ZERO_VALUE_END_WITH
    remove:
      - MESSAGE_NAMES_UPPER_CAMEL_CASE
```

### protoc

protoc provides basic syntax validation:

- Ensures proto files are syntactically correct
- Validates import statements
- Checks field numbers and types

## File-Specific Rules

Apply different configurations to different proto files:

```json
{
  "rules": [
    {
      "pattern": "api/v1/*.proto",
      "linter": "protobuf",
      "rules": {
        "forceTool": "buf",
        "disabledChecks": ["FIELD_LOWER_SNAKE_CASE"]
      }
    },
    {
      "pattern": "internal/*.proto",
      "linter": "protobuf",
      "rules": {
        "preferredTools": ["protolint"]
      }
    }
  ]
}
```

## Common Issues and Solutions

### Issue: buf not found

**Solution**: Install buf:
```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64 -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

### Issue: Import not found

**Solution**: Ensure your buf.yaml or protoc include paths are correctly configured:
```yaml
# buf.yaml
version: v1
deps:
  - buf.build/googleapis/googleapis
```

### Issue: Workspace not detected

**Solution**: Create a buf.work.yaml at your repository root:
```yaml
version: v1
directories:
  - api
  - proto
```

## Integration with CI/CD

### GitHub Actions

```yaml
- name: Setup buf
  uses: bufbuild/buf-setup-action@v1

- name: Run CCFeedback
  run: |
    go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
    ccfeedback --config .claude/ccfeedback.json
```

### GitLab CI

```yaml
lint:
  image: golang:latest
  before_script:
    - curl -sSL https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64 -o /usr/local/bin/buf
    - chmod +x /usr/local/bin/buf
    - go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
  script:
    - ccfeedback --config .claude/ccfeedback.json
```

## Performance Characteristics

| Check Type | Typical Speed | Notes |
|------------|---------------|-------|
| Syntax Check (protoc) | ~50ms | Basic validation only |
| buf lint | ~200ms | Comprehensive linting |
| protolint | ~150ms | Alternative rule set |
| Large files (>1MB) | Skipped | Configure with maxFileSize |

## Example Output

```text
> Write operation feedback:
- [ccfeedback:api/v1/service.proto]: api/v1/service.proto:4:1: Package name "example" should be
  suffixed with a correctly formed version, such as "example.v1". (PACKAGE_VERSION_SUFFIX)
  api/v1/service.proto:8:3: Field name "ID" should be lower_snake_case, such as "id". (FIELD_LOWER_SNAKE_CASE)
  api/v1/service.proto:12:1: Service name "ExampleAPI" should be suffixed with "Service". (SERVICE_SUFFIX)

⚠️  Found 3 warning(s) - consider fixing
📝 NON-BLOCKING: Issues detected but you can continue
```

## Best Practices

1. **Use buf**: It provides the most comprehensive and modern linting experience
2. **Version your APIs**: Use package versioning (e.g., `example.v1`)
3. **Configure in VCS**: Keep buf.yaml in version control
4. **Start strict**: Begin with default rules and only disable when necessary
5. **Document exceptions**: When disabling rules, document why in comments
6. **Use workspaces**: For multi-module projects, use buf.work.yaml

## Protobuf-Specific Features

### Automatic Tool Selection

CCFeedback automatically selects the best available tool:

```json
{
  "config": {
    "preferredTools": ["buf", "protolint", "protoc"]
  }
}
```

### Workspace Detection

The linter automatically detects buf workspaces and adjusts paths accordingly.

### Custom Import Paths

Configure custom import paths for protoc:

```json
{
  "config": {
    "protocArgs": ["--proto_path=./vendor"]
  }
}
```

This ensures your Protocol Buffer files maintain high quality and follow best practices.