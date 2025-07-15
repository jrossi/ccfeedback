package markdown

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jrossi/ccfeedback/linters"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"

	markdown "github.com/teekennedy/goldmark-markdown"
)

// MarkdownLinter handles markdown file linting, formatting, and front matter validation
type MarkdownLinter struct {
	parser  goldmark.Markdown
	rules   []MarkdownRule
	schemas map[string]*jsonschema.Schema
}

// MarkdownRule defines the interface for markdown linting rules
type MarkdownRule interface {
	Check(doc ast.Node, source []byte, filePath string) []linters.Issue
	Name() string
}

// NewMarkdownLinter creates a new markdown linter with all standard rules
func NewMarkdownLinter() *MarkdownLinter {
	// Create parser with front matter support
	frontmatterExtension := &frontmatter.Extender{}
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			frontmatterExtension,
		),
	)

	// Initialize all linting rules
	rules := []MarkdownRule{
		&HeadingHierarchyRule{},
		&ListIndentationRule{},
		&CodeBlockRule{},
		&LineLengthRule{MaxLength: 120},
		&TrailingWhitespaceRule{},
		&EmphasisConsistencyRule{},
		&BlankLineRule{},
	}

	return &MarkdownLinter{
		parser:  md,
		rules:   rules,
		schemas: make(map[string]*jsonschema.Schema),
	}
}

// Name returns the linter name
func (l *MarkdownLinter) Name() string {
	return "markdown"
}

// CanHandle returns true for markdown files
func (l *MarkdownLinter) CanHandle(filePath string) bool {
	return strings.HasSuffix(filePath, ".md") || strings.HasSuffix(filePath, ".markdown")
}

// Lint performs comprehensive linting on a markdown file
func (l *MarkdownLinter) Lint(_ context.Context, filePath string, content []byte) (*linters.LintResult, error) {
	result := &linters.LintResult{
		Success: true,
		Issues:  []linters.Issue{},
	}

	// Parse markdown with front matter
	reader := text.NewReader(content)
	parserCtx := parser.NewContext()
	document := l.parser.Parser().Parse(reader, parser.WithContext(parserCtx))

	// Extract front matter if present
	frontMatterData := frontmatter.Get(parserCtx)
	if frontMatterData != nil {
		// Convert to our concrete FrontMatter type
		var fm FrontMatter
		if err := frontMatterData.Decode(&fm); err != nil {
			// If decode fails, just skip validation
			// This could happen with non-standard front matter
		} else {
			// Validate front matter against schema
			if schemaIssues := l.validateFrontMatter(filePath, &fm); len(schemaIssues) > 0 {
				result.Issues = append(result.Issues, schemaIssues...)
			}
		}
	}

	// Apply all linting rules
	for _, rule := range l.rules {
		issues := rule.Check(document, content, filePath)
		result.Issues = append(result.Issues, issues...)
	}

	// Generate formatted output using a new renderer instance for thread safety
	var formatted bytes.Buffer
	formatter := markdown.NewRenderer()
	if err := formatter.Render(&formatted, content, document); err != nil {
		return nil, fmt.Errorf("failed to format markdown: %w", err)
	}
	result.Formatted = formatted.Bytes()

	// Check if formatting changed (indicates formatting issues)
	if !bytes.Equal(content, result.Formatted) {
		result.Issues = append(result.Issues, linters.Issue{
			File:     filePath,
			Line:     1,
			Column:   1,
			Severity: "warning",
			Message:  "File requires formatting to meet standards",
			Rule:     "formatting",
		})
	}

	// Determine overall success
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			result.Success = false
			break
		}
	}

	return result, nil
}

// FrontMatter represents structured front matter data
type FrontMatter struct {
	Title       string            `yaml:"title" json:"title"`
	Date        string            `yaml:"date" json:"date"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Author      string            `yaml:"author" json:"author"`
	Description string            `yaml:"description" json:"description"`
	Metadata    map[string]string `yaml:",inline" json:"-"`
}

// validateFrontMatter validates front matter against JSON schemas
func (l *MarkdownLinter) validateFrontMatter(format string, data *FrontMatter) []linters.Issue {
	// For now, just validate that front matter is well-formed
	// TODO: Add JSON schema validation when schemas are configured
	return []linters.Issue{}
}

// HeadingHierarchyRule ensures proper heading level progression (H1 → H2 → H3)
type HeadingHierarchyRule struct{}

func (r *HeadingHierarchyRule) Name() string {
	return "heading-hierarchy"
}

func (r *HeadingHierarchyRule) Check(doc ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue
	var lastLevel int
	var hasH1 bool

	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := node.(*ast.Heading); ok && entering {
			if heading.Level == 1 {
				if hasH1 {
					issues = append(issues, linters.Issue{
						File:     filePath,
						Line:     getLineNumber(source, heading),
						Column:   1,
						Severity: "warning",
						Message:  "Multiple H1 headings found, consider using H2 for subsequent sections",
						Rule:     r.Name(),
					})
				}
				hasH1 = true
			}

			if lastLevel > 0 && heading.Level > lastLevel+1 {
				issues = append(issues, linters.Issue{
					File:     filePath,
					Line:     getLineNumber(source, heading),
					Column:   1,
					Severity: "error",
					Message:  fmt.Sprintf("Heading level %d skips level %d (should not skip levels)", heading.Level, lastLevel+1),
					Rule:     r.Name(),
				})
			}
			lastLevel = heading.Level
		}
		return ast.WalkContinue, nil
	})

	return issues
}

// ListIndentationRule ensures consistent 2-space indentation for nested lists
type ListIndentationRule struct{}

func (r *ListIndentationRule) Name() string {
	return "list-indentation"
}

func (r *ListIndentationRule) Check(_ ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue
	lines := strings.Split(string(source), "\n")

	// Check list indentation using simple pattern matching
	listPattern := regexp.MustCompile(`^(\s*)[-*+]|\d+\.\s`)

	for i, line := range lines {
		if matches := listPattern.FindStringSubmatch(line); matches != nil {
			indent := len(matches[1])
			if indent%2 != 0 {
				issues = append(issues, linters.Issue{
					File:     filePath,
					Line:     i + 1,
					Column:   1,
					Severity: "warning",
					Message:  "List items should use 2-space indentation for nesting",
					Rule:     r.Name(),
				})
			}
		}
	}

	return issues
}

// CodeBlockRule ensures code blocks have language specifications
type CodeBlockRule struct{}

func (r *CodeBlockRule) Name() string {
	return "code-block-language"
}

func (r *CodeBlockRule) Check(doc ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue

	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if codeBlock, ok := node.(*ast.FencedCodeBlock); ok && entering {
			if codeBlock.Info == nil || codeBlock.Info.Value(source) == nil || len(codeBlock.Info.Value(source)) == 0 {
				issues = append(issues, linters.Issue{
					File:     filePath,
					Line:     getLineNumber(source, codeBlock),
					Column:   1,
					Severity: "warning",
					Message:  "Code blocks should specify a language for syntax highlighting",
					Rule:     r.Name(),
				})
			}
		}
		return ast.WalkContinue, nil
	})

	return issues
}

// LineLengthRule checks for lines exceeding maximum length
type LineLengthRule struct {
	MaxLength int
}

func (r *LineLengthRule) Name() string {
	return "line-length"
}

func (r *LineLengthRule) Check(_ ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue
	lines := strings.Split(string(source), "\n")

	for i, line := range lines {
		if len(line) > r.MaxLength {
			issues = append(issues, linters.Issue{
				File:     filePath,
				Line:     i + 1,
				Column:   r.MaxLength + 1,
				Severity: "warning",
				Message:  fmt.Sprintf("Line exceeds maximum length of %d characters (%d)", r.MaxLength, len(line)),
				Rule:     r.Name(),
			})
		}
	}

	return issues
}

// TrailingWhitespaceRule checks for trailing whitespace
type TrailingWhitespaceRule struct{}

func (r *TrailingWhitespaceRule) Name() string {
	return "trailing-whitespace"
}

func (r *TrailingWhitespaceRule) Check(_ ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue
	lines := strings.Split(string(source), "\n")

	for i, line := range lines {
		if len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
			issues = append(issues, linters.Issue{
				File:     filePath,
				Line:     i + 1,
				Column:   len(line),
				Severity: "error",
				Message:  "Line has trailing whitespace",
				Rule:     r.Name(),
			})
		}
	}

	return issues
}

// EmphasisConsistencyRule ensures consistent emphasis markers
type EmphasisConsistencyRule struct{}

func (r *EmphasisConsistencyRule) Name() string {
	return "emphasis-consistency"
}

func (r *EmphasisConsistencyRule) Check(doc ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue

	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if emphasis, ok := node.(*ast.Emphasis); ok && entering {
			// Check if using * for italic (preferred)
			line := getLineNumber(source, emphasis)
			if line > 0 {
				lines := strings.Split(string(source), "\n")
				if line <= len(lines) {
					lineText := lines[line-1]
					if strings.Contains(lineText, "_") && !strings.Contains(lineText, "*") {
						issues = append(issues, linters.Issue{
							File:     filePath,
							Line:     line,
							Column:   1,
							Severity: "info",
							Message:  "Prefer * for italic emphasis over _",
							Rule:     r.Name(),
						})
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})

	return issues
}

// BlankLineRule ensures proper spacing around elements
type BlankLineRule struct{}

func (r *BlankLineRule) Name() string {
	return "blank-line-spacing"
}

func (r *BlankLineRule) Check(_ ast.Node, source []byte, filePath string) []linters.Issue {
	var issues []linters.Issue
	lines := strings.Split(string(source), "\n")

	// Check for excessive blank lines (more than 2 consecutive)
	consecutiveBlank := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveBlank++
			if consecutiveBlank > 2 {
				issues = append(issues, linters.Issue{
					File:     filePath,
					Line:     i + 1,
					Column:   1,
					Severity: "warning",
					Message:  "More than 2 consecutive blank lines",
					Rule:     r.Name(),
				})
			}
		} else {
			consecutiveBlank = 0
		}
	}

	return issues
}

// getLineNumber calculates the line number for an AST node
func getLineNumber(source []byte, node ast.Node) int {
	// Try to get line information from different node types
	switch n := node.(type) {
	case *ast.Heading:
		if n.Lines().Len() > 0 {
			segment := n.Lines().At(0)
			lineCount := 1
			for i := 0; i < segment.Start && i < len(source); i++ {
				if source[i] == '\n' {
					lineCount++
				}
			}
			return lineCount
		}
		// Fallback: search for heading text
		lines := strings.Split(string(source), "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), strings.Repeat("#", n.Level)) {
				return i + 1
			}
		}
	case *ast.FencedCodeBlock:
		if n.Lines().Len() > 0 {
			segment := n.Lines().At(0)
			lineCount := 1
			for i := 0; i < segment.Start && i < len(source); i++ {
				if source[i] == '\n' {
					lineCount++
				}
			}
			return lineCount
		}
	case *ast.List:
		if n.Lines().Len() > 0 {
			segment := n.Lines().At(0)
			lineCount := 1
			for i := 0; i < segment.Start && i < len(source); i++ {
				if source[i] == '\n' {
					lineCount++
				}
			}
			return lineCount
		}
	}

	// For other nodes, try to find a parent with line info
	parent := node.Parent()
	if parent != nil {
		return getLineNumber(source, parent)
	}

	return 1
}
