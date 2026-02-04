// Package posix_bre implements the POSIX Basic Regular Expression flavor.
// This follows IEEE Std 1003.1 (POSIX.1) without GNU extensions.
//
// Key differences from POSIX ERE:
//   - Groups use \( and \) instead of ( and )
//   - Interval expressions use \{n,m\} instead of {n,m}
//   - No alternation operator (| is literal)
//   - No + or ? quantifiers (these are GNU extensions as \+ and \?)
//   - Back-references \1-\9 are supported (unlike ERE)
//   - ( ) { } + ? | are literal characters
package posix_bre

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// POSIXBRE is the POSIX Basic Regular Expression flavor implementation.
type POSIXBRE struct{}

// Ensure POSIXBRE implements the Flavor interface.
var _ flavor.Flavor = (*POSIXBRE)(nil)

// Name returns the flavor identifier.
func (p *POSIXBRE) Name() string {
	return "posix-bre"
}

// Description returns a human-readable description.
func (p *POSIXBRE) Description() string {
	return "POSIX Basic Regular Expressions (IEEE Std 1003.1) - uses \\( \\) for groups"
}

// Parse parses a POSIX BRE pattern and returns an AST.
func (p *POSIXBRE) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid flags for POSIX BRE.
// POSIX BRE has no inline flags; flags are external (e.g., grep -i).
func (p *POSIXBRE) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{}
}

// SupportedFeatures returns the feature capabilities of POSIX BRE.
func (p *POSIXBRE) SupportedFeatures() flavor.FeatureSet {
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
		POSIXClasses:          true, // Key feature of POSIX BRE
		BalancedGroups:        false,
		InlineModifiers:       false,
		Comments:              false,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the POSIX BRE flavor with the registry.
func init() {
	flavor.Register(&POSIXBRE{})
}
