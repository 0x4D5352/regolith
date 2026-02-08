package unescape

import "testing"

func TestJavaStringLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Primary use case: double-escaped regex patterns
		{name: "double backslash", input: `\\`, want: `\`},
		{name: "escaped dot", input: `\\.`, want: `\.`},
		{name: "full pattern", input: `^/\\.[^/]+/xinetd$`, want: `^/\.[^/]+/xinetd$`},
		{name: "double-double backslash", input: `\\\\`, want: `\\`},

		// Standard string escapes
		{name: "newline", input: `\n`, want: "\n"},
		{name: "tab", input: `\t`, want: "\t"},
		{name: "carriage return", input: `\r`, want: "\r"},
		{name: "backspace", input: `\b`, want: "\b"},
		{name: "form feed", input: `\f`, want: "\f"},
		{name: "double quote", input: `\"`, want: `"`},
		{name: "single quote", input: `\'`, want: `'`},

		// Unicode escapes
		{name: "unicode A", input: `\u0041`, want: "A"},
		{name: "unicode emoji", input: `\u00E9`, want: "\u00E9"},
		{name: "unicode too few digits", input: `\u04`, want: `\u04`},
		{name: "unicode at end", input: `\u`, want: `\u`},

		// Octal escapes
		{name: "octal A", input: `\0101`, want: "A"},
		{name: "octal NUL", input: `\0`, want: "\x00"},
		{name: "octal max 255", input: `\0377`, want: "\xff"},
		{name: "octal partial", input: `\07`, want: "\x07"},

		// Regex escapes pass through
		{name: "digit escape", input: `\d`, want: `\d`},
		{name: "word escape", input: `\w`, want: `\w`},
		{name: "space escape", input: `\s`, want: `\s`},
		{name: "boundary escape", input: `\B`, want: `\B`},

		// Edge cases
		{name: "empty", input: "", want: ""},
		{name: "no backslashes", input: "abc", want: "abc"},
		{name: "trailing backslash", input: `abc\`, want: `abc\`},
		{name: "mixed", input: `\\d+\\.\\d+`, want: `\d+\.\d+`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := JavaStringLiteral(tc.input)
			if got != tc.want {
				t.Errorf("JavaStringLiteral(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestContainsDoubleEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "has double backslash", input: `\\d`, want: true},
		{name: "has double backslash in pattern", input: `^/\\.[^/]+$`, want: true},
		{name: "single backslash", input: `\d`, want: false},
		{name: "no backslash", input: "abc", want: false},
		{name: "empty", input: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsDoubleEscapes(tc.input)
			if got != tc.want {
				t.Errorf("ContainsDoubleEscapes(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
