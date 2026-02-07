// Package parser provides regex parsing functionality.
// For backward compatibility, AST types are aliased from internal/ast.
package parser

import "github.com/0x4d5352/regolith/internal/ast"

// Type aliases for backward compatibility
// These allow existing code to continue using parser.Regexp, parser.Match, etc.

type Node = ast.Node
type Regexp = ast.Regexp
type Match = ast.Match
type MatchFragment = ast.MatchFragment
type Literal = ast.Literal
type AnyCharacter = ast.AnyCharacter
type Anchor = ast.Anchor
type Subexp = ast.Subexp
type Repeat = ast.Repeat
type Charset = ast.Charset
type CharsetItem = ast.CharsetItem
type CharsetLiteral = ast.CharsetLiteral
type CharsetRange = ast.CharsetRange
type Escape = ast.Escape
type BackReference = ast.BackReference
type UnicodePropertyEscape = ast.UnicodePropertyEscape
type ParserState = ast.ParserState

// Function aliases
var NewParserState = ast.NewParserState

// Anchor type constants (re-exported for compatibility)
const (
	AnchorStart           = ast.AnchorStart
	AnchorEnd             = ast.AnchorEnd
	AnchorWordBoundary    = ast.AnchorWordBoundary
	AnchorNonWordBoundary = ast.AnchorNonWordBoundary
	AnchorStringStart     = ast.AnchorStringStart
	AnchorStringEnd       = ast.AnchorStringEnd
	AnchorAbsoluteEnd     = ast.AnchorAbsoluteEnd
	AnchorWordStart       = ast.AnchorWordStart
	AnchorWordEnd         = ast.AnchorWordEnd
)

// Group type constants (re-exported for compatibility)
const (
	GroupCapture            = ast.GroupCapture
	GroupNonCapture         = ast.GroupNonCapture
	GroupPositiveLookahead  = ast.GroupPositiveLookahead
	GroupNegativeLookahead  = ast.GroupNegativeLookahead
	GroupPositiveLookbehind = ast.GroupPositiveLookbehind
	GroupNegativeLookbehind = ast.GroupNegativeLookbehind
	GroupNamedCapture       = ast.GroupNamedCapture
	GroupAtomic             = ast.GroupAtomic
)

// Future AST types (re-exported for compatibility)
// These are placeholders for when flavors are implemented
type POSIXClass = ast.POSIXClass
type AtomicGroup = ast.AtomicGroup
type Conditional = ast.Conditional
type RecursiveRef = ast.RecursiveRef
type BalancedGroup = ast.BalancedGroup
type Comment = ast.Comment
type QuotedLiteral = ast.QuotedLiteral
type InlineModifier = ast.InlineModifier
type BranchReset = ast.BranchReset
type BacktrackControl = ast.BacktrackControl
type PatternOption = ast.PatternOption
type Callout = ast.Callout

// POSIX class name constants (re-exported for compatibility)
const (
	POSIXAlnum  = ast.POSIXAlnum
	POSIXAlpha  = ast.POSIXAlpha
	POSIXBlank  = ast.POSIXBlank
	POSIXCntrl  = ast.POSIXCntrl
	POSIXDigit  = ast.POSIXDigit
	POSIXGraph  = ast.POSIXGraph
	POSIXLower  = ast.POSIXLower
	POSIXPrint  = ast.POSIXPrint
	POSIXPunct  = ast.POSIXPunct
	POSIXSpace  = ast.POSIXSpace
	POSIXUpper  = ast.POSIXUpper
	POSIXXdigit = ast.POSIXXdigit
)
