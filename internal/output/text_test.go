package output

import (
	"io"
	"strings"
	"testing"

	"github.com/muesli/termenv"

	"github.com/0x4d5352/regolith/internal/ast"
)

// sampleRegexp builds a small AST with enough variety to exercise
// headers, bold spans, and inline code spans in RenderText's output.
func sampleRegexp() *ast.Regexp {
	return &ast.Regexp{
		Matches: []*ast.Match{
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Literal{Text: "foo"}},
				{Content: &ast.Escape{EscapeType: "digit", Code: "d"}},
			}},
			{Fragments: []*ast.MatchFragment{
				{Content: &ast.Anchor{AnchorType: ast.AnchorStart}},
			}},
		},
	}
}

func ansiOutput() *termenv.Output {
	return termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI))
}

func plainOutput() *termenv.Output {
	return termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
}

// TestRenderTextMarkdownMatchesLegacy is the regression fence for the
// shared walker: as long as RenderText(..., markdown=true, ...) byte-
// matches RenderMarkdown for a representative pattern, the two modes
// cannot drift. Broken walker tests will surface in markdown_test.go;
// this test catches the narrower "did the text wrapper bypass the
// walker" failure mode.
func TestRenderTextMarkdownMatchesLegacy(t *testing.T) {
	root := sampleRegexp()
	want := RenderMarkdown(root, "foo\\d|^", "javascript")
	got := RenderText(root, "foo\\d|^", "javascript", true, nil)
	if got != want {
		t.Errorf("RenderText markdown mode drifted from RenderMarkdown\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRenderTextANSIContainsEscapeCodes(t *testing.T) {
	got := RenderText(sampleRegexp(), "foo\\d|^", "javascript", false, ansiOutput())
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape codes in output, got:\n%s", got)
	}
	if strings.HasPrefix(got, "# Regex:") {
		t.Errorf("expected Markdown header to be stripped in ANSI mode, got prefix:\n%s", got[:minLen(40, len(got))])
	}
	if !strings.Contains(got, "Regex: foo\\d|^") {
		t.Errorf("expected compact banner 'Regex: <pattern>', got:\n%s", got)
	}
}

func TestRenderTextAsciiProfileStripsEscapes(t *testing.T) {
	got := RenderText(sampleRegexp(), "foo\\d|^", "javascript", false, plainOutput())
	if strings.Contains(got, "\x1b[") {
		t.Errorf("expected no ANSI codes under Ascii profile, got:\n%s", got)
	}
	if !strings.Contains(got, "Regex: foo\\d|^") {
		t.Errorf("expected banner present under Ascii profile, got:\n%s", got)
	}
	if strings.Contains(got, "**") {
		t.Errorf("expected `**` bold runs to be rewritten under Ascii profile (termenv drops codes), got:\n%s", got)
	}
}

func TestRenderTextANSIStripsBoldMarkers(t *testing.T) {
	got := RenderText(sampleRegexp(), "foo\\d|^", "javascript", false, ansiOutput())
	if strings.Contains(got, "**") {
		t.Errorf("expected `**` markers to be replaced with ANSI bold, got:\n%s", got)
	}
}

func TestScanReplaceUnbalancedDelim(t *testing.T) {
	// Scanner must not panic on an unterminated delimiter; it leaves
	// the trailing prefix alone and returns the string otherwise
	// unchanged. This guards against broken walker output — we want
	// a degraded render, not a crash.
	in := "ok **bold** but **unterminated"
	got := scanReplace(in, "**", func(s string) string { return "<" + s + ">" })
	want := "ok <bold> but **unterminated"
	if got != want {
		t.Errorf("scanReplace unbalanced: got %q, want %q", got, want)
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
