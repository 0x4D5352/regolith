package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor"
	"github.com/0x4d5352/regolith/internal/flavor/dotnet"
	_ "github.com/0x4d5352/regolith/internal/flavor/gnugrep_bre"
	_ "github.com/0x4d5352/regolith/internal/flavor/gnugrep_ere"
	"github.com/0x4d5352/regolith/internal/flavor/java"
	"github.com/0x4d5352/regolith/internal/flavor/pcre"
	"github.com/0x4d5352/regolith/internal/flavor/posix_bre"
	"github.com/0x4d5352/regolith/internal/flavor/posix_ere"
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

// TestPOSIXEREGoldenFiles tests POSIX ERE patterns against golden file outputs
func TestPOSIXEREGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/posix-ere"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	ere := &posix_ere.POSIXERE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"literal", "abc"},
		{"alternation", "a|b|c"},
		{"charset", "[a-z]"},
		{"posix-alpha", "[[:alpha:]]"},
		{"posix-digit", "[[:digit:]]"},
		{"posix-alnum", "[[:alnum:]]"},
		{"posix-space", "[[:space:]]"},
		{"posix-multiple", "[[:alpha:][:digit:]]"},
		{"posix-negated", "[^[:digit:]]"},
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-range", "a{2,5}"},
		{"group", "(abc)"},
		{"anchor", "^start$"},
		{"complex-email", "[[:alnum:]]+@[[:alnum:]]+"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ere.Parse(tc.pattern)
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

// TestPOSIXEREIntegration tests complete POSIX ERE rendering pipeline
func TestPOSIXEREIntegration(t *testing.T) {
	ere := &posix_ere.POSIXERE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"identifier", "[[:alpha:]_][[:alnum:]_]*"},
		{"phone", "[0-9]{3}-[0-9]{4}"},
		{"date", "[0-9]{4}-[0-9]{2}-[0-9]{2}"},
		{"hex", "[[:xdigit:]]+"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestPOSIXBREGoldenFiles tests POSIX BRE patterns against golden file outputs
func TestPOSIXBREGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/posix-bre"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	bre := &posix_bre.POSIXBRE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic literals - note + ? | ( ) { } are literal in BRE!
		{"literal", "abc"},
		{"literal-special", "a+b?c|d"},
		{"literal-parens", "(abc)"},

		// BRE groups with \( \)
		{"group", `\(abc\)`},
		{"group-nested", `\(\(a\)\(b\)\)`},

		// Character sets
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// POSIX classes
		{"posix-alpha", "[[:alpha:]]"},
		{"posix-digit", "[[:digit:]]"},
		{"posix-alnum", "[[:alnum:]]"},
		{"posix-space", "[[:space:]]"},
		{"posix-multiple", "[[:alpha:][:digit:]]"},
		{"posix-negated", "[^[:digit:]]"},

		// Quantifiers (only * and \{n,m\} in BRE)
		{"quantifier-star", "a*"},
		{"quantifier-exact", `a\{3\}`},
		{"quantifier-min", `a\{2,\}`},
		{"quantifier-range", `a\{2,5\}`},

		// Back-references (BRE supports these!)
		{"backref", `\(a\)\1`},
		{"backref-word", `\([[:alpha:]]*\) \1`},

		// Anchors
		{"anchor", "^start$"},

		// Complex BRE patterns
		{"complex-word", `\([[:alpha:]]*\)\1`},
		{"complex-phone", `[0-9]\{3\}-[0-9]\{4\}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for pattern %q: %v", tc.pattern, err)
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

// TestPOSIXBREIntegration tests complete POSIX BRE rendering pipeline
func TestPOSIXBREIntegration(t *testing.T) {
	bre := &posix_bre.POSIXBRE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"identifier", "[[:alpha:]_][[:alnum:]_]*"},
		{"phone", `[0-9]\{3\}-[0-9]\{4\}`},
		{"date", `[0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}`},
		{"hex", "[[:xdigit:]]*"},
		{"word-repeat", `\([[:alpha:]]*\) \1`},
		{"literal-operators", "1+2=3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestJavaGoldenFiles tests Java patterns against golden file outputs
func TestJavaGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/java"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	javaFlavor := &java.Java{}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic patterns
		{"literal", "abc"},
		{"alternation", "a|b|c"},
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// Groups
		{"group-capture", "(abc)"},
		{"group-non-capture", "(?:abc)"},
		{"group-named", "(?<name>abc)"},
		{"group-atomic", "(?>abc)"},

		// Lookahead and lookbehind
		{"lookahead-positive", "(?=abc)"},
		{"lookahead-negative", "(?!abc)"},
		{"lookbehind-positive", "(?<=abc)"},
		{"lookbehind-negative", "(?<!abc)"},

		// Quantifiers
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-question", "a?"},
		{"quantifier-exact", "a{3}"},
		{"quantifier-range", "a{2,5}"},

		// Possessive quantifiers
		{"possessive-star", "a*+"},
		{"possessive-plus", "a++"},
		{"possessive-question", "a?+"},

		// Non-greedy quantifiers
		{"lazy-star", "a*?"},
		{"lazy-plus", "a+?"},

		// Java-specific escapes
		{"escape-horizontal-ws", `\h`},
		{"escape-vertical-ws", `\v`},
		{"escape-linebreak", `\R`},
		{"escape-grapheme", `\X`},
		{"escape-bell", `\a`},
		{"escape-char", `\e`},

		// Anchors
		{"anchor-line", "^start$"},
		{"anchor-input", `\Astart\z`},
		{"anchor-word", `\bword\b`},
		{"anchor-grapheme-boundary", `\b{g}`},
		{"anchor-grapheme-boundary-context", `\b{g}test\b{g}`},

		// Unicode properties
		{"unicode-letter", `\p{L}`},
		{"unicode-upper", `\p{Lu}`},
		{"unicode-negated", `\P{N}`},
		{"unicode-posix-lower", `\p{Lower}`},
		{"unicode-java", `\p{javaLowerCase}`},

		// Back-references
		{"backref-number", `(a)\1`},
		{"backref-named", `(?<n>a)\k<n>`},

		// Quoted literals
		{"quoted-literal", `\Q[a-z]+\E`},
		{"quoted-context", `foo\Q***\Ebar`},

		// Comments
		{"comment", `(?#this is a comment)`},
		{"comment-context", `foo(?#match foo)bar`},

		// Inline modifiers
		{"modifier-global", `(?i)abc`},
		{"modifier-scoped", `(?i:abc)`},
		{"modifier-enable-disable", `(?i-m)abc`},

		// Complex patterns
		{"complex-email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"complex-date", `(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := javaFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for %q: %v", tc.pattern, err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				err := os.WriteFile(goldenPath, []byte(svg), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if svg != string(expected) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}

// TestJavaIntegration tests complete Java rendering pipeline
func TestJavaIntegration(t *testing.T) {
	javaFlavor := &java.Java{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"date-named", `(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`},
		{"phone", `\d{3}-\d{3}-\d{4}`},
		{"url", `https?://[a-zA-Z0-9.-]+(?:/[a-zA-Z0-9./-]*)?`},
		{"quoted-special", `\Q$100.00\E`},
		{"atomic-greedy", `(?>a+)b`},
		{"possessive-pattern", `"[^"]*+"`},
		{"unicode-words", `\p{L}+`},
		{"modifier-case", `(?i:hello) world`},
		{"comment-pattern", `\d+(?#digits only)\.\d+`},
		{"grapheme-boundary", `\b{g}\w+\b{g}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := javaFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestDotNetGoldenFiles tests .NET patterns against golden file outputs
func TestDotNetGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/dotnet"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	dotnetFlavor := &dotnet.DotNet{}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic patterns
		{"literal", "abc"},
		{"alternation", "a|b|c"},
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// Groups
		{"group-capture", "(abc)"},
		{"group-non-capture", "(?:abc)"},
		{"group-named-angle", "(?<name>abc)"},
		{"group-named-quote", "(?'name'abc)"},
		{"group-atomic", "(?>abc)"},

		// Lookahead and lookbehind
		{"lookahead-positive", "(?=abc)"},
		{"lookahead-negative", "(?!abc)"},
		{"lookbehind-positive", "(?<=abc)"},
		{"lookbehind-negative", "(?<!abc)"},

		// Balanced groups (unique to .NET)
		{"balanced-capture", "(?<Close-Open>a)"},
		{"balanced-capture-quote", "(?'Close-Open'a)"},
		{"balanced-non-capture", "(?<-Open>a)"},
		{"balanced-non-capture-quote", "(?'-Open'a)"},
		{"balanced-parens", `\((?:[^()]|(?<O>\()|(?<-O>\)))*\)`},

		// Quantifiers
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-question", "a?"},
		{"quantifier-exact", "a{3}"},
		{"quantifier-range", "a{2,5}"},

		// Non-greedy quantifiers
		{"lazy-star", "a*?"},
		{"lazy-plus", "a+?"},

		// Escapes
		{"escape-digit", `\d`},
		{"escape-word", `\w`},
		{"escape-whitespace", `\s`},
		{"escape-vertical-tab", `\v`},
		{"escape-bell", `\a`},
		{"escape-char", `\e`},

		// Anchors
		{"anchor-line", "^start$"},
		{"anchor-input", `\Astart\z`},
		{"anchor-word", `\bword\b`},

		// Unicode properties
		{"unicode-letter", `\p{L}`},
		{"unicode-upper", `\p{Lu}`},
		{"unicode-negated", `\P{N}`},

		// Back-references (both syntaxes)
		{"backref-number", `(a)\1`},
		{"backref-named-angle", `(?<n>a)\k<n>`},
		{"backref-named-quote", `(?'n'a)\k'n'`},

		// Quoted literals
		{"quoted-literal", `\Q[a-z]+\E`},
		{"quoted-context", `foo\Q***\Ebar`},

		// Comments
		{"comment", `(?#this is a comment)`},
		{"comment-context", `foo(?#match foo)bar`},

		// Inline modifiers (.NET: i, m, s, n, x)
		{"modifier-global", `(?i)abc`},
		{"modifier-scoped", `(?i:abc)`},
		{"modifier-enable-disable", `(?i-m)abc`},
		{"modifier-explicit-capture", `(?n)abc`},

		// Unlimited lookbehind (unique to .NET)
		{"lookbehind-variable", `(?<=a+)b`},
		{"lookbehind-alternation", `(?<=ab|abc)x`},

		// Conditional patterns
		{"conditional-number", `(?(1)yes|no)`},
		{"conditional-name", `(?(name)yes|no)`},
		{"conditional-assertion", `(?(?=\d)yes|no)`},
		{"conditional-no-else", `(?(1)yes)`},
		{"conditional-balanced-idiom", `(?(Open)(?!))`},

		// Complex patterns
		{"complex-email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"complex-balanced", `(?<Open>\()(?<Close-Open>\))`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := dotnetFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for %q: %v", tc.pattern, err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				err := os.WriteFile(goldenPath, []byte(svg), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if svg != string(expected) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}

// TestDotNetIntegration tests complete .NET rendering pipeline
func TestDotNetIntegration(t *testing.T) {
	dotnetFlavor := &dotnet.DotNet{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"date-named", `(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`},
		{"phone", `\d{3}-\d{3}-\d{4}`},
		{"url", `https?://[a-zA-Z0-9.-]+(?:/[a-zA-Z0-9./-]*)?`},
		{"quoted-special", `\Q$100.00\E`},
		{"atomic-greedy", `(?>a+)b`},
		{"unicode-words", `\p{L}+`},
		{"modifier-case", `(?i:hello) world`},
		{"comment-pattern", `\d+(?#digits only)\.\d+`},
		// .NET unique features
		{"balanced-simple", `(?<Open>\()(?<Close-Open>\))`},
		{"named-quote-syntax", `(?'word'\w+)`},
		{"backref-quote-syntax", `(?'n'a)\k'n'`},
		{"unlimited-lookbehind", `(?<=\w+)x`},
		{"explicit-capture-mode", `(?n)(a)(?<named>b)`},
		// Conditional patterns
		{"conditional-number", `(?(1)yes|no)`},
		{"conditional-name", `(?(name)yes|no)`},
		{"conditional-assertion-lookahead", `(?(?=\d)yes|no)`},
		{"conditional-assertion-lookbehind", `(?(?<=\w)yes|no)`},
		{"conditional-no-else", `(?(1)yes)`},
		{"conditional-balanced-idiom", `(?(Open)(?!))`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := dotnetFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestPCREGoldenFiles tests PCRE patterns against golden file outputs
func TestPCREGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/pcre"

	// Create golden directory if it doesn't exist
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	pcreFlavor := &pcre.PCRE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic patterns
		{"literal", "abc"},
		{"alternation", "a|b|c"},
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// Groups
		{"group-capture", "(abc)"},
		{"group-non-capture", "(?:abc)"},
		{"group-named-perl", "(?<name>abc)"},
		{"group-named-perl-alt", "(?'name'abc)"},
		{"group-named-python", "(?P<name>abc)"},
		{"group-atomic", "(?>abc)"},
		{"group-atomic-alt", "(*atomic:abc)"},

		// Lookahead and lookbehind
		{"lookahead-positive", "(?=abc)"},
		{"lookahead-positive-alt", "(*pla:abc)"},
		{"lookahead-negative", "(?!abc)"},
		{"lookahead-negative-alt", "(*nla:abc)"},
		{"lookbehind-positive", "(?<=abc)"},
		{"lookbehind-positive-alt", "(*plb:abc)"},
		{"lookbehind-negative", "(?<!abc)"},
		{"lookbehind-negative-alt", "(*nlb:abc)"},

		// Recursive patterns
		{"recursive-whole", "(?R)"},
		{"recursive-zero", "(?0)"},
		{"recursive-number", "(?1)"},
		{"recursive-name-perl", "(?&name)"},
		{"recursive-name-python", "(?P>name)"},
		{"recursive-relative-fwd", "(?+1)"},
		{"recursive-relative-back", "(?-1)"},
		{"recursive-oniguruma", `\g<1>`},
		{"recursive-oniguruma-name", `\g<name>`},

		// Conditional patterns
		{"conditional-number", "(?(1)yes|no)"},
		{"conditional-number-no-else", "(?(1)yes)"},
		{"conditional-name", "(?(name)yes|no)"},
		{"conditional-name-angle", "(?(<name>)yes|no)"},
		{"conditional-name-quote", "(?('name')yes|no)"},
		{"conditional-recursion", "(?(R)yes|no)"},
		{"conditional-recursion-num", "(?(R1)yes|no)"},
		{"conditional-define", "(?(DEFINE)(?<digit>[0-9]))"},
		{"conditional-assertion", "(?(?=a)yes|no)"},

		// Branch reset
		{"branch-reset", "(?|a|b)"},
		{"branch-reset-groups", "(?|(red)|(green)|(blue))"},

		// Backtracking control
		{"backtrack-fail", "(*FAIL)"},
		{"backtrack-accept", "(*ACCEPT)"},
		{"backtrack-mark", "(*MARK:test)"},
		{"backtrack-commit", "(*COMMIT)"},
		{"backtrack-prune", "(*PRUNE)"},
		{"backtrack-skip", "(*SKIP)"},
		{"backtrack-skip-name", "(*SKIP:label)"},
		{"backtrack-then", "(*THEN)"},

		// Quantifiers
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-question", "a?"},
		{"quantifier-exact", "a{3}"},
		{"quantifier-range", "a{2,5}"},
		{"quantifier-zero-to-m", "a{,5}"},

		// Possessive quantifiers
		{"possessive-star", "a*+"},
		{"possessive-plus", "a++"},
		{"possessive-question", "a?+"},
		{"possessive-range", "a{2,5}+"},

		// Non-greedy quantifiers
		{"lazy-star", "a*?"},
		{"lazy-plus", "a+?"},

		// POSIX classes
		{"posix-alpha", "[[:alpha:]]"},
		{"posix-digit", "[[:digit:]]"},
		{"posix-alnum", "[[:alnum:]]"},
		{"posix-space", "[[:space:]]"},
		{"posix-word", "[[:word:]]"},
		{"posix-negated", "[[:^alpha:]]"},

		// Escapes
		{"escape-digit", `\d`},
		{"escape-word", `\w`},
		{"escape-whitespace", `\s`},
		{"escape-horizontal-ws", `\h`},
		{"escape-vertical-ws", `\v`},
		{"escape-linebreak", `\R`},
		{"escape-grapheme", `\X`},
		{"escape-non-newline", `\N`},
		{"escape-hex-ext", `\x{1F600}`},
		{"escape-octal-ext", `\o{101}`},

		// Anchors
		{"anchor-line", "^start$"},
		{"anchor-input", `\Astart\z`},
		{"anchor-word", `\bword\b`},
		{"anchor-reset-match", `foo\Kbar`},

		// Unicode properties
		{"unicode-letter", `\p{L}`},
		{"unicode-upper", `\p{Lu}`},
		{"unicode-negated", `\P{N}`},

		// Back-references
		{"backref-number", `(a)\1`},
		{"backref-named-k", `(?<n>a)\k<n>`},
		{"backref-named-k-alt", `(?'n'a)\k'n'`},
		{"backref-named-python", `(?P<n>a)(?P=n)`},
		{"backref-g", `(a)\g{1}`},

		// Quoted literals
		{"quoted-literal", `\Qhello\E`},
		{"quoted-context", `foo\Q***\Ebar`},

		// Comments
		{"comment", `(?#this is a comment)`},
		{"comment-context", `foo(?#match foo)bar`},

		// Inline modifiers
		{"modifier-global", `(?i)abc`},
		{"modifier-scoped", `(?i:abc)`},
		{"modifier-enable-disable", `(?i-m)abc`},

		// Non-atomic lookaround
		{"non-atomic-lookahead-short", "(?*abc)"},
		{"non-atomic-lookbehind-short", "(?<*abc)"},
		{"non-atomic-lookahead-long", "(*napla:abc)"},
		{"non-atomic-lookbehind-long", "(*naplb:abc)"},

		// Script runs
		{"script-run-full", "(*script_run:abc)"},
		{"script-run-short", "(*sr:abc)"},
		{"atomic-script-run-full", "(*atomic_script_run:abc)"},
		{"atomic-script-run-short", "(*asr:abc)"},

		// Pattern start options
		{"option-utf", "(*UTF)abc"},
		{"option-multiple", "(*UTF)(*UCP)abc"},
		{"option-limit", "(*LIMIT_MATCH=100)abc"},
		{"option-combined", "(*UTF)(*LIMIT_MATCH=100)(*CRLF)abc"},
		{"option-newline", "(*ANYCRLF)abc"},
		{"option-optimization", "(*NO_AUTO_POSSESS)abc"},

		// Callouts
		{"callout-default", "a(?C)b"},
		{"callout-number", "a(?C1)b"},
		{"callout-string-dq", `a(?C"test")b`},
		{"callout-string-brace", "a(?C{test})b"},
		{"callout-escaped-delimiter", `a(?C"say ""hi""")b`},

		// Complex patterns
		{"complex-balanced-parens", `\((?:[^()]|(?R))*\)`},
		{"complex-define-use", `(?(DEFINE)(?<d>[0-9]))(?&d)+`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := pcreFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for %q: %v", tc.pattern, err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				err := os.WriteFile(goldenPath, []byte(svg), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if svg != string(expected) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}

// TestPCREIntegration tests complete PCRE rendering pipeline
func TestPCREIntegration(t *testing.T) {
	pcreFlavor := &pcre.PCRE{}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"date-named", `(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})`},
		{"phone", `\d{3}-\d{3}-\d{4}`},
		{"url", `https?://[a-zA-Z0-9.-]+(?:/[a-zA-Z0-9./-]*)?`},
		{"quoted-special", `\Q$100.00\E`},
		{"atomic-greedy", `(?>a+)b`},
		{"possessive-pattern", `"[^"]*+"`},
		{"unicode-words", `\p{L}+`},
		{"modifier-case", `(?i:hello) world`},
		{"comment-pattern", `\d+(?#digits only)\.\d+`},
		// PCRE unique features
		{"recursive-parens", `\((?:[^()]|(?R))*\)`},
		{"conditional-simple", `(a)?(?(1)b|c)`},
		{"branch-reset-colors", `(?|(red)|(green)|(blue))`},
		{"backtrack-control", `a(*SKIP)b|c`},
		{"define-pattern", `(?(DEFINE)(?<d>[0-9]))(?&d)+`},
		// Non-atomic lookaround
		{"non-atomic-lookahead", `(?*foo)bar`},
		{"non-atomic-lookbehind", `(?<*foo)bar`},
		// Script runs
		{"script-run", `(*script_run:\d+)`},
		{"atomic-script-run", `(*asr:\w+)`},
		// Pattern start options
		{"option-utf-ucp", `(*UTF)(*UCP)\d+`},
		{"option-limit", `(*LIMIT_MATCH=100)\w+`},
		// Callouts
		{"callout-in-pattern", `\d+(?C1)\.\d+`},
		{"callout-string", `a(?C"check")b`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := pcreFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestGNUGrepBREGoldenFiles tests GNU grep BRE patterns against golden file outputs
func TestGNUGrepBREGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/gnugrep-bre"

	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	breFlavor, ok := flavor.Get("gnugrep-bre")
	if !ok {
		t.Fatal("gnugrep-bre flavor not registered")
	}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic literals - note + ? | ( ) { } are literal in BRE
		{"literal", "abc"},
		{"literal-special", "a+b?c"},

		// BRE groups with \( \)
		{"group", `\(abc\)`},
		{"group-nested", `\(\(a\)\(b\)\)`},

		// Character sets
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// POSIX classes
		{"posix-alpha", "[[:alpha:]]"},
		{"posix-digit", "[[:digit:]]"},
		{"posix-multiple", "[[:alpha:][:digit:]]"},
		{"posix-negated", "[^[:digit:]]"},

		// BRE quantifiers
		{"quantifier-star", "a*"},
		{"quantifier-exact", `a\{3\}`},
		{"quantifier-min", `a\{2,\}`},
		{"quantifier-range", `a\{2,5\}`},

		// GNU extension: quantifiers
		{"gnu-plus", `a\+`},
		{"gnu-question", `a\?`},
		{"gnu-at-most", `a\{,5\}`},

		// GNU extension: alternation
		{"gnu-alternation", `cat\|dog`},
		{"gnu-alternation-multiple", `one\|two\|three`},

		// GNU extensions: word boundaries
		{"word-start-end", `\<word\>`},
		{"word-boundary", `\bword\b`},
		{"non-word-boundary", `\Bword\B`},

		// GNU extensions: character class shorthands
		{"escape-word", `\w`},
		{"escape-non-word", `\W`},
		{"escape-whitespace", `\s`},
		{"escape-non-whitespace", `\S`},

		// Back-references
		{"backref", `\(a\)\1`},

		// Anchors
		{"anchor", "^start$"},

		// Complex GNU BRE patterns
		{"complex-word-match", `\<\w\+\>`},
		{"complex-email", `\w\+@\w\+\.\w\+`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := breFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for %q: %v", tc.pattern, err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				err := os.WriteFile(goldenPath, []byte(svg), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if svg != string(expected) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}

// TestGNUGrepBREIntegration tests complete GNU grep BRE rendering pipeline
func TestGNUGrepBREIntegration(t *testing.T) {
	breFlavor, ok := flavor.Get("gnugrep-bre")
	if !ok {
		t.Fatal("gnugrep-bre flavor not registered")
	}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"identifier", `[[:alpha:]_][[:alnum:]_]*`},
		{"phone", `[0-9]\{3\}-[0-9]\{4\}`},
		{"word-match", `\<\w\+\>`},
		{"email", `\w\+@\w\+\.\w\+`},
		{"alternation-groups", `\(foo\)\|\(bar\)\|\(baz\)`},
		{"literal-operators", "1+2=3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := breFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}

// TestGNUGrepAliasConsistency verifies that "gnugrep" and "gnugrep-bre" produce identical SVG output
func TestGNUGrepAliasConsistency(t *testing.T) {
	gnugrep, ok := flavor.Get("gnugrep")
	if !ok {
		t.Fatal("gnugrep flavor not registered")
	}
	gnugrepBRE, ok := flavor.Get("gnugrep-bre")
	if !ok {
		t.Fatal("gnugrep-bre flavor not registered")
	}

	patterns := []string{
		"abc",
		`\(hello\)\+`,
		`cat\|dog`,
		`\<\w\+\>`,
		`\bword\b`,
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			ast1, err := gnugrep.Parse(pattern)
			if err != nil {
				t.Fatalf("gnugrep parse error: %v", err)
			}

			ast2, err := gnugrepBRE.Parse(pattern)
			if err != nil {
				t.Fatalf("gnugrep-bre parse error: %v", err)
			}

			r := New(nil)
			svg1 := r.Render(ast1)
			svg2 := r.Render(ast2)

			if svg1 != svg2 {
				t.Error("gnugrep and gnugrep-bre produced different SVG output")
			}
		})
	}
}

// TestGNUGrepEREGoldenFiles tests GNU grep ERE patterns against golden file outputs
func TestGNUGrepEREGoldenFiles(t *testing.T) {
	goldenDir := "testdata/golden/gnugrep-ere"

	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("failed to create golden directory: %v", err)
	}

	ereFlavor, ok := flavor.Get("gnugrep-ere")
	if !ok {
		t.Fatal("gnugrep-ere flavor not registered")
	}

	testCases := []struct {
		name    string
		pattern string
	}{
		// Basic literals
		{"literal", "abc"},

		// ERE groups
		{"group", "(abc)"},
		{"group-nested", "((a)(b))"},

		// ERE alternation
		{"alternation", "a|b|c"},
		{"alternation-groups", "(foo)|(bar)"},

		// Character sets
		{"charset", "[a-z]"},
		{"charset-negated", "[^0-9]"},

		// POSIX classes
		{"posix-alpha", "[[:alpha:]]"},
		{"posix-digit", "[[:digit:]]"},
		{"posix-multiple", "[[:alpha:][:digit:]]"},
		{"posix-negated", "[^[:digit:]]"},

		// ERE quantifiers
		{"quantifier-star", "a*"},
		{"quantifier-plus", "a+"},
		{"quantifier-question", "a?"},
		{"quantifier-exact", "a{3}"},
		{"quantifier-min", "a{2,}"},
		{"quantifier-range", "a{2,5}"},

		// GNU extension: at most m
		{"gnu-at-most", "a{,5}"},

		// GNU extensions: word boundaries
		{"word-start-end", `\<word\>`},
		{"word-boundary", `\bword\b`},
		{"non-word-boundary", `\Bword\B`},

		// GNU extensions: character class shorthands
		{"escape-word", `\w`},
		{"escape-non-word", `\W`},
		{"escape-whitespace", `\s`},
		{"escape-non-whitespace", `\S`},

		// GNU extension: back-references in ERE
		{"backref", `(a)\1`},
		{"backref-multiple", `(a)(b)\1\2`},

		// Anchors
		{"anchor", "^start$"},

		// Complex GNU ERE patterns
		{"complex-word-match", `\<\w+\>`},
		{"complex-email", `\w+@\w+\.\w+`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ereFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error for %q: %v", tc.pattern, err)
			}

			r := New(nil)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, tc.name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				err := os.WriteFile(goldenPath, []byte(svg), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if svg != string(expected) {
				t.Errorf("SVG output differs from golden file %s", goldenPath)
				t.Logf("Run with GOLDEN_UPDATE=1 to update golden files")
			}
		})
	}
}

// TestGNUGrepEREIntegration tests complete GNU grep ERE rendering pipeline
func TestGNUGrepEREIntegration(t *testing.T) {
	ereFlavor, ok := flavor.Get("gnugrep-ere")
	if !ok {
		t.Fatal("gnugrep-ere flavor not registered")
	}

	testCases := []struct {
		name    string
		pattern string
	}{
		{"identifier", "[[:alpha:]_][[:alnum:]_]*"},
		{"phone", "[0-9]{3}-[0-9]{4}"},
		{"word-match", `\<\w+\>`},
		{"email", `\w+@\w+\.\w+`},
		{"url-like", "(https?://)?[a-z0-9.-]+"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ereFlavor.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			validateSVG(t, svg)
		})
	}
}
