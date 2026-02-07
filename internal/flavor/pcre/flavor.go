// Package pcre provides PCRE (Perl Compatible Regular Expressions) support.
// PCRE is the most feature-rich regex flavor, supporting recursive patterns,
// conditional patterns, backtracking control verbs, and more.
package pcre

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor"
)

func init() {
	flavor.Register(&PCRE{})
}

// PCRE implements the Flavor interface for PCRE (Perl Compatible Regular Expressions)
type PCRE struct{}

func (f *PCRE) Name() string {
	return "pcre"
}

func (f *PCRE) Description() string {
	return "Perl Compatible Regular Expressions (PCRE2) - the most feature-rich regex flavor"
}

func (f *PCRE) Parse(pattern string) (*ast.Regexp, error) {
	state := ast.NewParserState()
	result, err := Parse("", []byte(pattern), GlobalStore("state", state))
	if err != nil {
		return nil, err
	}
	return result.(*ast.Regexp), nil
}

func (f *PCRE) SupportedFlags() []flavor.FlagInfo {
	return []flavor.FlagInfo{
		{Char: 'i', Name: "caseless", Description: "Case-insensitive matching"},
		{Char: 'm', Name: "multiline", Description: "^ and $ match at newlines"},
		{Char: 's', Name: "dotall", Description: ". matches newlines"},
		{Char: 'x', Name: "extended", Description: "Ignore whitespace and allow comments"},
		{Char: 'J', Name: "dupnames", Description: "Allow duplicate named groups"},
		{Char: 'U', Name: "ungreedy", Description: "Invert greediness of quantifiers"},
		{Char: 'n', Name: "no_auto_capture", Description: "Plain (...) groups are non-capturing"},
	}
}

func (f *PCRE) SupportedFeatures() flavor.FeatureSet {
	return flavor.FeatureSet{
		Lookahead:             true,
		Lookbehind:            true,
		LookbehindUnlimited:   false, // PCRE has some restrictions
		NamedGroups:           true,
		AtomicGroups:          true,
		PossessiveQuantifiers: true,
		RecursivePatterns:     true,
		ConditionalPatterns:   true,
		UnicodeProperties:     true,
		POSIXClasses:          true,
		BalancedGroups:        false, // .NET only
		InlineModifiers:       true,
		Comments:              true,
		BranchReset:           true,
		BacktrackingControl:   true,
		Callouts:              true,
		ScriptRuns:            true,
		NonAtomicLookaround:   true,
		PatternStartOptions:   true,
	}
}
