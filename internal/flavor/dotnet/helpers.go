package dotnet

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor/helpers"
)

func getString(v any) string { return helpers.GetString(v) }
func parseInt(v any) int     { return helpers.ParseInt(v) }

// makeEscape creates an Escape node from an escape code character.
// .NET escape sequences:
// - \d, \D, \w, \W, \s, \S - standard character classes
// - \n, \r, \t, \f - standard control characters
// - \a (bell), \e (escape), \v (vertical tab)
// Note: Unlike Java, .NET \v is vertical tab, not vertical whitespace class
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
	case "v":
		// In .NET, \v is vertical tab (unlike Java where it's vertical whitespace class)
		escape.EscapeType = "vertical_tab"
		escape.Value = "vertical tab"

	default:
		escape.EscapeType = "literal"
		escape.Value = code
	}

	return escape
}

// makeAnchor creates an Anchor node from an anchor code.
// .NET supports: \b (word boundary), \B (non-word boundary), \A (start), \Z (end before \n), \z (absolute end)
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
	default:
		return &ast.Anchor{AnchorType: code}
	}
}
