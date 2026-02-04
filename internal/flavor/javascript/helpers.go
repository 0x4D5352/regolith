package javascript

import (
	"strconv"

	"github.com/0x4d5352/regolith/internal/ast"
)

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

// makeEscape creates an Escape node from an escape code character
func makeEscape(code string) *ast.Escape {
	escape := &ast.Escape{Code: code}

	switch code {
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
	case "b":
		escape.EscapeType = "word_boundary"
		escape.Value = "word boundary"
	case "B":
		escape.EscapeType = "non_word_boundary"
		escape.Value = "non-word boundary"
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
	case "v":
		escape.EscapeType = "vertical_tab"
		escape.Value = "vertical tab"
	default:
		escape.EscapeType = "literal"
		escape.Value = code
	}

	return escape
}
