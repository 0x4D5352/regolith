package output

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
)

func TestRenderMarkdown_Header(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
		},
	}
	got := RenderMarkdown(root, "a", "javascript")
	if !strings.HasPrefix(got, "# Regex: `a`\n\n**Flavor:** JavaScript\n\n") {
		t.Errorf("unexpected header:\n%s", got)
	}
}

func TestFormatFlavorName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"javascript", "JavaScript"},
		{"java", "Java"},
		{"dotnet", ".NET"},
		{"pcre", "PCRE"},
		{"posix-bre", "POSIX BRE"},
		{"posix-ere", "POSIX ERE"},
		{"gnugrep-bre", "GNU grep BRE"},
		{"gnugrep-ere", "GNU grep ERE"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatFlavorName(tt.input)
			if got != tt.want {
				t.Errorf("formatFlavorName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderMarkdown_Literal(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}},
			}},
		},
	}
	got := RenderMarkdown(root, "foo", "javascript")
	if !strings.Contains(got, "- Literal `foo`") {
		t.Errorf("expected Literal `foo`, got:\n%s", got)
	}
}

func TestRenderMarkdown_EmptyLiteral(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: ""}},
			}},
		},
	}
	got := RenderMarkdown(root, "", "javascript")
	if !strings.Contains(got, "- Literal ``") {
		t.Errorf("expected Literal ``, got:\n%s", got)
	}
}

func TestRenderMarkdown_AnyCharacter(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}},
			}},
		},
	}
	got := RenderMarkdown(root, ".", "javascript")
	if !strings.Contains(got, "- Any character") {
		t.Errorf("expected Any character, got:\n%s", got)
	}
}

func TestRenderMarkdown_AllAnchors(t *testing.T) {
	tests := []struct {
		anchorType string
		want       string
	}{
		{ast.AnchorStart, "Start of line"},
		{ast.AnchorEnd, "End of line"},
		{ast.AnchorWordBoundary, "Word boundary"},
		{ast.AnchorNonWordBoundary, "Non-word boundary"},
		{ast.AnchorStringStart, "Start of string"},
		{ast.AnchorStringEnd, "End of string"},
		{ast.AnchorAbsoluteEnd, "Absolute end of string"},
		{ast.AnchorWordStart, "Start of word"},
		{ast.AnchorWordEnd, "End of word"},
		{ast.AnchorGraphemeClusterBoundary, "Grapheme cluster boundary"},
	}
	for _, tt := range tests {
		t.Run(tt.anchorType, func(t *testing.T) {
			root := &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Anchor{AnchorType: tt.anchorType}},
					}},
				},
			}
			got := RenderMarkdown(root, `\b`, "javascript")
			if !strings.Contains(got, "- "+tt.want) {
				t.Errorf("expected %q, got:\n%s", tt.want, got)
			}
		})
	}
}

func TestRenderMarkdown_Escape(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Escape{EscapeType: "digit", Code: "d", Value: "digit (0-9)"}},
			}},
		},
	}
	got := RenderMarkdown(root, `\d`, "javascript")
	if !strings.Contains(got, "- Escape: digit (0-9)") {
		t.Errorf("expected Escape: digit (0-9), got:\n%s", got)
	}
}

func TestRenderMarkdown_Charset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Inverted: false,
					Items: []ast.CharsetItem{
						&ast.CharsetRange{First: "a", Last: "z"},
						&ast.CharsetLiteral{Text: "0"},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[a-z0]", "javascript")
	if !strings.Contains(got, "**Character class**: one of") {
		t.Errorf("expected character class header, got:\n%s", got)
	}
	if !strings.Contains(got, "Range `a` to `z`") {
		t.Errorf("expected range, got:\n%s", got)
	}
	if !strings.Contains(got, "`0`") {
		t.Errorf("expected charset literal, got:\n%s", got)
	}
}

func TestRenderMarkdown_NegatedCharset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Inverted: true,
					Items: []ast.CharsetItem{
						&ast.CharsetLiteral{Text: "x"},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[^x]", "javascript")
	if !strings.Contains(got, "**Negated character class**: none of") {
		t.Errorf("expected negated class header, got:\n%s", got)
	}
}

func TestRenderMarkdown_CharsetWithSetExpression(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					SetExpression: &ast.CharsetIntersection{
						Operands: []ast.Node{
							&ast.Charset{Items: []ast.CharsetItem{&ast.CharsetLiteral{Text: "a"}}},
							&ast.Charset{Items: []ast.CharsetItem{&ast.CharsetLiteral{Text: "b"}}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[[a]&&[b]]", "javascript")
	if !strings.Contains(got, "**Character class**: one of") {
		t.Errorf("expected character class header, got:\n%s", got)
	}
	if !strings.Contains(got, "**Intersection**") {
		t.Errorf("expected intersection, got:\n%s", got)
	}
}

func TestRenderMarkdown_POSIXClass(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.POSIXClass{Name: "alpha", Negated: false},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[[:alpha:]]", "posix-ere")
	if !strings.Contains(got, "POSIX [:alpha:]") {
		t.Errorf("expected POSIX [:alpha:], got:\n%s", got)
	}
}

func TestRenderMarkdown_NegatedPOSIXClass(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.POSIXClass{Name: "digit", Negated: true},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[[:^digit:]]", "pcre")
	if !strings.Contains(got, "NOT POSIX [:digit:]") {
		t.Errorf("expected NOT POSIX [:digit:], got:\n%s", got)
	}
}

func TestRenderMarkdown_UnicodeProperty(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.UnicodePropertyEscape{Property: "Letter", Negated: false},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, `[\p{Letter}]`, "java")
	if !strings.Contains(got, `Unicode property \p{Letter}`) {
		t.Errorf("expected unicode property, got:\n%s", got)
	}
}

func TestRenderMarkdown_NegatedUnicodeProperty(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.UnicodePropertyEscape{Property: "L", Negated: true}},
			}},
		},
	}
	got := RenderMarkdown(root, `\P{L}`, "java")
	if !strings.Contains(got, `NOT Unicode property \P{L}`) {
		t.Errorf("expected negated unicode property, got:\n%s", got)
	}
}

func TestRenderMarkdown_AllSubexpTypes(t *testing.T) {
	tests := []struct {
		name      string
		groupType string
		number    int
		groupName string
		want      string
	}{
		{"capture", ast.GroupCapture, 1, "", "**Capture group #1**"},
		{"named", ast.GroupNamedCapture, 2, "word", `**Named capture group #2 "word"**`},
		{"non-capture", ast.GroupNonCapture, 0, "", "**Non-capturing group**"},
		{"positive lookahead", ast.GroupPositiveLookahead, 0, "", "**Positive lookahead**"},
		{"negative lookahead", ast.GroupNegativeLookahead, 0, "", "**Negative lookahead**"},
		{"positive lookbehind", ast.GroupPositiveLookbehind, 0, "", "**Positive lookbehind**"},
		{"negative lookbehind", ast.GroupNegativeLookbehind, 0, "", "**Negative lookbehind**"},
		{"atomic", ast.GroupAtomic, 0, "", "**Atomic group**"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Subexp{
							GroupType: tt.groupType,
							Number:    tt.number,
							Name:      tt.groupName,
							Regexp: &ast.Regexp{
								Matches: []*ast.Match{
									{Fragments: []*ast.MatchFragment{
										{Content: &ast.Literal{Text: "x"}},
									}},
								},
							},
						}},
					}},
				},
			}
			got := RenderMarkdown(root, "(x)", "javascript")
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q, got:\n%s", tt.want, got)
			}
		})
	}
}

func TestRenderMarkdown_BackReferenceNumbered(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BackReference{Number: 1}},
			}},
		},
	}
	got := RenderMarkdown(root, `\1`, "javascript")
	if !strings.Contains(got, "Back-reference #1") {
		t.Errorf("expected Back-reference #1, got:\n%s", got)
	}
}

func TestRenderMarkdown_BackReferenceNamed(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BackReference{Name: "word"}},
			}},
		},
	}
	got := RenderMarkdown(root, `\k<word>`, "pcre")
	if !strings.Contains(got, `Back-reference "word"`) {
		t.Errorf("expected Back-reference \"word\", got:\n%s", got)
	}
}

func TestRenderMarkdown_QuantifierVariants(t *testing.T) {
	tests := []struct {
		name   string
		repeat ast.Repeat
		want   string
	}{
		{"0+ greedy", ast.Repeat{Min: 0, Max: -1, Greedy: true}, "Quantifier: 0 or more (greedy)"},
		{"1+ greedy", ast.Repeat{Min: 1, Max: -1, Greedy: true}, "Quantifier: 1 or more (greedy)"},
		{"optional greedy", ast.Repeat{Min: 0, Max: 1, Greedy: true}, "Quantifier: optional (greedy)"},
		{"0+ lazy", ast.Repeat{Min: 0, Max: -1, Greedy: false}, "Quantifier: 0 or more (lazy)"},
		{"1+ lazy", ast.Repeat{Min: 1, Max: -1, Greedy: false}, "Quantifier: 1 or more (lazy)"},
		{"exact 3", ast.Repeat{Min: 3, Max: 3, Greedy: true}, "Quantifier: exactly 3 times"},
		{"range 2-5", ast.Repeat{Min: 2, Max: 5, Greedy: true}, "Quantifier: 2 to 5 times (greedy)"},
		{"2+ greedy", ast.Repeat{Min: 2, Max: -1, Greedy: true}, "Quantifier: 2 or more (greedy)"},
		{"1+ possessive", ast.Repeat{Min: 1, Max: -1, Greedy: true, Possessive: true}, "Quantifier: 1 or more (possessive)"},
		{"0+ possessive", ast.Repeat{Min: 0, Max: -1, Greedy: true, Possessive: true}, "Quantifier: 0 or more (possessive)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repeat := tt.repeat
			root := &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Literal{Text: "a"}, Repeat: &repeat},
					}},
				},
			}
			got := RenderMarkdown(root, "a*", "javascript")
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q, got:\n%s", tt.want, got)
			}
		})
	}
}

func TestRenderMarkdown_QuantifierExactNoModifier(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}, Repeat: &ast.Repeat{Min: 3, Max: 3, Greedy: true}},
			}},
		},
	}
	got := RenderMarkdown(root, "a{3}", "javascript")
	if !strings.Contains(got, "Quantifier: exactly 3 times") {
		t.Errorf("expected exact quantifier, got:\n%s", got)
	}
	// Exact quantifiers should NOT have greedy/lazy/possessive
	if strings.Contains(got, "(greedy)") || strings.Contains(got, "(lazy)") || strings.Contains(got, "(possessive)") {
		t.Errorf("exact quantifier should not have modifier, got:\n%s", got)
	}
}

func TestRenderMarkdown_QuantifierAsSibling(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}, Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true}},
			}},
		},
	}
	got := RenderMarkdown(root, "a+", "javascript")
	lines := strings.Split(got, "\n")

	var literalLine, quantLine int
	for i, line := range lines {
		if strings.Contains(line, "Literal `a`") {
			literalLine = i
		}
		if strings.Contains(line, "Quantifier:") {
			quantLine = i
		}
	}
	if literalLine == 0 || quantLine == 0 {
		t.Fatalf("missing literal or quantifier line in:\n%s", got)
	}
	// Quantifier should be the line right after the literal
	if quantLine != literalLine+1 {
		t.Errorf("quantifier should be sibling (next line), literalLine=%d, quantLine=%d", literalLine, quantLine)
	}
	// Both should have the same indentation
	literalIndent := len(lines[literalLine]) - len(strings.TrimLeft(lines[literalLine], " "))
	quantIndent := len(lines[quantLine]) - len(strings.TrimLeft(lines[quantLine], " "))
	if literalIndent != quantIndent {
		t.Errorf("quantifier indent (%d) != literal indent (%d)", quantIndent, literalIndent)
	}
}

func TestRenderMarkdown_SingleMatchUnwraps(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "abc"}},
			}},
		},
	}
	got := RenderMarkdown(root, "abc", "javascript")
	if strings.Contains(got, "Alternation") {
		t.Errorf("single match should not show alternation, got:\n%s", got)
	}
	if strings.Contains(got, "Sequence") {
		t.Errorf("single fragment should not show sequence, got:\n%s", got)
	}
}

func TestRenderMarkdown_MultiMatchAlternation(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "b"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "c"}},
			}},
		},
	}
	got := RenderMarkdown(root, "a|b|c", "javascript")
	if !strings.Contains(got, "**Alternation** (3 branches)") {
		t.Errorf("expected alternation with 3 branches, got:\n%s", got)
	}
}

func TestRenderMarkdown_MultiFragmentSequence(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
				{Content: &ast.Literal{Text: "b"}},
			}},
		},
	}
	got := RenderMarkdown(root, "ab", "javascript")
	if !strings.Contains(got, "**Sequence**") {
		t.Errorf("expected Sequence for multi-fragment match, got:\n%s", got)
	}
}

func TestRenderMarkdown_Conditional(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Conditional{
					Condition: &ast.BackReference{Number: 1},
					TrueMatch: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "yes"}},
							}},
						},
					},
					FalseMatch: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "no"}},
							}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?(1)yes|no)", "pcre")
	if !strings.Contains(got, "**Conditional**") {
		t.Errorf("expected Conditional, got:\n%s", got)
	}
	if !strings.Contains(got, "Back-reference #1") {
		t.Errorf("expected condition, got:\n%s", got)
	}
	if !strings.Contains(got, "Literal `yes`") {
		t.Errorf("expected true branch, got:\n%s", got)
	}
	if !strings.Contains(got, "Literal `no`") {
		t.Errorf("expected false branch, got:\n%s", got)
	}
}

func TestRenderMarkdown_RecursiveRef(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.RecursiveRef{Target: "R"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?R)", "pcre")
	if !strings.Contains(got, "Recurse whole pattern") {
		t.Errorf("expected Recurse whole pattern, got:\n%s", got)
	}
}

func TestRenderMarkdown_RecursiveRefGroup(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.RecursiveRef{Target: "2"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?2)", "pcre")
	if !strings.Contains(got, "Recurse group 2") {
		t.Errorf("expected Recurse group 2, got:\n%s", got)
	}
}

func TestRenderMarkdown_RecursiveRefZero(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.RecursiveRef{Target: "0"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?0)", "pcre")
	if !strings.Contains(got, "Recurse whole pattern") {
		t.Errorf("expected Recurse whole pattern for target 0, got:\n%s", got)
	}
}

func TestRenderMarkdown_Comment(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Comment{Text: "a comment"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?#a comment)", "pcre")
	if !strings.Contains(got, `Comment: "a comment"`) {
		t.Errorf("expected comment, got:\n%s", got)
	}
}

func TestRenderMarkdown_QuotedLiteral(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.QuotedLiteral{Text: "foo.bar"}},
			}},
		},
	}
	got := RenderMarkdown(root, `\Qfoo.bar\E`, "pcre")
	if !strings.Contains(got, "Literal `foo.bar`") {
		t.Errorf("expected Literal `foo.bar`, got:\n%s", got)
	}
}

func TestRenderMarkdown_InlineModifier(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.InlineModifier{Enable: "im", Disable: "s"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?im-s)", "pcre")
	if !strings.Contains(got, "Flags: +im -s") {
		t.Errorf("expected Flags: +im -s, got:\n%s", got)
	}
}

func TestRenderMarkdown_InlineModifierScoped(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.InlineModifier{
					Enable: "i",
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "abc"}},
							}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?i:abc)", "pcre")
	if !strings.Contains(got, "Flags: +i") {
		t.Errorf("expected Flags: +i, got:\n%s", got)
	}
	if !strings.Contains(got, "Literal `abc`") {
		t.Errorf("expected body literal, got:\n%s", got)
	}
}

func TestRenderMarkdown_BranchReset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BranchReset{
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}},
							}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?|a)", "pcre")
	if !strings.Contains(got, "**Branch reset**") {
		t.Errorf("expected Branch reset, got:\n%s", got)
	}
}

func TestRenderMarkdown_BalancedGroup(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BalancedGroup{
					Name:      "open",
					OtherName: "close",
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "x"}},
							}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?<open-close>x)", "dotnet")
	if !strings.Contains(got, `**Balanced group** "open" (pop "close")`) {
		t.Errorf("expected balanced group, got:\n%s", got)
	}
}

func TestRenderMarkdown_BacktrackControl(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BacktrackControl{Verb: "FAIL"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(*FAIL)", "pcre")
	if !strings.Contains(got, "Backtrack: FAIL") {
		t.Errorf("expected Backtrack: FAIL, got:\n%s", got)
	}
}

func TestRenderMarkdown_BacktrackControlWithArg(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.BacktrackControl{Verb: "MARK", Arg: "name"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(*MARK:name)", "pcre")
	if !strings.Contains(got, "Backtrack: MARK(name)") {
		t.Errorf("expected Backtrack: MARK(name), got:\n%s", got)
	}
}

func TestRenderMarkdown_PatternOptions(t *testing.T) {
	root := &ast.Regexp{
		Options: []*ast.PatternOption{
			{Name: "UTF"},
			{Name: "LIMIT_MATCH", Value: "1000"},
		},
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
			}},
		},
	}
	got := RenderMarkdown(root, "(*UTF)(*LIMIT_MATCH=1000)a", "pcre")
	if !strings.Contains(got, "Option: UTF") {
		t.Errorf("expected Option: UTF, got:\n%s", got)
	}
	if !strings.Contains(got, "Option: LIMIT_MATCH=1000") {
		t.Errorf("expected Option: LIMIT_MATCH=1000, got:\n%s", got)
	}
}

func TestRenderMarkdown_CalloutNumeric(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Callout{Number: 42, Text: ""}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?C42)", "pcre")
	if !strings.Contains(got, "Callout #42") {
		t.Errorf("expected Callout #42, got:\n%s", got)
	}
}

func TestRenderMarkdown_CalloutZero(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Callout{Number: 0, Text: ""}},
			}},
		},
	}
	got := RenderMarkdown(root, "(?C)", "pcre")
	if !strings.Contains(got, "Callout #0") {
		t.Errorf("expected Callout #0, got:\n%s", got)
	}
}

func TestRenderMarkdown_CalloutString(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Callout{Number: -1, Text: "hello"}},
			}},
		},
	}
	got := RenderMarkdown(root, `(?C"hello")`, "pcre")
	if !strings.Contains(got, `Callout "hello"`) {
		t.Errorf("expected Callout \"hello\", got:\n%s", got)
	}
}

func TestRenderMarkdown_CharsetIntersection(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					SetExpression: &ast.CharsetIntersection{
						Operands: []ast.Node{
							&ast.Charset{Items: []ast.CharsetItem{&ast.CharsetRange{First: "a", Last: "z"}}},
							&ast.UnicodePropertyEscape{Property: "Letter"},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, `[[a-z]&&\p{Letter}]`, "javascript")
	if !strings.Contains(got, "**Intersection**") {
		t.Errorf("expected Intersection, got:\n%s", got)
	}
}

func TestRenderMarkdown_CharsetSubtraction(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					SetExpression: &ast.CharsetSubtraction{
						Operands: []ast.Node{
							&ast.Charset{Items: []ast.CharsetItem{&ast.CharsetRange{First: "a", Last: "z"}}},
							&ast.Charset{Items: []ast.CharsetItem{&ast.CharsetLiteral{Text: "m"}}},
						},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, `[[a-z]--[m]]`, "javascript")
	if !strings.Contains(got, "**Subtraction**") {
		t.Errorf("expected Subtraction, got:\n%s", got)
	}
}

func TestRenderMarkdown_StringDisjunction(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.CharsetStringDisjunction{Strings: []string{"abc", "def"}},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, `[\q{abc|def}]`, "javascript")
	if !strings.Contains(got, `String disjunction: "abc", "def"`) {
		t.Errorf("expected string disjunction, got:\n%s", got)
	}
}

func TestRenderMarkdown_NestedGroups(t *testing.T) {
	inner := &ast.Subexp{
		GroupType: ast.GroupCapture,
		Number:    2,
		Regexp: &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "b"}},
				}},
			},
		},
	}
	outer := &ast.Subexp{
		GroupType: ast.GroupCapture,
		Number:    1,
		Regexp: &ast.Regexp{
			Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "a"}},
					{Content: inner},
				}},
			},
		},
	}
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: outer},
			}},
		},
	}
	got := RenderMarkdown(root, "(a(b))", "javascript")

	// Verify indentation depth: outer group at level 0, sequence at level 1,
	// inner group at level 2, inner literal at level 3
	if !strings.Contains(got, "- **Capture group #1**") {
		t.Errorf("expected outer capture group, got:\n%s", got)
	}
	if !strings.Contains(got, "    - **Capture group #2**") {
		t.Errorf("expected inner capture group at indent 2, got:\n%s", got)
	}
	if !strings.Contains(got, "      - Literal `b`") {
		t.Errorf("expected inner literal at indent 3, got:\n%s", got)
	}
}

func TestRenderMarkdown_ComplexPattern(t *testing.T) {
	// Simulates: foo([a-z]+)?
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}},
				{Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{
							{Fragments: []*ast.MatchFragment{
								{Content: &ast.Charset{
									Items: []ast.CharsetItem{
										&ast.CharsetRange{First: "a", Last: "z"},
									},
								}, Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true}},
							}},
						},
					},
				}, Repeat: &ast.Repeat{Min: 0, Max: 1, Greedy: true}},
			}},
		},
	}
	got := RenderMarkdown(root, "foo([a-z]+)?", "javascript")

	expected := `# Regex: ` + "`foo([a-z]+)?`" + `

**Flavor:** JavaScript

- **Sequence**
  - Literal ` + "`foo`" + `
  - **Capture group #1**
    - **Character class**: one of
      - Range ` + "`a` to `z`" + `
    - Quantifier: 1 or more (greedy)
  - Quantifier: optional (greedy)
`
	if got != expected {
		t.Errorf("complex pattern mismatch.\nwant:\n%s\ngot:\n%s", expected, got)
	}
}

func TestRenderMarkdown_EscapeInCharset(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.Escape{EscapeType: "digit", Code: "d", Value: "digit (0-9)"},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, `[\d]`, "javascript")
	if !strings.Contains(got, "Escape: digit (0-9)") {
		t.Errorf("expected escape in charset, got:\n%s", got)
	}
}
