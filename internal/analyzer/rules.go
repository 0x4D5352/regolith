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

// ================================================================================
// Redundancy and Correctness Rules
// ================================================================================

// checkTrailingWildcard flags .* at the end of a match with no trailing anchor.
// In search-semantics engines (the common case), a trailing .* consumes to end of
// string but doesn't affect whether a match is found — the engine already reports
// the match without it.
func checkTrailingWildcard(match *ast.Match, findings *[]*Finding) {
	frags := match.Fragments
	if len(frags) == 0 {
		return
	}
	last := frags[len(frags)-1]
	if isUnboundedDot(last) {
		*findings = append(*findings, &Finding{
			ID:          "trailing-wildcard",
			Category:    CategoryRedundancy,
			Severity:    SeverityInfo,
			Title:       "Trailing .* without anchor",
			Description: "Trailing .* is redundant when using search (not full-match) semantics.",
			Suggestion:  "Remove the trailing .* or add a $ anchor if full-match is intended.",
			Node:        last,
		})
	}
}

// checkLeadingWildcard flags .* at the start of a match with no leading anchor.
// In search-semantics engines the engine already tries all starting positions, so
// a leading .* just adds unnecessary backtracking without changing what is matched.
func checkLeadingWildcard(match *ast.Match, findings *[]*Finding) {
	frags := match.Fragments
	if len(frags) == 0 {
		return
	}
	first := frags[0]
	if isUnboundedDot(first) {
		*findings = append(*findings, &Finding{
			ID:          "leading-wildcard",
			Category:    CategoryRedundancy,
			Severity:    SeverityInfo,
			Title:       "Leading .* without anchor",
			Description: "Leading .* is redundant when using search semantics, as the engine already tries all starting positions.",
			Suggestion:  "Remove the leading .* or add a ^ anchor if full-match is intended.",
			Node:        first,
		})
	}
}

// checkSingleCharClass flags character classes with exactly one non-negated literal
// member. [a] is precisely equivalent to the literal a, so the brackets carry no
// additional meaning and only add visual noise.
//
// Negated classes ([^a]) are intentionally excluded: they represent a large set of
// characters and cannot be simplified to a single literal.
// Classes with a SetExpression (intersection/subtraction) are also excluded since
// they carry structural meaning beyond their items list.
func checkSingleCharClass(node ast.Node, findings *[]*Finding) {
	cs, ok := node.(*ast.Charset)
	if !ok {
		return
	}
	if cs.Inverted || cs.SetExpression != nil {
		return
	}
	if len(cs.Items) == 1 {
		if _, isLiteral := cs.Items[0].(*ast.CharsetLiteral); isLiteral {
			*findings = append(*findings, &Finding{
				ID:          "single-char-class",
				Category:    CategoryRedundancy,
				Severity:    SeverityInfo,
				Title:       "Single-character class",
				Description: "A character class with a single literal member is equivalent to the literal itself.",
				Suggestion:  "Replace the character class with the literal character.",
				Node:        cs,
			})
		}
	}
}

// checkEmptyAlternative detects alternation branches with no fragments.
// An empty branch (foo|) matches the empty string, which may be a typo (missing
// the second alternative) or intentional (making the whole group optional).
// Either way it deserves a warning so the author can confirm intent.
func checkEmptyAlternative(r *ast.Regexp, findings *[]*Finding) {
	if len(r.Matches) < 2 {
		return
	}
	for _, m := range r.Matches {
		if len(m.Fragments) == 0 {
			*findings = append(*findings, &Finding{
				ID:          "empty-alternative",
				Category:    CategoryCorrectness,
				Severity:    SeverityWarning,
				Title:       "Empty alternative branch",
				Description: "An alternation contains an empty branch, which matches the empty string. This may be intentional (making the group optional) or a typo.",
				Node:        r,
			})
			return // Report once per Regexp node
		}
	}
}

// checkQuantifiedAssertion detects quantifiers applied to zero-width assertions.
// Anchors like ^, $, \b match a position rather than consuming characters, so
// repeating them (e.g. ^*) is always either a no-op or an error.
func checkQuantifiedAssertion(frag *ast.MatchFragment, findings *[]*Finding) {
	if frag.Repeat == nil {
		return
	}
	isAssertion := false
	switch frag.Content.(type) {
	case *ast.Anchor:
		isAssertion = true
	}
	if isAssertion {
		*findings = append(*findings, &Finding{
			ID:          "quantified-assertion",
			Category:    CategoryCorrectness,
			Severity:    SeverityWarning,
			Title:       "Quantified assertion",
			Description: "A quantifier on a zero-width assertion (anchor) is redundant or erroneous — assertions don't consume characters.",
			Node:        frag,
		})
	}
}

// isUnboundedDot returns true if the fragment is an AnyCharacter (.) with an
// unbounded upper repeat limit, i.e. .* or .+.
func isUnboundedDot(frag *ast.MatchFragment) bool {
	if frag.Repeat == nil || frag.Repeat.Max != -1 {
		return false
	}
	_, isDot := frag.Content.(*ast.AnyCharacter)
	return isDot
}
