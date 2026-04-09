package output

import (
	"encoding/json"
	"testing"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/ast"
)

func TestRenderAnalysisJSON(t *testing.T) {
	report := &analyzer.AnalysisReport{
		Pattern: ".*.*=.*",
		Flavor:  "javascript",
		Findings: []*analyzer.Finding{
			{
				ID: "adjacent-unbounded", Category: analyzer.CategoryBacktracking,
				Severity: analyzer.SeverityError, Title: "Adjacent unbounded quantifiers",
				Description: "Bad for performance.", Suggestion: "Combine into one.",
				Node: &ast.AnyCharacter{},
			},
		},
	}

	got, err := RenderAnalysisJSON(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, got)
	}

	if parsed["pattern"] != ".*.*=.*" {
		t.Errorf("expected pattern .*.*=.*, got %v", parsed["pattern"])
	}

	findings, ok := parsed["findings"].([]any)
	if !ok || len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %v", parsed["findings"])
	}

	finding := findings[0].(map[string]any)
	if finding["id"] != "adjacent-unbounded" {
		t.Errorf("expected id adjacent-unbounded, got %v", finding["id"])
	}
	if finding["severity"] != "error" {
		t.Errorf("expected severity error, got %v", finding["severity"])
	}
}
