// Package java implements the Java regex flavor.
// This supports java.util.regex.Pattern features including atomic groups,
// possessive quantifiers, inline modifiers, and Unicode properties.
package java

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

// Java is the Java regex flavor implementation.
type Java struct{}

// Ensure Java implements the Flavor interface.
var _ flavor.Flavor = (*Java)(nil)

// Name returns the flavor identifier.
func (j *Java) Name() string {
	return "java"
}

// Description returns a human-readable description.
func (j *Java) Description() string {
	return "Java (java.util.regex.Pattern) regular expressions"
}

// Parse parses a Java regex pattern and returns an AST.
func (j *Java) Parse(pattern string) (*ast.Regexp, error) {
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

// SupportedFlags returns information about valid flags for Java.
func (j *Java) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{
		{Char: 'd', Name: "UNIX_LINES", Description: "Only \\n is recognized as line terminator"},
		{Char: 'i', Name: "CASE_INSENSITIVE", Description: "Case-insensitive matching (US-ASCII)"},
		{Char: 'm', Name: "MULTILINE", Description: "^ and $ match at line boundaries"},
		{Char: 's', Name: "DOTALL", Description: ". matches any character including line terminators"},
		{Char: 'u', Name: "UNICODE_CASE", Description: "Unicode-aware case folding"},
		{Char: 'x', Name: "COMMENTS", Description: "Permit whitespace and comments in pattern"},
		{Char: 'U', Name: "UNICODE_CHARACTER_CLASS", Description: "Unicode version of predefined character classes"},
	}
}

// SupportedFeatures returns the feature capabilities of Java regex.
func (j *Java) SupportedFeatures() flavor.FeatureSet {
	return flavor.FeatureSet{
		Lookahead:             true,
		Lookbehind:            true,
		LookbehindUnlimited:   false, // Java lookbehind requires fixed-width patterns
		NamedGroups:           true,
		AtomicGroups:          true,
		PossessiveQuantifiers: true,
		RecursivePatterns:     false,
		ConditionalPatterns:   false,
		UnicodeProperties:     true,
		POSIXClasses:          true, // Via \p{Lower}, \p{Upper}, etc.
		BalancedGroups:        false,
		InlineModifiers:       true,
		Comments:              true,
		BranchReset:           false,
		BacktrackingControl:   false,
	}
}

// init registers the Java flavor with the registry.
func init() {
	flavor.Register(&Java{})
}
