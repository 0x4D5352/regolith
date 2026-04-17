package renderer

import (
	"testing"

	javascript "github.com/0x4d5352/regolith/internal/flavor/javascript"
)

// BenchmarkRender measures end-to-end rendering throughput. It covers
// the hot path hit by every `regolith --format svg` invocation: attribute
// construction (svgAttrs), path command joining (joinPath), and the
// overall tree-to-string traversal.
//
// Running the bench pre- and post-refactor quantifies the Tier 1.3/1.4
// wins (O(n) strings.Builder for paths, strings.Builder-backed attribute
// writer instead of per-attribute fmt.Sprintf + slice append + Join).
func BenchmarkRender(b *testing.B) {
	cases := []struct {
		name    string
		pattern string
	}{
		// Minimal shape — catches the per-element cost without
		// dominating it with tree walks.
		{"simple", "foo|bar"},
		// Groups, alternation, and a quantifier — exercises subexpression
		// nesting and the path builder for loop arrows.
		{"grouped", "a(b|c)*d"},
		// Deeper tree with character classes and anchors — closer to a
		// real-world production diagram.
		{"realistic", "^(?:\\d{3}-)?(?:\\d{3}|\\(\\d{3}\\))-?\\d{4}$"},
	}

	flavor := &javascript.JavaScript{}
	for _, tc := range cases {
		ast, err := flavor.Parse(tc.pattern)
		if err != nil {
			b.Fatalf("parse %q: %v", tc.pattern, err)
		}
		r := New(nil)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = r.Render(ast)
			}
		})
	}
}

// BenchmarkSVGAttrs isolates the attribute-building path. Every Render
// method on every SVG element type funnels through svgAttrs, so even a
// small per-call improvement compounds across a realistic diagram.
func BenchmarkSVGAttrs(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		r := &Rect{
			X: 10, Y: 20, Width: 100, Height: 40,
			Rx: 4, Ry: 4,
			Fill: "#fee2e2", Stroke: "#000", StrokeWidth: 1.5,
			Class: "literal",
		}
		_ = r.Render()
	}
}
