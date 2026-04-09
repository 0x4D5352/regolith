package output

import (
	"strings"
	"testing"
	"time"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/ast"
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

	got := RenderAnalysisText(report, false)

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

	got := RenderAnalysisText(report, true)

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

	got := RenderAnalysisText(report, false)

	if !strings.Contains(got, "Benchmark") {
		t.Error("expected Benchmark section")
	}
	if !strings.Contains(got, "regexp2") {
		t.Error("expected engine name")
	}
}
