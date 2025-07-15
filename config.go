package ccfeedback

import (
	"encoding/json"
	"path/filepath"
	"time"
)

// AppConfig represents the complete configuration for ccfeedback
type AppConfig struct {
	// Global settings
	Parallel *ParallelConfig `json:"parallel,omitempty"`
	Timeout  *Duration       `json:"timeout,omitempty"`

	// Linter configurations keyed by linter name
	Linters map[string]LinterConfig `json:"linters,omitempty"`

	// Rule overrides by file pattern
	Rules []RuleOverride `json:"rules,omitempty"`
}

// ParallelConfig controls parallel execution settings
type ParallelConfig struct {
	MaxWorkers      *int  `json:"maxWorkers,omitempty"`
	DisableParallel *bool `json:"disableParallel,omitempty"`
}

// LinterConfig represents configuration for a specific linter
type LinterConfig struct {
	Enabled *bool           `json:"enabled,omitempty"`
	Config  json.RawMessage `json:"config,omitempty"`
}

// RuleOverride applies linter-specific rules based on file patterns
type RuleOverride struct {
	Pattern string          `json:"pattern"` // glob pattern for files
	Linter  string          `json:"linter"`  // which linter this applies to
	Rules   json.RawMessage `json:"rules"`   // linter-specific rule configuration
}

// Duration is a wrapper around time.Duration for JSON unmarshaling
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler for Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

// MarshalJSON implements json.Marshaler for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// MarkdownConfig represents markdown linter specific configuration
type MarkdownConfig struct {
	MaxLineLength      *int             `json:"maxLineLength,omitempty"`
	RequireFrontmatter *bool            `json:"requireFrontmatter,omitempty"`
	FrontmatterSchema  *json.RawMessage `json:"frontmatterSchema,omitempty"`
	DisabledRules      []string         `json:"disabledRules,omitempty"`
	MaxBlankLines      *int             `json:"maxBlankLines,omitempty"`
	ListIndentSize     *int             `json:"listIndentSize,omitempty"`
}

// GolangConfig represents golang linter specific configuration
type GolangConfig struct {
	GolangciConfig *string   `json:"golangciConfig,omitempty"` // path to golangci.yml
	DisabledChecks []string  `json:"disabledChecks,omitempty"`
	TestTimeout    *Duration `json:"testTimeout,omitempty"`
}

// NewAppConfig creates a new AppConfig with default values
func NewAppConfig() *AppConfig {
	return &AppConfig{
		Linters: make(map[string]LinterConfig),
		Rules:   []RuleOverride{},
	}
}

// Merge combines two configs, with other taking precedence
func (c *AppConfig) Merge(other *AppConfig) {
	if other == nil {
		return
	}

	// Merge parallel config
	if other.Parallel != nil {
		if c.Parallel == nil {
			c.Parallel = &ParallelConfig{}
		}
		if other.Parallel.MaxWorkers != nil {
			c.Parallel.MaxWorkers = other.Parallel.MaxWorkers
		}
		if other.Parallel.DisableParallel != nil {
			c.Parallel.DisableParallel = other.Parallel.DisableParallel
		}
	}

	// Merge timeout
	if other.Timeout != nil {
		c.Timeout = other.Timeout
	}

	// Merge linters
	if c.Linters == nil {
		c.Linters = make(map[string]LinterConfig)
	}
	for name, linterConfig := range other.Linters {
		existing, exists := c.Linters[name]
		if !exists {
			c.Linters[name] = linterConfig
		} else {
			// Merge linter config
			if linterConfig.Enabled != nil {
				existing.Enabled = linterConfig.Enabled
			}
			if linterConfig.Config != nil {
				existing.Config = linterConfig.Config
			}
			c.Linters[name] = existing
		}
	}

	// Append rules (don't merge, later rules take precedence)
	c.Rules = append(c.Rules, other.Rules...)
}

// GetLinterConfig returns the configuration for a specific linter
func (c *AppConfig) GetLinterConfig(name string) (json.RawMessage, bool) {
	if c.Linters == nil {
		return nil, false
	}
	linterConfig, ok := c.Linters[name]
	if !ok || linterConfig.Config == nil {
		return nil, false
	}
	return linterConfig.Config, true
}

// IsLinterEnabled checks if a linter is enabled
func (c *AppConfig) IsLinterEnabled(name string) bool {
	if c.Linters == nil {
		return true // default to enabled
	}
	linterConfig, ok := c.Linters[name]
	if !ok || linterConfig.Enabled == nil {
		return true // default to enabled
	}
	return *linterConfig.Enabled
}

// GetRuleOverrides returns all rule overrides that match the given file path for a specific linter
func (c *AppConfig) GetRuleOverrides(filePath, linterName string) []json.RawMessage {
	if len(c.Rules) == 0 {
		return nil
	}

	var overrides []json.RawMessage
	for _, rule := range c.Rules {
		// Check if this rule applies to the given linter
		if rule.Linter != linterName && rule.Linter != "*" {
			continue
		}

		// Check if the pattern matches the file path
		matched, err := filepath.Match(rule.Pattern, filePath)
		if err != nil {
			// Invalid pattern, skip
			continue
		}

		if !matched {
			// Also check against just the filename
			matched, _ = filepath.Match(rule.Pattern, filepath.Base(filePath))
		}

		if matched {
			overrides = append(overrides, rule.Rules)
		}
	}

	return overrides
}
