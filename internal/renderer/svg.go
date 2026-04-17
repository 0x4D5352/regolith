package renderer

import (
	"html"
	"strconv"
	"strings"
)

// fmtFloat formats a float64 for SVG attributes with consistent cross-platform output.
// It uses fixed-decimal formatting with trailing zero trimming to eliminate
// FMA vs non-FMA precision artifacts (e.g. "68.80000000000001" -> "68.8").
func fmtFloat(v float64) string {
	s := strconv.FormatFloat(v, 'f', 10, 64)
	// Trim trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// ================================================================================
// Attribute builder
// ================================================================================

// svgAttrs incrementally builds a space-separated list of SVG attribute
// key="value" pairs into a single strings.Builder. Every Render method
// used to do the same dance manually with a []string + append + Sprintf
// + strings.Join — concentrating that logic here avoids repeated
// allocations and collapses ~30 lines of attribute shuffling per element
// into ~10. Attribute emission order is preserved exactly, so the output
// is byte-identical to the previous fmt.Sprintf-based version.
type svgAttrs struct {
	b   strings.Builder
	any bool
}

func (a *svgAttrs) sep() {
	if a.any {
		a.b.WriteByte(' ')
	}
	a.any = true
}

// Num writes name="<fmtFloat(v)>". Use Num for required numeric attrs
// (x, y, width, ...) that must always be emitted regardless of value.
func (a *svgAttrs) Num(name string, v float64) {
	a.sep()
	a.b.WriteString(name)
	a.b.WriteString(`="`)
	a.b.WriteString(fmtFloat(v))
	a.b.WriteByte('"')
}

// NumPositive writes name="<fmtFloat(v)>" only when v > 0. Mirrors the
// `if v > 0 { attrs = append(attrs, ...) }` guard every Render method
// previously open-coded for optional dimensions like rx/ry/stroke-width.
func (a *svgAttrs) NumPositive(name string, v float64) {
	if v > 0 {
		a.Num(name, v)
	}
}

// Str writes name="<v>" only when v is non-empty. The value is inserted
// verbatim — callers are expected to pass already-safe data (CSS color
// names, class names, url(#...) marker refs) exactly as the previous
// fmt.Sprintf("fill=%q", ...) did.
func (a *svgAttrs) Str(name, v string) {
	if v != "" {
		a.sep()
		a.b.WriteString(name)
		a.b.WriteString(`="`)
		a.b.WriteString(v)
		a.b.WriteByte('"')
	}
}

// StrAlways is the unconditional counterpart to Str — used for required
// fixed-value attributes like xmlns="http://www.w3.org/2000/svg" or
// fill="none".
func (a *svgAttrs) StrAlways(name, v string) {
	a.sep()
	a.b.WriteString(name)
	a.b.WriteString(`="`)
	a.b.WriteString(v)
	a.b.WriteByte('"')
}

func (a *svgAttrs) String() string { return a.b.String() }
func (a *svgAttrs) empty() bool    { return !a.any }

// ================================================================================
// Element types
// ================================================================================

// SVGElement is the interface for all SVG elements
type SVGElement interface {
	Render() string
}

// Group represents an SVG <g> element
type Group struct {
	Class     string
	Transform string
	Children  []SVGElement
}

func (g *Group) Render() string {
	var a svgAttrs
	a.Str("class", g.Class)
	a.Str("transform", g.Transform)

	var children strings.Builder
	for _, child := range g.Children {
		children.WriteString(child.Render())
	}

	var out strings.Builder
	out.WriteString("<g")
	if !a.empty() {
		out.WriteByte(' ')
		out.WriteString(a.String())
	}
	out.WriteByte('>')
	out.WriteString(children.String())
	out.WriteString("</g>")
	return out.String()
}

// Rect represents an SVG <rect> element
type Rect struct {
	X, Y            float64
	Width, Height   float64
	Rx, Ry          float64 // Corner radius
	Fill            string
	Stroke          string
	StrokeWidth     float64
	StrokeDashArray string // e.g. "6,3" for dashed borders on annotation overlays
	Class           string
}

func (r *Rect) Render() string {
	var a svgAttrs
	a.Num("x", r.X)
	a.Num("y", r.Y)
	a.Num("width", r.Width)
	a.Num("height", r.Height)
	a.NumPositive("rx", r.Rx)
	a.NumPositive("ry", r.Ry)
	a.Str("fill", r.Fill)
	a.Str("stroke", r.Stroke)
	a.NumPositive("stroke-width", r.StrokeWidth)
	a.Str("stroke-dasharray", r.StrokeDashArray)
	a.Str("class", r.Class)
	return "<rect " + a.String() + "/>"
}

// Circle represents an SVG <circle> element, used for severity badge icons
// in the analysis annotation overlay.
type Circle struct {
	Cx, Cy float64
	R      float64
	Fill   string
	Stroke string
	Class  string
}

func (c *Circle) Render() string {
	var a svgAttrs
	a.Num("cx", c.Cx)
	a.Num("cy", c.Cy)
	a.Num("r", c.R)
	a.Str("fill", c.Fill)
	a.Str("stroke", c.Stroke)
	a.Str("class", c.Class)
	return "<circle " + a.String() + "/>"
}

// Text represents an SVG <text> element
type Text struct {
	X, Y       float64
	Content    string
	FontFamily string
	FontSize   float64
	Fill       string
	Anchor     string // text-anchor: start, middle, end
	Class      string
	Spans      []*TSpan // Optional tspan children
}

func (t *Text) Render() string {
	var a svgAttrs
	a.Num("x", t.X)
	a.Num("y", t.Y)
	a.Str("font-family", t.FontFamily)
	a.NumPositive("font-size", t.FontSize)
	a.Str("fill", t.Fill)
	a.Str("text-anchor", t.Anchor)
	a.Str("class", t.Class)

	var content string
	if len(t.Spans) > 0 {
		var spans strings.Builder
		for _, span := range t.Spans {
			spans.WriteString(span.Render())
		}
		content = spans.String()
	} else {
		content = html.EscapeString(t.Content)
	}

	return "<text " + a.String() + ">" + content + "</text>"
}

// TSpan represents an SVG <tspan> element inside text
type TSpan struct {
	Content string
	Class   string
	Fill    string
}

func (ts *TSpan) Render() string {
	var a svgAttrs
	a.Str("class", ts.Class)
	a.Str("fill", ts.Fill)

	var out strings.Builder
	out.WriteString("<tspan")
	if !a.empty() {
		out.WriteByte(' ')
		out.WriteString(a.String())
	}
	out.WriteByte('>')
	out.WriteString(html.EscapeString(ts.Content))
	out.WriteString("</tspan>")
	return out.String()
}

// Path represents an SVG <path> element
type Path struct {
	D           string // Path data
	Fill        string
	Stroke      string
	StrokeWidth float64
	Class       string
}

func (p *Path) Render() string {
	var a svgAttrs
	a.StrAlways("d", p.D)
	// Path must always emit a fill attribute: the explicit value when
	// set, otherwise "none" so the path is drawn as an outline only.
	if p.Fill != "" {
		a.StrAlways("fill", p.Fill)
	} else {
		a.StrAlways("fill", "none")
	}
	a.Str("stroke", p.Stroke)
	a.NumPositive("stroke-width", p.StrokeWidth)
	a.Str("class", p.Class)
	return "<path " + a.String() + "/>"
}

// Line represents an SVG <line> element
type Line struct {
	X1, Y1      float64
	X2, Y2      float64
	Stroke      string
	StrokeWidth float64
	Class       string
	// MarkerStart / MarkerEnd reference marker definitions in the
	// surrounding <defs> block (e.g. "url(#start-arrow)"). Empty means
	// no marker is drawn at that end of the line.
	MarkerStart string
	MarkerEnd   string
}

func (l *Line) Render() string {
	var a svgAttrs
	a.Num("x1", l.X1)
	a.Num("y1", l.Y1)
	a.Num("x2", l.X2)
	a.Num("y2", l.Y2)
	a.Str("stroke", l.Stroke)
	a.NumPositive("stroke-width", l.StrokeWidth)
	a.Str("marker-start", l.MarkerStart)
	a.Str("marker-end", l.MarkerEnd)
	a.Str("class", l.Class)
	return "<line " + a.String() + "/>"
}

// Title represents an SVG <title> element (for tooltips)
type Title struct {
	Content string
}

func (t *Title) Render() string {
	return "<title>" + html.EscapeString(t.Content) + "</title>"
}

// SVG represents the root <svg> element
type SVG struct {
	Width   float64
	Height  float64
	ViewBox string
	// Defs is the content of an optional <defs> block rendered before
	// the <style> block. Used for shared definitions like <marker>
	// elements for connector terminators.
	Defs     string
	Style    string
	Children []SVGElement
}

func (s *SVG) Render() string {
	var a svgAttrs
	a.StrAlways("xmlns", "http://www.w3.org/2000/svg")
	a.NumPositive("width", s.Width)
	a.NumPositive("height", s.Height)
	a.Str("viewBox", s.ViewBox)

	var children strings.Builder
	if s.Defs != "" {
		children.WriteString("<defs>")
		children.WriteString(s.Defs)
		children.WriteString("</defs>")
	}
	if s.Style != "" {
		children.WriteString("<style>")
		children.WriteString(s.Style)
		children.WriteString("</style>")
	}
	for _, child := range s.Children {
		children.WriteString(child.Render())
	}

	return "<svg " + a.String() + ">" + children.String() + "</svg>"
}
