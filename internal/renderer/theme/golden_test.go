package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor/javascript"
	"github.com/0x4d5352/regolith/internal/renderer"
)

// goldenPattern is a single representative regex used to render a
// visual sample under every theme. It deliberately exercises a
// cross-section of node types that the renderer draws differently —
// anchors, a non-capturing group, nested alternation, a repeat with
// bounds, an escape, a character class, and a word-boundary escape —
// so a theme regression (wrong stroke color on literals, missing
// subexp panel, charset fill bleeding through) shows up as a diff in
// the golden file rather than a silent palette drift.
//
// The pattern stays in a flavor (JavaScript) that every developer
// working on this repo is fluent in, so reviewing a golden diff does
// not require guessing what the pattern means.
const goldenPattern = `^(?:abc|\d{2,4})[a-z]+\b`

// TestThemeGoldens renders goldenPattern under every registered
// theme and compares the result against a stored SVG. Strict mode:
// a missing golden file is a failure, not an auto-populate, so new
// themes require a deliberate `GOLDEN_UPDATE=1` run before they pass.
func TestThemeGoldens(t *testing.T) {
	js := &javascript.JavaScript{}
	ast, err := js.Parse(goldenPattern)
	if err != nil {
		t.Fatalf("parse %q: %v", goldenPattern, err)
	}

	goldenDir := filepath.Join("testdata", "golden")

	for _, name := range List() {
		t.Run(name, func(t *testing.T) {
			th, ok := Get(name)
			if !ok {
				t.Fatalf("Get(%q): not registered", name)
			}

			cfg := renderer.DefaultConfig()
			th.Apply(cfg)
			r := renderer.New(cfg)
			svg := r.Render(ast)

			goldenPath := filepath.Join(goldenDir, name+".svg")

			if os.Getenv("GOLDEN_UPDATE") == "1" {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", goldenDir, err)
				}
				if err := os.WriteFile(goldenPath, []byte(svg), 0o644); err != nil {
					t.Fatalf("write golden %s: %v", goldenPath, err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}
			if svg != string(want) {
				t.Errorf("SVG differs from %s (run with GOLDEN_UPDATE=1 to update)", goldenPath)
			}
		})
	}
}

// TestDefaultGolden is the control: rendering the same pattern with
// the untouched DefaultConfig must also be stable. If a theme file
// accidentally mutated package-level state at init() time, this test
// would catch it by drifting independently of any theme's own golden.
func TestDefaultGolden(t *testing.T) {
	js := &javascript.JavaScript{}
	ast, err := js.Parse(goldenPattern)
	if err != nil {
		t.Fatalf("parse %q: %v", goldenPattern, err)
	}

	cfg := renderer.DefaultConfig()
	r := renderer.New(cfg)
	svg := r.Render(ast)

	goldenPath := filepath.Join("testdata", "golden", "default.svg")

	if os.Getenv("GOLDEN_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(svg), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
	}
	if svg != string(want) {
		t.Errorf("SVG differs from %s (run with GOLDEN_UPDATE=1 to update)", goldenPath)
	}
}
