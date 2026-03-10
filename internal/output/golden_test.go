package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor"

	_ "github.com/0x4d5352/regolith/internal/flavor/javascript"
)

var goldenPatterns = []struct {
	name    string
	pattern string
}{
	{"alternation", "a|b|c"},
	{"charset", "[a-zA-Z0-9]"},
	{"quantifiers", "a*b+c?d{2,5}"},
	{"groups", `(foo)(?:bar)(?<name>baz)`},
	{"anchors", "^hello$"},
	{"escapes", `\d\w\s`},
	{"complex-email", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
	{"lookahead", "foo(?=bar)(?!baz)"},
	{"backreference", `(\w+)\s+\1`},
	{"nested-groups", "(a(b(c)))"},
}

func TestGoldenJSON(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	goldenDir := filepath.Join("testdata", "golden", "json")

	for _, tc := range goldenPatterns {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := f.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			got, err := RenderJSON(ast, tc.pattern, "javascript")
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			got += "\n"

			goldenPath := filepath.Join(goldenDir, tc.name+".json")

			if os.Getenv("GOLDEN_UPDATE") != "" {
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file %s not found (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("output does not match golden file %s\ngot:\n%s\nwant:\n%s", goldenPath, got, string(want))
			}
		})
	}
}

func TestGoldenMarkdown(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	goldenDir := filepath.Join("testdata", "golden", "markdown")

	for _, tc := range goldenPatterns {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := f.Parse(tc.pattern)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			got := RenderMarkdown(ast, tc.pattern, "javascript")

			goldenPath := filepath.Join(goldenDir, tc.name+".md")

			if os.Getenv("GOLDEN_UPDATE") != "" {
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file %s not found (run with GOLDEN_UPDATE=1 to create): %v", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("output does not match golden file %s\ngot:\n%s\nwant:\n%s", goldenPath, got, string(want))
			}
		})
	}
}
