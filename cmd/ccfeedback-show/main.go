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
	// Define flags
	debug := flag.Bool("debug", false, "Enable debug output")
	configFile := flag.String("config", "", "Path to configuration file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ccfeedback-show [options] <file>...\n\n")
		fmt.Fprintf(os.Stderr, "Show which configuration rules would apply to the given files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check for required arguments
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: show-actions requires at least one file path\n")
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

	// Process each file
	for _, filePath := range flag.Args() {
		if err := showActionsForFile(filePath, ruleEngine, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)
			os.Exit(1)
		}
	}
}

// showActionsForFile displays which configuration rules would apply to the given file
func showActionsForFile(filePath string, ruleEngine *ccfeedback.LintingRuleEngine, debug bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("\n=== Configuration Analysis for: %s ===\n", absPath)

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

	switch ext {
	case ".go":
		applicableLinters = append(applicableLinters, "golang")
		fmt.Printf("✓ golang linter (handles .go files)\n")
	case ".md", ".markdown":
		applicableLinters = append(applicableLinters, "markdown")
		fmt.Printf("✓ markdown linter (handles .md files)\n")
	default:
		fmt.Printf("ℹ️  No linters configured for %s files\n", ext)
		return nil
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

	// Show config file locations if in debug mode
	if debug {
		fmt.Printf("\n--- Configuration Sources ---\n")
		fmt.Printf("Configuration files are loaded in this order (later files override earlier):\n")
		fmt.Printf("1. ~/.claude/ccfeedback.json\n")
		fmt.Printf("2. .claude/ccfeedback.json (project root)\n")
		fmt.Printf("3. .claude/ccfeedback.local.json (project root)\n")
		fmt.Printf("4. --config flag (if specified)\n")
	}

	return nil
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

