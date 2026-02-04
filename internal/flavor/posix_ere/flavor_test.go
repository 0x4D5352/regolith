package posix_ere

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestPOSIXEREFlavorName(t *testing.T) {
	ere := &POSIXERE{}
	if ere.Name() != "posix-ere" {
		t.Errorf("expected name 'posix-ere', got '%s'", ere.Name())
	}
}

func TestPOSIXEREFlavorDescription(t *testing.T) {
	ere := &POSIXERE{}
	desc := ere.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(desc, "POSIX") {
		t.Error("description should mention POSIX")
	}
}

func TestPOSIXEREFlavorSupportedFlags(t *testing.T) {
	ere := &POSIXERE{}
	flags := ere.SupportedFlags()

	// POSIX ERE has no inline flags
	if len(flags) != 0 {
		t.Errorf("POSIX ERE should have no inline flags, got %d", len(flags))
	}
}

func TestPOSIXEREFlavorSupportedFeatures(t *testing.T) {
	ere := &POSIXERE{}
	features := ere.SupportedFeatures()

	// POSIX ERE should support POSIX classes
	if !features.POSIXClasses {
		t.Error("POSIX ERE should support POSIX classes")
	}

	// POSIX ERE should NOT support these
	if features.Lookahead {
		t.Error("POSIX ERE should not support lookahead")
	}
	if features.Lookbehind {
		t.Error("POSIX ERE should not support lookbehind")
	}
	if features.NamedGroups {
		t.Error("POSIX ERE should not support named groups")
	}
	if features.AtomicGroups {
		t.Error("POSIX ERE should not support atomic groups")
	}
	if features.PossessiveQuantifiers {
		t.Error("POSIX ERE should not support possessive quantifiers")
	}
	if features.RecursivePatterns {
		t.Error("POSIX ERE should not support recursive patterns")
	}
	if features.ConditionalPatterns {
		t.Error("POSIX ERE should not support conditional patterns")
	}
	if features.UnicodeProperties {
		t.Error("POSIX ERE should not support Unicode properties")
	}
	if features.BalancedGroups {
		t.Error("POSIX ERE should not support balanced groups")
	}
	if features.InlineModifiers {
		t.Error("POSIX ERE should not support inline modifiers")
	}
	if features.Comments {
		t.Error("POSIX ERE should not support comments")
	}
}

func TestPOSIXEREFlavorRegistered(t *testing.T) {
	// The POSIX ERE flavor should be registered via init()
	f, ok := flavor.Get("posix-ere")
	if !ok {
		t.Fatal("POSIX ERE flavor not registered")
	}
	if f.Name() != "posix-ere" {
		t.Errorf("expected name 'posix-ere', got '%s'", f.Name())
	}
}

func TestPOSIXEREFlavorInList(t *testing.T) {
	list := flavor.List()
	found := false
	for _, name := range list {
		if name == "posix-ere" {
			found = true
			break
		}
	}
	if !found {
		t.Error("POSIX ERE flavor not found in List()")
	}
}

func TestPOSIXEREParseValidPatterns(t *testing.T) {
	ere := &POSIXERE{}

	tests := []struct {
		name    string
		pattern string
	}{
		// Basic literals
		{"simple literal", "abc"},
		{"numbers", "123"},
		{"mixed", "abc123"},

		// Alternation
		{"alternation", "a|b|c"},
		{"alternation with literals", "cat|dog|bird"},

		// Groups (capturing only in ERE)
		{"simple group", "(abc)"},
		{"nested groups", "((a)(b))"},
		{"group with alternation", "(a|b)"},

		// Character classes
		{"simple charset", "[abc]"},
		{"negated charset", "[^abc]"},
		{"range charset", "[a-z]"},
		{"multiple ranges", "[a-zA-Z0-9]"},
		{"mixed charset", "[a-z0-9_]"},

		// POSIX character classes
		{"posix alpha", "[[:alpha:]]"},
		{"posix digit", "[[:digit:]]"},
		{"posix alnum", "[[:alnum:]]"},
		{"posix space", "[[:space:]]"},
		{"posix upper", "[[:upper:]]"},
		{"posix lower", "[[:lower:]]"},
		{"posix punct", "[[:punct:]]"},
		{"posix xdigit", "[[:xdigit:]]"},
		{"posix print", "[[:print:]]"},
		{"posix graph", "[[:graph:]]"},
		{"posix cntrl", "[[:cntrl:]]"},
		{"posix blank", "[[:blank:]]"},
		{"multiple posix classes", "[[:alpha:][:digit:]]"},
		{"negated posix class", "[^[:digit:]]"},
		{"mixed posix and range", "[[:alpha:]0-9]"},

		// Quantifiers
		{"star", "a*"},
		{"plus", "a+"},
		{"question", "a?"},
		{"exact count", "a{3}"},
		{"min count", "a{3,}"},
		{"range count", "a{3,5}"},

		// Anchors
		{"start anchor", "^abc"},
		{"end anchor", "abc$"},
		{"both anchors", "^abc$"},

		// Any character
		{"dot", "."},
		{"dot with quantifier", ".*"},
		{"dot in pattern", "a.b"},

		// Escaped metacharacters
		{"escaped dot", `\.`},
		{"escaped star", `\*`},
		{"escaped plus", `\+`},
		{"escaped question", `\?`},
		{"escaped paren", `\(`},
		{"escaped bracket", `\[`},
		{"escaped pipe", `\|`},
		{"escaped caret", `\^`},
		{"escaped dollar", `\$`},
		{"escaped backslash", `\\`},

		// Complex patterns
		{"email-like", "[[:alnum:]]+@[[:alnum:]]+"},
		{"phone-like", "[0-9]{3}-[0-9]{4}"},
		{"url-like", "(http|https)://[[:alnum:]]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Errorf("unexpected error for pattern %q: %v", tc.pattern, err)
			}
			if result == nil {
				t.Errorf("expected non-nil AST for pattern %q", tc.pattern)
			}
		})
	}
}

func TestPOSIXEREParseErrors(t *testing.T) {
	ere := &POSIXERE{}

	tests := []struct {
		name        string
		pattern     string
		errContains string
	}{
		// JavaScript-style escapes should fail
		{"digit escape", `\d`, "not supported"},
		{"non-digit escape", `\D`, "not supported"},
		{"word escape", `\w`, "not supported"},
		{"non-word escape", `\W`, "not supported"},
		{"whitespace escape", `\s`, "not supported"},
		{"non-whitespace escape", `\S`, "not supported"},

		// Word boundaries should fail
		{"word boundary", `\b`, "not supported"},
		{"non-word boundary", `\B`, "not supported"},

		// Back-references should fail
		{"back-reference 1", `(a)\1`, "not supported"},
		{"back-reference 2", `(a)(b)\2`, "not supported"},

		// Common escape sequences should produce helpful errors
		{"newline escape", `\n`, "not a standard POSIX ERE"},
		{"tab escape", `\t`, "not a standard POSIX ERE"},
		{"carriage return", `\r`, "not a standard POSIX ERE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ere.Parse(tc.pattern)
			if err == nil {
				t.Errorf("expected error for pattern %q, got nil", tc.pattern)
				return
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("error for pattern %q should contain %q, got: %v",
					tc.pattern, tc.errContains, err)
			}
		})
	}
}

func TestPOSIXEREParsePOSIXClasses(t *testing.T) {
	ere := &POSIXERE{}

	tests := []struct {
		pattern   string
		className string
		negated   bool
	}{
		{"[[:alpha:]]", "alpha", false},
		{"[[:digit:]]", "digit", false},
		{"[[:alnum:]]", "alnum", false},
		{"[[:space:]]", "space", false},
		{"[[:^alpha:]]", "alpha", true},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Navigate to the POSIX class
			if len(result.Matches) != 1 {
				t.Fatalf("expected 1 match, got %d", len(result.Matches))
			}
			match := result.Matches[0]
			if len(match.Fragments) != 1 {
				t.Fatalf("expected 1 fragment, got %d", len(match.Fragments))
			}
			charset, ok := match.Fragments[0].Content.(*ast.Charset)
			if !ok {
				t.Fatalf("expected Charset, got %T", match.Fragments[0].Content)
			}
			if len(charset.Items) != 1 {
				t.Fatalf("expected 1 charset item, got %d", len(charset.Items))
			}
			posixClass, ok := charset.Items[0].(*ast.POSIXClass)
			if !ok {
				t.Fatalf("expected POSIXClass, got %T", charset.Items[0])
			}
			if posixClass.Name != tc.className {
				t.Errorf("expected class name %q, got %q", tc.className, posixClass.Name)
			}
			if posixClass.Negated != tc.negated {
				t.Errorf("expected negated=%v, got %v", tc.negated, posixClass.Negated)
			}
		})
	}
}

func TestPOSIXEREParseQuantifiers(t *testing.T) {
	ere := &POSIXERE{}

	tests := []struct {
		pattern string
		min     int
		max     int
	}{
		{"a*", 0, -1},
		{"a+", 1, -1},
		{"a?", 0, 1},
		{"a{3}", 3, 3},
		{"a{3,}", 3, -1},
		{"a{3,5}", 3, 5},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			match := result.Matches[0]
			frag := match.Fragments[0]
			if frag.Repeat == nil {
				t.Fatal("expected Repeat, got nil")
			}
			if frag.Repeat.Min != tc.min {
				t.Errorf("expected Min=%d, got %d", tc.min, frag.Repeat.Min)
			}
			if frag.Repeat.Max != tc.max {
				t.Errorf("expected Max=%d, got %d", tc.max, frag.Repeat.Max)
			}
			// POSIX ERE quantifiers are always greedy
			if !frag.Repeat.Greedy {
				t.Error("expected Greedy=true")
			}
		})
	}
}

func TestPOSIXEREParseGroups(t *testing.T) {
	ere := &POSIXERE{}

	// Test that groups are numbered correctly
	result, err := ere.Parse("(a)(b)(c)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	match := result.Matches[0]
	for i, frag := range match.Fragments {
		subexp, ok := frag.Content.(*ast.Subexp)
		if !ok {
			t.Fatalf("fragment %d: expected Subexp, got %T", i, frag.Content)
		}
		expectedNum := i + 1
		if subexp.Number != expectedNum {
			t.Errorf("fragment %d: expected group number %d, got %d",
				i, expectedNum, subexp.Number)
		}
		if subexp.GroupType != "capture" {
			t.Errorf("fragment %d: expected group type 'capture', got %q",
				i, subexp.GroupType)
		}
	}
}

func TestPOSIXERENoNonCapturingGroups(t *testing.T) {
	ere := &POSIXERE{}

	// (?:...) syntax should not be recognized as non-capturing group
	// In POSIX ERE, (?: would be parsed differently
	result, err := ere.Parse("(?:abc)")
	if err != nil {
		// An error is acceptable - it means the pattern is not valid
		return
	}

	// If it parsed, check what we got
	if result != nil {
		match := result.Matches[0]
		if len(match.Fragments) > 0 {
			// Check it's not a non-capturing group
			if subexp, ok := match.Fragments[0].Content.(*ast.Subexp); ok {
				if subexp.GroupType == "non_capture" {
					t.Error("POSIX ERE should not support non-capturing groups")
				}
			}
		}
	}
}

func TestPOSIXERECharsetEdgeCases(t *testing.T) {
	ere := &POSIXERE{}

	tests := []struct {
		name    string
		pattern string
	}{
		// Dash at start/end is literal
		{"dash at start", "[-a]"},
		{"dash at end", "[a-]"},
		// Note: ] at start of charset (e.g., []a]) is valid POSIX but not currently supported
		// This is a known limitation that would require special grammar handling
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Errorf("unexpected error for pattern %q: %v", tc.pattern, err)
			}
		})
	}
}
