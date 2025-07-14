package ccfeedback

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Executor handles the execution of hooks and processing of responses
type Executor struct {
	handler  *Handler
	timeout  time.Duration
	registry *Registry
}

// NewExecutor creates a new hook executor
func NewExecutor(ruleEngine RuleEngine) *Executor {
	return &Executor{
		handler:  NewHandler(ruleEngine),
		timeout:  60 * time.Second, // Default 60 second timeout
		registry: NewRegistry(),
	}
}

// Execute runs the hook processing with the configured handler
func (e *Executor) Execute(ctx context.Context) error {
	_, err := e.ExecuteWithExitCode(ctx)
	return err
}

// ExecuteWithExitCode runs the hook processing and returns the appropriate exit code
func (e *Executor) ExecuteWithExitCode(ctx context.Context) (int, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Process the input and get the response
	response, err := e.handler.ProcessInputWithResponse(ctx)
	if err != nil {
		return 1, err
	}

	// Check if this is a PostToolUse hook by examining the handler's last processed message
	if e.handler.IsPostToolUseHook() {
		// For PostToolUse hooks, return exit code 1 if there's feedback
		if response != nil && hasResponseFeedback(response) {
			return 1, nil
		}
		// Otherwise return success
		return int(ExitSuccess), nil
	}

	// Determine exit code based on response
	if response != nil && response.Decision == "block" {
		return int(ExitBlocking), nil
	}

	return int(ExitSuccess), nil
}

// ExecuteWithReader processes hook messages from a custom reader
func (e *Executor) ExecuteWithReader(ctx context.Context, reader io.Reader) error {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Parse the message
	msg, err := e.handler.parser.ParseHookMessage(data)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Process the message
	response, err := e.handler.ProcessMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}

	// Write response if needed
	if response != nil {
		responseData, err := e.handler.parser.MarshalHookResponse(response)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		_, err = os.Stdout.Write(responseData)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	return nil
}

// SetTimeout updates the execution timeout
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetRuleEngine updates the rule engine
func (e *Executor) SetRuleEngine(engine RuleEngine) {
	e.handler.SetRuleEngine(engine)
}

// GetRegistry returns the hook registry
func (e *Executor) GetRegistry() *Registry {
	return e.registry
}

// HookRunner provides utilities for running external hooks
type HookRunner struct {
	timeout time.Duration
}

// NewHookRunner creates a new hook runner
func NewHookRunner(timeout time.Duration) *HookRunner {
	return &HookRunner{
		timeout: timeout,
	}
}

// RunHook executes an external hook program
func (r *HookRunner) RunHook(ctx context.Context, hookPath string, input []byte) ([]byte, error) {
	// Create command
	cmd := exec.CommandContext(ctx, hookPath)

	// Set up stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start hook: %w", err)
	}

	// Write input to stdin
	go func() {
		defer stdin.Close()
		_, _ = stdin.Write(input)
	}()

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Return stderr for exit code 2
				if exitErr.ExitCode() == int(ExitBlocking) {
					return exitErr.Stderr, nil
				}
			}
			return nil, fmt.Errorf("hook execution failed: %w", err)
		}
	}

	// Get stdout
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	return output, nil
}

// ChainExecutor allows chaining multiple rule engines in sequence
type ChainExecutor struct {
	executors []*Executor
}

// NewChainExecutor creates a new chain executor
func NewChainExecutor(executors ...*Executor) *ChainExecutor {
	return &ChainExecutor{
		executors: executors,
	}
}

// Execute runs all executors in sequence
func (c *ChainExecutor) Execute(ctx context.Context) error {
	for i, executor := range c.executors {
		if err := executor.Execute(ctx); err != nil {
			return fmt.Errorf("executor %d failed: %w", i, err)
		}
	}
	return nil
}

// hasResponseFeedback determines if a HookResponse contains meaningful feedback
func hasResponseFeedback(response *HookResponse) bool {
	if response == nil {
		return false
	}

	// Check if any feedback fields have content
	return response.Message != "" ||
		response.Reason != "" ||
		response.StopReason != "" ||
		response.Decision != "" ||
		response.Continue != nil ||
		response.SuppressOutput != nil
}
