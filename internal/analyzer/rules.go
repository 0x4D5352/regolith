package analyzer

import (
	"github.com/0x4d5352/regolith/internal/ast"
)

// checkAdjacentUnbounded detects adjacent MatchFragments that both have
// unbounded repetition (Max == -1). This is the classic cause of
// catastrophic backtracking: the engine tries all ways to split the
// input between the two quantifiers.
func checkAdjacentUnbounded(fragments []*ast.MatchFragment, findings *[]*Finding) {
	for i := range len(fragments) - 1 {
		curr := fragments[i]
		next := fragments[i+1]

		if curr.Repeat == nil || next.Repeat == nil {
			continue
		}
		if curr.Repeat.Max != -1 || next.Repeat.Max != -1 {
			continue
		}

		// Both are unbounded — flag the second one as the problematic node
		// since it's the one that creates the combinatorial explosion with
		// its predecessor.
		*findings = append(*findings, &Finding{
			ID:       "adjacent-unbounded",
			Category: CategoryBacktracking,
			Severity: SeverityError,
			Title:    "Adjacent unbounded quantifiers",
			Description: "Adjacent unbounded quantifiers cause catastrophic backtracking. " +
				"The engine tries all combinations of splitting the input between them.",
			Suggestion: "Combine into a single quantifier, add a separating literal, or use a non-greedy/possessive modifier.",
			Node:       next,
		})
	}
}

// checkNestedQuantifier detects a MatchFragment with a quantifier whose
// content is a Subexp that itself contains quantified fragments.
// Example: (a+)+ — the outer + and inner + create exponential paths.
func checkNestedQuantifier(frag *ast.MatchFragment, findings *[]*Finding) {
	if frag.Repeat == nil {
		return
	}

	subexp, ok := frag.Content.(*ast.Subexp)
	if !ok {
		return
	}

	if containsQuantifier(subexp.Regexp) {
		*findings = append(*findings, &Finding{
			ID:       "nested-quantifier",
			Category: CategoryBacktracking,
			Severity: SeverityError,
			Title:    "Nested quantifiers",
			Description: "A quantifier applied to a group that itself contains quantifiers " +
				"creates exponential backtracking paths.",
			Suggestion: "Flatten the expression, use an atomic group (?>...), or use a possessive quantifier if the flavor supports it.",
			Node:       frag,
		})
	}
}

// containsQuantifier returns true if the Regexp subtree has any
// MatchFragment with a non-nil Repeat.
func containsQuantifier(r *ast.Regexp) bool {
	if r == nil {
		return false
	}
	for _, m := range r.Matches {
		for _, f := range m.Fragments {
			if f.Repeat != nil {
				return true
			}
			// Recurse into nested subexpressions
			if sub, ok := f.Content.(*ast.Subexp); ok {
				if containsQuantifier(sub.Regexp) {
					return true
				}
			}
		}
	}
	return false
}
