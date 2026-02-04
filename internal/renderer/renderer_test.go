package renderer

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/parser"
)

func TestRenderLiteral(t *testing.T) {
	ast, err := parser.ParseRegex("abc")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	// Check SVG contains expected elements
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG element")
	}
	if !strings.Contains(svg, `class="literal"`) {
		t.Error("expected literal class")
	}
	if !strings.Contains(svg, "abc") {
		t.Error("expected 'abc' text")
	}
}

func TestRenderAlternation(t *testing.T) {
	ast, err := parser.ParseRegex("a|b|c")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	// Check for alternation structure
	if !strings.Contains(svg, `class="regexp"`) {
		t.Error("expected regexp class for alternation")
	}

	// Should have three literals
	count := strings.Count(svg, `class="literal"`)
	if count != 3 {
		t.Errorf("expected 3 literal elements, got %d", count)
	}
}

func TestRenderCharset(t *testing.T) {
	ast, err := parser.ParseRegex("[abc]")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, `class="charset"`) {
		t.Error("expected charset class")
	}
	if !strings.Contains(svg, "One of:") {
		t.Error("expected 'One of:' label")
	}
}

func TestRenderNegatedCharset(t *testing.T) {
	ast, err := parser.ParseRegex("[^abc]")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, "None of:") {
		t.Error("expected 'None of:' label for negated charset")
	}
}

func TestRenderCharsetRange(t *testing.T) {
	ast, err := parser.ParseRegex("[a-z]")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	// HTML entities: &#34; is "
	if !strings.Contains(svg, `&#34;a&#34; - &#34;z&#34;`) {
		t.Error("expected range 'a' - 'z'")
	}
}

func TestRenderQuantifiers(t *testing.T) {
	tests := []struct {
		pattern  string
		hasLoop  bool
		hasSkip  bool
	}{
		{"a*", true, true},   // 0 or more
		{"a+", true, false},  // 1 or more
		{"a?", false, true},  // 0 or 1
		{"a{3}", true, false}, // exactly 3 (has loop for repeating, no skip since min=3)
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			hasLoop := strings.Contains(svg, `class="loop-path"`)
			hasSkip := strings.Contains(svg, `class="skip-path"`)

			if hasLoop != tc.hasLoop {
				t.Errorf("expected hasLoop=%v, got %v", tc.hasLoop, hasLoop)
			}
			if hasSkip != tc.hasSkip {
				t.Errorf("expected hasSkip=%v, got %v", tc.hasSkip, hasSkip)
			}
		})
	}
}

func TestRenderCaptureGroup(t *testing.T) {
	ast, err := parser.ParseRegex("(abc)")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, `class="subexp"`) {
		t.Error("expected subexp class")
	}
	if !strings.Contains(svg, "group #1") {
		t.Error("expected 'group #1' label")
	}
}

func TestRenderNonCaptureGroup(t *testing.T) {
	ast, err := parser.ParseRegex("(?:abc)")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, "non-capturing group") {
		t.Error("expected 'non-capturing group' label")
	}
}

func TestRenderLookahead(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{"(?=abc)", "positive lookahead"},
		{"(?!abc)", "negative lookahead"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label", tc.label)
			}
		})
	}
}

func TestRenderLookbehind(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{"(?<=abc)", "positive lookbehind"},
		{"(?<!abc)", "negative lookbehind"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label", tc.label)
			}
		})
	}
}

func TestRenderNamedCaptureGroup(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{"(?<username>\\w+)", "group #1 &#39;username&#39;"}, // HTML entities for quotes
		{"(?<year>\\d+)", "group #1 &#39;year&#39;"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label in SVG, got: %s", tc.label, svg)
			}
		})
	}
}

func TestRenderNamedBackReference(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{`\k<word>`, "back reference &#39;word&#39;"},
		{`\k<name>`, "back reference &#39;name&#39;"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label in SVG, got: %s", tc.label, svg)
			}
		})
	}
}

func TestRenderUnicodePropertyEscape(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{`\p{Letter}`, "Unicode Letter"},
		{`\p{L}`, "Unicode L"},
		{`\P{Number}`, "NOT Unicode Number"},
		{`\p{Script=Greek}`, "Unicode Script=Greek"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label in SVG, got: %s", tc.label, svg)
			}
		})
	}
}

func TestRenderNewFlags(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{"/abc/s", "dotAll"},
		{"/abc/d", "hasIndices"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label in SVG, got: %s", tc.label, svg)
			}
		})
	}
}

func TestRenderAnchors(t *testing.T) {
	ast, err := parser.ParseRegex("^abc$")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, "Start of line") {
		t.Error("expected 'Start of line' anchor")
	}
	if !strings.Contains(svg, "End of line") {
		t.Error("expected 'End of line' anchor")
	}
}

func TestRenderEscapes(t *testing.T) {
	tests := []struct {
		pattern string
		label   string
	}{
		{`\d`, "digit"},
		{`\D`, "non-digit"},
		{`\w`, "word"},
		{`\W`, "non-word"},
		{`\s`, "white space"},
		{`\S`, "non-white space"},
		{`\n`, "new line"},
		{`\t`, "tab"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := parser.ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			r := New(nil)
			svg := r.Render(ast)

			if !strings.Contains(svg, tc.label) {
				t.Errorf("expected '%s' label", tc.label)
			}
		})
	}
}

func TestRenderAnyCharacter(t *testing.T) {
	ast, err := parser.ParseRegex("a.b")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, "any character") {
		t.Error("expected 'any character' label")
	}
}

func TestRenderBackReference(t *testing.T) {
	ast, err := parser.ParseRegex(`(a)\1`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, "back reference #1") {
		t.Error("expected 'back reference #1' label")
	}
}

func TestRenderFlags(t *testing.T) {
	// Parse pattern with flags - need to set flags manually since grammar expects /pattern/flags format
	ast, err := parser.ParseRegex("abc")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Set flags manually for testing
	ast.Flags = "gi"

	r := New(nil)
	svg := r.Render(ast)

	if !strings.Contains(svg, `class="flags"`) {
		t.Error("expected flags class")
	}
	if !strings.Contains(svg, "Flags:") {
		t.Error("expected 'Flags:' label")
	}
	if !strings.Contains(svg, "global") {
		t.Error("expected 'global' flag")
	}
	if !strings.Contains(svg, "ignore case") {
		t.Error("expected 'ignore case' flag")
	}
}

func TestRenderAllFlags(t *testing.T) {
	ast, err := parser.ParseRegex("test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ast.Flags = "gimuy"

	r := New(nil)
	svg := r.Render(ast)

	expectedFlags := []string{"global", "ignore case", "multiline", "unicode", "sticky"}
	for _, flag := range expectedFlags {
		if !strings.Contains(svg, flag) {
			t.Errorf("expected '%s' flag", flag)
		}
	}
}

func TestCustomConfig(t *testing.T) {
	ast, err := parser.ParseRegex("abc")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	cfg := DefaultConfig()
	cfg.LiteralFill = "#ff0000"
	cfg.FontSize = 20

	r := New(cfg)
	svg := r.Render(ast)

	if !strings.Contains(svg, "#ff0000") {
		t.Error("expected custom literal fill color")
	}
	if !strings.Contains(svg, "font-size: 20px") {
		t.Error("expected custom font size")
	}
}

func TestSVGStructure(t *testing.T) {
	ast, err := parser.ParseRegex("a")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	// Check for required SVG structure
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("expected SVG to start with <svg")
	}
	if !strings.Contains(svg, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("expected SVG namespace")
	}
	if !strings.Contains(svg, "viewBox") {
		t.Error("expected viewBox attribute")
	}
	if !strings.Contains(svg, "<style>") {
		t.Error("expected style element")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("expected SVG to end with </svg>")
	}
}

func TestComplexPattern(t *testing.T) {
	// Test a more complex pattern
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	ast, err := parser.ParseRegex(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	r := New(nil)
	svg := r.Render(ast)

	// Just verify it renders without panicking and produces valid SVG
	if !strings.Contains(svg, "<svg") {
		t.Error("expected valid SVG output")
	}
}
