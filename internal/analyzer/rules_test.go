package analyzer

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/ast"
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
