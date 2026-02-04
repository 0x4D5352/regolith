// Package javascript implements the JavaScript (ECMAScript) regex flavor.
// This supports ES2018+ features including lookbehind, named groups, and Unicode properties.
package javascript

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// JavaScript is the JavaScript regex flavor implementation.
type JavaScript struct{}

// Ensure JavaScript implements the Flavor interface.
var _ flavor.Flavor = (*JavaScript)(nil)

// Name returns the flavor identifier.
func (j *JavaScript) Name() string {
	return "javascript"
}

// Description returns a human-readable description.
func (j *JavaScript) Description() string {
	return "JavaScript (ECMAScript 2018+) regular expressions"
}

// Parse parses a JavaScript regex pattern and returns an AST.
func (j *JavaScript) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid flags for JavaScript.
func (j *JavaScript) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{
		{Char: 'd', Name: "hasIndices", Description: "Generate indices for substring matches"},
		{Char: 'g', Name: "global", Description: "Find all matches rather than stopping after the first"},
		{Char: 'i', Name: "ignoreCase", Description: "Case-insensitive matching"},
		{Char: 'm', Name: "multiline", Description: "^ and $ match line boundaries"},
		{Char: 's', Name: "dotAll", Description: ". matches newlines"},
		{Char: 'u', Name: "unicode", Description: "Enable full Unicode matching"},
		{Char: 'y', Name: "sticky", Description: "Matches only from the lastIndex property"},
	}
}

// SupportedFeatures returns the feature capabilities of JavaScript regex.
func (j *JavaScript) SupportedFeatures() flavor.FeatureSet {
	return flavor.FeatureSet{
		Lookahead:             true,
		Lookbehind:            true,
		LookbehindUnlimited:   false, // JavaScript lookbehind has limitations
		NamedGroups:           true,
		AtomicGroups:          false,
		PossessiveQuantifiers: false,
		RecursivePatterns:     false,
		ConditionalPatterns:   false,
		UnicodeProperties:     true,
		POSIXClasses:          false,
		BalancedGroups:        false,
		InlineModifiers:       false,
		Comments:              false,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the JavaScript flavor with the registry.
func init() {
	flavor.Register(&JavaScript{})
}
