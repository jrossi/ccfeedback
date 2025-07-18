---
title: "Gismo"
linkTitle: "Home"
type: "docs"
---

# Gismo

High-performance Go library and CLI tool for handling Claude Code hooks with built-in linting capabilities.

## Overview

Gismo serves as a hook processor that validates and analyzes code before and after tool execution in
Claude Code workflows. It provides comprehensive linting support for multiple languages and offers both library
and CLI interfaces.

## Key Features

- **Multi-language linting**: Go, Python, JavaScript, Markdown, and JSON
- **High performance**: Sub-microsecond message parsing with optimized execution
- **Flexible configuration**: Hierarchical configuration with pattern-based overrides
- **Hook integration**: Full Claude Code hook lifecycle support
- **Extensible architecture**: Pluggable rule engines and composite processing

## Quick Start

Get started with Gismo in minutes:

1. **[Install Gismo](installation/)** - Multiple installation options
2. **[Quick Start Guide](quickstart/)** - Basic usage examples
3. **[Configuration](configuration/)** - Set up your linting rules

## Use Cases

- **Claude Code Hook Processing**: Validate code changes before and after tool execution
- **CI/CD Integration**: Automated code quality checks in build pipelines
- **Development Workflows**: Real-time linting during development
- **Team Standards**: Enforce consistent code quality across teams

## Performance

Gismo is built for speed:
- Message parsing: ~700ns per message
- Rule evaluation: <1ns for simple rules
- Go linting: ~100ms enhanced / ~4μs fallback
- Full pipeline: ~22ns handler processing

[Get Started](installation/){.btn .btn-primary .btn-lg}
[View on GitHub](https://github.com/jrossi/gismo){.btn .btn-secondary .btn-lg}
<!-- Re-trigger deployment -->