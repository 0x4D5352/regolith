package output

// RenderText is the unified entry point for human-readable AST
// walks. It produces either Markdown (for writing to files) or an
// ANSI-styled terminal representation (for printing to stdout).
//
// The walker itself lives in markdown.go and always produces Markdown;
// ANSI mode is a post-pass that restyles the Markdown header and bold
// runs with termenv escape codes. Keeping a single walker means the two
// modes can never drift — the existing markdown_test.go golden fixtures
// act as a regression fence for the shared code path.

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"

	"github.com/0x4d5352/regolith/internal/ast"
)

// RenderText walks a Regexp AST and returns a human-readable
// description.
//
//   - markdown=true  → returns the Markdown outline produced by
//     RenderMarkdown. co is ignored.
//   - markdown=false → returns an ANSI-styled walk suitable for
//     writing to a terminal. co supplies the termenv profile
//     (Ascii / ANSI / ANSI256 / TrueColor) that decides whether
//     the output is actually colorized.
func RenderText(root *ast.Regexp, pattern, flavorName string, markdown bool, co *termenv.Output) string {
	md := RenderMarkdown(root, pattern, flavorName)
	if markdown {
		return md
	}
	return markdownToANSI(md, pattern, flavorName, co)
}

// markdownToANSI rewrites the Markdown output produced by RenderMarkdown
// into ANSI-styled text. The rewrite is deliberately narrow:
//
//  1. The document header (`# Regex: …` + `**Flavor:** …`) is replaced
//     with a compact single-line banner, matching the convention used
//     by RenderAnalysisText.
//  2. Inline `**bold**` runs are replaced with termenv bold.
//  3. Inline “ `code` “ runs are replaced with termenv faint to
//     distinguish literals and escape codes from the surrounding prose.
//
// The `- ` bullet prefix that RenderMarkdown emits for every line is
// preserved unchanged — it reads as an outline in any terminal.
func markdownToANSI(md, pattern, flavorName string, co *termenv.Output) string {
	header := co.String(fmt.Sprintf("Regex: %s", pattern)).Bold().String()
	flavorLine := fmt.Sprintf("%s %s",
		co.String("Flavor:").Bold().String(),
		formatFlavorName(flavorName))

	body := stripMarkdownHeader(md)
	body = replaceBoldRuns(body, co)
	body = replaceCodeRuns(body, co)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(flavorLine)
	sb.WriteString("\n\n")
	sb.WriteString(body)
	return sb.String()
}

// stripMarkdownHeader removes the leading `# Regex: …\n\n**Flavor:** …\n\n`
// block from RenderMarkdown's output so markdownToANSI can replace it
// with its own banner. Two blank lines always separate the header from
// the body, so we can locate the split deterministically.
func stripMarkdownHeader(md string) string {
	const sep = "\n\n"
	// First `\n\n` ends the `# Regex: ...` line. Second ends `**Flavor:** ...`.
	_, rest, ok := strings.Cut(md, sep)
	if !ok {
		return md
	}
	_, body, ok := strings.Cut(rest, sep)
	if !ok {
		return rest
	}
	return body
}

// replaceBoldRuns scans src for `**…**` spans and wraps their contents
// in termenv bold. Bold spans cannot nest in the walker's output, so a
// simple forward scan is sufficient.
func replaceBoldRuns(src string, co *termenv.Output) string {
	return scanReplace(src, "**", func(inner string) string {
		return co.String(inner).Bold().String()
	})
}

// replaceCodeRuns scans src for “ `…` “ spans and renders their
// contents with termenv faint. This dims literals and escape codes so
// the eye can separate them from the surrounding descriptive text
// without making them unreadable on dark backgrounds.
func replaceCodeRuns(src string, co *termenv.Output) string {
	return scanReplace(src, "`", func(inner string) string {
		return co.String(inner).Faint().String()
	})
}

// scanReplace walks src, locating matched pairs of delim. Each span's
// inner text is passed through wrap and the result is substituted in
// place. Unbalanced final delimiters are left as-is, which preserves
// any oddities in the walker's output rather than panicking.
func scanReplace(src, delim string, wrap func(string) string) string {
	var sb strings.Builder
	i := 0
	for {
		start := strings.Index(src[i:], delim)
		if start < 0 {
			sb.WriteString(src[i:])
			return sb.String()
		}
		absStart := i + start
		sb.WriteString(src[i:absStart])
		innerStart := absStart + len(delim)
		end := strings.Index(src[innerStart:], delim)
		if end < 0 {
			sb.WriteString(src[absStart:])
			return sb.String()
		}
		absEnd := innerStart + end
		inner := src[innerStart:absEnd]
		sb.WriteString(wrap(inner))
		i = absEnd + len(delim)
	}
}
