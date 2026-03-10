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
	if !strings.Contains(got, "Matches `foo` literally") {
		t.Errorf("expected Matches `foo` literally, got:\n%s", got)
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
	if !strings.Contains(got, "Matches `` literally") {
		t.Errorf("expected Matches `` literally, got:\n%s", got)
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
	if !strings.Contains(got, "Matches any character") {
		t.Errorf("expected Matches any character, got:\n%s", got)
	}
}

func TestRenderMarkdown_AllAnchors(t *testing.T) {
	tests := []struct {
		anchorType string
		want       string
	}{
		{ast.AnchorStart, "Asserts start of line"},
		{ast.AnchorEnd, "Asserts end of line"},
		{ast.AnchorWordBoundary, "Asserts word boundary"},
		{ast.AnchorNonWordBoundary, "Asserts non-word boundary"},
		{ast.AnchorStringStart, "Asserts start of string"},
		{ast.AnchorStringEnd, "Asserts end of string"},
		{ast.AnchorAbsoluteEnd, "Asserts absolute end of string"},
		{ast.AnchorWordStart, "Asserts start of word"},
		{ast.AnchorWordEnd, "Asserts end of word"},
		{ast.AnchorGraphemeClusterBoundary, "Asserts grapheme cluster boundary"},
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
	if !strings.Contains(got, "Matches any digit `\\d` (0-9)") {
		t.Errorf("expected escape description, got:\n%s", got)
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
	if !strings.Contains(got, "Matches one of the following:") {
		t.Errorf("expected character class header, got:\n%s", got)
	}
	if !strings.Contains(got, "`a` to `z` (lowercase letters)") {
		t.Errorf("expected range with description, got:\n%s", got)
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
	if !strings.Contains(got, "Matches any character NOT in:") {
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
	if !strings.Contains(got, "Matches one of the following:") {
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
	if !strings.Contains(got, "POSIX `[:alpha:]` (alphabetic characters)") {
		t.Errorf("expected POSIX [:alpha:] with description, got:\n%s", got)
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
	if !strings.Contains(got, "NOT POSIX `[:digit:]` (digits)") {
		t.Errorf("expected NOT POSIX [:digit:] with description, got:\n%s", got)
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
	if !strings.Contains(got, "Unicode property `\\p{Letter}`") {
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
	if !strings.Contains(got, "NOT Unicode property `\\P{L}`") {
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
		{"capture", ast.GroupCapture, 1, "", "**Capture group #1** -- captures matched text for back-reference as `\\1`"},
		{"named", ast.GroupNamedCapture, 2, "word", `**Named capture group #2 "word"** -- captures matched text for back-reference as`},
		{"non-capture", ast.GroupNonCapture, 0, "", "**Non-capturing group** -- groups without capturing"},
		{"positive lookahead", ast.GroupPositiveLookahead, 0, "", "**Positive lookahead** -- asserts what follows matches, without consuming characters"},
		{"negative lookahead", ast.GroupNegativeLookahead, 0, "", "**Negative lookahead** -- asserts what follows does NOT match"},
		{"positive lookbehind", ast.GroupPositiveLookbehind, 0, "", "**Positive lookbehind** -- asserts what precedes matches, without consuming characters"},
		{"negative lookbehind", ast.GroupNegativeLookbehind, 0, "", "**Negative lookbehind** -- asserts what precedes does NOT match"},
		{"atomic", ast.GroupAtomic, 0, "", "**Atomic group** -- matches without backtracking"},
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
	if !strings.Contains(got, "Matches the same text previously captured by group #1") {
		t.Errorf("expected back-reference description, got:\n%s", got)
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
	if !strings.Contains(got, `Matches the same text previously captured by group "word"`) {
		t.Errorf("expected named back-reference description, got:\n%s", got)
	}
}

func TestRenderMarkdown_QuantifierMerged(t *testing.T) {
	tests := []struct {
		name   string
		repeat ast.Repeat
		want   string
	}{
		{"0+ greedy", ast.Repeat{Min: 0, Max: -1, Greedy: true}, "Matches `a` literally, 0 or more times (greedy)"},
		{"1+ greedy", ast.Repeat{Min: 1, Max: -1, Greedy: true}, "Matches `a` literally, 1 or more times (greedy)"},
		{"optional greedy", ast.Repeat{Min: 0, Max: 1, Greedy: true}, "Matches `a` literally, optionally"},
		{"0+ lazy", ast.Repeat{Min: 0, Max: -1, Greedy: false}, "Matches `a` literally, 0 or more times (lazy)"},
		{"1+ lazy", ast.Repeat{Min: 1, Max: -1, Greedy: false}, "Matches `a` literally, 1 or more times (lazy)"},
		{"exact 3", ast.Repeat{Min: 3, Max: 3, Greedy: true}, "Matches `a` literally, exactly 3 times"},
		{"range 2-5", ast.Repeat{Min: 2, Max: 5, Greedy: true}, "Matches `a` literally, 2 to 5 times (greedy)"},
		{"2+ greedy", ast.Repeat{Min: 2, Max: -1, Greedy: true}, "Matches `a` literally, 2 or more times (greedy)"},
		{"1+ possessive", ast.Repeat{Min: 1, Max: -1, Greedy: true, Possessive: true}, "Matches `a` literally, 1 or more times (possessive)"},
		{"0+ possessive", ast.Repeat{Min: 0, Max: -1, Greedy: true, Possessive: true}, "Matches `a` literally, 0 or more times (possessive)"},
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
	if !strings.Contains(got, "exactly 3 times") {
		t.Errorf("expected exact quantifier, got:\n%s", got)
	}
	if strings.Contains(got, "(greedy)") || strings.Contains(got, "(lazy)") || strings.Contains(got, "(possessive)") {
		t.Errorf("exact quantifier should not have modifier, got:\n%s", got)
	}
}

func TestRenderMarkdown_QuantifierMergedSingleLine(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}, Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true}},
			}},
		},
	}
	got := RenderMarkdown(root, "a+", "javascript")
	// Quantifier should be merged into the literal line, not a separate sibling
	if strings.Contains(got, "Quantifier:") {
		t.Errorf("quantifier should be merged into node line, not separate, got:\n%s", got)
	}
	if !strings.Contains(got, "Matches `a` literally, 1 or more times (greedy)") {
		t.Errorf("expected merged line, got:\n%s", got)
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
	if !strings.Contains(got, "**Alternation** -- matches one of 3 branches:") {
		t.Errorf("expected alternation with 3 branches, got:\n%s", got)
	}
	if !strings.Contains(got, "**Branch 1:** Matches `a` literally") {
		t.Errorf("expected Branch 1, got:\n%s", got)
	}
	if !strings.Contains(got, "**Branch 2:** Matches `b` literally") {
		t.Errorf("expected Branch 2, got:\n%s", got)
	}
	if !strings.Contains(got, "**Branch 3:** Matches `c` literally") {
		t.Errorf("expected Branch 3, got:\n%s", got)
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
	if !strings.Contains(got, "**Conditional** -- matches based on a condition") {
		t.Errorf("expected Conditional with annotation, got:\n%s", got)
	}
	if !strings.Contains(got, "Matches the same text previously captured by group #1") {
		t.Errorf("expected condition, got:\n%s", got)
	}
	if !strings.Contains(got, "Matches `yes` literally") {
		t.Errorf("expected true branch, got:\n%s", got)
	}
	if !strings.Contains(got, "Matches `no` literally") {
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
	if !strings.Contains(got, "Matches `foo.bar` literally") {
		t.Errorf("expected Matches `foo.bar` literally, got:\n%s", got)
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
	if !strings.Contains(got, "Flags: +im -s -- modifies matching behavior") {
		t.Errorf("expected Flags with annotation, got:\n%s", got)
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
	if !strings.Contains(got, "Matches `abc` literally") {
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
	if !strings.Contains(got, "**Branch reset** -- resets group numbering for each branch") {
		t.Errorf("expected Branch reset with annotation, got:\n%s", got)
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

	if !strings.Contains(got, "- **Capture group #1** -- captures matched text for back-reference as `\\1`") {
		t.Errorf("expected outer capture group, got:\n%s", got)
	}
	if !strings.Contains(got, "    - **Capture group #2** -- captures matched text for back-reference as `\\2`") {
		t.Errorf("expected inner capture group at indent 2, got:\n%s", got)
	}
	if !strings.Contains(got, "      - Matches `b` literally") {
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
  - Matches ` + "`foo`" + ` literally
  - **Capture group #1** -- captures matched text for back-reference as ` + "`\\1`" + `, optionally
    - Matches one of the following, 1 or more times (greedy):
      - ` + "`a` to `z` (lowercase letters)" + `
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
	if !strings.Contains(got, "any digit `\\d` (0-9)") {
		t.Errorf("expected escape in charset, got:\n%s", got)
	}
}

func TestRenderMarkdown_CharsetQuantifierMerged(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.CharsetRange{First: "a", Last: "z"},
					},
				}, Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true}},
			}},
		},
	}
	got := RenderMarkdown(root, "[a-z]+", "javascript")
	if !strings.Contains(got, "Matches one of the following, 1 or more times (greedy):") {
		t.Errorf("expected charset with merged quantifier, got:\n%s", got)
	}
}

func TestRenderMarkdown_GroupedLiterals(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Charset{
					Items: []ast.CharsetItem{
						&ast.CharsetLiteral{Text: "."},
						&ast.CharsetLiteral{Text: "_"},
						&ast.CharsetLiteral{Text: "%"},
					},
				}},
			}},
		},
	}
	got := RenderMarkdown(root, "[._%]", "javascript")
	if !strings.Contains(got, "`.`, `_`, `%` (literal characters)") {
		t.Errorf("expected grouped literals, got:\n%s", got)
	}
}

func TestRenderMarkdown_AlternationWithMultiFragmentBranch(t *testing.T) {
	root := &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "a"}},
				{Content: &ast.Literal{Text: "b"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "c"}},
			}},
		},
	}
	got := RenderMarkdown(root, "ab|c", "javascript")
	if !strings.Contains(got, "**Alternation** -- matches one of 2 branches:") {
		t.Errorf("expected alternation, got:\n%s", got)
	}
	if !strings.Contains(got, "**Branch 1:**") {
		t.Errorf("expected Branch 1 header, got:\n%s", got)
	}
	if !strings.Contains(got, "**Branch 2:** Matches `c` literally") {
		t.Errorf("expected inlined Branch 2, got:\n%s", got)
	}
}
