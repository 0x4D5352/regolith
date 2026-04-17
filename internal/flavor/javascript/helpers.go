package javascript

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor/helpers"
)

// parseInt is referenced by name from the generated parser, so we keep
// a package-local alias that delegates to the shared implementation.
func parseInt(v any) int { return helpers.ParseInt(v) }

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
