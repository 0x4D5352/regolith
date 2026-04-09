package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/muesli/termenv"
)

// RenderAnalysisText formats an AnalysisReport as human-readable text.
// When markdown is true, the output uses Markdown formatting (headers,
// bold, code blocks) suitable for writing to a .md file.
//
// co controls whether ANSI escape codes are emitted. Pass a no-color
// output (Ascii profile) for plain text, or a real terminal output for
// colored output. Markdown mode always forces the Ascii profile.
func RenderAnalysisText(report *analyzer.AnalysisReport, markdown bool, co *termenv.Output) string {
	// Markdown output is meant for files — never emit ANSI codes.
	if markdown {
		co = termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
	}

	var sb strings.Builder

	if markdown {
		fmt.Fprintf(&sb, "# Analysis: `%s`\n\n", report.Pattern)
		fmt.Fprintf(&sb, "**Flavor:** %s\n\n", formatFlavorName(report.Flavor))
	} else {
		fmt.Fprintf(&sb, "Analysis: %s  (%s)\n\n", report.Pattern, report.Flavor)
	}

	if len(report.Findings) == 0 {
		msg := co.String("No issues found.").Foreground(termenv.ANSIColor(2)).String()
		sb.WriteString(msg + "\n")
	} else {
		renderFindings(&sb, report.Findings, markdown, co)
	}

	if report.BenchmarkSummary != nil {
		sb.WriteString("\n")
		renderBenchmarkText(&sb, report.BenchmarkSummary, markdown, co)
	}

	return sb.String()
}

// renderFindings groups findings by severity (Critical → Error → Warning →
// Info) and writes each non-empty group with a labeled header.
func renderFindings(sb *strings.Builder, findings []*analyzer.Finding, markdown bool, co *termenv.Output) {
	// Group by severity (highest first).
	groups := map[analyzer.Severity][]*analyzer.Finding{}
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}

	severityOrder := []analyzer.Severity{
		analyzer.SeverityCritical,
		analyzer.SeverityError,
		analyzer.SeverityWarning,
		analyzer.SeverityInfo,
	}

	severityLabel := map[analyzer.Severity]string{
		analyzer.SeverityCritical: "CRITICAL",
		analyzer.SeverityError:    "ERRORS",
		analyzer.SeverityWarning:  "WARNINGS",
		analyzer.SeverityInfo:     "INFO",
	}

	for _, sev := range severityOrder {
		group := groups[sev]
		if len(group) == 0 {
			continue
		}

		label := severityLabel[sev]
		if markdown {
			fmt.Fprintf(sb, "## %s (%d)\n\n", label, len(group))
		} else {
			styledLabel := co.String(label).Bold().Foreground(severityColor(sev)).String()
			fmt.Fprintf(sb, "%s (%d):\n", styledLabel, len(group))
		}

		for _, f := range group {
			if markdown {
				fmt.Fprintf(sb, "- **[%s]** %s\n", f.ID, f.Title)
				if f.Description != "" {
					fmt.Fprintf(sb, "  %s\n", f.Description)
				}
				if f.Suggestion != "" {
					fmt.Fprintf(sb, "  *Suggestion:* %s\n", f.Suggestion)
				}
			} else {
				styledID := co.String("[" + f.ID + "]").Faint().String()
				styledTitle := co.String(f.Title).Bold().String()
				fmt.Fprintf(sb, "  %s %s\n", styledID, styledTitle)
				if f.Description != "" {
					fmt.Fprintf(sb, "    %s\n", f.Description)
				}
				if f.Suggestion != "" {
					styledSugg := co.String("Suggestion: " + f.Suggestion).Faint().Italic().String()
					fmt.Fprintf(sb, "    %s\n", styledSugg)
				}
			}
			sb.WriteString("\n")
		}
	}
}

// renderBenchmarkText writes the benchmark summary section, including per-corpus
// timing tables with scaling classification derived via ClassifyScaling.
func renderBenchmarkText(sb *strings.Builder, summary *analyzer.BenchmarkSummary, markdown bool, co *termenv.Output) {
	engineLabel := summary.Engine
	if summary.IsFallback {
		engineLabel += " (fallback)"
	}

	if markdown {
		fmt.Fprintf(sb, "## Benchmark (%s)\n\n", engineLabel)
	} else {
		styledHeader := co.String(fmt.Sprintf("--- Benchmark (%s) ---", engineLabel)).Bold().String()
		fmt.Fprintf(sb, "%s\n", styledHeader)
	}

	// Sort corpus names for stable output.
	corpusNames := make([]string, 0, len(summary.Corpus))
	for name := range summary.Corpus {
		corpusNames = append(corpusNames, name)
	}
	sort.Strings(corpusNames)

	for _, name := range corpusNames {
		results := summary.Corpus[name]

		if markdown {
			fmt.Fprintf(sb, "### %s\n\n", name)
		} else {
			fmt.Fprintf(sb, "  Corpus: %s\n", name)
		}

		// Sort sizes ascending so the table reads smallest → largest.
		sizes := make([]int, 0, len(results))
		for size := range results {
			sizes = append(sizes, size)
		}
		sort.Ints(sizes)

		for _, size := range sizes {
			dur := results[size]
			// Negative durations are timeout sentinels (see benchmark.go).
			if dur < 0 {
				if markdown {
					fmt.Fprintf(sb, "- %d chars: **timeout**\n", size)
				} else {
					styledTimeout := co.String("timeout").Bold().Foreground(termenv.ANSIColor(1)).String()
					fmt.Fprintf(sb, "    %-12s %s\n", fmt.Sprintf("%d chars:", size), styledTimeout)
				}
			} else {
				if markdown {
					fmt.Fprintf(sb, "- %d chars: %s\n", size, formatDuration(dur))
				} else {
					fmt.Fprintf(sb, "    %-12s %s\n", fmt.Sprintf("%d chars:", size), formatDuration(dur))
				}
			}
		}

		scaling := analyzer.ClassifyScaling(results)
		if scaling != "unknown" {
			label := strings.ToUpper(scaling)
			if markdown {
				fmt.Fprintf(sb, "\n*Scaling:* **%s**\n\n", label)
			} else {
				styledScaling := co.String(label).Bold().Foreground(scalingColor(scaling)).String()
				fmt.Fprintf(sb, "    Scaling: %s\n", styledScaling)
			}
		}
		sb.WriteString("\n")
	}
}

// formatDuration produces a human-friendly duration string, picking units
// appropriate to the magnitude (ns, µs, ms, s).
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/1000)
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
