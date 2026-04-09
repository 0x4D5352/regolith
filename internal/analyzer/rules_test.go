package analyzer

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func TestRuleAdjacentUnbounded(t *testing.T) {
	tests := []struct {
		name         string
		fragments    []*ast.MatchFragment
		wantFindings int
	}{
		{
			name: "two adjacent .* triggers finding",
			fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
			},
			wantFindings: 1,
		},
		{
			name: "single .* is fine",
			fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
				{Content: &ast.Literal{Text: "="}, Repeat: nil},
			},
			wantFindings: 0,
		},
		{
			name: ".+ followed by .* also triggers",
			fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true}},
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
			},
			wantFindings: 1,
		},
		{
			name: "three adjacent unbounded triggers two findings",
			fragments: []*ast.MatchFragment{
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
				{Content: &ast.AnyCharacter{}, Repeat: &ast.Repeat{Min: 0, Max: -1, Greedy: true}},
			},
			wantFindings: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkAdjacentUnbounded(tc.fragments, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			for _, f := range findings {
				if f.ID != "adjacent-unbounded" {
					t.Errorf("unexpected finding ID: %s", f.ID)
				}
				if f.Severity != SeverityError {
					t.Errorf("expected SeverityError, got %v", f.Severity)
				}
			}
		})
	}
}

func TestRuleNestedQuantifier(t *testing.T) {
	tests := []struct {
		name         string
		node         ast.Node
		wantFindings int
	}{
		{
			name: "(a+)+ triggers finding",
			node: &ast.MatchFragment{
				Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{{
							Fragments: []*ast.MatchFragment{{
								Content: &ast.Literal{Text: "a"},
								Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true},
							}},
						}},
					},
				},
				Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true},
			},
			wantFindings: 1,
		},
		{
			name: "(abc)+ is fine — no inner quantifier",
			node: &ast.MatchFragment{
				Content: &ast.Subexp{
					GroupType: ast.GroupCapture,
					Number:    1,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{{
							Fragments: []*ast.MatchFragment{{
								Content: &ast.Literal{Text: "abc"},
								Repeat:  nil,
							}},
						}},
					},
				},
				Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true},
			},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			frag := tc.node.(*ast.MatchFragment)
			checkNestedQuantifier(frag, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
		})
	}
}

func TestRuleTrailingWildcard(t *testing.T) {
	unboundedDot := &ast.MatchFragment{
		Content: &ast.AnyCharacter{},
		Repeat:  &ast.Repeat{Min: 0, Max: -1, Greedy: true},
	}
	tests := []struct {
		name         string
		match        *ast.Match
		wantFindings int
		wantID       string
	}{
		{
			name: "trailing .* without anchor",
			match: &ast.Match{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
				unboundedDot,
			}},
			wantFindings: 1,
			wantID:       "trailing-wildcard",
		},
		{
			name: "trailing .* with anchor is fine",
			match: &ast.Match{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
				unboundedDot,
				{Content: &ast.Anchor{AnchorType: ast.AnchorEnd}, Repeat: nil},
			}},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkTrailingWildcard(tc.match, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			if tc.wantFindings > 0 && findings[0].ID != tc.wantID {
				t.Errorf("unexpected finding ID: %s", findings[0].ID)
			}
		})
	}
}

func TestRuleLeadingWildcard(t *testing.T) {
	unboundedDot := &ast.MatchFragment{
		Content: &ast.AnyCharacter{},
		Repeat:  &ast.Repeat{Min: 0, Max: -1, Greedy: true},
	}
	tests := []struct {
		name         string
		match        *ast.Match
		wantFindings int
		wantID       string
	}{
		{
			name: "leading .* without anchor",
			match: &ast.Match{Fragments: []*ast.MatchFragment{
				unboundedDot,
				{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
			}},
			wantFindings: 1,
			wantID:       "leading-wildcard",
		},
		{
			name: "leading anchor then .* is fine",
			match: &ast.Match{Fragments: []*ast.MatchFragment{
				{Content: &ast.Anchor{AnchorType: ast.AnchorStart}, Repeat: nil},
				unboundedDot,
				{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
			}},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkLeadingWildcard(tc.match, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			if tc.wantFindings > 0 && findings[0].ID != tc.wantID {
				t.Errorf("unexpected finding ID: %s", findings[0].ID)
			}
		})
	}
}

func TestRuleSingleCharClass(t *testing.T) {
	tests := []struct {
		name         string
		node         ast.Node
		wantFindings int
		wantID       string
	}{
		{
			name: "[a] triggers",
			node: &ast.Charset{
				Inverted: false,
				Items:    []ast.CharsetItem{&ast.CharsetLiteral{Text: "a"}},
			},
			wantFindings: 1,
			wantID:       "single-char-class",
		},
		{
			name: "[ab] does not trigger",
			node: &ast.Charset{
				Inverted: false,
				Items: []ast.CharsetItem{
					&ast.CharsetLiteral{Text: "a"},
					&ast.CharsetLiteral{Text: "b"},
				},
			},
			wantFindings: 0,
		},
		{
			name: "[^a] does not trigger — negation is meaningful",
			node: &ast.Charset{
				Inverted: true,
				Items:    []ast.CharsetItem{&ast.CharsetLiteral{Text: "a"}},
			},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkSingleCharClass(tc.node, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			if tc.wantFindings > 0 && findings[0].ID != tc.wantID {
				t.Errorf("unexpected finding ID: %s", findings[0].ID)
			}
		})
	}
}

func TestRuleEmptyAlternative(t *testing.T) {
	tests := []struct {
		name         string
		regexp       *ast.Regexp
		wantFindings int
		wantID       string
	}{
		{
			name: "empty branch in alternation",
			regexp: &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
					}},
					// Empty branch: the empty string
					{Fragments: []*ast.MatchFragment{}},
				},
			},
			wantFindings: 1,
			wantID:       "empty-alternative",
		},
		{
			name: "no empty branches",
			regexp: &ast.Regexp{
				Matches: []*ast.Match{
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Literal{Text: "foo"}, Repeat: nil},
					}},
					{Fragments: []*ast.MatchFragment{
						{Content: &ast.Literal{Text: "bar"}, Repeat: nil},
					}},
				},
			},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkEmptyAlternative(tc.regexp, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			if tc.wantFindings > 0 && findings[0].ID != tc.wantID {
				t.Errorf("unexpected finding ID: %s", findings[0].ID)
			}
		})
	}
}

func TestRuleQuantifiedAssertion(t *testing.T) {
	tests := []struct {
		name         string
		frag         *ast.MatchFragment
		wantFindings int
		wantID       string
	}{
		{
			name: "quantified anchor triggers",
			frag: &ast.MatchFragment{
				Content: &ast.Anchor{AnchorType: ast.AnchorStart},
				Repeat:  &ast.Repeat{Min: 0, Max: -1, Greedy: true},
			},
			wantFindings: 1,
			wantID:       "quantified-assertion",
		},
		{
			name: "unquantified anchor is fine",
			frag: &ast.MatchFragment{
				Content: &ast.Anchor{AnchorType: ast.AnchorStart},
				Repeat:  nil,
			},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkQuantifiedAssertion(tc.frag, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
			if tc.wantFindings > 0 && findings[0].ID != tc.wantID {
				t.Errorf("unexpected finding ID: %s", findings[0].ID)
			}
		})
	}
}

func TestRuleRedundantGroup(t *testing.T) {
	tests := []struct {
		name         string
		frag         *ast.MatchFragment
		wantFindings int
	}{
		{
			name: "(?:a) with no quantifier triggers",
			frag: &ast.MatchFragment{
				Content: &ast.Subexp{
					GroupType: ast.GroupNonCapture,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{{
							Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}, Repeat: nil},
							},
						}},
					},
				},
				Repeat: nil,
			},
			wantFindings: 1,
		},
		{
			name: "(?:a)+ is fine — quantifier justifies the group",
			frag: &ast.MatchFragment{
				Content: &ast.Subexp{
					GroupType: ast.GroupNonCapture,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{{
							Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}, Repeat: nil},
							},
						}},
					},
				},
				Repeat: &ast.Repeat{Min: 1, Max: -1, Greedy: true},
			},
			wantFindings: 0,
		},
		{
			name: "(?:ab) with multiple fragments is fine",
			frag: &ast.MatchFragment{
				Content: &ast.Subexp{
					GroupType: ast.GroupNonCapture,
					Regexp: &ast.Regexp{
						Matches: []*ast.Match{{
							Fragments: []*ast.MatchFragment{
								{Content: &ast.Literal{Text: "a"}, Repeat: nil},
								{Content: &ast.Literal{Text: "b"}, Repeat: nil},
							},
						}},
					},
				},
				Repeat: nil,
			},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkRedundantGroup(tc.frag, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
		})
	}
}

func TestRuleMissingAnchor(t *testing.T) {
	tests := []struct {
		name         string
		regexp       *ast.Regexp
		wantFindings int
	}{
		{
			name: "no anchors at all triggers",
			regexp: &ast.Regexp{Matches: []*ast.Match{{
				Fragments: []*ast.MatchFragment{
					{Content: &ast.Literal{Text: "hello"}, Repeat: nil},
				},
			}}},
			wantFindings: 1,
		},
		{
			name: "has start anchor is fine",
			regexp: &ast.Regexp{Matches: []*ast.Match{{
				Fragments: []*ast.MatchFragment{
					{Content: &ast.Anchor{AnchorType: ast.AnchorStart}, Repeat: nil},
					{Content: &ast.Literal{Text: "hello"}, Repeat: nil},
				},
			}}},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkMissingAnchor(tc.regexp, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
		})
	}
}

func TestRulePossessiveOpportunity(t *testing.T) {
	tests := []struct {
		name         string
		frag         *ast.MatchFragment
		features     flavor.FeatureSet
		wantFindings int
	}{
		{
			name: "greedy quantifier on literal in possessive-capable flavor",
			frag: &ast.MatchFragment{
				Content: &ast.Literal{Text: "a"},
				Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true},
			},
			features:     flavor.FeatureSet{PossessiveQuantifiers: true},
			wantFindings: 1,
		},
		{
			name: "greedy quantifier in non-possessive flavor",
			frag: &ast.MatchFragment{
				Content: &ast.Literal{Text: "a"},
				Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true},
			},
			features:     flavor.FeatureSet{PossessiveQuantifiers: false},
			wantFindings: 0,
		},
		{
			name: "already possessive is fine",
			frag: &ast.MatchFragment{
				Content: &ast.Literal{Text: "a"},
				Repeat:  &ast.Repeat{Min: 1, Max: -1, Greedy: true, Possessive: true},
			},
			features:     flavor.FeatureSet{PossessiveQuantifiers: true},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkPossessiveOpportunity(tc.frag, tc.features, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
		})
	}
}

func TestRuleOverlappingAlternatives(t *testing.T) {
	tests := []struct {
		name         string
		regexp       *ast.Regexp
		wantFindings int
	}{
		{
			name: "identical branches trigger",
			regexp: &ast.Regexp{Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{{Content: &ast.Literal{Text: "a"}}}},
				{Fragments: []*ast.MatchFragment{{Content: &ast.Literal{Text: "a"}}}},
			}},
			wantFindings: 1,
		},
		{
			name: "different branches are fine",
			regexp: &ast.Regexp{Matches: []*ast.Match{
				{Fragments: []*ast.MatchFragment{{Content: &ast.Literal{Text: "a"}}}},
				{Fragments: []*ast.MatchFragment{{Content: &ast.Literal{Text: "b"}}}},
			}},
			wantFindings: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var findings []*Finding
			checkOverlappingAlternatives(tc.regexp, &findings)
			if len(findings) != tc.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantFindings)
			}
		})
	}
}
