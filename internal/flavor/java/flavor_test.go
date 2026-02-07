package java

import (
	"testing"
)

func TestBasicParsing(t *testing.T) {
	j := &Java{}

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
		{"named group", "(?<name>abc)", false},
		{"atomic group", "(?>abc)", false},
		{"positive lookahead", "(?=abc)", false},
		{"negative lookahead", "(?!abc)", false},
		{"positive lookbehind", "(?<=abc)", false},
		{"negative lookbehind", "(?<!abc)", false},
		{"anchors", "^hello$", false},
		{"escape sequences", `\d\w\s`, false},
		{"back reference", `(a)\1`, false},
		{"named back reference", `(?<n>a)\k<n>`, false},
		{"unicode property", `\p{L}\P{N}`, false},
		{"possessive quantifier", "a++", false},
		{"non-greedy quantifier", "a+?", false},
		{"interval", "a{2,5}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestJavaSpecificEscapes(t *testing.T) {
	j := &Java{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Java-specific character class escapes
		{"horizontal whitespace", `\h`, false},
		{"non-horizontal whitespace", `\H`, false},
		{"vertical whitespace", `\v`, false},
		{"non-vertical whitespace", `\V`, false},
		{"linebreak matcher", `\R`, false},
		{"grapheme cluster", `\X`, false},
		// Java control characters
		{"bell escape", `\a`, false},
		{"escape char", `\e`, false},
		// Standard escapes
		{"tab", `\t`, false},
		{"newline", `\n`, false},
		{"carriage return", `\r`, false},
		{"form feed", `\f`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestJavaAnchors(t *testing.T) {
	j := &Java{}

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
		{"end of previous match", `\G`, false},
		{"grapheme cluster boundary", `\b{g}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestJavaUnicodeProperties(t *testing.T) {
	j := &Java{}

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
		// POSIX-style
		{"posix lower", `\p{Lower}`, false},
		{"posix upper", `\p{Upper}`, false},
		{"posix alpha", `\p{Alpha}`, false},
		{"posix digit", `\p{Digit}`, false},
		{"posix ascii", `\p{ASCII}`, false},
		// Java-specific
		{"java lowercase", `\p{javaLowerCase}`, false},
		{"java uppercase", `\p{javaUpperCase}`, false},
		// Script and block
		{"script latin", `\p{IsLatin}`, false},
		{"block greek", `\p{InGreek}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestPossessiveQuantifiers(t *testing.T) {
	j := &Java{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"possessive star", "a*+", false},
		{"possessive plus", "a++", false},
		{"possessive question", "a?+", false},
		{"possessive interval exact", "a{3}+", false},
		{"possessive interval min", "a{3,}+", false},
		{"possessive interval range", "a{3,5}+", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestQuotedLiterals(t *testing.T) {
	j := &Java{}

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
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestComments(t *testing.T) {
	j := &Java{}

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
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestInlineModifiers(t *testing.T) {
	j := &Java{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Global modifiers
		{"enable case insensitive", `(?i)abc`, false},
		{"enable multiple", `(?im)abc`, false},
		{"disable flag", `(?-i)abc`, false},
		{"enable and disable", `(?i-m)abc`, false},
		{"all flags", `(?idmsuxU)abc`, false},
		// Scoped modifiers
		{"scoped enable", `(?i:abc)`, false},
		{"scoped multiple", `(?im:abc)`, false},
		{"scoped enable and disable", `(?i-m:abc)`, false},
		{"scoped in context", `foo(?i:bar)baz`, false},
		{"nested scoped", `(?i:abc(?m:def))`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}
