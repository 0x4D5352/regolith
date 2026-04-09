package output

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/ast"
	"github.com/muesli/termenv"
)

func TestRenderAnalysisText(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: ".*.*=.*",
		Flavor:  "javascript",
		Findings: []*analyzer.Finding{
			{
				ID: "adjacent-unbounded", Category: analyzer.CategoryBacktracking,
				Severity: analyzer.SeverityError, Title: "Adjacent unbounded quantifiers",
				Description: "Adjacent unbounded quantifiers cause catastrophic backtracking.",
				Node:        &ast.AnyCharacter{},
			},
			{
				ID: "trailing-wildcard", Category: analyzer.CategoryRedundancy,
				Severity: analyzer.SeverityInfo, Title: "Trailing .* without anchor",
				Description: "Trailing .* is redundant in search mode.",
				Node:        &ast.AnyCharacter{},
			},
		},
	}

	noColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
	got := RenderAnalysisText(report, false, noColor)

	if !strings.Contains(got, "ERRORS") {
		t.Error("expected ERRORS section")
	}
	if !strings.Contains(got, "adjacent-unbounded") {
		t.Error("expected adjacent-unbounded finding")
	}
	if !strings.Contains(got, "trailing-wildcard") {
		t.Error("expected trailing-wildcard finding")
	}
}

func TestRenderAnalysisTextMarkdown(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: ".*.*=.*",
		Flavor:  "javascript",
		Findings: []*analyzer.Finding{
			{
				ID: "adjacent-unbounded", Category: analyzer.CategoryBacktracking,
				Severity: analyzer.SeverityError, Title: "Adjacent unbounded quantifiers",
				Node: &ast.AnyCharacter{},
			},
		},
	}

	noColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
	got := RenderAnalysisText(report, true, noColor)

	if !strings.Contains(got, "# Analysis:") {
		t.Error("expected markdown header")
	}
}

func TestRenderAnalysisTextWithBenchmark(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: "hello", Flavor: "javascript",
		Findings: nil,
		BenchmarkSummary: &analyzer.BenchmarkSummary{
			Engine: "regexp2",
			Corpus: map[string]map[int]time.Duration{
				"repeated": {10: 100 * time.Microsecond, 100: 1 * time.Millisecond},
			},
		},
	}

	noColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.Ascii))
	got := RenderAnalysisText(report, false, noColor)

	if !strings.Contains(got, "Benchmark") {
		t.Error("expected Benchmark section")
	}
	if !strings.Contains(got, "regexp2") {
		t.Error("expected engine name")
	}
}

func TestRenderAnalysisTextWithColor(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: ".*.*=.*",
		Flavor:  "javascript",
		Findings: []*analyzer.Finding{
			{
				ID: "adjacent-unbounded", Category: analyzer.CategoryBacktracking,
				Severity: analyzer.SeverityError, Title: "Adjacent unbounded quantifiers",
				Description: "Catastrophic backtracking.",
				Suggestion:  "Fix it.",
				Node:        &ast.AnyCharacter{},
			},
			{
				ID: "no-anchors", Category: analyzer.CategoryPerformance,
				Severity: analyzer.SeverityInfo, Title: "Pattern has no anchors",
				Node: &ast.AnyCharacter{},
			},
		},
	}

	withColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI))
	got := RenderAnalysisText(report, false, withColor)

	// Severity headers should contain ANSI escape codes.
	if !strings.Contains(got, "\033[") {
		t.Error("expected ANSI escape codes in colored output")
	}

	// Content should still be present.
	if !strings.Contains(got, "ERRORS") {
		t.Error("expected ERRORS section")
	}
	if !strings.Contains(got, "adjacent-unbounded") {
		t.Error("expected finding ID")
	}
}

func TestRenderAnalysisTextMarkdownIgnoresColor(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: ".*",
		Flavor:  "javascript",
		Findings: []*analyzer.Finding{
			{
				ID: "trailing-wildcard", Category: analyzer.CategoryRedundancy,
				Severity: analyzer.SeverityInfo, Title: "Trailing .* without anchor",
				Node: &ast.AnyCharacter{},
			},
		},
	}

	withColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI))
	got := RenderAnalysisText(report, true, withColor)

	if strings.Contains(got, "\033[") {
		t.Error("expected no ANSI codes in markdown output even with ANSI profile")
	}
}

func TestRenderAnalysisTextBenchmarkWithColor(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: "hello", Flavor: "javascript",
		Findings: nil,
		BenchmarkSummary: &analyzer.BenchmarkSummary{
			Engine: "regexp2",
			Corpus: map[string]map[int]time.Duration{
				"repeated": {10: 100 * time.Microsecond, 100: -1},
			},
		},
	}

	withColor := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI))
	got := RenderAnalysisText(report, false, withColor)

	// "timeout" should be styled.
	if !strings.Contains(got, "\033[") {
		t.Error("expected ANSI codes in benchmark output")
	}
	if !strings.Contains(got, "timeout") {
		t.Error("expected timeout label")
	}
}
