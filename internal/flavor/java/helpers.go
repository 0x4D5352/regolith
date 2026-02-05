package java

import (
	"strconv"

	"github.com/0x4d5352/regolith/internal/ast"
)

// getString converts PEG match result to string
func getString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case []any:
		var s string
		for _, b := range val {
			if byteSlice, ok := b.([]byte); ok {
				s += string(byteSlice)
			}
		}
		return s
	case string:
		return val
	default:
		return ""
	}
}

// parseInt parses an integer from PEG match result
func parseInt(v any) int {
	var s string
	switch val := v.(type) {
	case []byte:
		s = string(val)
	case []any:
		for _, b := range val {
			s += string(b.([]byte))
		}
	case string:
		s = val
	default:
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// makeEscape creates an Escape node from an escape code character.
// Java has additional escapes compared to JavaScript:
// - \a (bell), \e (escape)
// - \h/\H (horizontal whitespace)
// - \v/\V (vertical whitespace - NOTE: different from JS \v which is vertical tab!)
// - \R (linebreak)
// - \X (grapheme cluster)
func makeEscape(code string) *ast.Escape {
	escape := &ast.Escape{Code: code}

	switch code {
	// Standard character class escapes
	case "d":
		escape.EscapeType = "digit"
		escape.Value = "digit"
	case "D":
		escape.EscapeType = "non_digit"
		escape.Value = "non-digit"
	case "w":
		escape.EscapeType = "word"
		escape.Value = "word"
	case "W":
		escape.EscapeType = "non_word"
		escape.Value = "non-word"
	case "s":
		escape.EscapeType = "whitespace"
		escape.Value = "white space"
	case "S":
		escape.EscapeType = "non_whitespace"
		escape.Value = "non-white space"

	// Java-specific horizontal whitespace (space, tab, and other horizontal ws)
	case "h":
		escape.EscapeType = "horizontal_whitespace"
		escape.Value = "horizontal white space"
	case "H":
		escape.EscapeType = "non_horizontal_whitespace"
		escape.Value = "non-horizontal white space"

	// Java vertical whitespace (newline characters)
	// NOTE: In JavaScript \v is vertical tab, but in Java \v is vertical whitespace class
	case "v":
		escape.EscapeType = "vertical_whitespace"
		escape.Value = "vertical white space"
	case "V":
		escape.EscapeType = "non_vertical_whitespace"
		escape.Value = "non-vertical white space"

	// Control characters
	case "n":
		escape.EscapeType = "newline"
		escape.Value = "new line"
	case "r":
		escape.EscapeType = "carriage_return"
		escape.Value = "carriage return"
	case "t":
		escape.EscapeType = "tab"
		escape.Value = "tab"
	case "f":
		escape.EscapeType = "form_feed"
		escape.Value = "form feed"
	case "a":
		escape.EscapeType = "bell"
		escape.Value = "bell"
	case "e":
		escape.EscapeType = "escape_char"
		escape.Value = "escape"

	// Special matchers
	case "R":
		escape.EscapeType = "linebreak"
		escape.Value = "line break"
	case "X":
		escape.EscapeType = "grapheme"
		escape.Value = "grapheme cluster"

	default:
		escape.EscapeType = "literal"
		escape.Value = code
	}

	return escape
}

// makeAnchor creates an Anchor node from an anchor code.
// Java supports additional anchors: \A, \Z, \z, \G
func makeAnchor(code string) *ast.Anchor {
	switch code {
	case "b":
		return &ast.Anchor{AnchorType: ast.AnchorWordBoundary}
	case "B":
		return &ast.Anchor{AnchorType: ast.AnchorNonWordBoundary}
	case "A":
		return &ast.Anchor{AnchorType: ast.AnchorStringStart}
	case "Z":
		return &ast.Anchor{AnchorType: ast.AnchorStringEnd}
	case "z":
		return &ast.Anchor{AnchorType: ast.AnchorAbsoluteEnd}
	case "G":
		return &ast.Anchor{AnchorType: "end_of_previous_match"}
	default:
		return &ast.Anchor{AnchorType: code}
	}
}
