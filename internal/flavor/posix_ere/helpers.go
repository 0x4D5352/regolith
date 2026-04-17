package posix_ere

import "github.com/0x4d5352/regolith/internal/flavor/helpers"

func parseInt(v any) int { return helpers.ParseInt(v) }

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
