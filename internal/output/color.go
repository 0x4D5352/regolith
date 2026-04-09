package output

import (
	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/muesli/termenv"
)

// ResolveColorProfile maps a --color flag value ("auto", "always",
// "never") to the corresponding termenv Profile. "auto" detects
// terminal capabilities from the environment.
func ResolveColorProfile(mode string) termenv.Profile {
	switch mode {
	case "always":
		return termenv.ANSI
	case "never":
		return termenv.Ascii
	default:
		return termenv.ColorProfile()
	}
}

// severityColor returns the ANSI color for a finding severity level:
// red for critical/error, yellow for warning, blue for info.
func severityColor(sev analyzer.Severity) termenv.ANSIColor {
	switch sev {
	case analyzer.SeverityCritical, analyzer.SeverityError:
		return termenv.ANSIColor(1) // red
	case analyzer.SeverityWarning:
		return termenv.ANSIColor(3) // yellow
	default:
		return termenv.ANSIColor(4) // blue
	}
}

// scalingColor returns the ANSI color for a benchmark scaling
// classification: red for exponential, yellow for superlinear,
// green for linear, default for anything else.
func scalingColor(scaling string) termenv.ANSIColor {
	switch scaling {
	case "exponential":
		return termenv.ANSIColor(1) // red
	case "superlinear":
		return termenv.ANSIColor(3) // yellow
	case "linear":
		return termenv.ANSIColor(2) // green
	default:
		return termenv.ANSIColor(0) // default
	}
}
