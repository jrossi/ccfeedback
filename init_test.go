package ccfeedback_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jrossi/ccfeedback"
)

func TestInitCommand(t *testing.T) {
	// Save original home directory
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name          string
		options       ccfeedback.InitOptions
		setupSettings string // Initial settings.json content
		validate      func(t *testing.T, settingsPath string)
	}{
		{
			name: "create new settings file",
			options: ccfeedback.InitOptions{
				DryRun: true,
				Force:  true,
			},
			setupSettings: "", // No existing file
			validate: func(t *testing.T, settingsPath string) {
				// With dry-run, file should not be created
				if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
					t.Error("Expected settings file to not exist with dry-run")
				}
			},
		},
		{
			name: "update existing settings",
			options: ccfeedback.InitOptions{
				Force: true,
			},
			setupSettings: `{
				"permissions": {
					"allow": ["Read", "Write"]
				}
			}`,
			validate: func(t *testing.T, settingsPath string) {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("Failed to read settings: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("Failed to parse settings: %v", err)
				}

				// Check that permissions are preserved
				if perms, ok := settings["permissions"].(map[string]interface{}); ok {
					if allow, ok := perms["allow"].([]interface{}); ok {
						if len(allow) != 2 {
							t.Errorf("Expected permissions to be preserved")
						}
					}
				}

				// Check that hooks were added
				if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
					if postToolUse, ok := hooks["PostToolUse"].([]interface{}); ok {
						if len(postToolUse) == 0 {
							t.Error("Expected PostToolUse hooks to be added")
						}
					} else {
						t.Error("Expected PostToolUse hooks to be added")
					}
				} else {
					t.Error("Expected hooks section to be added")
				}
			},
		},
		{
			name: "preserve existing ccfeedback hook",
			options: ccfeedback.InitOptions{
				Force: true,
			},
			setupSettings: `{
				"hooks": {
					"PostToolUse": [
						{
							"matcher": "",
							"hooks": [
								{
									"type": "command",
									"command": "ccfeedback",
									"timeout": 30000
								}
							]
						}
					]
				}
			}`,
			validate: func(t *testing.T, settingsPath string) {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("Failed to read settings: %v", err)
				}

				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err != nil {
					t.Fatalf("Failed to parse settings: %v", err)
				}

				// Check that the hook was updated with correct timeout
				hooks := settings["hooks"].(map[string]interface{})
				postToolUse := hooks["PostToolUse"].([]interface{})
				group := postToolUse[0].(map[string]interface{})
				hooksList := group["hooks"].([]interface{})
				hook := hooksList[0].(map[string]interface{})

				if hook["timeout"].(float64) != 60000 {
					t.Errorf("Expected timeout to be updated to 60000, got %v", hook["timeout"])
				}
			},
		},
		{
			name: "dry run mode",
			options: ccfeedback.InitOptions{
				DryRun: true,
				Force:  true,
			},
			setupSettings: `{"test": "value"}`,
			validate: func(t *testing.T, settingsPath string) {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("Failed to read settings: %v", err)
				}

				// Original content should be unchanged
				if string(data) != `{"test": "value"}` {
					t.Error("Settings file should not be modified in dry-run mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)

			// Create .claude directory
			claudeDir := filepath.Join(tmpDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("Failed to create .claude directory: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, "settings.json")

			// Setup initial settings if provided
			if tt.setupSettings != "" {
				if err := os.WriteFile(settingsPath, []byte(tt.setupSettings), 0600); err != nil {
					t.Fatalf("Failed to write initial settings: %v", err)
				}
			}

			// Change to temp dir for project-level tests
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(oldWd); err != nil {
					t.Errorf("Failed to restore working directory: %v", err)
				}
			}()

			// Run the command
			err = ccfeedback.InitCommand(tt.options)
			if err != nil {
				t.Fatalf("InitCommand failed: %v", err)
			}

			// Validate results
			tt.validate(t, settingsPath)
		})
	}
}

func TestInitCommandProjectOnly(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	// Run init command with project-only flag
	options := ccfeedback.InitOptions{
		ProjectOnly: true,
		Force:       true,
	}

	err = ccfeedback.InitCommand(options)
	if err != nil {
		t.Fatalf("InitCommand failed: %v", err)
	}

	// Check that project settings were created
	projectSettings := filepath.Join(".claude", "settings.json")
	if _, err := os.Stat(projectSettings); os.IsNotExist(err) {
		t.Error("Expected project settings to be created")
	}

	// Read and validate the settings
	data, err := os.ReadFile(projectSettings)
	if err != nil {
		t.Fatalf("Failed to read project settings: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to parse settings: %v", err)
	}

	// Check that hooks were added
	if _, ok := settings["hooks"]; !ok {
		t.Error("Expected hooks to be added to project settings")
	}
}
