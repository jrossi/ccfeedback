---
title: "JSON Validation"
linkTitle: "JSON"
weight: 50
description: >
  JSON syntax validation and schema checking
---

# JSON Validation

CCFeedback provides comprehensive JSON validation using high-performance go-json parsing with optional
JSON schema validation for structured data validation.

## Features

- **High-Performance Parsing**: Uses go-json for 2-3x faster parsing than standard library
- **Syntax Validation**: Comprehensive JSON syntax checking
- **Schema Validation**: JSON Schema v7 support for structure validation
- **JSONL Support**: JSON Lines format validation
- **Fast Execution**: Optimized for large JSON files

## Basic Configuration

```json
{
  "linters": {
    "json": {
      "enabled": true,
      "config": {
        "validateSchema": false,
        "allowComments": false
      }
    }
  }
}
```

## Advanced Configuration

### Schema Validation

```json
{
  "linters": {
    "json": {
      "enabled": true,
      "config": {
        "validateSchema": true,
        "schemaPath": "schemas/config.json",
        "allowComments": false,
        "strictMode": true
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
      "pattern": "config/*.json",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/config-schema.json"
      }
    },
    {
      "pattern": "data/*.jsonl",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/data-schema.json",
        "format": "jsonl"
      }
    },
    {
      "pattern": "package.json",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/package-schema.json"
      }
    }
  ]
}
```

## JSON Schema Examples

### Configuration Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Application Configuration",
  "type": "object",
  "properties": {
    "server": {
      "type": "object",
      "properties": {
        "host": {
          "type": "string",
          "format": "hostname"
        },
        "port": {
          "type": "integer",
          "minimum": 1,
          "maximum": 65535
        }
      },
      "required": ["host", "port"]
    },
    "database": {
      "type": "object",
      "properties": {
        "url": {
          "type": "string",
          "format": "uri"
        },
        "timeout": {
          "type": "string",
          "pattern": "^[0-9]+[smh]$"
        }
      },
      "required": ["url"]
    }
  },
  "required": ["server", "database"],
  "additionalProperties": false
}
```

### API Response Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "API Response",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "error"]
    },
    "data": {
      "type": "object"
    },
    "error": {
      "type": "object",
      "properties": {
        "code": {
          "type": "integer"
        },
        "message": {
          "type": "string"
        }
      },
      "required": ["code", "message"]
    }
  },
  "required": ["status"],
  "if": {
    "properties": {
      "status": {
        "const": "error"
      }
    }
  },
  "then": {
    "required": ["error"]
  },
  "else": {
    "required": ["data"]
  }
}
```

## Validation Features

### Syntax Checking

| Check | Description |
|-------|-------------|
| `syntax` | Valid JSON syntax |
| `encoding` | UTF-8 encoding validation |
| `structure` | Proper nesting and brackets |

### Schema Validation

| Feature | Description |
|---------|-------------|
| `type` | Data type validation |
| `format` | String format validation |
| `constraints` | Min/max, length constraints |
| `patterns` | Regular expression matching |
| `conditionals` | If/then/else logic |

### Performance Features

| Feature | Description |
|---------|-------------|
| `streaming` | Large file streaming |
| `parallel` | Parallel validation |
| `caching` | Schema caching |

## Common Use Cases

### Configuration Files

```json
{
  "rules": [
    {
      "pattern": "configs/*.json",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/app-config.json",
        "strictMode": true
      }
    }
  ]
}
```

### API Documentation

```json
{
  "rules": [
    {
      "pattern": "api/openapi.json",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/openapi-3.0.json"
      }
    }
  ]
}
```

### Data Files

```json
{
  "rules": [
    {
      "pattern": "data/*.jsonl",
      "linter": "json",
      "rules": {
        "validateSchema": true,
        "schemaPath": "schemas/record.json",
        "format": "jsonl"
      }
    }
  ]
}
```

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check only changed JSON files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.json$\|\.jsonl$')
if [ -n "$changed_files" ]; then
    for file in $changed_files; do
        ccfeedback --file "$file"
    done
fi
```

### Makefile Integration

```makefile
.PHONY: validate-json lint-config

validate-json:
	ccfeedback --config .claude/ccfeedback.json --filter="*.json"

lint-config:
	find config -name "*.json" -exec ccfeedback --file {} \;

validate-schemas:
	find schemas -name "*.json" -exec jsonschema --check {} \;
```

### CI/CD Pipeline

```yaml
name: JSON Validation
on:
  push:
    paths: ['**/*.json', '**/*.jsonl']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Install ccfeedback
      run: go install github.com/jrossi-claude/ccfeedback/cmd/ccfeedback@latest

    - name: Validate JSON files
      run: ccfeedback --config .claude/ccfeedback.json

    - name: Validate schemas
      run: |
        pip install jsonschema
        find schemas -name "*.json" -exec jsonschema {} \;
```

## Performance Optimization

### Large Files

```json
{
  "config": {
    "streamingMode": true,
    "maxFileSize": "100MB"
  }
}
```

### Batch Processing

```json
{
  "config": {
    "batchSize": 100,
    "parallel": true
  }
}
```

## Error Handling

### Common Validation Errors

| Error | Description | Solution |
|-------|-------------|----------|
| `Syntax Error` | Invalid JSON syntax | Check brackets, quotes, commas |
| `Schema Violation` | Data doesn't match schema | Review schema requirements |
| `Type Mismatch` | Wrong data type | Check expected vs actual types |
| `Missing Required` | Required field missing | Add required properties |

### Error Messages

CCFeedback provides detailed error messages:

```
JSON validation failed at line 15, column 8:
- Expected string, got number for property 'name'
- Required property 'email' is missing
- Value '999' exceeds maximum of 100 for property 'age'
```

## Troubleshooting

### Common Issues

#### Schema Not Found

Ensure schema file exists and path is correct:

```bash
ls -la schemas/config.json
```

#### Performance Issues

For large files, enable streaming mode:

```json
{
  "config": {
    "streamingMode": true,
    "validateSchema": false
  }
}
```

#### Memory Usage

Limit file size or disable schema validation:

```json
{
  "config": {
    "maxFileSize": "10MB",
    "validateSchema": false
  }
}
```

### Debug Mode

Enable verbose validation output:

```bash
ccfeedback --config .claude/ccfeedback.json --verbose
```

## Best Practices

### Schema Design

1. **Clear structure**: Use descriptive property names and types
2. **Validation rules**: Add appropriate constraints and formats
3. **Error messages**: Include helpful descriptions
4. **Versioning**: Version your schemas for compatibility

### File Organization

1. **Schema directory**: Keep schemas in dedicated directory
2. **Naming conventions**: Use consistent schema naming
3. **Documentation**: Document schema purpose and usage
4. **Testing**: Test schemas with valid and invalid data

### Performance

1. **Schema caching**: Reuse schemas for multiple files
2. **Streaming mode**: Use for large files
3. **Selective validation**: Only validate when necessary
4. **Parallel processing**: Enable for multiple files

### Team Standards

```json
{
  "linters": {
    "json": {
      "enabled": true,
      "config": {
        "validateSchema": true,
        "strictMode": true
      }
    }
  },
  "rules": [
    {
      "pattern": "config/*.json",
      "linter": "json",
      "rules": {
        "schemaPath": "schemas/config.json"
      }
    },
    {
      "pattern": "data/*.json",
      "linter": "json",
      "rules": {
        "schemaPath": "schemas/data.json"
      }
    },
    {
      "pattern": "test-data/**",
      "linter": "json",
      "rules": {
        "validateSchema": false
      }
    }
  ]
}
```

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [JSON Schema](https://json-schema.org/) - Official JSON Schema documentation
- [go-json](https://github.com/goccy/go-json) - High-performance JSON library