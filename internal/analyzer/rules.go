package analyzer

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
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

// ================================================================================
// Performance Hint Rules
// ================================================================================

// checkRedundantGroup flags non-capturing groups around a single element
// with no quantifier. (?:a) is equivalent to a — the group wrapper adds
// no grouping, alternation, or repetition benefit.
func checkRedundantGroup(frag *ast.MatchFragment, findings *[]*Finding) {
	if frag.Repeat != nil {
		return // Quantifier justifies the group
	}
	sub, ok := frag.Content.(*ast.Subexp)
	if !ok || sub.GroupType != ast.GroupNonCapture {
		return
	}
	if sub.Regexp == nil || len(sub.Regexp.Matches) != 1 {
		return // Alternation justifies the group
	}
	if len(sub.Regexp.Matches[0].Fragments) != 1 {
		return // Multiple fragments justify the group
	}
	*findings = append(*findings, &Finding{
		ID:          "redundant-group",
		Category:    CategoryRedundancy,
		Severity:    SeverityInfo,
		Title:       "Redundant non-capturing group",
		Description: "A non-capturing group around a single element with no quantifier is redundant.",
		Suggestion:  "Remove the group wrapper.",
		Node:        frag,
	})
}

// checkMissingAnchor flags patterns that have no anchors at all. Adding
// ^ or $ anchors can help the engine skip impossible starting positions
// and improve matching performance.
func checkMissingAnchor(r *ast.Regexp, findings *[]*Finding) {
	if hasAnchor(r) {
		return
	}
	*findings = append(*findings, &Finding{
		ID:          "missing-anchor",
		Category:    CategoryPerformance,
		Severity:    SeverityInfo,
		Title:       "Pattern has no anchors",
		Description: "Adding ^ or $ anchors can help the engine skip impossible starting positions and improve performance.",
		Node:        r,
	})
}

// hasAnchor returns true if the Regexp subtree contains any Anchor node.
func hasAnchor(r *ast.Regexp) bool {
	if r == nil {
		return false
	}
	for _, m := range r.Matches {
		for _, f := range m.Fragments {
			if _, ok := f.Content.(*ast.Anchor); ok {
				return true
			}
			if sub, ok := f.Content.(*ast.Subexp); ok {
				if hasAnchor(sub.Regexp) {
					return true
				}
			}
		}
	}
	return false
}

// checkPossessiveOpportunity flags greedy quantifiers on content that
// cannot match the same characters as what follows, where a possessive
// quantifier would eliminate backtracking. Only applies to flavors that
// support possessive quantifiers.
func checkPossessiveOpportunity(frag *ast.MatchFragment, features flavor.FeatureSet, findings *[]*Finding) {
	if !features.PossessiveQuantifiers {
		return
	}
	if frag.Repeat == nil || !frag.Repeat.Greedy || frag.Repeat.Possessive {
		return
	}
	// Heuristic: literals, character classes, and escapes with greedy
	// quantifiers are safe candidates for possessive quantifiers because
	// their match set is fixed and cannot overlap with arbitrary continuations.
	switch frag.Content.(type) {
	case *ast.Literal, *ast.Charset, *ast.Escape:
		*findings = append(*findings, &Finding{
			ID:          "possessive-opportunity",
			Category:    CategoryPerformance,
			Severity:    SeverityInfo,
			Title:       "Possessive quantifier opportunity",
			Description: "This greedy quantifier could be made possessive to eliminate backtracking, since its content is a fixed character class.",
			Suggestion:  "Use a possessive quantifier (e.g., a++ instead of a+) if backtracking into this part is never needed.",
			Node:        frag,
		})
	}
}

// checkAtomicOpportunity flags quantified groups where an atomic group
// (?>...) would prevent unnecessary backtracking. Only applies to flavors
// that support atomic groups.
func checkAtomicOpportunity(frag *ast.MatchFragment, features flavor.FeatureSet, findings *[]*Finding) {
	if !features.AtomicGroups {
		return
	}
	sub, ok := frag.Content.(*ast.Subexp)
	if !ok || sub.GroupType == ast.GroupAtomic {
		return // Already atomic or not a group
	}
	if frag.Repeat == nil || frag.Repeat.Max == 1 {
		return // No quantifier or non-repeating
	}
	// Heuristic: quantified groups that don't use backreferences are
	// candidates for atomic groups, since the match is typically final
	// and the engine need not remember backtracking points inside.
	if !containsBackReference(sub.Regexp) {
		*findings = append(*findings, &Finding{
			ID:          "atomic-opportunity",
			Category:    CategoryPerformance,
			Severity:    SeverityInfo,
			Title:       "Atomic group opportunity",
			Description: "This quantified group could be wrapped in an atomic group (?>...) to prevent the engine from backtracking into it.",
			Suggestion:  "Use (?>...) if the group's match is always final.",
			Node:        frag,
		})
	}
}

// containsBackReference returns true if the subtree has any BackReference node.
func containsBackReference(r *ast.Regexp) bool {
	if r == nil {
		return false
	}
	for _, m := range r.Matches {
		for _, f := range m.Fragments {
			if _, ok := f.Content.(*ast.BackReference); ok {
				return true
			}
			if sub, ok := f.Content.(*ast.Subexp); ok {
				if containsBackReference(sub.Regexp) {
					return true
				}
			}
		}
	}
	return false
}

// checkOverlappingAlternatives uses a simple heuristic to detect
// alternation branches that start with the same literal prefix. A full
// overlap analysis would require building first-sets for each branch,
// which is deferred as a TODO for a more complete implementation.
//
// TODO: build first-set analysis for richer overlap detection.
func checkOverlappingAlternatives(r *ast.Regexp, findings *[]*Finding) {
	if len(r.Matches) < 2 {
		return
	}

	// Extract the first literal or content type of each branch to use as
	// a signature for overlap detection.
	type branchSig struct {
		nodeType string
		text     string
	}
	sigs := make([]branchSig, len(r.Matches))
	for i, m := range r.Matches {
		if len(m.Fragments) > 0 {
			content := m.Fragments[0].Content
			switch c := content.(type) {
			case *ast.Literal:
				sigs[i] = branchSig{"literal", c.Text}
			case *ast.AnyCharacter:
				sigs[i] = branchSig{"any", ""}
			default:
				sigs[i] = branchSig{content.Type(), ""}
			}
		}
	}

	// Report once when a duplicate signature is found.
	seen := map[branchSig]bool{}
	for _, sig := range sigs {
		if sig.nodeType == "" {
			continue
		}
		if seen[sig] {
			*findings = append(*findings, &Finding{
				ID:          "overlapping-alternatives",
				Category:    CategoryBacktracking,
				Severity:    SeverityWarning,
				Title:       "Potentially overlapping alternatives",
				Description: "Multiple alternation branches start with the same pattern, which may cause unnecessary backtracking.",
				Suggestion:  "Factor out the common prefix or reorder branches.",
				Node:        r,
			})
			return
		}
		seen[sig] = true
	}
}

// checkUnreachableAlternative detects branches that are fully subsumed by
// an earlier branch. Uses a simple heuristic: if an earlier branch is .*
// (matches everything), any later branch is unreachable.
//
// TODO: extend to detect non-.* subsumption cases (e.g., .+ before [a-z]+).
func checkUnreachableAlternative(r *ast.Regexp, findings *[]*Finding) {
	if len(r.Matches) < 2 {
		return
	}

	for i, m := range r.Matches {
		if i == len(r.Matches)-1 {
			break // Last branch can't subsume anything after it
		}
		// A branch consisting of only .* matches every string, making all
		// subsequent branches unreachable.
		if len(m.Fragments) == 1 && isUnboundedDot(m.Fragments[0]) {
			*findings = append(*findings, &Finding{
				ID:          "unreachable-alternative",
				Category:    CategoryCorrectness,
				Severity:    SeverityWarning,
				Title:       "Unreachable alternative",
				Description: "A branch matching .* appears before other branches, making them unreachable.",
				Node:        r,
			})
			return
		}
	}
}
