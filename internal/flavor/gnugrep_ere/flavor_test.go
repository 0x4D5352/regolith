package gnugrep_ere

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestGNUGrepEREFlavorName(t *testing.T) {
	ere := &GNUGrepERE{}
	if ere.Name() != "gnugrep-ere" {
		t.Errorf("expected name 'gnugrep-ere', got '%s'", ere.Name())
	}
}

func TestGNUGrepEREFlavorDescription(t *testing.T) {
	ere := &GNUGrepERE{}
	desc := ere.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(desc, "GNU") {
		t.Error("description should mention GNU")
	}
	if !strings.Contains(desc, "ERE") || !strings.Contains(desc, "Extended") {
		t.Error("description should mention ERE or Extended")
	}
	if !strings.Contains(desc, "grep -E") {
		t.Error("description should mention 'grep -E'")
	}
}

func TestGNUGrepEREFlavorSupportedFlags(t *testing.T) {
	ere := &GNUGrepERE{}
	flags := ere.SupportedFlags()

	// GNU grep has no inline flags
	if len(flags) != 0 {
		t.Errorf("GNU grep ERE should have no inline flags, got %d", len(flags))
	}
}

func TestGNUGrepEREFlavorSupportedFeatures(t *testing.T) {
	ere := &GNUGrepERE{}
	features := ere.SupportedFeatures()

	// GNU ERE should support POSIX classes
	if !features.POSIXClasses {
		t.Error("GNU ERE should support POSIX classes")
	}

	// GNU ERE should NOT support these
	if features.Lookahead {
		t.Error("GNU ERE should not support lookahead")
	}
	if features.Lookbehind {
		t.Error("GNU ERE should not support lookbehind")
	}
	if features.NamedGroups {
		t.Error("GNU ERE should not support named groups")
	}
	if features.AtomicGroups {
		t.Error("GNU ERE should not support atomic groups")
	}
}

func TestGNUGrepEREFlavorRegistered(t *testing.T) {
	f, ok := flavor.Get("gnugrep-ere")
	if !ok {
		t.Fatal("gnugrep-ere flavor not registered")
	}
	if f.Name() != "gnugrep-ere" {
		t.Errorf("expected name 'gnugrep-ere', got '%s'", f.Name())
	}
}

func TestGNUGrepEREFlavorInList(t *testing.T) {
	list := flavor.List()
	found := false
	for _, name := range list {
		if name == "gnugrep-ere" {
			found = true
			break
		}
	}
	if !found {
		t.Error("gnugrep-ere flavor not found in List()")
	}
}

func TestGNUGrepEREParseValidPatterns(t *testing.T) {
	ere := &GNUGrepERE{}

	tests := []struct {
		name    string
		pattern string
	}{
		// Basic literals
		{"simple literal", "abc"},
		{"numbers", "123"},

		// ERE Groups with ()
		{"simple group", "(abc)"},
		{"nested groups", "((a)(b))"},

		// ERE Quantifiers (unescaped)
		{"star", "a*"},
		{"plus", "a+"},
		{"question", "a?"},
		{"exact count", "a{3}"},
		{"min count", "a{3,}"},
		{"range count", "a{3,5}"},

		// GNU extension: {,m} for "at most m"
		{"at most 5", "a{,5}"},

		// ERE Alternation (unescaped |)
		{"alternation", "cat|dog"},
		{"alternation multiple", "one|two|three"},
		{"alternation with groups", "(foo)|(bar)"},

		// GNU extension: word boundaries
		{"word boundary", `\bword\b`},
		{"non-word boundary", `\Brat\B`},
		{"word start", `\<hello`},
		{"word end", `hello\>`},
		{"word both", `\<hello\>`},

		// GNU extension: character class shorthands
		{"word char", `\w`},
		{"word chars", `\w+`},
		{"non-word char", `\W`},
		{"whitespace", `\s`},
		{"non-whitespace", `\S`},

		// GNU extension: back-references in ERE
		{"back-reference", `(word)\1`},
		{"multiple back-references", `(a)(b)\1\2`},

		// POSIX character classes
		{"posix alpha", "[[:alpha:]]"},
		{"posix digit", "[[:digit:]]"},
		{"posix alnum", "[[:alnum:]]"},

		// Anchors
		{"start anchor", "^abc"},
		{"end anchor", "abc$"},

		// Any character
		{"dot", "."},
		{"dot with star", ".*"},

		// Escaped metacharacters
		{"escaped dot", `\.`},
		{"escaped star", `\*`},
		{"escaped plus", `\+`},
		{"escaped question", `\?`},
		{"escaped pipe", `\|`},
		{"escaped parens", `\(\)`},

		// GNU extension: \] and \}
		{"escaped bracket", `\]`},
		{"escaped brace", `\}`},

		// Complex patterns
		{"word match", `\<\w+\>`},
		{"email-like", `\w+@\w+\.\w+`},
		{"url-like", "(https?://)?[a-z0-9.-]+"},
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

func TestGNUGrepEREAlternation(t *testing.T) {
	ere := &GNUGrepERE{}

	result, err := ere.Parse("cat|dog|bird")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 alternatives
	if len(result.Matches) != 3 {
		t.Errorf("expected 3 alternatives, got %d", len(result.Matches))
	}
}

func TestGNUGrepEREQuantifiers(t *testing.T) {
	ere := &GNUGrepERE{}

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
		{"a{,5}", 0, 5}, // GNU extension
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := ere.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			frag := result.Matches[0].Fragments[0]
			if frag.Repeat == nil {
				t.Fatal("expected Repeat")
			}
			if frag.Repeat.Min != tc.min {
				t.Errorf("expected Min=%d, got %d", tc.min, frag.Repeat.Min)
			}
			if frag.Repeat.Max != tc.max {
				t.Errorf("expected Max=%d, got %d", tc.max, frag.Repeat.Max)
			}
			// GNU ERE quantifiers are always greedy
			if !frag.Repeat.Greedy {
				t.Error("expected Greedy=true")
			}
		})
	}
}

func TestGNUGrepEREWordBoundaries(t *testing.T) {
	ere := &GNUGrepERE{}

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
			result, err := ere.Parse(tc.pattern)
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

func TestGNUGrepERECharacterClassShorthands(t *testing.T) {
	ere := &GNUGrepERE{}

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
			result, err := ere.Parse(tc.pattern)
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

func TestGNUGrepEREGroups(t *testing.T) {
	ere := &GNUGrepERE{}

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

func TestGNUGrepEREBackReferences(t *testing.T) {
	ere := &GNUGrepERE{}

	// GNU extension: back-references in ERE
	tests := []struct {
		pattern     string
		backRefNums []int
	}{
		{`(a)\1`, []int{1}},
		{`(a)(b)\1\2`, []int{1, 2}},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			result, err := ere.Parse(tc.pattern)
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

func TestGNUGrepEREPOSIXClasses(t *testing.T) {
	ere := &GNUGrepERE{}

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

func TestGNUGrepEREDifferentFromBRE(t *testing.T) {
	ere := &GNUGrepERE{}

	// In ERE, these metacharacters work WITHOUT backslash
	// This is a key difference from BRE

	t.Run("unescaped + is quantifier", func(t *testing.T) {
		result, err := ere.Parse("a+")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Error("expected + to be a quantifier")
		}
	})

	t.Run("unescaped ? is quantifier", func(t *testing.T) {
		result, err := ere.Parse("a?")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Error("expected ? to be a quantifier")
		}
	})

	t.Run("unescaped | is alternation", func(t *testing.T) {
		result, err := ere.Parse("a|b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Matches) != 2 {
			t.Errorf("expected 2 alternatives, got %d", len(result.Matches))
		}
	})

	t.Run("unescaped () is group", func(t *testing.T) {
		result, err := ere.Parse("(abc)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		_, ok := frag.Content.(*ast.Subexp)
		if !ok {
			t.Errorf("expected Subexp, got %T", frag.Content)
		}
	})

	t.Run("unescaped {} is interval", func(t *testing.T) {
		result, err := ere.Parse("a{2,3}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		if frag.Repeat == nil {
			t.Error("expected {} to be interval expression")
		}
		if frag.Repeat.Min != 2 || frag.Repeat.Max != 3 {
			t.Errorf("expected {2,3}, got {%d,%d}", frag.Repeat.Min, frag.Repeat.Max)
		}
	})
}
