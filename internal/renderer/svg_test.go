package renderer

import (
	"strings"
	"testing"
)

func TestRectStrokeDashArray(t *testing.T) {
	r := &Rect{
		X: 10, Y: 20, Width: 100, Height: 50,
		Fill: "none", Stroke: "#e53e3e", StrokeWidth: 2.5,
		StrokeDashArray: "6,3",
	}
	got := r.Render()
	if !strings.Contains(got, `stroke-dasharray="6,3"`) {
		t.Errorf("expected stroke-dasharray in output, got: %s", got)
	}
}

func TestCircleRender(t *testing.T) {
	c := &Circle{
		Cx: 100, Cy: 50, R: 11,
		Fill: "#e53e3e",
	}
	got := c.Render()
	want := `<circle cx="100" cy="50" r="11" fill="#e53e3e"/>`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
