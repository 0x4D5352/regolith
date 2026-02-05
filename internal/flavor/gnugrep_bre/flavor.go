// Package gnugrep_bre implements the GNU grep Basic Regular Expression flavor.
// This extends POSIX BRE with GNU extensions including:
//   - \+ for one-or-more
//   - \? for zero-or-one
//   - \| for alternation
//   - \{,m\} for "at most m"
//   - \b, \B for word boundaries
//   - \<, \> for word start/end
//   - \w, \W, \s, \S for character classes
package gnugrep_bre

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// GNUGrepBRE is the GNU grep BRE flavor implementation.
type GNUGrepBRE struct {
	name string // "gnugrep" or "gnugrep-bre"
}

// Ensure GNUGrepBRE implements the Flavor interface.
var _ flavor.Flavor = (*GNUGrepBRE)(nil)

// Name returns the flavor identifier.
func (g *GNUGrepBRE) Name() string {
	return g.name
}

// Description returns a human-readable description.
func (g *GNUGrepBRE) Description() string {
	if g.name == "gnugrep" {
		return "GNU grep default mode (BRE with GNU extensions)"
	}
	return "GNU grep Basic Regular Expressions (BRE with GNU extensions)"
}

// Parse parses a GNU BRE pattern and returns an AST.
func (g *GNUGrepBRE) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid flags for GNU grep BRE.
// GNU grep has no inline flags; flags are external (e.g., grep -i).
func (g *GNUGrepBRE) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{}
}

// SupportedFeatures returns the feature capabilities of GNU grep BRE.
func (g *GNUGrepBRE) SupportedFeatures() flavor.FeatureSet {
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

// init registers the GNU grep BRE flavor with the registry.
// Registers as both "gnugrep" (default) and "gnugrep-bre" (explicit).
func init() {
	flavor.Register(&GNUGrepBRE{name: "gnugrep"})
	flavor.Register(&GNUGrepBRE{name: "gnugrep-bre"})
}
