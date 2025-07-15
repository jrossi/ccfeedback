---
title: "JavaScript Linting"
linkTitle: "JavaScript"
weight: 40
description: >
  JavaScript and TypeScript linting with ESLint integration
---

# JavaScript Linting

CCFeedback provides JavaScript and TypeScript linting through ESLint integration with support for modern
JavaScript features and TypeScript syntax.

## Features

- **ESLint Integration**: Full ESLint compatibility with existing configurations
- **TypeScript Support**: Native TypeScript syntax and type checking
- **Modern JavaScript**: ES6+, JSX, and modern JavaScript features
- **Configurable Rules**: Extensive rule configuration and plugin support
- **Fast Execution**: Optimized for individual file processing

## Basic Configuration

```json
{
  "linters": {
    "javascript": {
      "enabled": true,
      "config": {
        "eslintConfig": ".eslintrc.js",
        "timeout": "30s"
      }
    }
  }
}
```

## Advanced Configuration

### Custom ESLint Settings

```json
{
  "linters": {
    "javascript": {
      "enabled": true,
      "config": {
        "eslintConfig": ".eslintrc.json",
        "disabledRules": ["no-console", "prefer-const"],
        "enabledRules": ["eqeqeq", "curly"],
        "timeout": "60s",
        "extensions": [".js", ".jsx", ".ts", ".tsx"]
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
      "pattern": "*.test.js",
      "linter": "javascript",
      "rules": {
        "disabledRules": ["no-magic-numbers"],
        "timeout": "45s"
      }
    },
    {
      "pattern": "src/**/*.ts",
      "linter": "javascript",
      "rules": {
        "eslintConfig": ".eslintrc.typescript.json"
      }
    },
    {
      "pattern": "scripts/**",
      "linter": "javascript",
      "rules": {
        "disabledRules": ["no-console", "no-process-exit"]
      }
    }
  ]
}
```

## ESLint Configuration

### Basic .eslintrc.json

```json
{
  "env": {
    "browser": true,
    "es2021": true,
    "node": true
  },
  "extends": [
    "eslint:recommended"
  ],
  "parserOptions": {
    "ecmaVersion": 12,
    "sourceType": "module"
  },
  "rules": {
    "indent": ["error", 2],
    "linebreak-style": ["error", "unix"],
    "quotes": ["error", "single"],
    "semi": ["error", "always"]
  }
}
```

### TypeScript Configuration

```json
{
  "env": {
    "browser": true,
    "es2021": true,
    "node": true
  },
  "extends": [
    "eslint:recommended",
    "@typescript-eslint/recommended"
  ],
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 12,
    "sourceType": "module",
    "project": "./tsconfig.json"
  },
  "plugins": [
    "@typescript-eslint"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/explicit-function-return-type": "warn",
    "@typescript-eslint/no-explicit-any": "error"
  }
}
```

### React Configuration

```json
{
  "env": {
    "browser": true,
    "es2021": true
  },
  "extends": [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended"
  ],
  "parserOptions": {
    "ecmaFeatures": {
      "jsx": true
    },
    "ecmaVersion": 12,
    "sourceType": "module"
  },
  "plugins": [
    "react",
    "react-hooks"
  ],
  "rules": {
    "react/prop-types": "off",
    "react-hooks/rules-of-hooks": "error",
    "react-hooks/exhaustive-deps": "warn"
  },
  "settings": {
    "react": {
      "version": "detect"
    }
  }
}
```

## Common Rule Categories

### Code Quality

| Rule | Description |
|------|-------------|
| `no-unused-vars` | Detect unused variables |
| `no-undef` | Detect undefined variables |
| `eqeqeq` | Require strict equality |
| `curly` | Require curly braces |

### Style Consistency

| Rule | Description |
|------|-------------|
| `indent` | Consistent indentation |
| `quotes` | Quote style enforcement |
| `semi` | Semicolon requirements |
| `comma-dangle` | Trailing comma rules |

### Best Practices

| Rule | Description |
|------|-------------|
| `no-console` | Prevent console.log |
| `no-debugger` | Prevent debugger statements |
| `prefer-const` | Prefer const over let |
| `no-var` | Prevent var usage |

## Integration Examples

### Package.json Scripts

```json
{
  "scripts": {
    "lint": "ccfeedback --config .claude/ccfeedback.json",
    "lint:js": "eslint src/**/*.{js,jsx,ts,tsx}",
    "lint:fix": "eslint --fix src/**/*.{js,jsx,ts,tsx}"
  }
}
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check only changed JavaScript/TypeScript files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(js|jsx|ts|tsx)$')
if [ -n "$changed_files" ]; then
    for file in $changed_files; do
        ccfeedback --file "$file"
    done
fi
```

### CI/CD Pipeline

```yaml
name: JavaScript Quality
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Install ccfeedback
      run: go install github.com/jrossi-claude/ccfeedback/cmd/ccfeedback@latest

    - name: Lint JavaScript
      run: ccfeedback --config .claude/ccfeedback.json
```

## Project-Specific Configurations

### Node.js Backend

```json
{
  "linters": {
    "javascript": {
      "config": {
        "eslintConfig": ".eslintrc.backend.json",
        "disabledRules": ["no-console"],
        "enabledRules": ["no-process-exit"]
      }
    }
  }
}
```

### Frontend Application

```json
{
  "linters": {
    "javascript": {
      "config": {
        "eslintConfig": ".eslintrc.frontend.json",
        "disabledRules": ["no-console"],
        "enabledRules": ["react/prop-types"]
      }
    }
  }
}
```

### Library Development

```json
{
  "linters": {
    "javascript": {
      "config": {
        "eslintConfig": ".eslintrc.lib.json",
        "enabledRules": ["jsdoc/require-jsdoc"],
        "timeout": "45s"
      }
    }
  }
}
```

## Performance Optimization

### Fast Mode

```json
{
  "config": {
    "timeout": "15s",
    "disabledRules": ["import/no-cycle"]
  }
}
```

### Comprehensive Analysis

```json
{
  "config": {
    "timeout": "120s",
    "enabledRules": ["complexity", "max-depth"]
  }
}
```

## Troubleshooting

### Common Issues

#### ESLint Not Found

Install ESLint and required plugins:

```bash
npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

#### Configuration Conflicts

Check for conflicting rules in different config files:

```bash
eslint --print-config src/file.js
```

#### Performance Issues

Reduce rule complexity or increase timeout:

```json
{
  "config": {
    "timeout": "60s",
    "disabledRules": ["import/no-cycle"]
  }
}
```

### Debug Mode

Enable verbose ESLint output:

```bash
ccfeedback --config .claude/ccfeedback.json --verbose
```

## Best Practices

### Configuration Strategy

1. **Start with recommended**: Use eslint:recommended as base
2. **Add incrementally**: Introduce new rules gradually
3. **Project-specific configs**: Different rules for different project types
4. **Team consensus**: Agree on style preferences

### Code Quality

1. **Fix errors first**: Address syntax and logic errors
2. **Style consistency**: Enforce consistent formatting
3. **Best practices**: Enable modern JavaScript patterns
4. **Type safety**: Use TypeScript rules for type checking

### Team Standards

```json
{
  "linters": {
    "javascript": {
      "enabled": true,
      "config": {
        "eslintConfig": ".eslintrc.json",
        "timeout": "30s"
      }
    }
  },
  "rules": [
    {
      "pattern": "src/**/*.{js,jsx}",
      "linter": "javascript",
      "rules": {
        "eslintConfig": ".eslintrc.react.json"
      }
    },
    {
      "pattern": "src/**/*.{ts,tsx}",
      "linter": "javascript",
      "rules": {
        "eslintConfig": ".eslintrc.typescript.json"
      }
    },
    {
      "pattern": "tests/**",
      "linter": "javascript",
      "rules": {
        "disabledRules": ["no-magic-numbers"]
      }
    }
  ]
}
```

## Related Documentation

- [Configuration Guide](/docs/configuration/) - General configuration options
- [ESLint Documentation](https://eslint.org/docs/) - Comprehensive ESLint guide
- [TypeScript ESLint](https://typescript-eslint.io/) - TypeScript-specific rules