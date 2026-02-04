package parser

// Node is the interface all AST nodes implement
type Node interface {
	Type() string
}

// Regexp is the root node representing the entire regex
type Regexp struct {
	Matches []*Match // Alternation branches
	Flags   string   // Optional flags (gimuy)
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

// Anchor represents ^, $, \b, \B
type Anchor struct {
	AnchorType string // "start", "end", "word_boundary", "non_word_boundary"
}

func (a *Anchor) Type() string { return "anchor" }

// Subexp represents a group: (), (?:), (?=), (?!), (?<=), (?<!), (?<name>)
type Subexp struct {
	GroupType string  // "capture", "non_capture", "positive_lookahead", "negative_lookahead", "positive_lookbehind", "negative_lookbehind", "named_capture"
	Number    int     // Capture group number (0 if non-capture/lookbehind)
	Name      string  // Group name for named capture groups (empty otherwise)
	Regexp    *Regexp // The contained expression
}

func (s *Subexp) Type() string { return "subexp" }

// Repeat represents quantifiers: *, +, ?, {n}, {n,}, {n,m}
type Repeat struct {
	Min    int  // Minimum repetitions
	Max    int  // Maximum repetitions (-1 for unbounded)
	Greedy bool // true if greedy, false if non-greedy (has trailing ?)
}

func (r *Repeat) Type() string { return "repeat" }

// Charset represents a character class: [abc], [^abc], [a-z]
type Charset struct {
	Inverted bool          // true if negated [^...]
	Items    []CharsetItem // Contents of the charset
}

func (c *Charset) Type() string { return "charset" }

// CharsetItem can be a literal, escape, or range within a charset
type CharsetItem interface {
	Node
	isCharsetItem()
}

// CharsetLiteral is a literal character within a charset
type CharsetLiteral struct {
	Text string
}

func (cl *CharsetLiteral) Type() string    { return "charset_literal" }
func (cl *CharsetLiteral) isCharsetItem()  {}

// CharsetRange represents a range like a-z within a charset
type CharsetRange struct {
	First string // Starting character
	Last  string // Ending character
}

func (cr *CharsetRange) Type() string    { return "charset_range" }
func (cr *CharsetRange) isCharsetItem()  {}

// Escape represents escape sequences: \d, \w, \s, \n, etc.
type Escape struct {
	EscapeType string // "digit", "word", "whitespace", "newline", etc.
	Code       string // The original escape code (e.g., "d", "w", "n")
	Value      string // Display value/description
}

func (e *Escape) Type() string    { return "escape" }
func (e *Escape) isCharsetItem()  {}

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
