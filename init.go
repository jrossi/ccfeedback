package ccfeedback

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// ClaudeSettings represents the structure of Claude's settings.json
type ClaudeSettings struct {
	Permissions *PermissionsConfig     `json:"permissions,omitempty"`
	Hooks       map[string][]HookGroup `json:"hooks,omitempty"`
	// Preserve any other fields
	Extra map[string]json.RawMessage `json:"-"`
}

// PermissionsConfig represents Claude's permission settings
type PermissionsConfig struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// HookGroup represents a group of hooks with a matcher
type HookGroup struct {
	Matcher string             `json:"matcher"`
	Hooks   []ClaudeHookConfig `json:"hooks"`
}

// ClaudeHookConfig represents a single hook configuration in Claude settings
type ClaudeHookConfig struct {
	Type            string `json:"type"`
	Command         string `json:"command"`
	Timeout         int    `json:"timeout,omitempty"`
	ContinueOnError bool   `json:"continueOnError,omitempty"`
}

// InitOptions contains options for the init command
type InitOptions struct {
	GlobalOnly  bool
	ProjectOnly bool
	DryRun      bool
	Force       bool
}

// InitCommand handles the ccfeedback init command
func InitCommand(options InitOptions) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Determine which settings files to update
	var settingsPaths []string
	if !options.ProjectOnly {
		globalPath := filepath.Join(homeDir, ".claude", "settings.json")
		settingsPaths = append(settingsPaths, globalPath)
	}
	if !options.GlobalOnly {
		projectPath := filepath.Join(".claude", "settings.json")
		settingsPaths = append(settingsPaths, projectPath)
	}

	// Check if ccfeedback is in PATH
	if !isCCFeedbackAvailable() {
		fmt.Fprintf(os.Stderr, "Warning: ccfeedback command not found in PATH\n")
		fmt.Fprintf(os.Stderr, "Make sure ccfeedback is installed and available in your PATH\n\n")
	}

	// Track if any changes were made
	changesMade := false

	// Process each settings file
	for _, settingsPath := range settingsPaths {
		fmt.Printf("Processing: %s\n", settingsPath)
		wasModified, err := processSettingsFile(settingsPath, options)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", settingsPath, err)
		}
		if wasModified {
			changesMade = true
		}
		fmt.Println()
	}

	// Show next steps only if changes were actually made
	if !options.DryRun && changesMade {
		showNextSteps(options)
	}

	return nil
}

// processSettingsFile handles a single settings file
func processSettingsFile(settingsPath string, options InitOptions) (bool, error) {
	// Read existing settings
	settings, extraFields, err := readClaudeSettings(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read settings: %w", err)
	}

	// Store original for comparison
	originalJSON, _ := marshalClaudeSettings(settings, extraFields)

	// Propose changes
	modified := proposeHookChanges(settings)

	// Marshal the modified settings
	modifiedJSON, err := marshalClaudeSettings(modified, extraFields)
	if err != nil {
		return false, fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Check if anything changed
	if string(originalJSON) == string(modifiedJSON) {
		fmt.Println("✓ CCFeedback hook is already configured correctly")
		return false, nil
	}

	// Display changes
	fmt.Println("\nProposed changes:")
	displayChanges(originalJSON, modifiedJSON)

	if options.DryRun {
		fmt.Println("\n(Dry run - no changes were made)")
		return false, nil
	}

	// Ask for confirmation unless forced
	if !options.Force {
		fmt.Print("\nApply these changes? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("Canceled - no changes made")
			return false, nil
		}
	}

	// Backup existing file if it exists
	if _, err := os.Stat(settingsPath); err == nil {
		backupPath := fmt.Sprintf("%s.backup-%s", settingsPath, time.Now().Format("20060102-150405"))
		if err := copyFile(settingsPath, backupPath); err != nil {
			return false, fmt.Errorf("failed to backup existing settings: %w", err)
		}
		fmt.Printf("✓ Created backup: %s\n", backupPath)
	}

	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the new settings
	if err := os.WriteFile(settingsPath, modifiedJSON, 0600); err != nil {
		return false, fmt.Errorf("failed to write settings: %w", err)
	}

	fmt.Printf("✓ Updated: %s\n", settingsPath)
	return true, nil
}

// readClaudeSettings reads and parses Claude settings.json
func readClaudeSettings(path string) (*ClaudeSettings, map[string]json.RawMessage, error) {
	settings := &ClaudeSettings{
		Hooks: make(map[string][]HookGroup),
		Extra: make(map[string]json.RawMessage),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty settings for new file
			return settings, make(map[string]json.RawMessage), nil
		}
		return nil, nil, err
	}

	// First unmarshal to preserve unknown fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract known fields
	extraFields := make(map[string]json.RawMessage)
	for key, value := range raw {
		switch key {
		case "permissions":
			if err := json.Unmarshal(value, &settings.Permissions); err != nil {
				return nil, nil, fmt.Errorf("invalid permissions: %w", err)
			}
		case "hooks":
			if err := json.Unmarshal(value, &settings.Hooks); err != nil {
				return nil, nil, fmt.Errorf("invalid hooks: %w", err)
			}
		default:
			extraFields[key] = value
		}
	}

	return settings, extraFields, nil
}

// proposeHookChanges adds or updates ccfeedback hook configuration
func proposeHookChanges(settings *ClaudeSettings) *ClaudeSettings {
	// Make a copy
	modified := &ClaudeSettings{
		Permissions: settings.Permissions,
		Hooks:       make(map[string][]HookGroup),
		Extra:       settings.Extra,
	}

	// Copy existing hooks
	for event, groups := range settings.Hooks {
		modified.Hooks[event] = make([]HookGroup, len(groups))
		copy(modified.Hooks[event], groups)
	}

	// Check if PostToolUse exists
	postToolUseGroups, exists := modified.Hooks["PostToolUse"]
	if !exists {
		postToolUseGroups = []HookGroup{}
	}

	// Look for existing ccfeedback hook
	ccfeedbackFound := false
	for i, group := range postToolUseGroups {
		if group.Matcher == "" {
			for j, hook := range group.Hooks {
				if hook.Type == "command" && hook.Command == "ccfeedback" {
					// Update existing hook with recommended settings
					postToolUseGroups[i].Hooks[j] = ClaudeHookConfig{
						Type:            "command",
						Command:         "ccfeedback",
						Timeout:         60000,
						ContinueOnError: false,
					}
					ccfeedbackFound = true
					break
				}
			}
		}
	}

	// If not found, add it
	if !ccfeedbackFound {
		// Look for existing empty matcher group
		emptyMatcherIndex := -1
		for i, group := range postToolUseGroups {
			if group.Matcher == "" {
				emptyMatcherIndex = i
				break
			}
		}

		ccfeedbackHook := ClaudeHookConfig{
			Type:            "command",
			Command:         "ccfeedback",
			Timeout:         60000,
			ContinueOnError: false,
		}

		if emptyMatcherIndex >= 0 {
			// Add to existing group
			postToolUseGroups[emptyMatcherIndex].Hooks = append(
				postToolUseGroups[emptyMatcherIndex].Hooks,
				ccfeedbackHook,
			)
		} else {
			// Create new group
			postToolUseGroups = append(postToolUseGroups, HookGroup{
				Matcher: "",
				Hooks:   []ClaudeHookConfig{ccfeedbackHook},
			})
		}
	}

	modified.Hooks["PostToolUse"] = postToolUseGroups
	return modified
}

// marshalClaudeSettings marshals settings back to JSON preserving extra fields
func marshalClaudeSettings(settings *ClaudeSettings, extraFields map[string]json.RawMessage) ([]byte, error) {
	// Build the final object
	result := make(map[string]interface{})

	// Add extra fields first
	for key, value := range extraFields {
		var v interface{}
		if err := json.Unmarshal(value, &v); err != nil {
			result[key] = value
		} else {
			result[key] = v
		}
	}

	// Add known fields (these override extras if there's a conflict)
	if settings.Permissions != nil {
		result["permissions"] = settings.Permissions
	}
	if len(settings.Hooks) > 0 {
		result["hooks"] = settings.Hooks
	}

	// Marshal with nice formatting
	return json.MarshalIndent(result, "", "  ")
}

// displayChanges shows a diff-style comparison of the changes
func displayChanges(original, modified []byte) {
	fmt.Println("\n📝 Proposed Changes:")
	fmt.Println("==================================================")

	if len(original) == 0 {
		// New file - show as additions
		fmt.Println("Creating new settings.json:")
		fmt.Println()
		lines := strings.Split(string(modified), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("+ %s\n", line)
			}
		}
	} else {
		// Existing file - show actual diff
		var origSettings, modSettings map[string]interface{}
		if err := json.Unmarshal(original, &origSettings); err != nil {
			// Fallback to simple display
			fmt.Println("Error parsing original settings")
			return
		}
		if err := json.Unmarshal(modified, &modSettings); err != nil {
			// Fallback to simple display
			fmt.Println("Error parsing modified settings")
			return
		}

		// Check if hooks section exists in original
		origHooks, hasOrigHooks := origSettings["hooks"].(map[string]interface{})
		modHooks := modSettings["hooks"].(map[string]interface{})

		if !hasOrigHooks {
			// Adding hooks section for the first time
			fmt.Println("Adding new 'hooks' section:")
			fmt.Println()
			hookJSON, _ := json.MarshalIndent(map[string]interface{}{
				"hooks": modHooks,
			}, "", "  ")
			lines := strings.Split(string(hookJSON), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("+ %s\n", line)
				}
			}
		} else {
			// Modifying existing hooks
			origPostToolUse, hasOrigPTU := origHooks["PostToolUse"].([]interface{})
			modPostToolUse := modHooks["PostToolUse"].([]interface{})

			if !hasOrigPTU {
				fmt.Println("Adding 'PostToolUse' to existing hooks:")
				fmt.Println()
			} else {
				fmt.Println("Modifying 'PostToolUse' hooks:")
				fmt.Println()
				// Show what's being removed
				origJSON, _ := json.MarshalIndent(map[string]interface{}{
					"PostToolUse": origPostToolUse,
				}, "", "  ")
				lines := strings.Split(string(origJSON), "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Printf("- %s\n", line)
					}
				}
				fmt.Println()
			}

			// Show what's being added
			modJSON, _ := json.MarshalIndent(map[string]interface{}{
				"PostToolUse": modPostToolUse,
			}, "", "  ")
			lines := strings.Split(string(modJSON), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("+ %s\n", line)
				}
			}
		}

		// Check for other preserved fields
		preservedCount := 0
		for key := range origSettings {
			if key != "hooks" {
				preservedCount++
			}
		}
		if preservedCount > 0 {
			fmt.Printf("\n✓ Preserving %d other configuration field(s)\n", preservedCount)
		}
	}
	fmt.Println("==================================================")
}

// showNextSteps displays instructions for next steps
func showNextSteps(options InitOptions) {
	fmt.Println("\n✅ CCFeedback has been configured for Claude Code!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Create a ccfeedback configuration file:")
	fmt.Println("   - Global config: ~/.claude/ccfeedback.json")
	fmt.Println("   - Project config: .claude/ccfeedback.json")
	fmt.Println("\n2. Example ccfeedback.json:")
	fmt.Println(`{
  "linters": {
    "golang": {
      "enabled": true,
      "config": {
        "golangciConfig": ".golangci.yml"
      }
    },
    "markdown": {
      "enabled": true,
      "config": {
        "maxLineLength": 120
      }
    }
  }
}`)
	fmt.Println("\n3. Test your setup:")
	fmt.Println("   ccfeedback show-actions <file>")
}

// isCCFeedbackAvailable checks if ccfeedback is in PATH
func isCCFeedbackAvailable() bool {
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, path := range paths {
		fullPath := filepath.Join(path, "ccfeedback")
		if _, err := os.Stat(fullPath); err == nil {
			return true
		}
		// Check with .exe extension on Windows
		if _, err := os.Stat(fullPath + ".exe"); err == nil {
			return true
		}
	}
	return false
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0600)
}
