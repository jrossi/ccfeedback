---
title: "Markdown Linting"
linkTitle: "Markdown"
weight: 30
description: >
  Built-in markdown linting with frontmatter validation
---

# Markdown Linting

CCFeedback provides comprehensive markdown linting using the goldmark parser with support for frontmatter
validation and documentation standards enforcement.

## Features

- **Built-in Parser**: Uses goldmark for fast, accurate markdown processing
- **Frontmatter Support**: YAML frontmatter parsing and JSON schema validation
- **Style Enforcement**: Line length, heading structure, list formatting
- **Performance**: High-speed processing with minimal overhead
- **Configurable Rules**: Flexible rule configuration for different document types

## Basic Configuration

```json
{
  "linters": {
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120,
        "requireFrontmatter": false,
        "maxBlankLines": 2
      }
    }
  }
}
```

## Advanced Configuration

### Documentation Standards

```json
{
  "linters": {
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 80,
        "requireFrontmatter": true,
        "frontmatterSchema": "docs/schema.json",
        "maxBlankLines": 1,
        "listIndentSize": 2,
        "disabledRules": []
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
      "pattern": "docs/**/*.md",
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
    },
    {
      "pattern": "CHANGELOG.md",
      "linter": "markdown",
      "rules": {
        "maxLineLength": 200,
        "disabledRules": ["heading-structure"]
      }
    }
  ]
}
```

## Available Rules

### Line Formatting

| Rule | Description | Default |
|------|-------------|---------|
| `line-length` | Maximum line length | 120 |
| `trailing-whitespace` | Trailing whitespace detection | enabled |
| `blank-lines` | Maximum consecutive blank lines | 2 |

### Document Structure

| Rule | Description | Default |
|------|-------------|---------|
| `heading-structure` | Proper heading hierarchy | enabled |
| `frontmatter-required` | Require YAML frontmatter | false |
| `frontmatter-schema` | JSON schema validation | none |

### List Formatting

| Rule | Description | Default |
|------|-------------|---------|
| `list-indent` | Consistent list indentation | 2 spaces |
| `list-marker` | Consistent list markers | enabled |

### Link Validation

| Rule | Description | Default |
|------|-------------|---------|
| `link-format` | Proper link formatting | enabled |
| `reference-links` | Reference link validation | enabled |

## Frontmatter Schema Validation

### Schema Definition

Create a JSON schema for frontmatter validation:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1
    },
    "description": {
      "type": "string",
      "minLength": 1
    },
    "weight": {
      "type": "integer",
      "minimum": 1
    },
    "draft": {
      "type": "boolean"
    }
  },
  "required": ["title", "description"],
  "additionalProperties": false
}
```

### Usage Example

```json
{
  "linters": {
    "markdown": {
      "config": {
        "requireFrontmatter": true,
        "frontmatterSchema": "docs/frontmatter-schema.json"
      }
    }
  }
}
```

Valid frontmatter:

```yaml
---
title: "Getting Started"
description: "Quick start guide for new users"
weight: 10
draft: false
---
```

## Configuration Examples

### Blog Posts

```json
{
  "rules": [
    {
      "pattern": "blog/*.md",
      "linter": "markdown",
      "rules": {
        "requireFrontmatter": true,
        "frontmatterSchema": "schemas/blog-post.json",
        "maxLineLength": 80,
        "maxBlankLines": 1
      }
    }
  ]
}
```

### Technical Documentation

```json
{
  "rules": [
    {
      "pattern": "docs/**/*.md",
      "linter": "markdown",
      "rules": {
        "requireFrontmatter": true,
        "frontmatterSchema": "schemas/documentation.json",
        "maxLineLength": 100,
        "disabledRules": []
      }
    }
  ]
}
```

### README Files

```json
{
  "rules": [
    {
      "pattern": "**/README.md",
      "linter": "markdown",
      "rules": {
        "requireFrontmatter": false,
        "maxLineLength": 120,
        "disabledRules": ["heading-structure"]
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

# Check only changed markdown files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.md$')
if [ -n "$changed_files" ]; then
    for file in $changed_files; do
        ccfeedback --file "$file"
    done
fi
```

### Makefile Integration

```makefile
.PHONY: lint-docs check-docs

lint-docs:
	ccfeedback --config .claude/ccfeedback.json --filter="*.md"

check-docs:
	find docs -name "*.md" -exec ccfeedback --file {} \;

fix-docs:
	find docs -name "*.md" -exec prettier --write {} \;
```

### CI/CD Pipeline

```yaml
name: Documentation Quality
on:
  push:
    paths: ['docs/**', '*.md']

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Install ccfeedback
      run: go install github.com/jrossi-claude/ccfeedback/cmd/ccfeedback@latest

    - name: Lint documentation
      run: ccfeedback --config .claude/ccfeedback.json
```

## Common Rule Configurations

### Strict Documentation

```json
{
  "config": {
    "maxLineLength": 80,
    "requireFrontmatter": true,
    "maxBlankLines": 1,
    "disabledRules": []
  }
}
```

### Relaxed Standards

```json
{
  "config": {
    "maxLineLength": 120,
    "requireFrontmatter": false,
    "maxBlankLines": 3,
    "disabledRules": ["heading-structure"]
  }
}
```

### Blog-Specific

```json
{
  "config": {
    "maxLineLength": 100,
    "requireFrontmatter": true,
    "frontmatterSchema": "blog-schema.json",
    "disabledRules": ["line-length"]
  }
}
```

## Troubleshooting

### Common Issues

#### Frontmatter Validation Errors

Check schema syntax and required fields:

```bash
# Validate schema file
jsonschema validate --instance frontmatter.yaml schema.json
```

#### Line Length Issues

Configure appropriate line length for content type:

```json
{
  "config": {
    "maxLineLength": 120,
    "disabledRules": ["line-length"]
  }
}
```

#### Performance with Large Files

Use file size limits or exclude large files:

```json
{
  "rules": [
    {
      "pattern": "large-docs/**",
      "linter": "markdown",
      "rules": {
        "enabled": false
      }
    }
  ]
}
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
ccfeedback --config .claude/ccfeedback.json --verbose
```

## Best Practices

### Documentation Strategy

1. **Consistent formatting**: Use standard line lengths and spacing
2. **Frontmatter schemas**: Define clear schemas for different document types
3. **Pattern-based rules**: Different standards for different document purposes
4. **Link validation**: Ensure all links are properly formatted

### Content Quality

1. **Clear structure**: Use proper heading hierarchy
2. **Readable length**: Keep lines at reasonable length for readability
3. **Consistent style**: Use consistent list formatting and spacing
4. **Metadata validation**: Validate frontmatter for structured content

### Team Standards

```json
{
  "linters": {
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 100,
        "requireFrontmatter": false,
        "maxBlankLines": 2
      }
    }
  },
  "rules": [
    {
      "pattern": "docs/**/*.md",
      "linter": "markdown",
      "rules": {
        "requireFrontmatter": true,
        "maxLineLength": 80
      }
    },
    {
      "pattern": "*.md",
      "linter": "markdown",
      "rules": {
        "maxLineLength": 120
      }
    }
  ]
}
```

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [Frontmatter Schema Examples](https://json-schema.org/examples.html) - JSON Schema documentation
- [Goldmark Documentation](https://github.com/yuin/goldmark) - Markdown parser details