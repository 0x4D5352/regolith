// Package gnugrep_ere implements the GNU grep Extended Regular Expression flavor.
// This extends POSIX ERE with GNU extensions including:
//   - {,m} for "at most m"
//   - \b, \B for word boundaries
//   - \<, \> for word start/end
//   - \w, \W, \s, \S for character classes
//   - \1-\9 back-references (GNU extension to ERE)
package gnugrep_ere

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// GNUGrepERE is the GNU grep ERE flavor implementation.
type GNUGrepERE struct{}

// Ensure GNUGrepERE implements the Flavor interface.
var _ flavor.Flavor = (*GNUGrepERE)(nil)

// Name returns the flavor identifier.
func (g *GNUGrepERE) Name() string {
	return "gnugrep-ere"
}

// Description returns a human-readable description.
func (g *GNUGrepERE) Description() string {
	return "GNU grep Extended Regular Expressions (ERE with GNU extensions, like grep -E)"
}

// Parse parses a GNU ERE pattern and returns an AST.
func (g *GNUGrepERE) Parse(pattern string) (*ast.Regexp, error) {
	state := ast.NewParserState()

	result, err := Parse("", []byte(pattern), GlobalStore("state", state))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	regexp, ok := result.(*ast.Regexp)
	if !ok {
		return nil, fmt.Errorf("unexpected parse result type: %T", result)
	}

	return regexp, nil
}

// SupportedFlags returns information about valid flags for GNU grep ERE.
// GNU grep has no inline flags; flags are external (e.g., grep -i).
func (g *GNUGrepERE) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{}
}

// SupportedFeatures returns the feature capabilities of GNU grep ERE.
func (g *GNUGrepERE) SupportedFeatures() flavor.FeatureSet {
	return flavor.FeatureSet{
		Lookahead:             false,
		Lookbehind:            false,
		LookbehindUnlimited:   false,
		NamedGroups:           false,
		AtomicGroups:          false,
		PossessiveQuantifiers: false,
		RecursivePatterns:     false,
		ConditionalPatterns:   false,
		UnicodeProperties:     false,
		POSIXClasses:          true,
		BalancedGroups:        false,
		InlineModifiers:       false,
		Comments:              false,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the GNU grep ERE flavor with the registry.
func init() {
	flavor.Register(&GNUGrepERE{})
}
