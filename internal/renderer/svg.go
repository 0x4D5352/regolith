package renderer

import (
	"fmt"
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
	var attrs []string
	if g.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, g.Class))
	}
	if g.Transform != "" {
		attrs = append(attrs, fmt.Sprintf(`transform="%s"`, g.Transform))
	}

	var children strings.Builder
	for _, child := range g.Children {
		children.WriteString(child.Render())
	}

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf("<g%s>%s</g>", attrStr, children.String())
}

// Rect represents an SVG <rect> element
type Rect struct {
	X, Y          float64
	Width, Height float64
	Rx, Ry        float64 // Corner radius
	Fill          string
	Stroke        string
	StrokeWidth   float64
	Class         string
}

func (r *Rect) Render() string {
	var attrs []string
	attrs = append(attrs, `x="`+fmtFloat(r.X)+`"`)
	attrs = append(attrs, `y="`+fmtFloat(r.Y)+`"`)
	attrs = append(attrs, `width="`+fmtFloat(r.Width)+`"`)
	attrs = append(attrs, `height="`+fmtFloat(r.Height)+`"`)

	if r.Rx > 0 {
		attrs = append(attrs, `rx="`+fmtFloat(r.Rx)+`"`)
	}
	if r.Ry > 0 {
		attrs = append(attrs, `ry="`+fmtFloat(r.Ry)+`"`)
	}
	if r.Fill != "" {
		attrs = append(attrs, fmt.Sprintf(`fill="%s"`, r.Fill))
	}
	if r.Stroke != "" {
		attrs = append(attrs, fmt.Sprintf(`stroke="%s"`, r.Stroke))
	}
	if r.StrokeWidth > 0 {
		attrs = append(attrs, `stroke-width="`+fmtFloat(r.StrokeWidth)+`"`)
	}
	if r.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, r.Class))
	}

	return fmt.Sprintf("<rect %s/>", strings.Join(attrs, " "))
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
	var attrs []string
	attrs = append(attrs, `x="`+fmtFloat(t.X)+`"`)
	attrs = append(attrs, `y="`+fmtFloat(t.Y)+`"`)

	if t.FontFamily != "" {
		attrs = append(attrs, fmt.Sprintf(`font-family="%s"`, t.FontFamily))
	}
	if t.FontSize > 0 {
		attrs = append(attrs, `font-size="`+fmtFloat(t.FontSize)+`"`)
	}
	if t.Fill != "" {
		attrs = append(attrs, fmt.Sprintf(`fill="%s"`, t.Fill))
	}
	if t.Anchor != "" {
		attrs = append(attrs, fmt.Sprintf(`text-anchor="%s"`, t.Anchor))
	}
	if t.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, t.Class))
	}

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

	return fmt.Sprintf("<text %s>%s</text>", strings.Join(attrs, " "), content)
}

// TSpan represents an SVG <tspan> element inside text
type TSpan struct {
	Content string
	Class   string
	Fill    string
}

func (ts *TSpan) Render() string {
	var attrs []string
	if ts.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, ts.Class))
	}
	if ts.Fill != "" {
		attrs = append(attrs, fmt.Sprintf(`fill="%s"`, ts.Fill))
	}

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf("<tspan%s>%s</tspan>", attrStr, html.EscapeString(ts.Content))
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
	var attrs []string
	attrs = append(attrs, fmt.Sprintf(`d="%s"`, p.D))

	if p.Fill != "" {
		attrs = append(attrs, fmt.Sprintf(`fill="%s"`, p.Fill))
	} else {
		attrs = append(attrs, `fill="none"`)
	}
	if p.Stroke != "" {
		attrs = append(attrs, fmt.Sprintf(`stroke="%s"`, p.Stroke))
	}
	if p.StrokeWidth > 0 {
		attrs = append(attrs, `stroke-width="`+fmtFloat(p.StrokeWidth)+`"`)
	}
	if p.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, p.Class))
	}

	return fmt.Sprintf("<path %s/>", strings.Join(attrs, " "))
}

// Line represents an SVG <line> element
type Line struct {
	X1, Y1      float64
	X2, Y2      float64
	Stroke      string
	StrokeWidth float64
	Class       string
}

func (l *Line) Render() string {
	var attrs []string
	attrs = append(attrs, `x1="`+fmtFloat(l.X1)+`"`)
	attrs = append(attrs, `y1="`+fmtFloat(l.Y1)+`"`)
	attrs = append(attrs, `x2="`+fmtFloat(l.X2)+`"`)
	attrs = append(attrs, `y2="`+fmtFloat(l.Y2)+`"`)

	if l.Stroke != "" {
		attrs = append(attrs, fmt.Sprintf(`stroke="%s"`, l.Stroke))
	}
	if l.StrokeWidth > 0 {
		attrs = append(attrs, `stroke-width="`+fmtFloat(l.StrokeWidth)+`"`)
	}
	if l.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, l.Class))
	}

	return fmt.Sprintf("<line %s/>", strings.Join(attrs, " "))
}

// Title represents an SVG <title> element (for tooltips)
type Title struct {
	Content string
}

func (t *Title) Render() string {
	return fmt.Sprintf("<title>%s</title>", html.EscapeString(t.Content))
}

// SVG represents the root <svg> element
type SVG struct {
	Width    float64
	Height   float64
	ViewBox  string
	Children []SVGElement
	Style    string
}

func (s *SVG) Render() string {
	var attrs []string
	attrs = append(attrs, `xmlns="http://www.w3.org/2000/svg"`)

	if s.Width > 0 {
		attrs = append(attrs, `width="`+fmtFloat(s.Width)+`"`)
	}
	if s.Height > 0 {
		attrs = append(attrs, `height="`+fmtFloat(s.Height)+`"`)
	}
	if s.ViewBox != "" {
		attrs = append(attrs, fmt.Sprintf(`viewBox="%s"`, s.ViewBox))
	}

	var children strings.Builder
	if s.Style != "" {
		children.WriteString(fmt.Sprintf("<style>%s</style>", s.Style))
	}
	for _, child := range s.Children {
		children.WriteString(child.Render())
	}

	return fmt.Sprintf("<svg %s>%s</svg>", strings.Join(attrs, " "), children.String())
}
