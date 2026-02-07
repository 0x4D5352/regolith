// Package ast defines the Abstract Syntax Tree nodes for regex patterns.
// These types are used by all regex flavor parsers.
package ast

// Node is the interface all AST nodes implement
type Node interface {
	Type() string
}

// Regexp is the root node representing the entire regex
type Regexp struct {
	Matches []*Match         // Alternation branches
	Flags   string           // Optional flags (flavor-dependent)
	Options []*PatternOption // PCRE pattern start options (nil for other flavors)
}

func (r *Regexp) Type() string { return "regexp" }

// Match represents a sequence of fragments (one branch of alternation)
type Match struct {
	Fragments []*MatchFragment
}

func (m *Match) Type() string { return "match" }

// MatchFragment represents a content node with optional repeat
type MatchFragment struct {
	Content Node    // Literal, Escape, Charset, Subexp, Anchor, AnyCharacter
	Repeat  *Repeat // nil if no quantifier
}

func (mf *MatchFragment) Type() string { return "match_fragment" }

// Literal represents one or more literal characters
type Literal struct {
	Text string
}

func (l *Literal) Type() string { return "literal" }

// AnyCharacter represents the . metacharacter
type AnyCharacter struct{}

func (a *AnyCharacter) Type() string { return "any_character" }

// Anchor represents ^, $, \b, \B, \A, \Z, \z, \<, \>, \b{g}
type Anchor struct {
	AnchorType string // "start", "end", "word_boundary", "non_word_boundary", "string_start", "string_end", "absolute_end", "word_start", "word_end", "grapheme_cluster_boundary"
}

func (a *Anchor) Type() string { return "anchor" }

// Anchor type constants
const (
	AnchorStart           = "start"             // ^
	AnchorEnd             = "end"               // $
	AnchorWordBoundary    = "word_boundary"     // \b
	AnchorNonWordBoundary = "non_word_boundary" // \B
	AnchorStringStart     = "string_start"      // \A
	AnchorStringEnd       = "string_end"        // \Z (before final newline)
	AnchorAbsoluteEnd     = "absolute_end"      // \z (absolute end)
	AnchorWordStart                  = "word_start"                  // \< (GNU)
	AnchorWordEnd                    = "word_end"                    // \> (GNU)
	AnchorGraphemeClusterBoundary    = "grapheme_cluster_boundary"   // \b{g} (Java)
)

// Subexp represents a group: (), (?:), (?=), (?!), (?<=), (?<!), (?<name>)
type Subexp struct {
	GroupType string  // "capture", "non_capture", "positive_lookahead", "negative_lookahead", "positive_lookbehind", "negative_lookbehind", "named_capture", "atomic"
	Number    int     // Capture group number (0 if non-capture/lookbehind)
	Name      string  // Group name for named capture groups (empty otherwise)
	Regexp    *Regexp // The contained expression
}

func (s *Subexp) Type() string { return "subexp" }

// Group type constants
const (
	GroupCapture           = "capture"
	GroupNonCapture        = "non_capture"
	GroupPositiveLookahead = "positive_lookahead"
	GroupNegativeLookahead = "negative_lookahead"
	GroupPositiveLookbehind = "positive_lookbehind"
	GroupNegativeLookbehind = "negative_lookbehind"
	GroupNamedCapture      = "named_capture"
	GroupAtomic            = "atomic"
)

// Repeat represents quantifiers: *, +, ?, {n}, {n,}, {n,m}
type Repeat struct {
	Min       int  // Minimum repetitions
	Max       int  // Maximum repetitions (-1 for unbounded)
	Greedy    bool // true if greedy, false if non-greedy (has trailing ?)
	Possessive bool // true for possessive quantifiers like *+, ++, ?+ (PCRE/Java)
}

func (r *Repeat) Type() string { return "repeat" }

// Charset represents a character class: [abc], [^abc], [a-z]
type Charset struct {
	Inverted bool          // true if negated [^...]
	Items    []CharsetItem // Contents of the charset
}

func (c *Charset) Type() string { return "charset" }

// CharsetItem can be a literal, escape, range, or POSIX class within a charset
type CharsetItem interface {
	Node
	isCharsetItem()
}

// CharsetLiteral is a literal character within a charset
type CharsetLiteral struct {
	Text string
}

func (cl *CharsetLiteral) Type() string   { return "charset_literal" }
func (cl *CharsetLiteral) isCharsetItem() {}

// CharsetRange represents a range like a-z within a charset
type CharsetRange struct {
	First string // Starting character
	Last  string // Ending character
}

func (cr *CharsetRange) Type() string   { return "charset_range" }
func (cr *CharsetRange) isCharsetItem() {}

// Escape represents escape sequences: \d, \w, \s, \n, etc.
type Escape struct {
	EscapeType string // "digit", "word", "whitespace", "newline", etc.
	Code       string // The original escape code (e.g., "d", "w", "n")
	Value      string // Display value/description
}

func (e *Escape) Type() string   { return "escape" }
func (e *Escape) isCharsetItem() {}

// BackReference represents \1 through \9 or \k<name>
type BackReference struct {
	Number int    // The group number being referenced (0 for named refs)
	Name   string // The group name for named backreferences (empty for numbered)
}

func (br *BackReference) Type() string { return "back_reference" }

// UnicodePropertyEscape represents \p{...} and \P{...}
type UnicodePropertyEscape struct {
	Property string // The property name (e.g., "Letter", "L", "Script=Greek")
	Negated  bool   // true for \P{...}, false for \p{...}
}

func (upe *UnicodePropertyEscape) Type() string { return "unicode_property_escape" }

// -----------------------------------------------------------------------------
// Future AST node types for other regex flavors
// These are placeholders that will be implemented as flavors are added
// -----------------------------------------------------------------------------

// POSIXClass represents POSIX character classes like [:alpha:], [:digit:]
// Used in: POSIX BRE, POSIX ERE, PCRE, GNU grep
type POSIXClass struct {
	Name    string // "alpha", "digit", "space", "alnum", etc.
	Negated bool   // [:^alpha:] in some implementations
}

func (pc *POSIXClass) Type() string   { return "posix_class" }
func (pc *POSIXClass) isCharsetItem() {}

// POSIX class name constants
const (
	POSIXAlnum  = "alnum"  // Alphanumeric
	POSIXAlpha  = "alpha"  // Alphabetic
	POSIXBlank  = "blank"  // Space or tab
	POSIXCntrl  = "cntrl"  // Control characters
	POSIXDigit  = "digit"  // Digits
	POSIXGraph  = "graph"  // Visible characters
	POSIXLower  = "lower"  // Lowercase
	POSIXPrint  = "print"  // Printable
	POSIXPunct  = "punct"  // Punctuation
	POSIXSpace  = "space"  // Whitespace
	POSIXUpper  = "upper"  // Uppercase
	POSIXXdigit = "xdigit" // Hex digits
)

// AtomicGroup represents (?>...) - non-backtracking groups
// Used in: PCRE, Java, .NET
type AtomicGroup struct {
	Regexp *Regexp
}

func (ag *AtomicGroup) Type() string { return "atomic_group" }

// Conditional represents conditional patterns (?(...)|...)
// Used in: PCRE
type Conditional struct {
	Condition  Node    // What to test (group number, name, or assertion)
	TrueMatch  *Regexp // Pattern if condition is true
	FalseMatch *Regexp // Pattern if condition is false (optional)
}

func (c *Conditional) Type() string { return "conditional" }

// RecursiveRef represents recursive pattern references (?R), (?1), (?&name)
// Used in: PCRE
type RecursiveRef struct {
	Target string // "R" for whole pattern, number for group, name for named group
}

func (rr *RecursiveRef) Type() string { return "recursive_ref" }

// BalancedGroup represents .NET balanced groups (?<name-otherName>...)
// Used in: .NET
type BalancedGroup struct {
	Name      string
	OtherName string
	Regexp    *Regexp
}

func (bg *BalancedGroup) Type() string { return "balanced_group" }

// Comment represents (?#...) comments in patterns
// Used in: PCRE, Java, .NET
type Comment struct {
	Text string
}

func (c *Comment) Type() string { return "comment" }

// QuotedLiteral represents \Q...\E quoted literal sequences
// Used in: PCRE, Java
type QuotedLiteral struct {
	Text string
}

func (ql *QuotedLiteral) Type() string { return "quoted_literal" }

// InlineModifier represents inline flag modifiers like (?i), (?m), (?s)
// Used in: PCRE, Java, .NET
type InlineModifier struct {
	Enable  string // Flags to enable (e.g., "im")
	Disable string // Flags to disable (e.g., "s")
	Regexp  *Regexp // Optional: scoped modifier (?i:...)
}

func (im *InlineModifier) Type() string { return "inline_modifier" }

// BranchReset represents branch reset groups (?|...)
// Used in: PCRE
type BranchReset struct {
	Regexp *Regexp
}

func (br *BranchReset) Type() string { return "branch_reset" }

// BacktrackControl represents backtracking control verbs (*PRUNE), (*SKIP), (*FAIL), etc.
// Used in: PCRE
type BacktrackControl struct {
	Verb string // "PRUNE", "SKIP", "FAIL", "ACCEPT", etc.
	Arg  string // Optional argument
}

func (bc *BacktrackControl) Type() string { return "backtrack_control" }

// PatternOption represents PCRE2 pattern start options like (*UTF), (*LIMIT_MATCH=d)
// Used in: PCRE
type PatternOption struct {
	Name  string // "UTF", "CR", "LIMIT_MATCH", etc.
	Value string // For LIMIT_* options, the numeric value; empty otherwise
}

func (po *PatternOption) Type() string { return "pattern_option" }

// Callout represents PCRE2 callout syntax (?C), (?Cn), (?C"text")
// Used in: PCRE
type Callout struct {
	Number int    // 0-255 for numeric callouts, -1 for string callouts
	Text   string // Content for string callouts (empty for numeric)
}

func (co *Callout) Type() string { return "callout" }

// -----------------------------------------------------------------------------
// Parser state (shared across flavors)
// -----------------------------------------------------------------------------

// ParserState tracks state during parsing
type ParserState struct {
	GroupCounter int // For numbering capture groups
}

// NewParserState creates a new parser state
func NewParserState() *ParserState {
	return &ParserState{GroupCounter: 0}
}

// NextGroupNumber returns the next capture group number
func (ps *ParserState) NextGroupNumber() int {
	ps.GroupCounter++
	return ps.GroupCounter
}
