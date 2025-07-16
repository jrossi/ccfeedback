package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jrossi/ccfeedback"
)

func main() {
	// Define global flags
	debug := flag.Bool("debug", false, "Enable debug output")
	configFile := flag.String("config", "", "Path to configuration file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ccfeedback-show [options] <command> [arguments]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  config          Show current configuration\n")
		fmt.Fprintf(os.Stderr, "  filter <file>   Show which rules and linters apply to a file\n")
		fmt.Fprintf(os.Stderr, "  setup           Show setup status and configuration paths\n")
		fmt.Fprintf(os.Stderr, "  linters         Show available linters and their status\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check for required command
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: no command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	configLoader, err := ccfeedback.NewConfigLoader()
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "Failed to create config loader: %v\n", err)
		}
		// Continue without config
		configLoader = nil
	}

	var appConfig *ccfeedback.AppConfig
	if configLoader != nil {
		if *configFile != "" {
			// Load specific config file
			appConfig, err = configLoader.LoadConfigWithPaths([]string{*configFile})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load config file %s: %v\n", *configFile, err)
				os.Exit(1)
			}
		} else {
			// Load default config files
			appConfig, err = configLoader.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Create linting config from app config
	lintingConfig := ccfeedback.LintingConfig{}
	if appConfig != nil {
		if appConfig.Parallel != nil {
			if appConfig.Parallel.MaxWorkers != nil {
				lintingConfig.MaxWorkers = *appConfig.Parallel.MaxWorkers
			}
			if appConfig.Parallel.DisableParallel != nil {
				lintingConfig.DisableParallel = *appConfig.Parallel.DisableParallel
			}
		}
	}

	// Create rule engine with linting capabilities
	ruleEngine := ccfeedback.NewLintingRuleEngineWithConfig(lintingConfig)

	// Set the app config if available
	if appConfig != nil {
		ruleEngine.SetAppConfig(appConfig)
	}

	// Handle commands
	command := flag.Args()[0]
	switch command {
	case "config":
		if err := showConfig(appConfig, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "filter":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Error: filter command requires a file path\n")
			os.Exit(1)
		}
		if err := showFilter(flag.Args()[1], ruleEngine, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "setup":
		if err := showSetup(appConfig, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "linters":
		if err := showLinters(ruleEngine, appConfig, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

// showConfig displays the current configuration
func showConfig(appConfig *ccfeedback.AppConfig, debug bool) error {
	fmt.Println("=== Current Configuration ===")

	if appConfig == nil {
		fmt.Println("\nNo configuration loaded.")
		return nil
	}

	// Pretty print the configuration
	configJSON, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println()
	fmt.Println(string(configJSON))

	if debug {
		fmt.Println("\n--- Configuration Sources ---")
		fmt.Println("Configuration files are loaded in this order (later files override earlier):")
		fmt.Println("1. ~/.claude/ccfeedback.json (global)")
		fmt.Println("2. .claude/ccfeedback.json (project)")
		fmt.Println("3. .claude/ccfeedback.local.json (local overrides)")
		fmt.Println("4. --config flag (if specified)")
	}

	return nil
}

// showFilter shows which rules and linters apply to a specific file
func showFilter(filePath string, ruleEngine *ccfeedback.LintingRuleEngine, debug bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("=== Filter Analysis for: %s ===\n", absPath)

	// Get the app config from the rule engine
	appConfig := ruleEngine.GetAppConfig()
	if appConfig == nil {
		fmt.Printf("\nNo configuration loaded.\n")
		return nil
	}

	// Determine which linters would handle this file
	fmt.Printf("\n--- Applicable Linters ---\n")
	ext := filepath.Ext(filePath)
	applicableLinters := []string{}

	// Check all possible linters
	linterMap := map[string][]string{
		".go":       {"golang"},
		".md":       {"markdown"},
		".markdown": {"markdown"},
		".js":       {"javascript"},
		".jsx":      {"javascript"},
		".ts":       {"javascript"},
		".tsx":      {"javascript"},
		".py":       {"python"},
		".rs":       {"rust"},
		".proto":    {"protobuf"},
		".json":     {"json"},
		".jsonc":    {"json"},
		".json5":    {"json"},
	}

	if linters, ok := linterMap[ext]; ok {
		for _, linter := range linters {
			applicableLinters = append(applicableLinters, linter)
			fmt.Printf("✓ %s linter (handles %s files)\n", linter, ext)
		}
	} else {
		fmt.Printf("ℹ️  No linters configured for %s files\n", ext)
	}

	// Show base configuration for each applicable linter
	for _, linterName := range applicableLinters {
		fmt.Printf("\n--- Base Configuration for %s ---\n", linterName)

		if linterConfig, exists := appConfig.GetLinterConfig(linterName); exists {
			// Pretty print the linter config
			var configMap map[string]interface{}
			if err := json.Unmarshal(linterConfig, &configMap); err == nil {
				for key, value := range configMap {
					fmt.Printf("  %s: %v\n", key, value)
				}
			} else {
				fmt.Printf("  Raw config: %s\n", string(linterConfig))
			}
		} else {
			fmt.Printf("  (default configuration)\n")
		}

		// Check if linter is enabled
		if appConfig.IsLinterEnabled(linterName) {
			fmt.Printf("  ✓ Linter is enabled\n")
		} else {
			fmt.Printf("  ✗ Linter is disabled\n")
		}
	}

	// Show which rules would apply
	fmt.Printf("\n--- Rule Hierarchy ---\n")
	fmt.Printf("Rules are applied in order. Later rules override earlier ones.\n\n")

	matchedRules := false
	for i, rule := range appConfig.Rules {
		// Check if this rule matches the file
		matched := MatchesPattern(rule.Pattern, absPath)

		if debug && !matched {
			fmt.Printf("   Pattern '%s' did not match '%s'\n", rule.Pattern, absPath)
		}

		if matched {
			matchedRules = true
			fmt.Printf("%d. Pattern: %s", i+1, rule.Pattern)
			if rule.Linter == "*" {
				fmt.Printf(" (applies to ALL linters)\n")
			} else {
				fmt.Printf(" (applies to %s linter)\n", rule.Linter)
			}

			// Show what this rule would override
			var overrideMap map[string]interface{}
			if err := json.Unmarshal(rule.Rules, &overrideMap); err == nil {
				for key, value := range overrideMap {
					fmt.Printf("   → %s: %v\n", key, value)
				}
			}
			fmt.Printf("\n")
		}
	}

	if !matchedRules {
		fmt.Printf("ℹ️  No pattern-based rules match this file.\n")
		fmt.Printf("   Base linter configuration will be used.\n")
	}

	// Show the final merged configuration for each linter
	for _, linterName := range applicableLinters {
		fmt.Printf("\n--- Final Configuration for %s ---\n", linterName)
		fmt.Printf("(After applying all matching rules)\n")

		// Get all overrides that would apply
		overrides := appConfig.GetRuleOverrides(absPath, linterName)

		// Start with base config
		finalConfig := make(map[string]interface{})
		if baseConfig, exists := appConfig.GetLinterConfig(linterName); exists {
			_ = json.Unmarshal(baseConfig, &finalConfig)
		}

		// Apply each override in order
		for _, override := range overrides {
			var overrideMap map[string]interface{}
			if err := json.Unmarshal(override, &overrideMap); err == nil {
				for k, v := range overrideMap {
					finalConfig[k] = v
				}
			}
		}

		// Display final config
		if len(finalConfig) > 0 {
			for key, value := range finalConfig {
				fmt.Printf("  %s: %v\n", key, value)
			}
		} else {
			fmt.Printf("  (default configuration)\n")
		}
	}

	return nil
}

// showSetup displays setup status and configuration paths
func showSetup(appConfig *ccfeedback.AppConfig, debug bool) error {
	fmt.Println("=== CCFeedback Setup Status ===")

	// Check ccfeedback binary
	fmt.Println("\n--- Binary Status ---")
	if isCCFeedbackAvailable() {
		fmt.Println("✓ ccfeedback is available in PATH")
	} else {
		fmt.Println("✗ ccfeedback not found in PATH")
		fmt.Println("  Run: go install github.com/jrossi/ccfeedback/cmd/ccfeedback")
	}

	// Check configuration files
	fmt.Println("\n--- Configuration Files ---")
	homeDir, _ := os.UserHomeDir()
	configs := []struct {
		path string
		desc string
	}{
		{filepath.Join(homeDir, ".claude", "ccfeedback.json"), "Global config"},
		{".claude/ccfeedback.json", "Project config"},
		{".claude/ccfeedback.local.json", "Local overrides"},
	}

	foundConfig := false
	for _, cfg := range configs {
		if _, err := os.Stat(cfg.path); err == nil {
			fmt.Printf("✓ %s: %s\n", cfg.desc, cfg.path)
			foundConfig = true
		} else {
			fmt.Printf("  %s: not found\n", cfg.desc)
		}
	}

	if !foundConfig {
		fmt.Println("\nℹ️  No configuration files found. Using defaults.")
	}

	// Check Claude settings
	fmt.Println("\n--- Claude Integration ---")
	claudeSettings := []string{
		filepath.Join(homeDir, ".claude", "settings.json"),
		".claude/settings.json",
	}

	foundSettings := false
	for _, settingsPath := range claudeSettings {
		if data, err := os.ReadFile(settingsPath); err == nil {
			var settings map[string]interface{}
			if err := json.Unmarshal(data, &settings); err == nil {
				if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
					if postToolUse, ok := hooks["PostToolUse"].([]interface{}); ok {
						for _, group := range postToolUse {
							if g, ok := group.(map[string]interface{}); ok {
								if hookList, ok := g["hooks"].([]interface{}); ok {
									for _, hook := range hookList {
										if h, ok := hook.(map[string]interface{}); ok {
											if cmd, ok := h["command"].(string); ok && strings.Contains(cmd, "ccfeedback") {
												fmt.Printf("✓ Hook configured in: %s\n", settingsPath)
												foundSettings = true
												break
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if !foundSettings {
		fmt.Println("✗ No ccfeedback hooks found in Claude settings")
		fmt.Println("  Run: ccfeedback init")
	}

	// Show current configuration summary
	if appConfig != nil {
		fmt.Println("\n--- Configuration Summary ---")

		// Count enabled linters
		enabledCount := 0
		totalCount := 0
		if appConfig.Linters != nil {
			for _, linterCfg := range appConfig.Linters {
				totalCount++
				// LinterConfig has json.RawMessage, so we need to check the raw config
				if linterCfgJSON, err := json.Marshal(linterCfg); err == nil {
					var cfg map[string]interface{}
					if err := json.Unmarshal(linterCfgJSON, &cfg); err == nil {
						if enabled, ok := cfg["enabled"].(bool); !ok || enabled {
							enabledCount++
						}
					} else {
						enabledCount++ // Default to enabled
					}
				} else {
					enabledCount++ // Default to enabled
				}
			}
		}
		fmt.Printf("Linters: %d enabled / %d configured\n", enabledCount, totalCount)

		// Count rules
		fmt.Printf("Rules: %d pattern-based overrides\n", len(appConfig.Rules))

		// Show timeout
		if appConfig.Timeout != nil {
			fmt.Printf("Timeout: %s\n", appConfig.Timeout.Duration)
		}

		// Show parallel config
		if appConfig.Parallel != nil {
			if appConfig.Parallel.DisableParallel != nil && *appConfig.Parallel.DisableParallel {
				fmt.Println("Parallel: disabled")
			} else if appConfig.Parallel.MaxWorkers != nil {
				fmt.Printf("Parallel: enabled (max %d workers)\n", *appConfig.Parallel.MaxWorkers)
			}
		}
	}

	if debug {
		fmt.Println("\n--- Environment ---")
		fmt.Printf("Current directory: %s\n", mustGetwd())
		fmt.Printf("Home directory: %s\n", homeDir)
		fmt.Printf("PATH: %s\n", os.Getenv("PATH"))
	}

	return nil
}

// showLinters displays available linters and their status
func showLinters(ruleEngine *ccfeedback.LintingRuleEngine, appConfig *ccfeedback.AppConfig, debug bool) error {
	fmt.Println("=== Available Linters ===")

	// Define all known linters with their details
	linters := []struct {
		name        string
		extensions  []string
		description string
		tool        string
	}{
		{
			name:        "golang",
			extensions:  []string{".go"},
			description: "Go code linter using golangci-lint",
			tool:        "golangci-lint",
		},
		{
			name:        "markdown",
			extensions:  []string{".md", ".markdown"},
			description: "Markdown linter checking formatting and style",
			tool:        "built-in",
		},
		{
			name:        "javascript",
			extensions:  []string{".js", ".jsx", ".ts", ".tsx"},
			description: "JavaScript/TypeScript linter using ESLint",
			tool:        "eslint",
		},
		{
			name:        "python",
			extensions:  []string{".py"},
			description: "Python linter using ruff",
			tool:        "ruff",
		},
		{
			name:        "rust",
			extensions:  []string{".rs"},
			description: "Rust linter using cargo check and clippy",
			tool:        "cargo",
		},
		{
			name:        "protobuf",
			extensions:  []string{".proto"},
			description: "Protocol Buffer linter using buf",
			tool:        "buf",
		},
		{
			name:        "json",
			extensions:  []string{".json", ".jsonc", ".json5"},
			description: "JSON syntax and formatting checker",
			tool:        "built-in",
		},
	}

	for _, linter := range linters {
		fmt.Printf("\n--- %s ---\n", linter.name)
		fmt.Printf("Description: %s\n", linter.description)
		fmt.Printf("File types: %s\n", strings.Join(linter.extensions, ", "))
		fmt.Printf("Tool: %s\n", linter.tool)

		// Check if enabled in config
		if appConfig != nil {
			if appConfig.IsLinterEnabled(linter.name) {
				fmt.Printf("Status: ✓ Enabled\n")
			} else {
				fmt.Printf("Status: ✗ Disabled\n")
			}

			// Show configuration if present
			if linterConfig, exists := appConfig.GetLinterConfig(linter.name); exists {
				var configMap map[string]interface{}
				if err := json.Unmarshal(linterConfig, &configMap); err == nil && len(configMap) > 0 {
					fmt.Println("Configuration:")
					for key, value := range configMap {
						if key != "enabled" { // Don't show enabled again
							fmt.Printf("  %s: %v\n", key, value)
						}
					}
				}
			}
		} else {
			fmt.Printf("Status: ✓ Enabled (default)\n")
		}

		// Check if tool is available
		if linter.tool != "built-in" {
			if isToolAvailable(linter.tool) {
				fmt.Printf("Tool availability: ✓ %s found\n", linter.tool)
			} else {
				fmt.Printf("Tool availability: ✗ %s not found in PATH\n", linter.tool)
			}
		}
	}

	if debug {
		fmt.Println("\n--- Debug Info ---")
		fmt.Printf("Total linters available: %d\n", len(linters))
		if appConfig != nil && appConfig.Linters != nil {
			fmt.Printf("Linters in config: %d\n", len(appConfig.Linters))
		}
	}

	return nil
}

// Helper functions

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

func isToolAvailable(tool string) bool {
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, path := range paths {
		fullPath := filepath.Join(path, tool)
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

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

// MatchesPattern checks if a file path matches a glob pattern
// It supports ** for matching any number of directories
func MatchesPattern(pattern, path string) bool {
	// For absolute paths, also try relative matching from current directory
	relPath := path
	if filepath.IsAbs(path) {
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, path); err == nil {
				relPath = rel
			}
		}
	}

	// Try both absolute and relative paths
	for _, p := range []string{path, relPath} {
		// First try direct match
		if matched, _ := filepath.Match(pattern, p); matched {
			return true
		}

		// Try matching against just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(p)); matched {
			return true
		}

		// Handle ** patterns
		if strings.Contains(pattern, "**") {
			if MatchesDoubleStarPattern(pattern, p) {
				return true
			}
		}
	}

	return false
}

// MatchesDoubleStarPattern handles patterns with ** for directory wildcards
func MatchesDoubleStarPattern(pattern, path string) bool {
	// Convert ** to a regex-like pattern
	// e.g., "internal/**/*.go" should match "internal/foo/bar.go"
	parts := strings.Split(pattern, "**")
	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// For patterns starting with **, match anywhere in path
		if prefix == "" && suffix != "" {
			// Pattern like "**/*.go" should match any .go file at any depth
			pathParts := strings.Split(path, "/")
			for i := range pathParts {
				subPath := strings.Join(pathParts[i:], "/")
				if matched, _ := filepath.Match(suffix, subPath); matched {
					return true
				}
			}
			// Also check just the filename
			return MatchesSimplePattern(suffix, filepath.Base(path))
		}

		// Check if path starts with prefix
		if prefix != "" && !strings.HasPrefix(path, prefix+"/") && path != prefix {
			return false
		}

		// Get the part after the prefix
		remainder := strings.TrimPrefix(path, prefix)
		remainder = strings.TrimPrefix(remainder, "/")

		// Check if the remainder matches the suffix pattern
		if suffix != "" {
			// For patterns like "*.go", we need to check the end of the path
			if strings.HasPrefix(suffix, "*") && !strings.Contains(suffix, "/") {
				return strings.HasSuffix(remainder, strings.TrimPrefix(suffix, "*"))
			}
			// For other patterns, try to match against the remainder
			if matched, _ := filepath.Match(suffix, remainder); matched {
				return true
			}
			// Also try matching just the filename part
			if matched, _ := filepath.Match(suffix, filepath.Base(remainder)); matched {
				return true
			}
		} else {
			// Pattern ends with **, matches everything under prefix
			return true
		}
	}

	return false
}

// MatchesSimplePattern is a helper for simple pattern matching
func MatchesSimplePattern(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}

