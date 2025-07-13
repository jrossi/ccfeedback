package linters

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// GoLinter handles Go file linting, formatting, and test running
type GoLinter struct {
	// Cache module roots to avoid repeated filesystem walks
	moduleCache map[string]*ModuleInfo
	mu          sync.RWMutex
}

// ModuleInfo contains information about a Go module
type ModuleInfo struct {
	Root      string // Module root directory (where go.mod is located)
	Path      string // Module path from go.mod
	GoModPath string // Full path to go.mod file
}

// NewGoLinter creates a new Go linter
func NewGoLinter() *GoLinter {
	return &GoLinter{
		moduleCache: make(map[string]*ModuleInfo),
	}
}

// Name returns the linter name
func (l *GoLinter) Name() string {
	return "go"
}

// CanHandle returns true for Go files
func (l *GoLinter) CanHandle(filePath string) bool {
	return strings.HasSuffix(filePath, ".go")
}

// Lint performs linting on a Go file
func (l *GoLinter) Lint(ctx context.Context, filePath string, content []byte) (*LintResult, error) {
	result := &LintResult{
		Success: true,
		Issues:  []Issue{},
	}

	// Skip generated files
	if bytes.Contains(content, []byte("// Code generated")) {
		return result, nil
	}

	// Skip test data files
	if strings.Contains(filePath, "/testdata/") {
		return result, nil
	}

	// Check formatting
	formatted, err := format.Source(content)
	if err != nil {
		result.Success = false
		result.Issues = append(result.Issues, Issue{
			File:     filePath,
			Line:     1,
			Column:   1,
			Severity: "error",
			Message:  fmt.Sprintf("Go syntax error: %v", err),
			Rule:     "syntax",
		})
		return result, nil
	}

	// Check if formatting is needed
	if !bytes.Equal(content, formatted) {
		result.Formatted = formatted
		result.Issues = append(result.Issues, Issue{
			File:     filePath,
			Line:     1,
			Column:   1,
			Severity: "warning",
			Message:  "File is not properly formatted with gofmt",
			Rule:     "format",
		})
	}

	// If this is a test file, run the tests
	if strings.HasSuffix(filePath, "_test.go") {
		if output, err := l.runTests(ctx, filePath); err != nil {
			result.Success = false
			result.Issues = append(result.Issues, Issue{
				File:     filePath,
				Line:     1,
				Column:   1,
				Severity: "error",
				Message:  fmt.Sprintf("Tests failed: %v", err),
				Rule:     "test",
			})
			result.TestOutput = output
		} else {
			result.TestOutput = output
		}
	} else {
		// For non-test files, check if corresponding test file exists and run it
		testFile := strings.TrimSuffix(filePath, ".go") + "_test.go"
		if _, err := os.Stat(testFile); err == nil {
			if output, err := l.runTests(ctx, testFile); err != nil {
				result.Success = false
				result.Issues = append(result.Issues, Issue{
					File:     testFile,
					Line:     1,
					Column:   1,
					Severity: "error",
					Message:  fmt.Sprintf("Tests failed: %v", err),
					Rule:     "test",
				})
				result.TestOutput = output
			} else {
				result.TestOutput = output
			}
		}
	}

	return result, nil
}

// FindModuleRoot walks up the directory tree to find go.mod
func (l *GoLinter) FindModuleRoot(startPath string) (*ModuleInfo, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check cache first
	l.mu.RLock()
	if moduleInfo, exists := l.moduleCache[absPath]; exists {
		l.mu.RUnlock()
		return moduleInfo, nil
	}
	l.mu.RUnlock()

	// Walk up the directory tree
	currentPath := absPath
	if info, err := os.Stat(currentPath); err == nil && !info.IsDir() {
		currentPath = filepath.Dir(currentPath)
	}

	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod file
			moduleInfo := &ModuleInfo{
				Root:      currentPath,
				GoModPath: goModPath,
			}

			// Read module path from go.mod
			if data, err := os.ReadFile(goModPath); err == nil {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "module ") {
						moduleInfo.Path = strings.TrimSpace(strings.TrimPrefix(line, "module "))
						break
					}
				}
			}

			// Cache the result
			l.mu.Lock()
			l.moduleCache[absPath] = moduleInfo
			l.mu.Unlock()

			return moduleInfo, nil
		}

		// Get parent directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached root of filesystem
			return nil, fmt.Errorf("go.mod not found")
		}
		currentPath = parent
	}
}

// runTests runs tests for a specific Go file
func (l *GoLinter) runTests(ctx context.Context, testFile string) (string, error) {
	// Find module root
	moduleInfo, err := l.FindModuleRoot(testFile)
	if err != nil {
		return "", fmt.Errorf("failed to find module root: %w", err)
	}

	// Calculate relative path from module root
	relPath, err := filepath.Rel(moduleInfo.Root, filepath.Dir(testFile))
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Convert to Unix-style path for go test
	testPath := "./" + filepath.ToSlash(relPath)

	// Run go test
	cmd := exec.CommandContext(ctx, "go", "test", "-v", testPath)
	cmd.Dir = moduleInfo.Root

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("go test failed: %w", err)
	}

	return output, nil
}

// FormatFile formats a Go file using gofmt
func (l *GoLinter) FormatFile(content []byte) ([]byte, error) {
	return format.Source(content)
}
