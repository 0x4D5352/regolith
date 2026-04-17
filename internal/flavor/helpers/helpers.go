// Package helpers provides shared PEG-action helpers used by multiple
// regex flavor parsers. Each flavor's generated parser calls these via
// a thin per-package wrapper so that the generated code keeps compiling
// without regenerating grammars.
package helpers

import (
	"fmt"
	"strconv"

	"github.com/0x4d5352/regolith/internal/ast"
)

// ParseInt converts a PEG match result to an int. The pigeon runtime
// yields either a single []byte, a []any of []byte chunks, or a string
// depending on grammar shape; non-matching types yield 0 so callers can
// treat them as "not an integer" without panicking.
func ParseInt(v any) int {
	s := collectString(v)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// GetString converts a PEG match result to its string form, joining
// []any byte-chunks when present. Returns "" for nil or unknown shapes.
func GetString(v any) string {
	if v == nil {
		return ""
	}
	return collectString(v)
}

// collectString is the shared tokenizer for the two shapes pigeon emits
// for text-valued match results. Extracting it avoids duplicating the
// same three-case switch in every flavor's helpers.go.
func collectString(v any) string {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case []any:
		// Pre-compute final length so we can build the string in a
		// single allocation instead of O(n) += concatenations.
		n := 0
		for _, b := range val {
			if bs, ok := b.([]byte); ok {
				n += len(bs)
			}
		}
		buf := make([]byte, 0, n)
		for _, b := range val {
			if bs, ok := b.([]byte); ok {
				buf = append(buf, bs...)
			}
		}
		return string(buf)
	case string:
		return val
	default:
		return ""
	}
}

// FinalizeParse wraps the (result, err) tuple returned by a flavor's
// generated Parse function, producing the uniform error-wrapping and
// type-assertion that every flavor previously open-coded.
//
// Callers use it via argument-passing:
//
//	return helpers.FinalizeParse(Parse("", []byte(pattern), opts...))
func FinalizeParse(result any, err error) (*ast.Regexp, error) {
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	regexp, ok := result.(*ast.Regexp)
	if !ok {
		return nil, fmt.Errorf("unexpected parse result type: %T", result)
	}
	return regexp, nil
}
