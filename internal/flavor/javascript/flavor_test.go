package javascript

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestJavaScriptFlavorName(t *testing.T) {
	js := &JavaScript{}
	if js.Name() != "javascript" {
		t.Errorf("expected name 'javascript', got '%s'", js.Name())
	}
}

func TestJavaScriptFlavorDescription(t *testing.T) {
	js := &JavaScript{}
	desc := js.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
}

func TestJavaScriptFlavorParse(t *testing.T) {
	js := &JavaScript{}

	tests := []struct {
		pattern string
		wantErr bool
	}{
		{"abc", false},
		{"a|b|c", false},
		{"(abc)", false},
		{"(?:abc)", false},
		{"(?=abc)", false},
		{"(?!abc)", false},
		{"(?<=abc)", false},
		{"(?<!abc)", false},
		{"(?<name>abc)", false},
		{`\k<name>`, false},
		{`\p{Letter}`, false},
		{`\P{Number}`, false},
		{"/pattern/gi", false},
		{"[a-z]", false},
		{`\d+`, false},
		{`\u{1F600}`, false},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := js.Parse(tc.pattern)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if ast == nil {
					t.Error("expected non-nil AST")
				}
			}
		})
	}
}

func TestJavaScriptFlavorSupportedFlags(t *testing.T) {
	js := &JavaScript{}
	flags := js.SupportedFlags()

	if len(flags) == 0 {
		t.Fatal("expected at least one supported flag")
	}

	// Check for expected flags
	expectedFlags := map[rune]bool{
		'd': false, // hasIndices
		'g': false, // global
		'i': false, // ignoreCase
		'm': false, // multiline
		's': false, // dotAll
		'u': false, // unicode
		'y': false, // sticky
	}

	for _, f := range flags {
		if _, exists := expectedFlags[f.Char]; exists {
			expectedFlags[f.Char] = true
		}
		// Verify each flag has a name and description
		if f.Name == "" {
			t.Errorf("flag '%c' has empty name", f.Char)
		}
		if f.Description == "" {
			t.Errorf("flag '%c' has empty description", f.Char)
		}
	}

	// Check all expected flags are present
	for char, found := range expectedFlags {
		if !found {
			t.Errorf("expected flag '%c' not found in supported flags", char)
		}
	}
}

func TestJavaScriptFlavorSupportedFeatures(t *testing.T) {
	js := &JavaScript{}
	features := js.SupportedFeatures()

	// JavaScript should support these
	if !features.Lookahead {
		t.Error("JavaScript should support lookahead")
	}
	if !features.Lookbehind {
		t.Error("JavaScript should support lookbehind")
	}
	if !features.NamedGroups {
		t.Error("JavaScript should support named groups")
	}
	if !features.UnicodeProperties {
		t.Error("JavaScript should support Unicode properties")
	}

	// JavaScript should NOT support these
	if features.AtomicGroups {
		t.Error("JavaScript should not support atomic groups")
	}
	if features.PossessiveQuantifiers {
		t.Error("JavaScript should not support possessive quantifiers")
	}
	if features.RecursivePatterns {
		t.Error("JavaScript should not support recursive patterns")
	}
	if features.ConditionalPatterns {
		t.Error("JavaScript should not support conditional patterns")
	}
	if features.POSIXClasses {
		t.Error("JavaScript should not support POSIX classes")
	}
	if features.BalancedGroups {
		t.Error("JavaScript should not support balanced groups")
	}
}

func TestJavaScriptFlavorRegistered(t *testing.T) {
	// The JavaScript flavor should be registered via init()
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("JavaScript flavor not registered")
	}
	if f.Name() != "javascript" {
		t.Errorf("expected name 'javascript', got '%s'", f.Name())
	}
}

func TestJavaScriptFlavorInList(t *testing.T) {
	list := flavor.List()
	found := false
	for _, name := range list {
		if name == "javascript" {
			found = true
			break
		}
	}
	if !found {
		t.Error("JavaScript flavor not found in List()")
	}
}
