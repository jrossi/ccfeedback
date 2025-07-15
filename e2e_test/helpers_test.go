package e2e_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildTestBinary builds the ccfeedback binary for testing
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Create temporary binary
	tmpDir, err := os.MkdirTemp("", "ccfeedback_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	binPath := filepath.Join(tmpDir, "ccfeedback_test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/ccfeedback")
	cmd.Dir = ".." // Go up one directory since we're in e2e_test

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v\nStderr: %s", err, stderr.String())
	}

	return binPath
}
