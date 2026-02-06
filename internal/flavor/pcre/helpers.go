package pcre

import (
	"strconv"
	"unicode"

	"github.com/0x4d5352/regolith/internal/ast"
)

// makeEscape creates an Escape node for a given escape code
func makeEscape(code string) *ast.Escape {
	escape := &ast.Escape{Code: code}

	switch code {
	// Character type escapes
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
		escape.Value = "whitespace"
	case "S":
		escape.EscapeType = "non_whitespace"
		escape.Value = "non-whitespace"
	case "h":
		escape.EscapeType = "horizontal_whitespace"
		escape.Value = "horizontal whitespace"
	case "H":
		escape.EscapeType = "non_horizontal_whitespace"
		escape.Value = "non-horizontal whitespace"
	case "v":
		escape.EscapeType = "vertical_whitespace"
		escape.Value = "vertical whitespace"
	case "V":
		escape.EscapeType = "non_vertical_whitespace"
		escape.Value = "non-vertical whitespace"
	case "N":
		escape.EscapeType = "non_newline"
		escape.Value = "non-newline"
	case "R":
		escape.EscapeType = "newline_sequence"
		escape.Value = "newline sequence"
	case "X":
		escape.EscapeType = "extended_grapheme"
		escape.Value = "extended grapheme cluster"

	// Control characters
	case "n":
		escape.EscapeType = "newline"
		escape.Value = "newline"
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
		escape.EscapeType = "alert"
		escape.Value = "alert (bell)"
	case "e":
		escape.EscapeType = "escape"
		escape.Value = "escape"

	default:
		escape.EscapeType = "literal"
		escape.Value = code
	}

	return escape
}

// makeAnchor creates an Anchor node for a given anchor code
func makeAnchor(code string) *ast.Anchor {
	anchor := &ast.Anchor{}

	switch code {
	case "b":
		anchor.AnchorType = ast.AnchorWordBoundary
	case "B":
		anchor.AnchorType = ast.AnchorNonWordBoundary
	case "A":
		anchor.AnchorType = ast.AnchorStringStart
	case "Z":
		anchor.AnchorType = ast.AnchorStringEnd
	case "z":
		anchor.AnchorType = ast.AnchorAbsoluteEnd
	case "G":
		anchor.AnchorType = "first_match_position"
	case "K":
		anchor.AnchorType = "reset_match_start"
	default:
		anchor.AnchorType = code
	}

	return anchor
}

// parseInt parses an interface{} containing digits into an int
func parseInt(v any) int {
	switch val := v.(type) {
	case []byte:
		n, _ := strconv.Atoi(string(val))
		return n
	case []any:
		str := ""
		for _, b := range val {
			if bs, ok := b.([]byte); ok {
				str += string(bs)
			}
		}
		n, _ := strconv.Atoi(str)
		return n
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}

// getString converts an interface{} to string
func getString(v any) string {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case []any:
		str := ""
		for _, b := range val {
			if bs, ok := b.([]byte); ok {
				str += string(bs)
			}
		}
		return str
	case string:
		return val
	default:
		return ""
	}
}

// isDigits checks if a string contains only digits
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
