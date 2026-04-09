package analyzer

import (
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor"
	_ "github.com/0x4d5352/regolith/internal/flavor/javascript"
)

func TestAnalyze(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	tests := []struct {
		name         string
		pattern      string
		wantIDs      []string
		wantMinCount int
	}{
		{
			name:         ".*.*=.* has adjacent-unbounded and trailing-wildcard",
			pattern:      ".*.*=.*",
			wantIDs:      []string{"adjacent-unbounded", "trailing-wildcard"},
			wantMinCount: 2,
		},
		{
			name:         "(a+)+ has nested-quantifier",
			pattern:      "(a+)+",
			wantIDs:      []string{"nested-quantifier"},
			wantMinCount: 1,
		},
		{
			// Use a fully-anchored literal so that missing-anchor is not raised,
			// allowing us to assert zero findings for a structurally clean pattern.
			name:         "anchored literal has no backtracking or correctness findings",
			pattern:      "^hello$",
			wantIDs:      nil,
			wantMinCount: 0,
		},
		{
			name:         "empty alternative detected",
			pattern:      "a|",
			wantIDs:      []string{"empty-alternative"},
			wantMinCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := f.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			report := Analyze(parsed, tc.pattern, "javascript", f.SupportedFeatures())
			if len(report.Findings) < tc.wantMinCount {
				t.Errorf("got %d findings, want at least %d", len(report.Findings), tc.wantMinCount)
			}
			for _, wantID := range tc.wantIDs {
				found := false
				for _, finding := range report.Findings {
					if finding.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected finding %q not found in %v", wantID, findingIDs(report.Findings))
				}
			}
		})
	}
}

func findingIDs(findings []*Finding) []string {
	ids := make([]string, len(findings))
	for i, f := range findings {
		ids[i] = f.ID
	}
	return ids
}
