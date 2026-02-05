package gnugrep_bre

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestGNUGrepBREFlavorNames(t *testing.T) {
	// Test "gnugrep" (default)
	gnugrep := &GNUGrepBRE{name: "gnugrep"}
	if gnugrep.Name() != "gnugrep" {
		t.Errorf("expected name 'gnugrep', got '%s'", gnugrep.Name())
	}

	// Test "gnugrep-bre" (explicit)
	gnugrepBRE := &GNUGrepBRE{name: "gnugrep-bre"}
	if gnugrepBRE.Name() != "gnugrep-bre" {
		t.Errorf("expected name 'gnugrep-bre', got '%s'", gnugrepBRE.Name())
	}
}

func TestGNUGrepBREFlavorDescriptions(t *testing.T) {
	// Test "gnugrep" description
	gnugrep := &GNUGrepBRE{name: "gnugrep"}
	desc := gnugrep.Description()
	if !strings.Contains(desc, "default") {
		t.Error("gnugrep description should mention 'default'")
	}
	if !strings.Contains(desc, "GNU") {
		t.Error("description should mention GNU")
	}

	// Test "gnugrep-bre" description
	gnugrepBRE := &GNUGrepBRE{name: "gnugrep-bre"}
	desc = gnugrepBRE.Description()
	if !strings.Contains(desc, "GNU") {
		t.Error("description should mention GNU")
	}
	if !strings.Contains(desc, "BRE") || !strings.Contains(desc, "Basic") {
		t.Error("description should mention BRE or Basic")
	}
}

func TestGNUGrepBREFlavorSupportedFlags(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}
	flags := bre.SupportedFlags()

	// GNU grep has no inline flags
	if len(flags) != 0 {
		t.Errorf("GNU grep BRE should have no inline flags, got %d", len(flags))
	}
}

func TestGNUGrepBREFlavorSupportedFeatures(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}
	features := bre.SupportedFeatures()

	// GNU BRE should support POSIX classes
	if !features.POSIXClasses {
		t.Error("GNU BRE should support POSIX classes")
	}

	// GNU BRE should NOT support these
	if features.Lookahead {
		t.Error("GNU BRE should not support lookahead")
	}
	if features.Lookbehind {
		t.Error("GNU BRE should not support lookbehind")
	}
	if features.NamedGroups {
		t.Error("GNU BRE should not support named groups")
	}
	if features.AtomicGroups {
		t.Error("GNU BRE should not support atomic groups")
	}
}

func TestGNUGrepBREFlavorsRegistered(t *testing.T) {
	// Both "gnugrep" and "gnugrep-bre" should be registered
	tests := []string{"gnugrep", "gnugrep-bre"}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			f, ok := flavor.Get(name)
			if !ok {
				t.Fatalf("%s flavor not registered", name)
			}
			if f.Name() != name {
				t.Errorf("expected name '%s', got '%s'", name, f.Name())
			}
		})
	}
}

func TestGNUGrepBREFlavorsInList(t *testing.T) {
	list := flavor.List()

	expected := []string{"gnugrep", "gnugrep-bre"}
	for _, name := range expected {
		found := false
		for _, n := range list {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s flavor not found in List()", name)
		}
	}
}

func TestGNUGrepBREParseValidPatterns(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	tests := []struct {
		name    string
		pattern string
	}{
		// Basic literals (same as POSIX BRE)
		{"simple literal", "abc"},
		{"numbers", "123"},

		// BRE Groups with \( \)
		{"simple group", `\(abc\)`},
		{"nested groups", `\(\(a\)\(b\)\)`},

		// Quantifiers (BRE style)
		{"star", "a*"},
		{"exact count", `a\{3\}`},
		{"min count", `a\{3,\}`},
		{"range count", `a\{3,5\}`},

		// GNU extension: \+ for one-or-more
		{"gnu plus", `a\+`},
		{"gnu plus word", `\(hello\)\+`},

		// GNU extension: \? for zero-or-one
		{"gnu question", `a\?`},
		{"gnu question word", `colou\?r`},

		// GNU extension: \| for alternation
		{"gnu alternation", `cat\|dog`},
		{"gnu alternation multiple", `one\|two\|three`},
		{"gnu alternation with groups", `\(foo\)\|\(bar\)`},

		// GNU extension: \{,m\} for "at most m"
		{"at most 5", `a\{,5\}`},

		// GNU extension: word boundaries
		{"word boundary", `\bword\b`},
		{"non-word boundary", `\Brat\B`},
		{"word start", `\<hello`},
		{"word end", `hello\>`},
		{"word both", `\<hello\>`},

		// GNU extension: character class shorthands
		{"word char", `\w`},
		{"word chars", `\w\+`},
		{"non-word char", `\W`},
		{"whitespace", `\s`},
		{"non-whitespace", `\S`},

		// POSIX character classes (inherited)
		{"posix alpha", "[[:alpha:]]"},
		{"posix digit", "[[:digit:]]"},
		{"posix alnum", "[[:alnum:]]"},

		// Back-references (inherited from BRE)
		{"back-reference", `\(word\)\1`},

		// Anchors
		{"start anchor", "^abc"},
		{"end anchor", "abc$"},

		// Any character
		{"dot", "."},
		{"dot with star", ".*"},

		// Escaped metacharacters
		{"escaped dot", `\.`},
		{"escaped star", `\*`},

		// GNU extension: \] and \}
		{"escaped bracket", `\]`},
		{"escaped brace", `\}`},

		// Complex patterns
		{"word match", `\<\w\+\>`},
		{"email-like", `\w\+@\w\+\.\w\+`},
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

func TestGNUGrepBREGNUExtensions(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	t.Run("alternation with \\|", func(t *testing.T) {
		result, err := bre.Parse(`cat\|dog`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have 2 alternatives
		if len(result.Matches) != 2 {
			t.Errorf("expected 2 alternatives, got %d", len(result.Matches))
		}
	})

	t.Run("\\+ quantifier", func(t *testing.T) {
		result, err := bre.Parse(`a\+`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Fatal("expected Repeat")
		}
		if frag.Repeat.Min != 1 || frag.Repeat.Max != -1 {
			t.Errorf("expected {1,-1}, got {%d,%d}", frag.Repeat.Min, frag.Repeat.Max)
		}
	})

	t.Run("\\? quantifier", func(t *testing.T) {
		result, err := bre.Parse(`a\?`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Fatal("expected Repeat")
		}
		if frag.Repeat.Min != 0 || frag.Repeat.Max != 1 {
			t.Errorf("expected {0,1}, got {%d,%d}", frag.Repeat.Min, frag.Repeat.Max)
		}
	})

	t.Run("\\{,m\\} quantifier", func(t *testing.T) {
		result, err := bre.Parse(`a\{,5\}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Fatal("expected Repeat")
		}
		if frag.Repeat.Min != 0 || frag.Repeat.Max != 5 {
			t.Errorf("expected {0,5}, got {%d,%d}", frag.Repeat.Min, frag.Repeat.Max)
		}
	})
}

func TestGNUGrepBREWordBoundaries(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	tests := []struct {
		pattern    string
		anchorType string
	}{
		{`\<`, "word_start"},
		{`\>`, "word_end"},
		{`\b`, "word_boundary"},
		{`\B`, "non_word_boundary"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			frag := result.Matches[0].Fragments[0]
			anchor, ok := frag.Content.(*ast.Anchor)
			if !ok {
				t.Fatalf("expected Anchor, got %T", frag.Content)
			}
			if anchor.AnchorType != tc.anchorType {
				t.Errorf("expected anchor type %q, got %q", tc.anchorType, anchor.AnchorType)
			}
		})
	}
}

func TestGNUGrepBRECharacterClassShorthands(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	tests := []struct {
		pattern    string
		escapeType string
		code       string
	}{
		{`\w`, "word", "w"},
		{`\W`, "non_word", "W"},
		{`\s`, "whitespace", "s"},
		{`\S`, "non_whitespace", "S"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			frag := result.Matches[0].Fragments[0]
			escape, ok := frag.Content.(*ast.Escape)
			if !ok {
				t.Fatalf("expected Escape, got %T", frag.Content)
			}
			if escape.EscapeType != tc.escapeType {
				t.Errorf("expected escape type %q, got %q", tc.escapeType, escape.EscapeType)
			}
			if escape.Code != tc.code {
				t.Errorf("expected code %q, got %q", tc.code, escape.Code)
			}
		})
	}
}

func TestGNUGrepBREGroups(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

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

func TestGNUGrepBREBackReferences(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	tests := []struct {
		pattern     string
		backRefNums []int
	}{
		{`\(a\)\1`, []int{1}},
		{`\(a\)\(b\)\1\2`, []int{1, 2}},
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

func TestGNUGrepBRELiteralMetacharacters(t *testing.T) {
	bre := &GNUGrepBRE{name: "gnugrep"}

	// In BRE, these are literal characters (not metacharacters)
	// Note: + and ? are special now due to GNU extensions \+ and \?
	// but unescaped + and ? are still literal
	tests := []struct {
		pattern  string
		expected string
	}{
		{"a+b", "a+b"},       // + without backslash is literal
		{"a?b", "a?b"},       // ? without backslash is literal
		{"(abc)", "(abc)"},   // () are literal
		{"a{2}", "a{2}"},     // {} without backslash are literal
		{"a{2,3}", "a{2,3}"}, // {} without backslash are literal
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := bre.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error for pattern %q: %v", tc.pattern, err)
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
