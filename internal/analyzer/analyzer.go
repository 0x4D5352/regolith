package analyzer

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// analysis holds walker state shared across rules. Rules that need visibility
// over the whole pattern (group numbering, backreference targets) read from
// here instead of re-walking the AST for each check.
type analysis struct {
	features     flavor.FeatureSet
	findings     []*Finding
	definedNums  map[int]bool    // All capture group numbers present in the pattern
	definedNames map[string]bool // All named capture groups present in the pattern
	usedNums     map[int]bool    // Group numbers referenced by backrefs / recursive refs
	usedNames    map[string]bool // Group names referenced by backrefs / recursive refs
}

// Analyze performs static analysis on a parsed regex AST, walking all
// nodes and applying rules at the appropriate scope. The flavorName and
// features parameters allow rules to gate on flavor-specific capabilities
// (e.g., possessive quantifiers, atomic groups).
func Analyze(root *ast.Regexp, pattern, flavorName string, features flavor.FeatureSet) *AnalysisReport {
	a := &analysis{
		features:     features,
		definedNums:  map[int]bool{},
		definedNames: map[string]bool{},
		usedNums:     map[int]bool{},
		usedNames:    map[string]bool{},
	}

	// Single pre-pass over the AST to collect group definitions and their
	// references. This lets global rules fire without re-walking and lets
	// walkFragment ask "is this group ever referenced?" in O(1).
	a.collectGroupMetadata(root)

	// Global rules: fire once against the whole pattern. These check
	// pattern-wide properties (presence of anchors, validity of every
	// backreference target) and would produce duplicate findings if run
	// per scope.
	checkMissingAnchor(root, &a.findings)
	a.checkInvalidBackReferences(root)

	// Per-scope rules: recurse through the AST.
	a.walkRegexp(root)

	return &AnalysisReport{
		Pattern:  pattern,
		Flavor:   flavorName,
		Findings: a.findings,
	}
}

// walkRegexp checks alternation-level rules then recurses into each branch.
// Rules at this level operate on the full set of alternatives (e.g., empty
// branches, overlapping or unreachable alternatives).
func (a *analysis) walkRegexp(r *ast.Regexp) {
	if r == nil {
		return
	}
	checkEmptyAlternative(r, &a.findings)
	checkOverlappingAlternatives(r, &a.findings)
	checkUnreachableAlternative(r, &a.findings)
	for _, m := range r.Matches {
		a.walkMatch(m)
	}
}

// walkMatch checks sequence-level rules then recurses into each fragment.
// Rules at this level need visibility over the full fragment list of a single
// branch (e.g., adjacent unbounded quantifiers, leading/trailing wildcards).
func (a *analysis) walkMatch(m *ast.Match) {
	if m == nil {
		return
	}
	checkAdjacentUnbounded(m.Fragments, &a.findings)
	checkTrailingWildcard(m, &a.findings)
	checkLeadingWildcard(m, &a.findings)
	checkRepeatedSingleToken(m, &a.findings)
	for _, frag := range m.Fragments {
		a.walkFragment(frag)
	}
}

// walkFragment checks fragment-level rules and recurses into any nested
// regexp structures. Rules here inspect a single quantified element
// (e.g., nested quantifiers, redundant groups, optimization opportunities).
// The switch at the end handles all node types that contain a nested *ast.Regexp
// so that the walker reaches every level of the AST.
func (a *analysis) walkFragment(frag *ast.MatchFragment) {
	if frag == nil {
		return
	}
	checkNestedQuantifier(frag, &a.findings)
	checkQuantifiedAssertion(frag, &a.findings)
	checkRedundantBoundedQuantifier(frag, &a.findings)
	checkSingleCharClass(frag.Content, &a.findings)
	checkRedundantGroup(frag, &a.findings)
	a.checkUselessCapture(frag)
	checkPossessiveOpportunity(frag, a.features, &a.findings)
	checkAtomicOpportunity(frag, a.features, &a.findings)

	// Recurse into all node types that can contain a nested regexp, so that
	// rules are applied uniformly regardless of nesting depth or flavor.
	switch n := frag.Content.(type) {
	case *ast.Subexp:
		a.walkRegexp(n.Regexp)
	case *ast.Conditional:
		a.walkRegexp(n.TrueMatch)
		a.walkRegexp(n.FalseMatch)
	case *ast.BranchReset:
		a.walkRegexp(n.Regexp)
	case *ast.BalancedGroup:
		a.walkRegexp(n.Regexp)
	case *ast.InlineModifier:
		// InlineModifier.Regexp is only non-nil for scoped modifiers like (?i:...).
		// A bare (?i) has a nil Regexp and walkRegexp handles that gracefully.
		a.walkRegexp(n.Regexp)
	}
}

// collectGroupMetadata walks the AST once, populating definedNums/definedNames
// (every capture group in the pattern) and usedNums/usedNames (every group
// number or name referenced by a backreference or recursive reference).
func (a *analysis) collectGroupMetadata(r *ast.Regexp) {
	if r == nil {
		return
	}
	for _, m := range r.Matches {
		for _, frag := range m.Fragments {
			if frag == nil {
				continue
			}
			switch n := frag.Content.(type) {
			case *ast.Subexp:
				if n.GroupType == ast.GroupCapture || n.GroupType == ast.GroupNamedCapture {
					if n.Number > 0 {
						a.definedNums[n.Number] = true
					}
					if n.Name != "" {
						a.definedNames[n.Name] = true
					}
				}
				a.collectGroupMetadata(n.Regexp)
			case *ast.BackReference:
				if n.Name != "" {
					a.usedNames[n.Name] = true
				} else if n.Number > 0 {
					a.usedNums[n.Number] = true
				}
			case *ast.RecursiveRef:
				// RecursiveRef.Target is either "R" (whole pattern), a numeric
				// string, or a name. Treat numeric targets as group references
				// and non-"R" non-numeric targets as name references.
				if n.Target != "" && n.Target != "R" {
					if num, ok := atoiSafe(n.Target); ok {
						a.usedNums[num] = true
					} else {
						a.usedNames[n.Target] = true
					}
				}
				// Recursive refs have no inner Regexp to descend into.
			case *ast.Conditional:
				a.collectGroupMetadata(n.TrueMatch)
				a.collectGroupMetadata(n.FalseMatch)
			case *ast.BranchReset:
				a.collectGroupMetadata(n.Regexp)
			case *ast.BalancedGroup:
				// .NET balanced groups define Name (and consume OtherName).
				if n.Name != "" {
					a.definedNames[n.Name] = true
				}
				if n.OtherName != "" {
					a.usedNames[n.OtherName] = true
				}
				a.collectGroupMetadata(n.Regexp)
			case *ast.InlineModifier:
				a.collectGroupMetadata(n.Regexp)
			}
		}
	}
}

// atoiSafe parses a non-negative decimal string into an int. Used for
// RecursiveRef.Target disambiguation between numeric and named targets.
func atoiSafe(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
