package posix_bre

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestPOSIXBREFlavorName(t *testing.T) {
	bre := &POSIXBRE{}
	if bre.Name() != "posix-bre" {
		t.Errorf("expected name 'posix-bre', got '%s'", bre.Name())
	}
}

func TestPOSIXBREFlavorDescription(t *testing.T) {
	bre := &POSIXBRE{}
	desc := bre.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(desc, "POSIX") {
		t.Error("description should mention POSIX")
	}
	if !strings.Contains(desc, "Basic") || !strings.Contains(desc, "\\(") {
		t.Error("description should mention Basic and escaped group syntax")
	}
}

func TestPOSIXBREFlavorSupportedFlags(t *testing.T) {
	bre := &POSIXBRE{}
	flags := bre.SupportedFlags()

	// POSIX BRE has no inline flags
	if len(flags) != 0 {
		t.Errorf("POSIX BRE should have no inline flags, got %d", len(flags))
	}
}

func TestPOSIXBREFlavorSupportedFeatures(t *testing.T) {
	bre := &POSIXBRE{}
	features := bre.SupportedFeatures()

	// POSIX BRE should support POSIX classes
	if !features.POSIXClasses {
		t.Error("POSIX BRE should support POSIX classes")
	}

	// POSIX BRE should NOT support these
	if features.Lookahead {
		t.Error("POSIX BRE should not support lookahead")
	}
	if features.Lookbehind {
		t.Error("POSIX BRE should not support lookbehind")
	}
	if features.NamedGroups {
		t.Error("POSIX BRE should not support named groups")
	}
	if features.AtomicGroups {
		t.Error("POSIX BRE should not support atomic groups")
	}
	if features.PossessiveQuantifiers {
		t.Error("POSIX BRE should not support possessive quantifiers")
	}
	if features.RecursivePatterns {
		t.Error("POSIX BRE should not support recursive patterns")
	}
	if features.ConditionalPatterns {
		t.Error("POSIX BRE should not support conditional patterns")
	}
	if features.UnicodeProperties {
		t.Error("POSIX BRE should not support Unicode properties")
	}
	if features.BalancedGroups {
		t.Error("POSIX BRE should not support balanced groups")
	}
	if features.InlineModifiers {
		t.Error("POSIX BRE should not support inline modifiers")
	}
	if features.Comments {
		t.Error("POSIX BRE should not support comments")
	}
}

func TestPOSIXBREFlavorRegistered(t *testing.T) {
	// The POSIX BRE flavor should be registered via init()
	f, ok := flavor.Get("posix-bre")
	if !ok {
		t.Fatal("POSIX BRE flavor not registered")
	}
	if f.Name() != "posix-bre" {
		t.Errorf("expected name 'posix-bre', got '%s'", f.Name())
	}
}

func TestPOSIXBREFlavorInList(t *testing.T) {
	list := flavor.List()
	found := false
	for _, name := range list {
		if name == "posix-bre" {
			found = true
			break
		}
	}
	if !found {
		t.Error("POSIX BRE flavor not found in List()")
	}
}

func TestPOSIXBREParseValidPatterns(t *testing.T) {
	bre := &POSIXBRE{}

	tests := []struct {
		name    string
		pattern string
	}{
		// Basic literals
		{"simple literal", "abc"},
		{"numbers", "123"},
		{"mixed", "abc123"},

		// Literal characters that are metacharacters in ERE
		{"literal plus", "a+b"},
		{"literal question", "a?b"},
		{"literal pipe", "a|b"},
		{"literal parentheses", "(abc)"},
		{"literal braces", "a{2,3}"},

		// BRE Groups with \( \)
		{"simple group", `\(abc\)`},
		{"nested groups", `\(\(a\)\(b\)\)`},
		{"group with content", `\(hello\)`},

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

		// Quantifiers (only * and \{n,m\} in BRE)
		{"star", "a*"},
		{"exact count", `a\{3\}`},
		{"min count", `a\{3,\}`},
		{"range count", `a\{3,5\}`},

		// Back-references (BRE supports these!)
		{"back-reference 1", `\(a\)\1`},
		{"back-reference 2", `\(a\)\(b\)\1\2`},
		{"back-reference word", `\(hello\) \1`},

		// Anchors
		{"start anchor", "^abc"},
		{"end anchor", "abc$"},
		{"both anchors", "^abc$"},

		// Any character
		{"dot", "."},
		{"dot with quantifier", ".*"},
		{"dot in pattern", "a.b"},

		// Escaped metacharacters (make them literal)
		{"escaped dot", `\.`},
		{"escaped star", `\*`},
		{"escaped bracket", `\[`},
		{"escaped caret", `\^`},
		{"escaped dollar", `\$`},
		{"escaped backslash", `\\`},

		// Complex patterns
		{"email-like", `[[:alnum:]]*@[[:alnum:]]*`},
		{"phone-like", `[0-9]\{3\}-[0-9]\{4\}`},
		{"word repeat", `\([[:alpha:]]*\) \1`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Errorf("unexpected error for pattern %q: %v", tc.pattern, err)
			}
			if result == nil {
				t.Errorf("expected non-nil AST for pattern %q", tc.pattern)
			}
		})
	}
}

func TestPOSIXBREParseErrors(t *testing.T) {
	bre := &POSIXBRE{}

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

		// GNU extensions should fail with helpful messages
		{"gnu plus", `a\+`, "GNU extension"},
		{"gnu question", `a\?`, "GNU extension"},
		{"gnu alternation", `a\|b`, "GNU extension"},

		// Common escape sequences should produce helpful errors
		{"newline escape", `\n`, "not a standard POSIX BRE"},
		{"tab escape", `\t`, "not a standard POSIX BRE"},
		{"carriage return", `\r`, "not a standard POSIX BRE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bre.Parse(tc.pattern)
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

func TestPOSIXBREParsePOSIXClasses(t *testing.T) {
	bre := &POSIXBRE{}

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
			result, err := bre.Parse(tc.pattern)
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

func TestPOSIXBREParseQuantifiers(t *testing.T) {
	bre := &POSIXBRE{}

	tests := []struct {
		pattern string
		min     int
		max     int
	}{
		{"a*", 0, -1},
		{`a\{3\}`, 3, 3},
		{`a\{3,\}`, 3, -1},
		{`a\{3,5\}`, 3, 5},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
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
			// POSIX BRE quantifiers are always greedy
			if !frag.Repeat.Greedy {
				t.Error("expected Greedy=true")
			}
		})
	}
}

func TestPOSIXBREParseGroups(t *testing.T) {
	bre := &POSIXBRE{}

	// Test that groups are numbered correctly
	result, err := bre.Parse(`\(a\)\(b\)\(c\)`)
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

func TestPOSIXBREParseBackReferences(t *testing.T) {
	bre := &POSIXBRE{}

	tests := []struct {
		pattern      string
		backRefNums  []int
	}{
		{`\(a\)\1`, []int{1}},
		{`\(a\)\(b\)\1\2`, []int{1, 2}},
		{`\(a\)\(b\)\(c\)\3\2\1`, []int{3, 2, 1}},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			match := result.Matches[0]
			var backRefs []int
			for _, frag := range match.Fragments {
				if br, ok := frag.Content.(*ast.BackReference); ok {
					backRefs = append(backRefs, br.Number)
				}
			}

			if len(backRefs) != len(tc.backRefNums) {
				t.Fatalf("expected %d back-references, got %d", len(tc.backRefNums), len(backRefs))
			}
			for i, expected := range tc.backRefNums {
				if backRefs[i] != expected {
					t.Errorf("back-reference %d: expected \\%d, got \\%d", i, expected, backRefs[i])
				}
			}
		})
	}
}

func TestPOSIXBRELiteralMetacharacters(t *testing.T) {
	bre := &POSIXBRE{}

	// In BRE, these are literal characters (not metacharacters)
	tests := []struct {
		pattern  string
		expected string
	}{
		{"a+b", "a+b"},           // + is literal
		{"a?b", "a?b"},           // ? is literal
		{"a|b", "a|b"},           // | is literal
		{"(abc)", "(abc)"},       // () are literal
		{"a{2}", "a{2}"},         // {} are literal
		{"a{2,3}", "a{2,3}"},     // {} are literal
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error for pattern %q: %v", tc.pattern, err)
			}

			// Should parse as a single match with a literal
			if len(result.Matches) != 1 {
				t.Fatalf("expected 1 match, got %d", len(result.Matches))
			}
			match := result.Matches[0]
			if len(match.Fragments) != 1 {
				t.Fatalf("expected 1 fragment, got %d", len(match.Fragments))
			}
			lit, ok := match.Fragments[0].Content.(*ast.Literal)
			if !ok {
				t.Fatalf("expected Literal, got %T", match.Fragments[0].Content)
			}
			if lit.Text != tc.expected {
				t.Errorf("expected literal %q, got %q", tc.expected, lit.Text)
			}
		})
	}
}

func TestPOSIXBRECharsetEdgeCases(t *testing.T) {
	bre := &POSIXBRE{}

	tests := []struct {
		name    string
		pattern string
	}{
		// Dash at start/end is literal
		{"dash at start", "[-a]"},
		{"dash at end", "[a-]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Errorf("unexpected error for pattern %q: %v", tc.pattern, err)
			}
		})
	}
}

func TestPOSIXBRENoAlternation(t *testing.T) {
	bre := &POSIXBRE{}

	// a|b should parse as literal "a|b", not as alternation
	result, err := bre.Parse("cat|dog")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be a single match, not two alternatives
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match (no alternation in BRE), got %d", len(result.Matches))
	}

	// The content should be a single literal "cat|dog"
	match := result.Matches[0]
	if len(match.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(match.Fragments))
	}
	lit, ok := match.Fragments[0].Content.(*ast.Literal)
	if !ok {
		t.Fatalf("expected Literal, got %T", match.Fragments[0].Content)
	}
	if lit.Text != "cat|dog" {
		t.Errorf("expected literal 'cat|dog', got %q", lit.Text)
	}
}
