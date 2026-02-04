package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/parser"
)

// Integration tests that verify the complete pipeline works
// and produces valid SVG output for real-world regex patterns

func TestIntegrationEmailPattern(t *testing.T) {
	// Simplified email pattern
	pattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationPhonePattern(t *testing.T) {
	// US phone number pattern
	pattern := `\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationURLPattern(t *testing.T) {
	// Simplified URL pattern (forward slashes must be escaped like in JS regex literals)
	pattern := `https?:\/\/[a-zA-Z0-9.-]+(?:\/[a-zA-Z0-9.\/_-]*)?`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationIPv4Pattern(t *testing.T) {
	// IPv4 address pattern (simplified)
	pattern := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationDatePattern(t *testing.T) {
	// Date pattern YYYY-MM-DD
	pattern := `\d{4}-\d{2}-\d{2}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationIdentifierPattern(t *testing.T) {
	// Programming language identifier
	pattern := `[a-zA-Z_][a-zA-Z0-9_]*`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationHexColorPattern(t *testing.T) {
	// Hex color pattern
	pattern := `#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationWithLookahead(t *testing.T) {
	// Password validation with lookahead
	pattern := `(?=.*\d)(?=.*[a-z])(?=.*[A-Z]).{8,}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationNestedGroups(t *testing.T) {
	// Nested groups
	pattern := `((a|b)(c|d))+`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

func TestIntegrationAllQuantifiers(t *testing.T) {
	// Pattern using all quantifier types
	pattern := `a*b+c?d{2}e{3,}f{4,5}`

	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	validateSVG(t, svg)
}

// validateSVG performs basic validation on SVG output
func validateSVG(t *testing.T, svg string) {
	t.Helper()

	if !strings.HasPrefix(svg, "<svg") {
		t.Error("SVG should start with <svg")
	}

	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("SVG should end with </svg>")
	}

	if !strings.Contains(svg, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("SVG should have xmlns attribute")
	}

	if !strings.Contains(svg, "viewBox") {
		t.Error("SVG should have viewBox attribute")
	}

	// Check for unbalanced tags (basic check)
	openTags := strings.Count(svg, "<g")
	closeTags := strings.Count(svg, "</g>")
	if openTags != closeTags {
		t.Errorf("unbalanced g tags: %d opens, %d closes", openTags, closeTags)
	}
}

// TestGoldenFiles tests against golden file outputs
func TestGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"literal", "abc"},
		{"alternation", "a|b|c"},
		{"charset", "[a-z]"},
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-question", "a?"},
		{"group", "(abc)"},
		{"anchor", "^start$"},
		{"escape-digit", `\d+`},
		{"complex", `^[a-z]+@[a-z]+\.[a-z]{2,}$`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			// If GOLDEN_UPDATE env var is set, update golden files
			if os.Getenv("GOLDEN_UPDATE") != "" {
				if err := os.WriteFile(goldenPath, []byte(svg), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			// Check if golden file exists
			golden, err := os.ReadFile(goldenPath)
			if os.IsNotExist(err) {
				// Golden file doesn't exist, create it
				if err := os.WriteFile(goldenPath, []byte(svg), 0644); err != nil {
					t.Fatalf("failed to create golden file: %v", err)
				}
				t.Logf("Created golden file: %s", goldenPath)
				return
			} else if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			// Compare with golden file
			if svg != string(golden) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}
