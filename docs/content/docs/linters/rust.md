---
title: "Rust"
linkTitle: "Rust"
weight: 60
description: >
  Rust linting with cargo clippy and rustfmt
---

# Rust Linting

CCFeedback provides comprehensive Rust linting support using cargo's built-in tools: `cargo clippy` for linting
and `cargo fmt` for formatting checks.

## Prerequisites

- Rust toolchain installed (cargo, rustc, clippy, rustfmt)
- Project must have a `Cargo.toml` file

## Features

- **Clippy Integration**: Runs cargo clippy with configurable lints
- **Format Checking**: Validates code formatting with rustfmt
- **Test Execution**: Automatically runs tests in files containing `#[test]`
- **Workspace Support**: Detects and handles Cargo workspaces
- **Fast Mode**: Uses `--no-deps` by default to lint only your code

## Configuration

### Basic Configuration

```json
{
  "linters": {
    "rust": {
      "enabled": true
    }
  }
}
```

### Advanced Configuration

```json
{
  "linters": {
    "rust": {
      "enabled": true,
      "config": {
        "noDeps": true,
        "allTargets": true,
        "allFeatures": false,
        "features": ["async", "serde"],
        "disabledLints": ["dead_code", "unused_variables"],
        "enabledLints": ["clippy::pedantic", "clippy::nursery"],
        "testTimeout": "5m",
        "clippyConfig": "path/to/clippy.toml",
        "rustfmtConfig": "path/to/rustfmt.toml"
      }
    }
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `noDeps` | boolean | `true` | Run clippy only on the given crate, without linting dependencies |
| `allTargets` | boolean | `true` | Check all targets (lib, bin, test, example, etc.) |
| `allFeatures` | boolean | `false` | Activate all available features |
| `features` | string[] | `[]` | List of features to activate |
| `disabledLints` | string[] | `[]` | Clippy lints to disable |
| `enabledLints` | string[] | `[]` | Additional clippy lints to enable |
| `testTimeout` | string | `"10m"` | Timeout for running tests |
| `clippyConfig` | string | - | Path to custom clippy.toml |
| `rustfmtConfig` | string | - | Path to custom rustfmt.toml |

## Clippy Lints

### Enabling Additional Lints

Enable stricter linting with clippy lint groups:

```json
{
  "enabledLints": [
    "clippy::pedantic",
    "clippy::nursery",
    "clippy::cargo"
  ]
}
```

### Disabling Specific Lints

Disable lints that don't apply to your project:

```json
{
  "disabledLints": [
    "dead_code",
    "unused_variables",
    "clippy::module_name_repetitions"
  ]
}
```

## File-Specific Rules

Apply different configurations to different files:

```json
{
  "rules": [
    {
      "pattern": "src/bin/*.rs",
      "linter": "rust",
      "rules": {
        "disabledLints": ["dead_code"]
      }
    },
    {
      "pattern": "**/tests/*.rs",
      "linter": "rust",
      "rules": {
        "disabledLints": ["clippy::unwrap_used"]
      }
    }
  ]
}
```

## Common Issues and Solutions

### Issue: Clippy not found

**Solution**: Install clippy with rustup:
```bash
rustup component add clippy
```

### Issue: Rustfmt not found

**Solution**: Install rustfmt with rustup:
```bash
rustup component add rustfmt
```

### Issue: Slow linting on large projects

**Solution**: Use `noDeps: true` (default) to skip dependency linting:
```json
{
  "config": {
    "noDeps": true
  }
}
```

### Issue: Tests timing out

**Solution**: Increase the test timeout:
```json
{
  "config": {
    "testTimeout": "30m"
  }
}
```

## Integration with CI/CD

### GitHub Actions

```yaml
- name: Setup Rust
  uses: actions-rs/toolchain@v1
  with:
    toolchain: stable
    components: clippy, rustfmt

- name: Run CCFeedback
  run: |
    go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
    ccfeedback --config .claude/ccfeedback.json
```

### GitLab CI

```yaml
lint:
  image: rust:latest
  before_script:
    - rustup component add clippy rustfmt
    - go install github.com/jrossi/ccfeedback/cmd/ccfeedback@latest
  script:
    - ccfeedback --config .claude/ccfeedback.json
```

## Performance Characteristics

| Check Type | Typical Speed | Notes |
|------------|---------------|-------|
| Syntax Check | ~100ms | Via cargo check |
| Clippy (no deps) | ~500ms | Fast mode, current crate only |
| Clippy (with deps) | ~5s+ | Depends on dependency count |
| Format Check | ~50ms | rustfmt --check |
| Test Execution | Variable | Depends on test count |

## Example Output

```text
> Write operation feedback:
- [ccfeedback:src/main.rs]: src/main.rs:5:9: warning: unused variable: `unused_var` (unused_variables)
  src/main.rs:12:1: warning: function is never used: `dead_fn` (dead_code)
  src/main.rs:1:1: warning: File is not properly formatted with rustfmt (rustfmt)

⚠️  Found 3 warning(s) - consider fixing
📝 NON-BLOCKING: Issues detected but you can continue
```

## Best Practices

1. **Use rustfmt.toml**: Define project-wide formatting rules
2. **Configure clippy.toml**: Set project-specific lint configurations
3. **Feature Flags**: Test with different feature combinations
4. **Workspace Mode**: Use workspace-wide configuration for consistency
5. **CI Integration**: Run CCFeedback in CI to catch issues early

## Rust-Specific Features

### Cargo Workspace Detection

CCFeedback automatically detects Cargo workspaces and applies linting appropriately:

```toml
# workspace Cargo.toml
[workspace]
members = ["crate1", "crate2"]
```

### Module Path Resolution

The linter correctly resolves Rust module paths for targeted test execution.

### Feature-Aware Linting

Configure feature-specific linting:

```json
{
  "config": {
    "features": ["tokio", "async-std"],
    "allFeatures": false
  }
}
```

This ensures your code is checked with the correct feature flags enabled.