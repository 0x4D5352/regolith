package javascript

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
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
		'v': false, // unicodeSets
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

func TestJavaScriptVFlagParsing(t *testing.T) {
	js := &JavaScript{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"v flag only", "/abc/v", false},
		{"v flag with others", "/abc/giv", false},
		{"u and v together", "/abc/uv", true},
		{"v and u together", "/abc/vu", true},
		{"u flag only", "/abc/u", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := js.Parse(tc.pattern)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error for u+v combination, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestJavaScriptVModeSetOperations(t *testing.T) {
	js := &JavaScript{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Intersection
		{"intersection basic", `[\w&&\d]`, false},
		{"intersection nested", `[[a-z]&&[aeiou]]`, false},
		{"intersection chained", `[\w&&\d&&[0-5]]`, false},
		// Subtraction
		{"subtraction basic", `[\w--[0-9]]`, false},
		{"subtraction unicode", `[\p{Letter}--\p{Script=Greek}]`, false},
		// Nested
		{"nested charset", `[[a-z][A-Z]]`, false},
		// Classic in v-mode still works
		{"classic charset", `[a-z]`, false},
		{"classic with escapes", `[\d\w]`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := js.Parse(tc.pattern)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil AST")
				}
			}
		})
	}
}

func TestJavaScriptVModeStringDisjunction(t *testing.T) {
	js := &JavaScript{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"single string", `[\q{abc}]`, false},
		{"multiple strings", `[\q{abc|def|ghi}]`, false},
		{"empty alternative", `[\q{abc|}]`, false},
		{"single char strings", `[\q{a|b|c}]`, false},
		{"in intersection", `[\q{abc|def}&&\p{ASCII}]`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := js.Parse(tc.pattern)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil AST")
				}
			}
		})
	}
}

func TestJavaScriptUnicodeSetsFeature(t *testing.T) {
	js := &JavaScript{}
	features := js.SupportedFeatures()
	if !features.UnicodeSets {
		t.Error("JavaScript should support UnicodeSets")
	}
}

func TestJavaScriptVModeASTStructure(t *testing.T) {
	js := &JavaScript{}

	t.Run("intersection produces CharsetIntersection", func(t *testing.T) {
		result, err := js.Parse(`[\w&&\d]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset, ok := frag.Content.(*ast.Charset)
		if !ok {
			t.Fatalf("expected *ast.Charset, got %T", frag.Content)
		}
		if charset.SetExpression == nil {
			t.Fatal("expected non-nil SetExpression")
		}
		inter, ok := charset.SetExpression.(*ast.CharsetIntersection)
		if !ok {
			t.Fatalf("expected *ast.CharsetIntersection, got %T", charset.SetExpression)
		}
		if len(inter.Operands) != 2 {
			t.Errorf("expected 2 operands, got %d", len(inter.Operands))
		}
	})

	t.Run("subtraction produces CharsetSubtraction", func(t *testing.T) {
		result, err := js.Parse(`[\w--[0-9]]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset, ok := frag.Content.(*ast.Charset)
		if !ok {
			t.Fatalf("expected *ast.Charset, got %T", frag.Content)
		}
		if charset.SetExpression == nil {
			t.Fatal("expected non-nil SetExpression")
		}
		sub, ok := charset.SetExpression.(*ast.CharsetSubtraction)
		if !ok {
			t.Fatalf("expected *ast.CharsetSubtraction, got %T", charset.SetExpression)
		}
		if len(sub.Operands) != 2 {
			t.Errorf("expected 2 operands, got %d", len(sub.Operands))
		}
	})

	t.Run("chained intersection has 3 operands", func(t *testing.T) {
		result, err := js.Parse(`[\p{Letter}&&\p{ASCII}&&[a-z]]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset := frag.Content.(*ast.Charset)
		inter := charset.SetExpression.(*ast.CharsetIntersection)
		if len(inter.Operands) != 3 {
			t.Errorf("expected 3 operands, got %d", len(inter.Operands))
		}
	})

	t.Run("negated intersection", func(t *testing.T) {
		result, err := js.Parse(`[^\w&&\d]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset := frag.Content.(*ast.Charset)
		if !charset.Inverted {
			t.Error("expected Inverted to be true")
		}
		if charset.SetExpression == nil {
			t.Fatal("expected non-nil SetExpression")
		}
	})

	t.Run("string disjunction produces CharsetStringDisjunction", func(t *testing.T) {
		result, err := js.Parse(`[\q{abc|def}]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset := frag.Content.(*ast.Charset)
		if len(charset.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(charset.Items))
		}
		sd, ok := charset.Items[0].(*ast.CharsetStringDisjunction)
		if !ok {
			t.Fatalf("expected *ast.CharsetStringDisjunction, got %T", charset.Items[0])
		}
		if len(sd.Strings) != 2 {
			t.Errorf("expected 2 strings, got %d", len(sd.Strings))
		}
		if sd.Strings[0] != "abc" || sd.Strings[1] != "def" {
			t.Errorf("expected [abc, def], got %v", sd.Strings)
		}
	})

	t.Run("classic charset has nil SetExpression", func(t *testing.T) {
		result, err := js.Parse(`[a-z]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset := frag.Content.(*ast.Charset)
		if charset.SetExpression != nil {
			t.Error("expected nil SetExpression for classic charset")
		}
		if len(charset.Items) == 0 {
			t.Error("expected non-empty Items for classic charset")
		}
	})

	t.Run("nested charset in intersection", func(t *testing.T) {
		result, err := js.Parse(`[[a-z]&&[aeiou]]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		frag := result.Matches[0].Fragments[0]
		charset := frag.Content.(*ast.Charset)
		inter := charset.SetExpression.(*ast.CharsetIntersection)
		// Both operands should be Charsets
		for i, op := range inter.Operands {
			if _, ok := op.(*ast.Charset); !ok {
				t.Errorf("operand %d: expected *ast.Charset, got %T", i, op)
			}
		}
	})
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
