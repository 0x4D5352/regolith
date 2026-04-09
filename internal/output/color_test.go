package output

import (
	"io"
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/muesli/termenv"
)

func TestResolveColorProfile(t *testing.T) {
	tests := []struct {
		mode string
		want termenv.Profile
	}{
		{"never", termenv.Ascii},
		{"always", termenv.ANSI},
	}

	for _, tc := range tests {
		t.Run(tc.mode, func(t *testing.T) {
			got := ResolveColorProfile(tc.mode)
			if got != tc.want {
				t.Errorf("ResolveColorProfile(%q) = %v, want %v", tc.mode, got, tc.want)
			}
		})
	}
}

func TestSeverityColor(t *testing.T) {
	tests := []struct {
		sev  analyzer.Severity
		want termenv.ANSIColor
	}{
		{analyzer.SeverityCritical, termenv.ANSIColor(1)},
		{analyzer.SeverityError, termenv.ANSIColor(1)},
		{analyzer.SeverityWarning, termenv.ANSIColor(3)},
		{analyzer.SeverityInfo, termenv.ANSIColor(4)},
	}

	for _, tc := range tests {
		t.Run(tc.sev.String(), func(t *testing.T) {
			got := severityColor(tc.sev)
			if got != tc.want {
				t.Errorf("severityColor(%v) = %v, want %v", tc.sev, got, tc.want)
			}
		})
	}
}

func TestScalingColor(t *testing.T) {
	tests := []struct {
		scaling string
		want    termenv.ANSIColor
	}{
		{"exponential", termenv.ANSIColor(1)},
		{"superlinear", termenv.ANSIColor(3)},
		{"linear", termenv.ANSIColor(2)},
		{"unknown", termenv.ANSIColor(0)},
	}

	for _, tc := range tests {
		t.Run(tc.scaling, func(t *testing.T) {
			got := scalingColor(tc.scaling)
			if got != tc.want {
				t.Errorf("scalingColor(%q) = %v, want %v", tc.scaling, got, tc.want)
			}
		})
	}
}

func TestStyledOutputContainsANSI(t *testing.T) {
	withColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI))
	styled := withColor.String("ERRORS").Bold().Foreground(termenv.ANSIColor(1)).String()

	if !strings.Contains(styled, "\033[") {
		t.Errorf("expected ANSI escape codes in styled output, got: %q", styled)
	}
	if !strings.Contains(styled, "ERRORS") {
		t.Errorf("expected text content preserved, got: %q", styled)
	}
}

func TestStyledOutputPlainWithAscii(t *testing.T) {
	noColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
	styled := noColor.String("ERRORS").Bold().Foreground(termenv.ANSIColor(1)).String()

	if strings.Contains(styled, "\033[") {
		t.Errorf("expected no ANSI codes with Ascii profile, got: %q", styled)
	}
	if styled != "ERRORS" {
		t.Errorf("expected plain text %q, got: %q", "ERRORS", styled)
	}
}
