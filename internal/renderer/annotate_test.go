package renderer

import (
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/flavor"
	_ "github.com/0x4d5352/regolith/internal/flavor/javascript"
)

func TestRenderAnnotated(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	parsed, err := f.Parse(".*.*=.*")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	report := analyzer.Analyze(parsed, ".*.*=.*", "javascript", f.SupportedFeatures())
	cfg := DefaultConfig()
	r := New(cfg)

	svg := r.RenderAnnotated(parsed, report)

	if !strings.HasPrefix(svg, "<svg") {
		t.Error("expected SVG output")
	}
	if !strings.Contains(svg, `stroke-dasharray="6,3"`) {
		t.Error("expected dashed border annotation")
	}
	if !strings.Contains(svg, `class="analysis-badge"`) {
		t.Error("expected analysis badge")
	}
	if !strings.Contains(svg, `class="analysis-legend"`) {
		t.Error("expected analysis legend")
	}
}

func TestRenderAnnotatedNoFindings(t *testing.T) {
	f, ok := flavor.Get("javascript")
	if !ok {
		t.Fatal("javascript flavor not registered")
	}

	parsed, err := f.Parse("hello")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	report := &analyzer.AnalysisReport{
		Pattern: "hello", Flavor: "javascript", Findings: nil,
	}

	cfg := DefaultConfig()
	r := New(cfg)
	svg := r.RenderAnnotated(parsed, report)

	if !strings.HasPrefix(svg, "<svg") {
		t.Error("expected SVG output")
	}
	if strings.Contains(svg, `stroke-dasharray="6,3"`) {
		t.Error("unexpected annotation in clean pattern")
	}
}
