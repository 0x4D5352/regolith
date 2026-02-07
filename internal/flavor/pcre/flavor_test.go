package pcre

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
)

func TestBasicParsing(t *testing.T) {
	p := &PCRE{}

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
		{"named group perl", "(?<name>abc)", false},
		{"named group perl alt", "(?'name'abc)", false},
		{"named group python", "(?P<name>abc)", false},
		{"atomic group", "(?>abc)", false},
		{"positive lookahead", "(?=abc)", false},
		{"negative lookahead", "(?!abc)", false},
		{"positive lookbehind", "(?<=abc)", false},
		{"negative lookbehind", "(?<!abc)", false},
		{"anchors", "^hello$", false},
		{"escape sequences", `\d\w\s`, false},
		{"back reference", `(a)\1`, false},
		{"named back reference k", `(?<n>a)\k<n>`, false},
		{"named back reference k alt", `(?'n'a)\k'n'`, false},
		{"named back reference python", `(?P<n>a)(?P=n)`, false},
		{"unicode property", `\p{L}\P{N}`, false},
		{"possessive quantifier", "a++", false},
		{"non-greedy quantifier", "a+?", false},
		{"interval", "a{2,5}", false},
		{"interval zero to m", "a{,5}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestRecursivePatterns(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"recurse whole pattern R", "(?R)", false},
		{"recurse whole pattern 0", "(?0)", false},
		{"recurse group by number", "(?1)", false},
		{"recurse group by name perl", "(?&name)", false},
		{"recurse group by name python", "(?P>name)", false},
		{"relative forward", "(?+1)", false},
		{"relative backward", "(?-1)", false},
		{"recursive in context", "a(?R)?b", false},
		{"oniguruma style number", `\g<1>`, false},
		{"oniguruma style name", `\g<name>`, false},
		{"oniguruma style number alt", `\g'1'`, false},
		{"oniguruma style name alt", `\g'name'`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestConditionalPatterns(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"conditional by number", "(?(1)yes|no)", false},
		{"conditional by number no else", "(?(1)yes)", false},
		{"conditional by name", "(?(name)yes|no)", false},
		{"conditional by name angle", "(?(<name>)yes|no)", false},
		{"conditional by name quote", "(?('name')yes|no)", false},
		{"conditional relative forward", "(?(+1)yes|no)", false},
		{"conditional relative backward", "(?(-1)yes|no)", false},
		{"conditional recursion", "(?(R)yes|no)", false},
		{"conditional recursion to group", "(?(R1)yes|no)", false},
		{"conditional recursion to name", "(?(R&name)yes|no)", false},
		{"conditional DEFINE", "(?(DEFINE)(?<digit>[0-9]))", false},
		{"conditional assertion lookahead", "(?(?=a)yes|no)", false},
		{"conditional assertion negative", "(?(?!a)yes|no)", false},
		{"conditional assertion lookbehind", "(?(?<=a)yes|no)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestBranchReset(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"simple branch reset", "(?|a|b)", false},
		{"branch reset with groups", "(?|(a)|(b))", false},
		{"branch reset complex", "(?|(red)|(green)|(blue))", false},
		{"branch reset in context", "before(?|a|b)after", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestBacktrackControl(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"fail", "(*FAIL)", false},
		{"fail abbrev", "(*F)", false},
		{"accept", "(*ACCEPT)", false},
		{"mark", "(*MARK:name)", false},
		{"commit", "(*COMMIT)", false},
		{"prune", "(*PRUNE)", false},
		{"prune with name", "(*PRUNE:name)", false},
		{"skip", "(*SKIP)", false},
		{"skip with name", "(*SKIP:name)", false},
		{"then", "(*THEN)", false},
		{"then with name", "(*THEN:name)", false},
		{"in context", "a(*SKIP)b|c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestPOSIXClasses(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"alpha", "[[:alpha:]]", false},
		{"digit", "[[:digit:]]", false},
		{"alnum", "[[:alnum:]]", false},
		{"space", "[[:space:]]", false},
		{"upper", "[[:upper:]]", false},
		{"lower", "[[:lower:]]", false},
		{"punct", "[[:punct:]]", false},
		{"xdigit", "[[:xdigit:]]", false},
		{"word", "[[:word:]]", false},
		{"ascii", "[[:ascii:]]", false},
		{"negated", "[[:^alpha:]]", false},
		{"multiple", "[[:alpha:][:digit:]]", false},
		{"mixed with range", "[[:alpha:]0-9]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestPCREAnchors(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"word boundary", `\b`, false},
		{"non-word boundary", `\B`, false},
		{"start of input", `\A`, false},
		{"end of input before newline", `\Z`, false},
		{"absolute end", `\z`, false},
		{"end of previous match", `\G`, false},
		{"reset match start", `\K`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestPCRESpecificEscapes(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Extended escapes
		{"horizontal whitespace", `\h`, false},
		{"non-horizontal whitespace", `\H`, false},
		{"vertical whitespace", `\v`, false},
		{"non-vertical whitespace", `\V`, false},
		{"linebreak sequence", `\R`, false},
		{"grapheme cluster", `\X`, false},
		{"non-newline", `\N`, false},
		// Extended hex
		{"hex extended", `\x{1F600}`, false},
		// Octal extended
		{"octal extended", `\o{101}`, false},
		// Unicode named
		{"unicode named", `\N{U+0041}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestAlternativeSyntax(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"atomic alt", "(*atomic:abc)", false},
		{"pla alt", "(*pla:abc)", false},
		{"positive_lookahead alt", "(*positive_lookahead:abc)", false},
		{"nla alt", "(*nla:abc)", false},
		{"negative_lookahead alt", "(*negative_lookahead:abc)", false},
		{"plb alt", "(*plb:abc)", false},
		{"positive_lookbehind alt", "(*positive_lookbehind:abc)", false},
		{"nlb alt", "(*nlb:abc)", false},
		{"negative_lookbehind alt", "(*negative_lookbehind:abc)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestPossessiveQuantifiers(t *testing.T) {
	p := &PCRE{}

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
		{"possessive interval zero to m", "a{,5}+", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestQuotedLiterals(t *testing.T) {
	p := &PCRE{}

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
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestComments(t *testing.T) {
	p := &PCRE{}

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
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestInlineModifiers(t *testing.T) {
	p := &PCRE{}

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
		{"pcre flags", `(?imsxJUn)abc`, false},
		// Scoped modifiers
		{"scoped enable", `(?i:abc)`, false},
		{"scoped multiple", `(?im:abc)`, false},
		{"scoped enable and disable", `(?i-m:abc)`, false},
		{"scoped in context", `foo(?i:bar)baz`, false},
		{"nested scoped", `(?i:abc(?m:def))`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestRecursiveRefAST(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name   string
		pattern string
		target string
	}{
		{"whole pattern R", "(?R)", "R"},
		{"whole pattern 0", "(?0)", "0"},
		{"by number", "(?1)", "1"},
		{"by name perl", "(?&foo)", "foo"},
		{"by name python", "(?P>foo)", "foo"},
		{"relative forward", "(?+2)", "+2"},
		{"relative backward", "(?-1)", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.pattern, err)
			}

			// Find the RecursiveRef in the AST
			if len(result.Matches) != 1 || len(result.Matches[0].Fragments) != 1 {
				t.Fatalf("Expected 1 match with 1 fragment, got %v", result)
			}

			ref, ok := result.Matches[0].Fragments[0].Content.(*ast.RecursiveRef)
			if !ok {
				t.Fatalf("Expected RecursiveRef, got %T", result.Matches[0].Fragments[0].Content)
			}

			if ref.Target != tt.target {
				t.Errorf("RecursiveRef.Target = %q, want %q", ref.Target, tt.target)
			}
		})
	}
}

func TestBacktrackControlAST(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		verb    string
		arg     string
	}{
		{"fail", "(*FAIL)", "FAIL", ""},
		{"fail abbrev", "(*F)", "FAIL", ""},
		{"accept", "(*ACCEPT)", "ACCEPT", ""},
		{"mark with arg", "(*MARK:test)", "MARK", "test"},
		{"skip with arg", "(*SKIP:label)", "SKIP", "label"},
		{"prune", "(*PRUNE)", "PRUNE", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.pattern, err)
			}

			// Find the BacktrackControl in the AST
			if len(result.Matches) != 1 || len(result.Matches[0].Fragments) != 1 {
				t.Fatalf("Expected 1 match with 1 fragment, got %v", result)
			}

			bc, ok := result.Matches[0].Fragments[0].Content.(*ast.BacktrackControl)
			if !ok {
				t.Fatalf("Expected BacktrackControl, got %T", result.Matches[0].Fragments[0].Content)
			}

			if bc.Verb != tt.verb {
				t.Errorf("BacktrackControl.Verb = %q, want %q", bc.Verb, tt.verb)
			}
			if bc.Arg != tt.arg {
				t.Errorf("BacktrackControl.Arg = %q, want %q", bc.Arg, tt.arg)
			}
		})
	}
}

func TestNonAtomicLookaround(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name      string
		pattern   string
		wantErr   bool
		groupType string
	}{
		{"napla short", "(?*abc)", false, "non_atomic_positive_lookahead"},
		{"napla long", "(*napla:abc)", false, "non_atomic_positive_lookahead"},
		{"napla full", "(*non_atomic_positive_lookahead:abc)", false, "non_atomic_positive_lookahead"},
		{"naplb short", "(?<*abc)", false, "non_atomic_positive_lookbehind"},
		{"naplb long", "(*naplb:abc)", false, "non_atomic_positive_lookbehind"},
		{"naplb full", "(*non_atomic_positive_lookbehind:abc)", false, "non_atomic_positive_lookbehind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(result.Matches) != 1 || len(result.Matches[0].Fragments) != 1 {
				t.Fatalf("Expected 1 match with 1 fragment, got %v", result)
			}

			subexp, ok := result.Matches[0].Fragments[0].Content.(*ast.Subexp)
			if !ok {
				t.Fatalf("Expected Subexp, got %T", result.Matches[0].Fragments[0].Content)
			}

			if subexp.GroupType != tt.groupType {
				t.Errorf("GroupType = %q, want %q", subexp.GroupType, tt.groupType)
			}
		})
	}
}

func TestScriptRuns(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name      string
		pattern   string
		wantErr   bool
		groupType string
	}{
		{"script_run full", "(*script_run:abc)", false, "script_run"},
		{"sr short", "(*sr:abc)", false, "script_run"},
		{"atomic_script_run full", "(*atomic_script_run:abc)", false, "atomic_script_run"},
		{"asr short", "(*asr:abc)", false, "atomic_script_run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(result.Matches) != 1 || len(result.Matches[0].Fragments) != 1 {
				t.Fatalf("Expected 1 match with 1 fragment, got %v", result)
			}

			subexp, ok := result.Matches[0].Fragments[0].Content.(*ast.Subexp)
			if !ok {
				t.Fatalf("Expected Subexp, got %T", result.Matches[0].Fragments[0].Content)
			}

			if subexp.GroupType != tt.groupType {
				t.Errorf("GroupType = %q, want %q", subexp.GroupType, tt.groupType)
			}
		})
	}
}

func TestPatternStartOptions(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
		opts    []struct {
			name  string
			value string
		}
	}{
		{"UTF", "(*UTF)abc", false, []struct{ name, value string }{{"UTF", ""}}},
		{"UCP", "(*UCP)abc", false, []struct{ name, value string }{{"UCP", ""}}},
		{"multiple", "(*UTF)(*UCP)abc", false, []struct{ name, value string }{{"UTF", ""}, {"UCP", ""}}},
		{"LIMIT_MATCH", "(*LIMIT_MATCH=100)abc", false, []struct{ name, value string }{{"LIMIT_MATCH", "100"}}},
		{"LIMIT_DEPTH", "(*LIMIT_DEPTH=50)abc", false, []struct{ name, value string }{{"LIMIT_DEPTH", "50"}}},
		{"LIMIT_HEAP", "(*LIMIT_HEAP=1000)abc", false, []struct{ name, value string }{{"LIMIT_HEAP", "1000"}}},
		{"CR", "(*CR)abc", false, []struct{ name, value string }{{"CR", ""}}},
		{"LF", "(*LF)abc", false, []struct{ name, value string }{{"LF", ""}}},
		{"CRLF", "(*CRLF)abc", false, []struct{ name, value string }{{"CRLF", ""}}},
		{"ANYCRLF", "(*ANYCRLF)abc", false, []struct{ name, value string }{{"ANYCRLF", ""}}},
		{"ANY", "(*ANY)abc", false, []struct{ name, value string }{{"ANY", ""}}},
		{"NUL", "(*NUL)abc", false, []struct{ name, value string }{{"NUL", ""}}},
		{"BSR_ANYCRLF", "(*BSR_ANYCRLF)abc", false, []struct{ name, value string }{{"BSR_ANYCRLF", ""}}},
		{"BSR_UNICODE", "(*BSR_UNICODE)abc", false, []struct{ name, value string }{{"BSR_UNICODE", ""}}},
		{"NOTEMPTY", "(*NOTEMPTY)abc", false, []struct{ name, value string }{{"NOTEMPTY", ""}}},
		{"NOTEMPTY_ATSTART", "(*NOTEMPTY_ATSTART)abc", false, []struct{ name, value string }{{"NOTEMPTY_ATSTART", ""}}},
		{"NO_AUTO_POSSESS", "(*NO_AUTO_POSSESS)abc", false, []struct{ name, value string }{{"NO_AUTO_POSSESS", ""}}},
		{"NO_DOTSTAR_ANCHOR", "(*NO_DOTSTAR_ANCHOR)abc", false, []struct{ name, value string }{{"NO_DOTSTAR_ANCHOR", ""}}},
		{"NO_JIT", "(*NO_JIT)abc", false, []struct{ name, value string }{{"NO_JIT", ""}}},
		{"NO_START_OPT", "(*NO_START_OPT)abc", false, []struct{ name, value string }{{"NO_START_OPT", ""}}},
		{"combined", "(*UTF)(*LIMIT_MATCH=100)(*CRLF)abc", false, []struct{ name, value string }{{"UTF", ""}, {"LIMIT_MATCH", "100"}, {"CRLF", ""}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(result.Options) != len(tt.opts) {
				t.Fatalf("Expected %d options, got %d", len(tt.opts), len(result.Options))
			}

			for i, want := range tt.opts {
				got := result.Options[i]
				if got.Name != want.name {
					t.Errorf("Option[%d].Name = %q, want %q", i, got.Name, want.name)
				}
				if got.Value != want.value {
					t.Errorf("Option[%d].Value = %q, want %q", i, got.Value, want.value)
				}
			}
		})
	}
}

func TestCallouts(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
		number  int
		text    string
	}{
		{"default callout", "(?C)", false, 0, ""},
		{"numbered callout", "(?C1)", false, 1, ""},
		{"numbered callout 255", "(?C255)", false, 255, ""},
		{"string callout dq", `(?C"hello")`, false, -1, "hello"},
		{"string callout sq", "(?C'hello')", false, -1, "hello"},
		{"string callout bt", "(?C`hello`)", false, -1, "hello"},
		{"string callout caret", "(?C^hello^)", false, -1, "hello"},
		{"string callout percent", "(?C%hello%)", false, -1, "hello"},
		{"string callout hash", "(?C#hello#)", false, -1, "hello"},
		{"string callout dollar", "(?C$hello$)", false, -1, "hello"},
		{"string callout brace", "(?C{hello})", false, -1, "hello"},
		{"escaped delimiter dq", `(?C"say ""hi""")`, false, -1, `say "hi"`},
		{"in context", `a(?C1)b`, false, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Find the Callout node
			var callout *ast.Callout
			for _, m := range result.Matches {
				for _, f := range m.Fragments {
					if c, ok := f.Content.(*ast.Callout); ok {
						callout = c
						break
					}
				}
			}

			if callout == nil {
				t.Fatalf("Expected Callout node in AST for %q", tt.pattern)
			}

			if callout.Number != tt.number {
				t.Errorf("Callout.Number = %d, want %d", callout.Number, tt.number)
			}
			if callout.Text != tt.text {
				t.Errorf("Callout.Text = %q, want %q", callout.Text, tt.text)
			}
		})
	}
}

func TestComplexPatterns(t *testing.T) {
	p := &PCRE{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Real-world PCRE patterns
		{"balanced parens", `\((?:[^()]|(?R))*\)`, false},
		{"nested quotes", `"(?:[^"\\]|\\.)*"`, false},
		{"IP address", `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`, false},
		{"email basic", `[\w.+-]+@[\w.-]+\.[\w]{2,}`, false},
		{"define and use", `(?(DEFINE)(?<digit>[0-9]))(?&digit)+`, false},
		{"conditional with groups", `((a)|(b))(?(2)then2|then3)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}
