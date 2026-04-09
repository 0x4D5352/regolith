package analyzer

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// Analyze performs static analysis on a parsed regex AST, walking all
// nodes and applying rules at the appropriate scope. The flavorName and
// features parameters allow rules to gate on flavor-specific capabilities
// (e.g., possessive quantifiers, atomic groups).
func Analyze(root *ast.Regexp, pattern, flavorName string, features flavor.FeatureSet) *AnalysisReport {
	report := &AnalysisReport{
		Pattern: pattern,
		Flavor:  flavorName,
	}
	walkRegexp(root, features, &report.Findings)
	return report
}

// walkRegexp checks alternation-level rules then recurses into each branch.
// Rules at this level operate on the full set of alternatives (e.g., empty
// branches, overlapping or unreachable alternatives, missing anchors).
func walkRegexp(r *ast.Regexp, features flavor.FeatureSet, findings *[]*Finding) {
	if r == nil {
		return
	}
	checkEmptyAlternative(r, findings)
	checkOverlappingAlternatives(r, findings)
	checkUnreachableAlternative(r, findings)
	checkMissingAnchor(r, findings)
	for _, m := range r.Matches {
		walkMatch(m, features, findings)
	}
}

// walkMatch checks sequence-level rules then recurses into each fragment.
// Rules at this level need visibility over the full fragment list of a single
// branch (e.g., adjacent unbounded quantifiers, leading/trailing wildcards).
func walkMatch(m *ast.Match, features flavor.FeatureSet, findings *[]*Finding) {
	if m == nil {
		return
	}
	checkAdjacentUnbounded(m.Fragments, findings)
	checkTrailingWildcard(m, findings)
	checkLeadingWildcard(m, findings)
	for _, frag := range m.Fragments {
		walkFragment(frag, features, findings)
	}
}

// walkFragment checks fragment-level rules and recurses into any nested
// regexp structures. Rules here inspect a single quantified element
// (e.g., nested quantifiers, redundant groups, optimization opportunities).
// The switch at the end handles all node types that contain a nested *ast.Regexp
// so that the walker reaches every level of the AST.
func walkFragment(frag *ast.MatchFragment, features flavor.FeatureSet, findings *[]*Finding) {
	if frag == nil {
		return
	}
	checkNestedQuantifier(frag, findings)
	checkQuantifiedAssertion(frag, findings)
	checkSingleCharClass(frag.Content, findings)
	checkRedundantGroup(frag, findings)
	checkPossessiveOpportunity(frag, features, findings)
	checkAtomicOpportunity(frag, features, findings)

	// Recurse into all node types that can contain a nested regexp, so that
	// rules are applied uniformly regardless of nesting depth or flavor.
	switch n := frag.Content.(type) {
	case *ast.Subexp:
		walkRegexp(n.Regexp, features, findings)
	case *ast.Conditional:
		walkRegexp(n.TrueMatch, features, findings)
		walkRegexp(n.FalseMatch, features, findings)
	case *ast.BranchReset:
		walkRegexp(n.Regexp, features, findings)
	case *ast.BalancedGroup:
		walkRegexp(n.Regexp, features, findings)
	case *ast.InlineModifier:
		// InlineModifier.Regexp is only non-nil for scoped modifiers like (?i:...).
		// A bare (?i) has a nil Regexp and walkRegexp handles that gracefully.
		walkRegexp(n.Regexp, features, findings)
	}
}
