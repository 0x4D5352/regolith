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
		{
			// (a{3}){5} is bounded on both axes — the fixed nested-quantifier
			// rule should NOT fire on this shape.
			name:         "bounded nested quantifier is not flagged",
			pattern:      "^(a{3}){5}$",
			wantIDs:      nil,
			wantMinCount: 0,
		},
		{
			// Useless capture: group defined, no backreference uses it.
			name:         "capturing group never referenced",
			pattern:      "^(foo)$",
			wantIDs:      []string{"useless-capture"},
			wantMinCount: 1,
		},
		{
			// Backreference targets a group that doesn't exist.
			name:         "invalid backreference detected",
			pattern:      "^(foo)\\2$",
			wantIDs:      []string{"invalid-backreference"},
			wantMinCount: 1,
		},
		{
			// Three consecutive \d triggers repeated-token.
			name:         "repeated single token detected",
			pattern:      "^\\d\\d\\d$",
			wantIDs:      []string{"repeated-token"},
			wantMinCount: 1,
		},
		{
			name:         "redundant {1} quantifier",
			pattern:      "^a{1}b$",
			wantIDs:      []string{"redundant-bounded-quantifier"},
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

// countByID returns how many findings carry the given rule ID.
func countByID(findings []*Finding, id string) int {
	n := 0
	for _, f := range findings {
		if f.ID == id {
			n++
		}
	}
	return n
}

// TestAnalyzeNoDuplicateGlobalFindings asserts that pattern-level rules fire
// exactly once per pattern, regardless of how many nested groups exist. This
// is a regression test for the bug where checkMissingAnchor ran at every
// walkRegexp call and produced one finding per Subexp.
func TestAnalyzeNoDuplicateGlobalFindings(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	// Patterns taken verbatim from test_cases.txt. They all lack anchors
	// and all contain several nested groups — prime conditions for the
	// duplicate-finding bug.
	patterns := []struct {
		name    string
		pattern string
	}{
		{"ipv4", `(?:(?:\d|[01]?\d\d|2[0-4]\d|25[0-5])\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d|\d)(?:\/\d{1,2})?`},
		{"mac", `[A-Fa-f\d]{2}(?:[:-][A-Fa-f\d]{2}){5}`},
		{"uuid", `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`},
		{"iso8601", `((?:19|20)\d\d)[- /.](0[1-9]|1[012])[- /.](0[1-9]|[12][0-9]|3[01])`},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := f.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			report := Analyze(parsed, tc.pattern, "javascript", f.SupportedFeatures())

			// missing-anchor should fire at most once. The UUID pattern
			// contains \b (word-boundary anchors), so it shouldn't fire
			// at all; the others have no anchors so it should fire once.
			if got := countByID(report.Findings, "missing-anchor"); got > 1 {
				t.Errorf("missing-anchor fired %d times, want at most 1", got)
			}

			// The nested-quantifier rule should NOT fire on bounded-only
			// nestings (MAC address, IPV4 non-capturing iterations).
			if tc.name == "mac" || tc.name == "ipv4" {
				if got := countByID(report.Findings, "nested-quantifier"); got > 0 {
					t.Errorf("nested-quantifier fired %d times on %s, want 0", got, tc.name)
				}
			}
		})
	}
}

// TestAnalyzeUUIDWordBoundariesSuppressMissingAnchor verifies that \b counts
// as an anchor for the missing-anchor rule — the UUID pattern uses \b between
// hex blocks and should not trigger the rule at all.
func TestAnalyzeUUIDWordBoundariesSuppressMissingAnchor(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}
	pattern := `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}`
	parsed, err := f.Parse(pattern)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	report := Analyze(parsed, pattern, "javascript", f.SupportedFeatures())
	if got := countByID(report.Findings, "missing-anchor"); got != 0 {
		t.Errorf("missing-anchor fired %d times, want 0 (\\b should count as an anchor)", got)
	}
}
