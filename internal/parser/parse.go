package parser

import "fmt"

// ParseRegex parses a regex pattern and returns the AST
func ParseRegex(pattern string) (*Regexp, error) {
	state := NewParserState()

	result, err := Parse("", []byte(pattern), GlobalStore("state", state))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	regexp, ok := result.(*Regexp)
	if !ok {
		return nil, fmt.Errorf("unexpected parse result type: %T", result)
	}

	return regexp, nil
}
