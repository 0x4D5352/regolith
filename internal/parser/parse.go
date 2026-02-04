package parser

import (
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/0x4d5352/regolith/internal/flavor/javascript"
)

// jsFlavor is the JavaScript flavor instance used for parsing.
var jsFlavor = &javascript.JavaScript{}

// ParseRegex parses a regex pattern using the JavaScript flavor and returns the AST.
// This function is maintained for backward compatibility.
// For explicit flavor selection, use the flavor package directly.
func ParseRegex(pattern string) (*ast.Regexp, error) {
	return jsFlavor.Parse(pattern)
}
