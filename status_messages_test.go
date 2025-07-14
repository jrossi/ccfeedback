package ccfeedback

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestLintingEngine_StatusMessages(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name             string
		fileContent      string
		expectedStderr   string
		expectedDecision string
	}{
		{
			name:             "clean_go_file",
			fileContent:      "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n",
			expectedStderr:   "Style clean",
			expectedDecision: "",
		},
		{
			name:             "go_file_with_formatting_issues",
			fileContent:      "package main\n\nfunc main() {\nfmt.Println(\"Hello\")\n}\n",
			expectedStderr:   "File is not properly formatted with gofmt",
			expectedDecision: "",
		},
		{
			name:             "go_file_with_syntax_error",
			fileContent:      "package main\n\nfunc main() {",
			expectedStderr:   "expected '}', found 'EOF'",
			expectedDecision: "block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test content to file
			if err := os.WriteFile(tmpFile.Name(), []byte(tt.fileContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Create linting engine
			engine := NewLintingRuleEngine()

			// Test PreToolUse (for blocking errors)
			if strings.Contains(tt.fileContent, "func main() {") && !strings.Contains(tt.fileContent, "}") {
				msg := &PreToolUseMessage{
					BaseHookMessage: BaseHookMessage{
						HookEventName: PreToolUseEvent,
					},
					ToolName: "Write",
					ToolInput: testConvertToRawMessage(map[string]interface{}{
						"file_path": tmpFile.Name(),
						"content":   tt.fileContent,
					}),
				}

				// Capture stderr
				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				response, _ := engine.EvaluatePreToolUse(context.Background(), msg)

				w.Close()
				os.Stderr = oldStderr

				// Read stderr
				buf := make([]byte, 4096)
				n, _ := r.Read(buf)
				stderr := string(buf[:n])

				// Check decision
				if response != nil && response.Decision != tt.expectedDecision {
					t.Errorf("Expected decision %q, got %q", tt.expectedDecision, response.Decision)
				}

				// Check stderr
				if !strings.Contains(stderr, tt.expectedStderr) {
					t.Errorf("Expected stderr to contain %q, got %q", tt.expectedStderr, stderr)
				}
			} else {
				// Test PostToolUse (for non-blocking feedback)
				msg := &PostToolUseMessage{
					BaseHookMessage: BaseHookMessage{
						HookEventName: PostToolUseEvent,
					},
					ToolName: "Write",
					ToolInput: testConvertToRawMessage(map[string]interface{}{
						"file_path": tmpFile.Name(),
					}),
				}

				// Capture stderr
				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				response, _ := engine.EvaluatePostToolUse(context.Background(), msg)

				w.Close()
				os.Stderr = oldStderr

				// Read stderr
				buf := make([]byte, 4096)
				n, _ := r.Read(buf)
				stderr := string(buf[:n])

				// Check stderr
				if !strings.Contains(stderr, tt.expectedStderr) {
					t.Errorf("Expected stderr to contain %q, got %q", tt.expectedStderr, stderr)
				}

				// For PostToolUse, check if response exists when issues are found
				if strings.Contains(tt.fileContent, "fmt.Println") && !strings.Contains(tt.fileContent, "\tfmt") {
					if response == nil || response.Message == "" {
						t.Error("Expected response with message for formatting issues")
					}
				}
			}
		})
	}
}

func TestExecutor_AllExitCodes(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name         string
		fileContent  string
		messageType  HookEventName
		expectedCode int
		toolName     string
	}{
		{
			name:         "exit_0_clean_file",
			fileContent:  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n",
			messageType:  PostToolUseEvent,
			expectedCode: 0,
			toolName:     "Write",
		},
		{
			name:         "exit_1_formatting_issues",
			fileContent:  "package main\n\nfunc main() {\nfmt.Println(\"Hello\")\n}\n",
			messageType:  PostToolUseEvent,
			expectedCode: 1,
			toolName:     "Write",
		},
		{
			name:         "exit_2_syntax_error",
			fileContent:  "package main\n\nfunc main() {",
			messageType:  PreToolUseEvent,
			expectedCode: 2,
			toolName:     "Write",
		},
		{
			name:         "exit_0_non_file_operation",
			fileContent:  "",
			messageType:  PostToolUseEvent,
			expectedCode: 0,
			toolName:     "Bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test content if needed
			if tt.fileContent != "" {
				if err := os.WriteFile(tmpFile.Name(), []byte(tt.fileContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Create engine and executor
			engine := NewLintingRuleEngine()
			executor := NewExecutor(engine)

			// Prepare message based on type
			var msg HookMessage
			if tt.messageType == PreToolUseEvent {
				msg = &PreToolUseMessage{
					BaseHookMessage: BaseHookMessage{
						HookEventName: PreToolUseEvent,
					},
					ToolName: tt.toolName,
					ToolInput: testConvertToRawMessage(map[string]interface{}{
						"file_path": tmpFile.Name(),
						"content":   tt.fileContent,
					}),
				}
			} else {
				input := map[string]interface{}{
					"file_path": tmpFile.Name(),
				}
				if tt.toolName == "Bash" {
					input = map[string]interface{}{
						"command": "echo test",
					}
				}
				msg = &PostToolUseMessage{
					BaseHookMessage: BaseHookMessage{
						HookEventName: PostToolUseEvent,
					},
					ToolName:  tt.toolName,
					ToolInput: testConvertToRawMessage(input),
				}
			}

			// Process message
			response, err := executor.handler.ProcessMessage(context.Background(), msg)
			if err != nil {
				t.Fatal(err)
			}

			// Determine expected exit code
			var expectedCode int
			if tt.messageType == PostToolUseEvent && executor.handler.IsPostToolUseHook() {
				if response != nil && hasResponseFeedback(response) {
					expectedCode = 1
				} else {
					expectedCode = 0
				}
			} else if response != nil && response.Decision == "block" {
				expectedCode = 2
			} else {
				expectedCode = 0
			}

			if expectedCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedCode, expectedCode)
			}
		})
	}
}
