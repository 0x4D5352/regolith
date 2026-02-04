package posix_ere

import (
	"strconv"
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

// POSIX class labels for rendering
var POSIXClassLabels = map[string]string{
	"alnum":  "alphanumeric",
	"alpha":  "alphabetic",
	"blank":  "blank (space/tab)",
	"cntrl":  "control character",
	"digit":  "digit",
	"graph":  "visible character",
	"lower":  "lowercase",
	"print":  "printable",
	"punct":  "punctuation",
	"space":  "whitespace",
	"upper":  "uppercase",
	"xdigit": "hex digit",
}
