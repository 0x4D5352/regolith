// Package dotnet implements the .NET regex flavor.
// This supports System.Text.RegularExpressions features including balanced groups,
// unlimited lookbehind, inline modifiers, and Unicode properties.
package dotnet

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// DotNet is the .NET regex flavor implementation.
type DotNet struct{}

// Ensure DotNet implements the Flavor interface.
var _ flavor.Flavor = (*DotNet)(nil)

// Name returns the flavor identifier.
func (d *DotNet) Name() string {
	return "dotnet"
}

// Description returns a human-readable description.
func (d *DotNet) Description() string {
	return ".NET (System.Text.RegularExpressions) regular expressions"
}

// Parse parses a .NET regex pattern and returns an AST.
func (d *DotNet) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid inline modifiers for .NET.
func (d *DotNet) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{
		{Char: 'i', Name: "IgnoreCase", Description: "Case-insensitive matching"},
		{Char: 'm', Name: "Multiline", Description: "^ and $ match at line boundaries"},
		{Char: 's', Name: "Singleline", Description: ". matches newline characters"},
		{Char: 'n', Name: "ExplicitCapture", Description: "Only named groups are captured"},
		{Char: 'x', Name: "IgnorePatternWhitespace", Description: "Ignore unescaped whitespace and allow # comments"},
	}
}

// SupportedFeatures returns the feature capabilities of .NET regex.
func (d *DotNet) SupportedFeatures() flavor.FeatureSet {
	return flavor.FeatureSet{
		Lookahead:             true,
		Lookbehind:            true,
		LookbehindUnlimited:   true, // .NET allows variable-length lookbehind!
		NamedGroups:           true,
		AtomicGroups:          true,
		PossessiveQuantifiers: true, // .NET 7+ has some support
		RecursivePatterns:     false,
		ConditionalPatterns:   true,
		UnicodeProperties:     true,
		POSIXClasses:          false, // .NET doesn't use POSIX syntax
		BalancedGroups:        true,  // Unique to .NET!
		InlineModifiers:       true,
		Comments:              true,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the .NET flavor with the registry.
func init() {
	flavor.Register(&DotNet{})
}
