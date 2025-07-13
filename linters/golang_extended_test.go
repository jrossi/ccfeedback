package linters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGoLinter_Name(t *testing.T) {
	linter := NewGoLinter()
	if got := linter.Name(); got != "go" {
		t.Errorf("Name() = %q, want %q", got, "go")
	}
}

func TestGoLinter_findGolangciLint_NotFound(t *testing.T) {
	// Test when golangci-lint is not found
	linter := &GoLinter{}

	// Temporarily modify PATH to ensure golangci-lint is not found
	oldPath := os.Getenv("PATH")
	oldHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PATH", oldPath)
		os.Setenv("HOME", oldHome)
	}()

	os.Setenv("PATH", "/nonexistent")
	os.Setenv("HOME", "/nonexistent")

	path := linter.findGolangciLint()
	if path != "" {
		t.Errorf("Expected empty path when golangci-lint not found, got %q", path)
	}

	// Call again to test once.Do behavior
	path2 := linter.findGolangciLint()
	if path2 != "" {
		t.Errorf("Expected empty path on second call, got %q", path2)
	}
}

func TestGoLinter_runTests(t *testing.T) {
	linter := NewGoLinter()

	// Create a temporary directory with a simple test file
	tmpDir, err := os.MkdirTemp("", "go_test_runner_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "example_test.go")
	testContent := `package example

import "testing"

func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Also need a go.mod file
	goModContent := `module example

go 1.19
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Run tests
	output, err := linter.runTests(context.Background(), tmpDir)
	if err != nil {
		t.Logf("runTests() error (may be expected): %v", err)
	}

	// Check output
	if output == "" {
		t.Log("No test output (tests may have passed)")
	}
}

func TestGoLinter_runTests_Failure(t *testing.T) {
	linter := NewGoLinter()

	// Create a temporary directory with a failing test
	tmpDir, err := os.MkdirTemp("", "go_test_fail_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a failing test file
	testFile := filepath.Join(tmpDir, "fail_test.go")
	testContent := `package example

import "testing"

func TestFailing(t *testing.T) {
	t.Error("This test always fails")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create go.mod
	goModContent := `module example

go 1.19
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Run tests
	output, err := linter.runTests(context.Background(), tmpDir)
	if err == nil {
		t.Error("Expected error for failing tests")
	}

	// Should have output about failure
	if output == "" {
		t.Error("Expected test output for failing tests")
	}
}

func TestGoLinter_runTests_NoTests(t *testing.T) {
	linter := NewGoLinter()

	// Create a temporary directory with no test files
	tmpDir, err := os.MkdirTemp("", "go_no_tests_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a regular Go file (not a test)
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := `package main

func main() {
	println("Hello")
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to write go file: %v", err)
	}

	// Create go.mod
	goModContent := `module example

go 1.19
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Run tests
	output, err := linter.runTests(context.Background(), tmpDir)
	// No tests is not an error
	if err != nil {
		t.Logf("runTests() returned error (expected for no tests): %v", err)
	}

	// Should handle no tests gracefully
	if output != "" {
		t.Logf("Got output: %s", output)
	}
}

func TestGoLinter_EnhancedLinting_ConfigVariations(t *testing.T) {
	linter := NewGoLinter()

	tests := []struct {
		name          string
		configFile    string
		configContent string
	}{
		{
			name:       "golangci.yml",
			configFile: ".golangci.yml",
			configContent: `linters:
  enable:
    - gofmt
    - govet
`,
		},
		{
			name:       "golangci.yaml",
			configFile: ".golangci.yaml",
			configContent: `linters:
  enable:
    - gofmt
`,
		},
		{
			name:       "golangci.toml",
			configFile: ".golangci.toml",
			configContent: `[linters]
enable = ["gofmt"]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if golangci-lint not available
			if linter.findGolangciLint() == "" {
				t.Skip("golangci-lint not available")
			}

			tmpDir, err := os.MkdirTemp("", "golangci_config_")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Write config file
			configPath := filepath.Join(tmpDir, tt.configFile)
			if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			// Write a simple Go file
			goFile := filepath.Join(tmpDir, "main.go")
			goContent := `package main

func main() {
    println("Hello")
}
`
			if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
				t.Fatalf("Failed to write go file: %v", err)
			}

			// Should detect config file
			result, err := linter.Lint(context.Background(), goFile, []byte(goContent))
			if err != nil {
				t.Fatalf("Lint() error = %v", err)
			}

			// Check that enhanced linting was attempted
			if result == nil {
				t.Error("Expected result, got nil")
			}
		})
	}
}
