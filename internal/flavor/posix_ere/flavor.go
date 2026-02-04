// Package posix_ere implements the POSIX Extended Regular Expression flavor.
// This follows IEEE Std 1003.1 (POSIX.1) without GNU extensions.
package posix_ere

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// POSIXERE is the POSIX Extended Regular Expression flavor implementation.
type POSIXERE struct{}

// Ensure POSIXERE implements the Flavor interface.
var _ flavor.Flavor = (*POSIXERE)(nil)

// Name returns the flavor identifier.
func (p *POSIXERE) Name() string {
	return "posix-ere"
}

// Description returns a human-readable description.
func (p *POSIXERE) Description() string {
	return "POSIX Extended Regular Expressions (IEEE Std 1003.1)"
}

// Parse parses a POSIX ERE pattern and returns an AST.
func (p *POSIXERE) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid flags for POSIX ERE.
// POSIX ERE has no inline flags; flags are external (e.g., grep -i).
func (p *POSIXERE) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{}
}

// SupportedFeatures returns the feature capabilities of POSIX ERE.
func (p *POSIXERE) SupportedFeatures() flavor.FeatureSet {
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
		POSIXClasses:          true, // Key feature of POSIX ERE
		BalancedGroups:        false,
		InlineModifiers:       false,
		Comments:              false,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the POSIX ERE flavor with the registry.
func init() {
	flavor.Register(&POSIXERE{})
}
