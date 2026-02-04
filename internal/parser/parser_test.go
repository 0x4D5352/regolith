package parser

import (
	"testing"
)

func TestParseLiteral(t *testing.T) {
	ast, err := ParseRegex("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	if len(ast.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(ast.Matches))
	}

	match := ast.Matches[0]
	if len(match.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(match.Fragments))
	}

	lit, ok := match.Fragments[0].Content.(*Literal)
	if !ok {
		t.Fatalf("expected Literal, got %T", match.Fragments[0].Content)
	}

	if lit.Text != "abc" {
		t.Errorf("expected 'abc', got '%s'", lit.Text)
	}
}

func TestParseAlternation(t *testing.T) {
	ast, err := ParseRegex("a|b|c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.Matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(ast.Matches))
	}

	// Check each alternative
	expected := []string{"a", "b", "c"}
	for i, m := range ast.Matches {
		if len(m.Fragments) != 1 {
			t.Errorf("match %d: expected 1 fragment, got %d", i, len(m.Fragments))
			continue
		}
		lit, ok := m.Fragments[0].Content.(*Literal)
		if !ok {
			t.Errorf("match %d: expected Literal, got %T", i, m.Fragments[0].Content)
			continue
		}
		if lit.Text != expected[i] {
			t.Errorf("match %d: expected '%s', got '%s'", i, expected[i], lit.Text)
		}
	}
}

func TestParseCaptureGroup(t *testing.T) {
	ast, err := ParseRegex("(abc)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(ast.Matches))
	}

	match := ast.Matches[0]
	if len(match.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(match.Fragments))
	}

	subexp, ok := match.Fragments[0].Content.(*Subexp)
	if !ok {
		t.Fatalf("expected Subexp, got %T", match.Fragments[0].Content)
	}

	if subexp.GroupType != "capture" {
		t.Errorf("expected 'capture', got '%s'", subexp.GroupType)
	}

	if subexp.Number != 1 {
		t.Errorf("expected group number 1, got %d", subexp.Number)
	}
}

func TestParseNonCaptureGroup(t *testing.T) {
	ast, err := ParseRegex("(?:abc)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	match := ast.Matches[0]
	subexp, ok := match.Fragments[0].Content.(*Subexp)
	if !ok {
		t.Fatalf("expected Subexp, got %T", match.Fragments[0].Content)
	}

	if subexp.GroupType != "non_capture" {
		t.Errorf("expected 'non_capture', got '%s'", subexp.GroupType)
	}

	if subexp.Number != 0 {
		t.Errorf("expected group number 0, got %d", subexp.Number)
	}
}

func TestParseQuantifiers(t *testing.T) {
	tests := []struct {
		pattern string
		min     int
		max     int
		greedy  bool
	}{
		{"a*", 0, -1, true},
		{"a+", 1, -1, true},
		{"a?", 0, 1, true},
		{"a*?", 0, -1, false},
		{"a+?", 1, -1, false},
		{"a??", 0, 1, false},
		{"a{3}", 3, 3, true},
		{"a{2,5}", 2, 5, true},
		{"a{2,}", 2, -1, true},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			if frag.Repeat == nil {
				t.Fatal("expected Repeat, got nil")
			}

			if frag.Repeat.Min != tc.min {
				t.Errorf("expected min %d, got %d", tc.min, frag.Repeat.Min)
			}
			if frag.Repeat.Max != tc.max {
				t.Errorf("expected max %d, got %d", tc.max, frag.Repeat.Max)
			}
			if frag.Repeat.Greedy != tc.greedy {
				t.Errorf("expected greedy %v, got %v", tc.greedy, frag.Repeat.Greedy)
			}
		})
	}
}

func TestParseCharset(t *testing.T) {
	ast, err := ParseRegex("[abc]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frag := ast.Matches[0].Fragments[0]
	charset, ok := frag.Content.(*Charset)
	if !ok {
		t.Fatalf("expected Charset, got %T", frag.Content)
	}

	if charset.Inverted {
		t.Error("expected non-inverted charset")
	}

	if len(charset.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(charset.Items))
	}
}

func TestParseNegatedCharset(t *testing.T) {
	ast, err := ParseRegex("[^abc]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frag := ast.Matches[0].Fragments[0]
	charset, ok := frag.Content.(*Charset)
	if !ok {
		t.Fatalf("expected Charset, got %T", frag.Content)
	}

	if !charset.Inverted {
		t.Error("expected inverted charset")
	}
}

func TestParseCharsetRange(t *testing.T) {
	ast, err := ParseRegex("[a-z]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frag := ast.Matches[0].Fragments[0]
	charset, ok := frag.Content.(*Charset)
	if !ok {
		t.Fatalf("expected Charset, got %T", frag.Content)
	}

	if len(charset.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(charset.Items))
	}

	rng, ok := charset.Items[0].(*CharsetRange)
	if !ok {
		t.Fatalf("expected CharsetRange, got %T", charset.Items[0])
	}

	if rng.First != "a" || rng.Last != "z" {
		t.Errorf("expected 'a'-'z', got '%s'-'%s'", rng.First, rng.Last)
	}
}

func TestParseAnyCharacter(t *testing.T) {
	ast, err := ParseRegex("a.b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frags := ast.Matches[0].Fragments
	if len(frags) != 3 {
		t.Fatalf("expected 3 fragments, got %d", len(frags))
	}

	_, ok := frags[1].Content.(*AnyCharacter)
	if !ok {
		t.Fatalf("expected AnyCharacter, got %T", frags[1].Content)
	}
}

func TestParseAnchors(t *testing.T) {
	ast, err := ParseRegex("^abc$")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frags := ast.Matches[0].Fragments
	if len(frags) != 3 {
		t.Fatalf("expected 3 fragments, got %d", len(frags))
	}

	startAnchor, ok := frags[0].Content.(*Anchor)
	if !ok {
		t.Fatalf("expected Anchor, got %T", frags[0].Content)
	}
	if startAnchor.AnchorType != "start" {
		t.Errorf("expected 'start', got '%s'", startAnchor.AnchorType)
	}

	endAnchor, ok := frags[2].Content.(*Anchor)
	if !ok {
		t.Fatalf("expected Anchor, got %T", frags[2].Content)
	}
	if endAnchor.AnchorType != "end" {
		t.Errorf("expected 'end', got '%s'", endAnchor.AnchorType)
	}
}

func TestParseEscapes(t *testing.T) {
	tests := []struct {
		pattern    string
		escapeType string
	}{
		{`\d`, "digit"},
		{`\D`, "non_digit"},
		{`\w`, "word"},
		{`\W`, "non_word"},
		{`\s`, "whitespace"},
		{`\S`, "non_whitespace"},
		{`\n`, "newline"},
		{`\t`, "tab"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			escape, ok := frag.Content.(*Escape)
			if !ok {
				t.Fatalf("expected Escape, got %T", frag.Content)
			}

			if escape.EscapeType != tc.escapeType {
				t.Errorf("expected '%s', got '%s'", tc.escapeType, escape.EscapeType)
			}
		})
	}
}

func TestParseLookbehind(t *testing.T) {
	tests := []struct {
		pattern   string
		groupType string
	}{
		{"(?<=abc)", "positive_lookbehind"},
		{"(?<!abc)", "negative_lookbehind"},
		{"(?<=a+)b", "positive_lookbehind"},
		{"(?<!\\d)x", "negative_lookbehind"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			subexp, ok := frag.Content.(*Subexp)
			if !ok {
				t.Fatalf("expected Subexp, got %T", frag.Content)
			}

			if subexp.GroupType != tc.groupType {
				t.Errorf("expected '%s', got '%s'", tc.groupType, subexp.GroupType)
			}

			// Lookbehind groups should not be numbered
			if subexp.Number != 0 {
				t.Errorf("expected group number 0 for lookbehind, got %d", subexp.Number)
			}
		})
	}
}

func TestParseNamedCaptureGroup(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		number  int
	}{
		{"(?<name>abc)", "name", 1},
		{"(?<_private>xyz)", "_private", 1},
		{"(?<Name123>def)", "Name123", 1},
		{"(?<a>x)(?<b>y)", "a", 1}, // First group
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			subexp, ok := frag.Content.(*Subexp)
			if !ok {
				t.Fatalf("expected Subexp, got %T", frag.Content)
			}

			if subexp.GroupType != "named_capture" {
				t.Errorf("expected 'named_capture', got '%s'", subexp.GroupType)
			}

			if subexp.Name != tc.name {
				t.Errorf("expected name '%s', got '%s'", tc.name, subexp.Name)
			}

			if subexp.Number != tc.number {
				t.Errorf("expected group number %d, got %d", tc.number, subexp.Number)
			}
		})
	}
}

func TestParseMultipleNamedGroups(t *testing.T) {
	ast, err := ParseRegex("(?<year>\\d+)-(?<month>\\d+)-(?<day>\\d+)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 5 fragments: year group, -, month group, -, day group
	frags := ast.Matches[0].Fragments
	if len(frags) != 5 {
		t.Fatalf("expected 5 fragments, got %d", len(frags))
	}

	// Check year group (first)
	yearGroup := frags[0].Content.(*Subexp)
	if yearGroup.Name != "year" || yearGroup.Number != 1 {
		t.Errorf("expected year group #1, got name='%s' number=%d", yearGroup.Name, yearGroup.Number)
	}

	// Check month group (third)
	monthGroup := frags[2].Content.(*Subexp)
	if monthGroup.Name != "month" || monthGroup.Number != 2 {
		t.Errorf("expected month group #2, got name='%s' number=%d", monthGroup.Name, monthGroup.Number)
	}

	// Check day group (fifth)
	dayGroup := frags[4].Content.(*Subexp)
	if dayGroup.Name != "day" || dayGroup.Number != 3 {
		t.Errorf("expected day group #3, got name='%s' number=%d", dayGroup.Name, dayGroup.Number)
	}
}

func TestParseNamedBackReference(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
	}{
		{`\k<word>`, "word"},
		{`\k<name>`, "name"},
		{`\k<_test>`, "_test"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			br, ok := frag.Content.(*BackReference)
			if !ok {
				t.Fatalf("expected BackReference, got %T", frag.Content)
			}

			if br.Name != tc.name {
				t.Errorf("expected name '%s', got '%s'", tc.name, br.Name)
			}

			if br.Number != 0 {
				t.Errorf("expected number 0 for named ref, got %d", br.Number)
			}
		})
	}
}

func TestParseNamedGroupWithBackReference(t *testing.T) {
	// Pattern: (?<word>\w+)\s+\k<word> - match repeated words
	ast, err := ParseRegex(`(?<word>\w+)\s+\k<word>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frags := ast.Matches[0].Fragments
	if len(frags) != 3 {
		t.Fatalf("expected 3 fragments, got %d", len(frags))
	}

	// First is named group
	group := frags[0].Content.(*Subexp)
	if group.Name != "word" {
		t.Errorf("expected group name 'word', got '%s'", group.Name)
	}

	// Second is \s+
	_, ok := frags[1].Content.(*Escape)
	if !ok {
		t.Errorf("expected Escape for \\s, got %T", frags[1].Content)
	}

	// Third is named backreference
	br := frags[2].Content.(*BackReference)
	if br.Name != "word" {
		t.Errorf("expected backref name 'word', got '%s'", br.Name)
	}
}

func TestParseUnicodePropertyEscape(t *testing.T) {
	tests := []struct {
		pattern  string
		property string
		negated  bool
	}{
		{`\p{Letter}`, "Letter", false},
		{`\p{L}`, "L", false},
		{`\p{Ll}`, "Ll", false},
		{`\P{Number}`, "Number", true},
		{`\P{N}`, "N", true},
		{`\p{Script=Greek}`, "Script=Greek", false},
		{`\p{Emoji}`, "Emoji", false},
		{`\P{ASCII}`, "ASCII", true},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			upe, ok := frag.Content.(*UnicodePropertyEscape)
			if !ok {
				t.Fatalf("expected UnicodePropertyEscape, got %T", frag.Content)
			}

			if upe.Property != tc.property {
				t.Errorf("expected property '%s', got '%s'", tc.property, upe.Property)
			}

			if upe.Negated != tc.negated {
				t.Errorf("expected negated=%v, got %v", tc.negated, upe.Negated)
			}
		})
	}
}

func TestParseAllFlags(t *testing.T) {
	tests := []struct {
		pattern string
		flags   string
	}{
		{"/abc/g", "g"},
		{"/abc/i", "i"},
		{"/abc/m", "m"},
		{"/abc/s", "s"},
		{"/abc/u", "u"},
		{"/abc/y", "y"},
		{"/abc/d", "d"},
		{"/abc/gimsuyd", "gimsuyd"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ast.Flags != tc.flags {
				t.Errorf("expected flags '%s', got '%s'", tc.flags, ast.Flags)
			}
		})
	}
}

func TestParseBracedUnicodeEscape(t *testing.T) {
	tests := []struct {
		pattern    string
		escapeType string
		code       string
	}{
		{`\u{1F600}`, "unicode_braced", `\u{1F600}`},     // Emoji (grinning face)
		{`\u{0041}`, "unicode_braced", `\u{0041}`},       // ASCII 'A'
		{`\u{10FFFF}`, "unicode_braced", `\u{10FFFF}`},   // Max code point
		{`\u{A}`, "unicode_braced", `\u{A}`},             // Single hex digit
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			esc, ok := frag.Content.(*Escape)
			if !ok {
				t.Fatalf("expected Escape, got %T", frag.Content)
			}

			if esc.EscapeType != tc.escapeType {
				t.Errorf("expected escape type '%s', got '%s'", tc.escapeType, esc.EscapeType)
			}

			if esc.Code != tc.code {
				t.Errorf("expected code '%s', got '%s'", tc.code, esc.Code)
			}
		})
	}
}

func TestParseBracedUnicodeInCharset(t *testing.T) {
	// Test braced Unicode escape inside a character class
	ast, err := ParseRegex(`[\u{1F600}-\u{1F64F}]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frag := ast.Matches[0].Fragments[0]
	charset, ok := frag.Content.(*Charset)
	if !ok {
		t.Fatalf("expected Charset, got %T", frag.Content)
	}

	// Should have a range
	if len(charset.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(charset.Items))
	}

	rng, ok := charset.Items[0].(*CharsetRange)
	if !ok {
		t.Fatalf("expected CharsetRange, got %T", charset.Items[0])
	}

	if rng.First != `\u{1F600}` {
		t.Errorf("expected first '\\u{1F600}', got '%s'", rng.First)
	}
}

// Edge case tests - document current behavior

func TestEdgeCaseQuantifierOnAnchor(t *testing.T) {
	// JavaScript throws SyntaxError for ^+, but we parse it for visualization purposes
	// This is a lenient approach - show what the user wrote
	tests := []struct {
		pattern string
		valid   bool
	}{
		{"^+", true},   // Parses (lenient)
		{"$*", true},   // Parses (lenient)
		{"^?", true},   // Parses (lenient)
		{"^{2}", true}, // Parses (lenient)
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			_, err := ParseRegex(tc.pattern)
			if tc.valid && err != nil {
				t.Errorf("expected pattern to parse, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("expected error, but pattern parsed")
			}
		})
	}
}

func TestEdgeCaseEmptyCharset(t *testing.T) {
	// In JavaScript, [] is an error, but [^] matches any character
	// Our parser accepts both for visualization purposes

	// Empty charset [] - parses (lenient)
	ast, err := ParseRegex("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	frag := ast.Matches[0].Fragments[0]
	charset := frag.Content.(*Charset)
	if len(charset.Items) != 0 {
		t.Errorf("expected 0 items in empty charset, got %d", len(charset.Items))
	}
	if charset.Inverted {
		t.Errorf("expected non-inverted empty charset")
	}
}

func TestEdgeCaseCharsetDashPositioning(t *testing.T) {
	tests := []struct {
		pattern  string
		numItems int
		desc     string
	}{
		{"[-a]", 2, "dash at start - treated as literal dash and 'a'"},
		{"[a-]", 2, "dash at end - treated as 'a' and literal dash"},
		{"[a-b]", 1, "normal range"},
		{"[a-b-c]", 3, "dash after range - range, literal dash, and 'c'"},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			ast, err := ParseRegex(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			frag := ast.Matches[0].Fragments[0]
			charset := frag.Content.(*Charset)
			if len(charset.Items) != tc.numItems {
				t.Errorf("%s: expected %d items, got %d", tc.desc, tc.numItems, len(charset.Items))
			}
		})
	}
}

func TestEdgeCaseMultiDigitBackreference(t *testing.T) {
	// Currently only supports \1-\9
	// \10 and above are treated as octal or literal

	// \1 works as backreference
	ast, err := ParseRegex(`\1`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	frag := ast.Matches[0].Fragments[0]
	if _, ok := frag.Content.(*BackReference); !ok {
		t.Errorf("expected BackReference for \\1, got %T", frag.Content)
	}

	// \10 is NOT a backreference - treated as \1 followed by 0
	ast, err = ParseRegex(`\10`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have 2 fragments: \1 (backref) and 0 (literal)
	if len(ast.Matches[0].Fragments) != 2 {
		t.Errorf("expected 2 fragments for \\10, got %d", len(ast.Matches[0].Fragments))
	}
}
