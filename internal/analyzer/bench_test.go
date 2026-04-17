package analyzer

import (
	"testing"

	javascript "github.com/0x4d5352/regolith/internal/flavor/javascript"
)

// BenchmarkAnalyze measures analyzer throughput on patterns that
// exercise different parts of the rule set: a clean pattern (fewest
// findings), a pattern with common ReDoS shapes (most rules fire), and
// one with many literal fragments (hot walkFragment loop).
//
// Running the bench pre- and post-refactor quantifies the Tier 1.5 and
// 2.1/2.2 wins: merged metadata+backref pass and quantifier-guarded
// fragment-rule calls.
func BenchmarkAnalyze(b *testing.B) {
	cases := []struct {
		name    string
		pattern string
	}{
		{"clean", "^hello world$"},
		// Multiple ReDoS signals — nested quantifier, adjacent unbounded,
		// useless capture, no anchor.
		{"redos", "(a+)+b.*c.*d(e|f)*"},
		// Lots of literal fragments — exercises the hot per-fragment
		// loop where the quantifier guard pays off.
		{"literal-heavy", "abcdefghij klmnopqrst uvwxyz0123456789"},
	}

	f := &javascript.JavaScript{}
	features := f.SupportedFeatures()
	for _, tc := range cases {
		ast, err := f.Parse(tc.pattern)
		if err != nil {
			b.Fatalf("parse %q: %v", tc.pattern, err)
		}
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = Analyze(ast, tc.pattern, f.Name(), features)
			}
		})
	}
}
