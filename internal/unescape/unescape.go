package unescape

import (
	"strings"
	"unicode/utf8"
)

// JavaStringLiteral processes a string as if it were the contents of a Java/C#
// string literal, converting escape sequences to their actual characters.
// Regex-specific escapes like \d, \w, \s are passed through unchanged so the
// result can be fed directly to a regex parser.
func JavaStringLiteral(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			i++
			continue
		}

		// We have a backslash followed by at least one character.
		next := s[i+1]
		switch next {
		case '\\':
			b.WriteByte('\\')
			i += 2
		case '"':
			b.WriteByte('"')
			i += 2
		case '\'':
			b.WriteByte('\'')
			i += 2
		case 'n':
			b.WriteByte('\n')
			i += 2
		case 't':
			b.WriteByte('\t')
			i += 2
		case 'r':
			b.WriteByte('\r')
			i += 2
		case 'b':
			b.WriteByte('\b')
			i += 2
		case 'f':
			b.WriteByte('\f')
			i += 2
		case 'u':
			// \uXXXX - exactly 4 hex digits required
			if i+5 < len(s) && isHexRun(s[i+2:i+6]) {
				r := hexToRune(s[i+2 : i+6])
				b.WriteRune(r)
				i += 6
			} else {
				// Not enough hex digits; pass through unchanged
				b.WriteByte('\\')
				b.WriteByte('u')
				i += 2
			}
		case '0':
			// Octal: \0 followed by up to 3 octal digits, value <= 255
			j := i + 2
			val := 0
			digits := 0
			for j < len(s) && digits < 3 && s[j] >= '0' && s[j] <= '7' {
				newVal := val*8 + int(s[j]-'0')
				if newVal > 255 {
					break
				}
				val = newVal
				digits++
				j++
			}
			b.WriteByte(byte(val))
			i = j
		default:
			// Unknown escape: pass through unchanged (preserves \d, \w, \s, etc.)
			b.WriteByte('\\')
			b.WriteByte(next)
			i += 2
		}
	}

	return b.String()
}

// ContainsDoubleEscapes reports whether s contains a double backslash (\\),
// which may indicate the pattern was copied from a string literal.
func ContainsDoubleEscapes(s string) bool {
	return strings.Contains(s, `\\`)
}

func isHexRun(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isHexDigit(s[i]) {
			return false
		}
	}
	return true
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hexToRune(s string) rune {
	var val rune
	for i := 0; i < len(s); i++ {
		val <<= 4
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			val |= rune(c - '0')
		case c >= 'a' && c <= 'f':
			val |= rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val |= rune(c-'A') + 10
		}
	}
	if !utf8.ValidRune(val) {
		val = utf8.RuneError
	}
	return val
}
