package dotnet

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
)

func TestBasicParsing(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"simple literal", "hello", false},
		{"alternation", "a|b|c", false},
		{"charset", "[abc]", false},
		{"quantifiers", "a*b+c?", false},
		{"groups", "(abc)", false},
		{"non-capturing group", "(?:abc)", false},
		{"atomic group", "(?>abc)", false},
		{"positive lookahead", "(?=abc)", false},
		{"negative lookahead", "(?!abc)", false},
		{"positive lookbehind", "(?<=abc)", false},
		{"negative lookbehind", "(?<!abc)", false},
		{"anchors", "^hello$", false},
		{"escape sequences", `\d\w\s`, false},
		{"back reference", `(a)\1`, false},
		{"unicode property", `\p{L}\P{N}`, false},
		{"non-greedy quantifier", "a+?", false},
		{"interval", "a{2,5}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestDotNetNamedGroups(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Angle bracket syntax
		{"named group angle brackets", "(?<name>abc)", false},
		{"named group with digits", "(?<name123>abc)", false},
		{"named group with underscore", "(?<_name>abc)", false},
		// Single quote syntax
		{"named group single quotes", "(?'name'abc)", false},
		{"named group single quotes with digits", "(?'name123'abc)", false},
		{"named group single quotes with underscore", "(?'_name'abc)", false},
		// Named backreferences
		{"named backref angle brackets", `(?<n>a)\k<n>`, false},
		{"named backref single quotes", `(?'n'a)\k'n'`, false},
		// Mixed syntax (valid in .NET)
		{"mixed syntax", `(?<name>a)\k'name'`, false},
		{"mixed syntax reverse", `(?'name'a)\k<name>`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestBalancedGroups(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Basic balanced groups
		{"capturing balanced group angle", "(?<Close-Open>a)", false},
		{"capturing balanced group quotes", "(?'Close-Open'a)", false},
		// Non-capturing balanced groups
		{"non-capturing balanced angle", "(?<-Open>a)", false},
		{"non-capturing balanced quotes", "(?'-Open'a)", false},
		// Balanced groups in context
		{"open then close", "(?<Open>a)(?<Close-Open>b)", false},
		{"multiple balance", "(?<O>a)(?<-O>b)(?<O>c)(?<-O>d)", false},
		// Real-world patterns
		{"simple parens", `\((?:[^()]|(?<Open>\()|(?<Close-Open>\)))*\)`, false},
		{"balanced braces", `\{(?:[^{}]|(?<Open>\{)|(?<Close-Open>\}))*\}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestBalancedGroupAST(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name      string
		pattern   string
		groupName string
		otherName string
	}{
		{"capturing balanced", "(?<Close-Open>a)", "Close", "Open"},
		{"non-capturing balanced", "(?<-Open>a)", "", "Open"},
		{"quotes capturing", "(?'Close-Open'a)", "Close", "Open"},
		{"quotes non-capturing", "(?'-Open'a)", "", "Open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := d.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.pattern, err)
			}

			if len(result.Matches) != 1 {
				t.Fatalf("expected 1 match, got %d", len(result.Matches))
			}
			if len(result.Matches[0].Fragments) != 1 {
				t.Fatalf("expected 1 fragment, got %d", len(result.Matches[0].Fragments))
			}

			bg, ok := result.Matches[0].Fragments[0].Content.(*ast.BalancedGroup)
			if !ok {
				t.Fatalf("expected BalancedGroup, got %T", result.Matches[0].Fragments[0].Content)
			}

			if bg.Name != tt.groupName {
				t.Errorf("expected Name=%q, got %q", tt.groupName, bg.Name)
			}
			if bg.OtherName != tt.otherName {
				t.Errorf("expected OtherName=%q, got %q", tt.otherName, bg.OtherName)
			}
		})
	}
}

func TestDotNetAnchors(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"word boundary", `\b`, false},
		{"non-word boundary", `\B`, false},
		{"start of input", `\A`, false},
		{"end of input (before newline)", `\Z`, false},
		{"absolute end", `\z`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestDotNetEscapes(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Standard escapes
		{"digit", `\d`, false},
		{"non-digit", `\D`, false},
		{"word", `\w`, false},
		{"non-word", `\W`, false},
		{"whitespace", `\s`, false},
		{"non-whitespace", `\S`, false},
		// Control characters
		{"tab", `\t`, false},
		{"newline", `\n`, false},
		{"carriage return", `\r`, false},
		{"form feed", `\f`, false},
		{"bell", `\a`, false},
		{"escape char", `\e`, false},
		{"vertical tab", `\v`, false}, // Note: \v is vertical tab in .NET, not vertical whitespace class
		// Hex and unicode
		{"hex escape", `\x41`, false},
		{"unicode escape", `\u0041`, false},
		// Octal
		{"octal", `\0101`, false},
		// Control character
		{"control char", `\cA`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestDotNetUnicodeProperties(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Unicode category
		{"unicode letter", `\p{L}`, false},
		{"unicode uppercase", `\p{Lu}`, false},
		{"unicode digit", `\p{Nd}`, false},
		// Negated
		{"not unicode letter", `\P{L}`, false},
		// Named blocks
		{"basic latin block", `\p{IsBasicLatin}`, false},
		{"greek block", `\p{IsGreek}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestInlineModifiers(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Global modifiers - .NET flags: i, m, s, n, x
		{"enable case insensitive", `(?i)abc`, false},
		{"enable multiline", `(?m)abc`, false},
		{"enable singleline", `(?s)abc`, false},
		{"enable explicit capture", `(?n)abc`, false},
		{"enable ignore whitespace", `(?x)abc`, false},
		{"enable multiple", `(?imsnx)abc`, false},
		{"disable flag", `(?-i)abc`, false},
		{"enable and disable", `(?i-m)abc`, false},
		// Scoped modifiers
		{"scoped enable", `(?i:abc)`, false},
		{"scoped multiple", `(?im:abc)`, false},
		{"scoped enable and disable", `(?i-m:abc)`, false},
		{"scoped in context", `foo(?i:bar)baz`, false},
		{"nested scoped", `(?i:abc(?m:def))`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestComments(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"simple comment", `(?#this is a comment)`, false},
		{"comment in pattern", `foo(?#match foo)bar`, false},
		{"empty comment", `(?#)`, false},
		{"comment with special chars", `(?#[a-z]+ matches...)`, false},
		{"multiple comments", `a(?#first)b(?#second)c`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestQuotedLiterals(t *testing.T) {
	d := &DotNet{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"simple quoted", `\Qhello\E`, false},
		{"quoted with metacharacters", `\Q[a-z]+\E`, false},
		{"quoted in context", `foo\Q***\Ebar`, false},
		{"quoted with special chars", `\Q($.*)\E`, false},
		{"empty quoted", `\Q\E`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestUnlimitedLookbehind(t *testing.T) {
	d := &DotNet{}

	// .NET allows variable-length lookbehind, which is unique
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Fixed-width lookbehind (should work in all flavors)
		{"fixed width", `(?<=abc)`, false},
		{"fixed with escapes", `(?<=\d\d\d)`, false},
		// Variable-width lookbehind (unique to .NET)
		{"variable width star", `(?<=a*)`, false},
		{"variable width plus", `(?<=a+)`, false},
		{"variable width question", `(?<=ab?)`, false},
		{"alternation different lengths", `(?<=ab|abc|abcd)`, false},
		{"bounded quantifier", `(?<=a{1,10})`, false},
		// Negative variable-width lookbehind
		{"negative variable width", `(?<!a+)`, false},
		{"negative alternation", `(?<!foo|foobar)`, false},
		// Complex patterns
		{"complex lookbehind", `(?<=prefix\s+)word`, false},
		{"lookbehind with groups", `(?<=(a|b)+)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestFlavorInfo(t *testing.T) {
	d := &DotNet{}

	if d.Name() != "dotnet" {
		t.Errorf("expected Name() = 'dotnet', got %q", d.Name())
	}

	features := d.SupportedFeatures()

	if !features.Lookahead {
		t.Error("expected Lookahead = true")
	}
	if !features.Lookbehind {
		t.Error("expected Lookbehind = true")
	}
	if !features.LookbehindUnlimited {
		t.Error("expected LookbehindUnlimited = true")
	}
	if !features.NamedGroups {
		t.Error("expected NamedGroups = true")
	}
	if !features.AtomicGroups {
		t.Error("expected AtomicGroups = true")
	}
	if !features.BalancedGroups {
		t.Error("expected BalancedGroups = true")
	}
	if !features.InlineModifiers {
		t.Error("expected InlineModifiers = true")
	}
	if !features.Comments {
		t.Error("expected Comments = true")
	}

	flags := d.SupportedFlags()
	if len(flags) != 5 {
		t.Errorf("expected 5 flags, got %d", len(flags))
	}
}

func TestComplexPatterns(t *testing.T) {
	d := &DotNet{}

	// Real-world .NET regex patterns
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Email validation
		{"email pattern", `^[\w.-]+@[\w.-]+\.\w+$`, false},
		// URL matching
		{"url pattern", `https?://[\w.-]+(?:/[\w./-]*)?`, false},
		// Balanced parentheses matching (classic .NET use case)
		{"balanced parens simple", `\((?:[^()]|(?<o>\()|(?<-o>\)))*\)`, false},
		// HTML tag matching
		{"html tag", `<(?<tag>\w+)[^>]*>(?<content>.*?)</\k<tag>>`, false},
		// IP address
		{"ip address", `\b(?:\d{1,3}\.){3}\d{1,3}\b`, false},
		// Mixed features
		{"mixed features", `(?i)(?<name>\w+)(?#comment)\k<name>`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}
