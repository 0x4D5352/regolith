package renderer

import (
	"fmt"
	"math"
	"strings"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/parser"
)

// Renderer handles rendering regex AST to SVG
type Renderer struct {
	Config       *Config
	subexpDepth  int // Tracks nesting depth for subexpressions
	nodeFindings map[parser.Node]*analyzer.Finding
}

// New creates a new Renderer with the given config
func New(cfg *Config) *Renderer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Renderer{Config: cfg}
}

// Render renders a regex AST to SVG
// Marker layout constants. These need to stay consistent with the
// marker definitions in getDefs() — startArrowReach is the polygon
// width and endDotRadius is the circle r attribute.
const (
	// startArrowReach is how far the start arrow extends rightward
	// from its line start point.
	startArrowReach = 10
	// endDotRadius is the radius of the end dot circle.
	endDotRadius = 3
	// visibleConnectorWidth is the visible line segment between a
	// connector marker and its adjacent content node. The same value
	// is used on both sides of the diagram so the spacing reads
	// symmetric, and it equals the horizontal gap between match items
	// so that in annotated mode the gap from the start arrow to an
	// annotation border matches the gap from the next item to that
	// border.
	visibleConnectorWidth = 10
)

// contentLeftMargin and contentRightMargin compute the per-side
// horizontal margin from the SVG viewBox edge to the first/last
// content node. They differ because the start arrow is anchored to its
// line start (extends rightward by startArrowReach) while the end dot
// is centered on its line end (extends both directions by endDotRadius).
//
//	[edgeClearance | start arrow | visible connector | content ... | visible connector | dot | edgeClearance]
//
// edgeClearance comes from cfg.Padding/2 so users who tune Padding
// also tune the breathing room at the diagram's edges.
func contentLeftMargin(padding float64) float64 {
	edgeClearance := padding / 2
	return edgeClearance + startArrowReach + visibleConnectorWidth
}

func contentRightMargin(padding float64) float64 {
	edgeClearance := padding / 2
	// visible connector + dot diameter (2*radius) + edge clearance
	return visibleConnectorWidth + 2*endDotRadius + edgeClearance
}

func (r *Renderer) Render(ast *parser.Regexp) string {
	rendered := r.renderRegexp(ast)

	// Add padding around the diagram. The content area is offset on
	// each side by contentLeftMargin / contentRightMargin, which
	// reserve space for the start/end markers and a visible connector
	// segment between the marker and the first/last content node.
	padding := r.Config.Padding
	leftMargin := contentLeftMargin(padding)
	rightMargin := contentRightMargin(padding)
	width := rendered.BBox.Width + leftMargin + rightMargin
	height := rendered.BBox.Height + 2*padding

	// Check for flags and render them
	var flagsElement SVGElement
	var flagsRendered RenderedNode
	flagsWidth := 0.0
	if ast.Flags != "" {
		flagsRendered = r.renderFlags(ast.Flags)
		flagsElement = flagsRendered.Element
		flagsWidth = flagsRendered.BBox.Width + padding
		width += flagsWidth
		// Ensure height accommodates flags box
		if flagsRendered.BBox.Height+2*padding > height {
			height = flagsRendered.BBox.Height + 2*padding
		}
	}

	// Check for pattern start options (PCRE)
	var bannerHeight float64
	var bannerElement SVGElement
	if len(ast.Options) > 0 {
		bannerRendered := r.renderPatternOptions(ast.Options)
		bannerElement = bannerRendered.Element
		bannerHeight = bannerRendered.BBox.Height + padding/2

		// Ensure the SVG is wide enough for the banner
		bannerWidth := bannerRendered.BBox.Width + 2*padding
		if bannerWidth > width {
			width = bannerWidth
		}

		height += bannerHeight
	}

	// Create start and end connectors. The start line spans from the
	// left edge clearance out to leftMargin (where content begins),
	// hosting the arrow marker plus a visible connector segment. The
	// end line mirrors this on the right with the dot marker.
	startX := padding / 2
	anchorY := bannerHeight + padding + rendered.BBox.AnchorY
	contentEndX := width - rightMargin - flagsWidth
	endLineLength := float64(visibleConnectorWidth + endDotRadius)

	startLine := &Line{
		X1:          startX,
		Y1:          anchorY,
		X2:          leftMargin,
		Y2:          anchorY,
		Stroke:      r.Config.Connector.Color,
		StrokeWidth: r.Config.Connector.StrokeWidth,
		MarkerStart: startMarkerRef(r.Config.Connector.StartMarker),
	}

	endLine := &Line{
		X1:          contentEndX,
		Y1:          anchorY,
		X2:          contentEndX + endLineLength,
		Y2:          anchorY,
		Stroke:      r.Config.Connector.Color,
		StrokeWidth: r.Config.Connector.StrokeWidth,
		MarkerEnd:   endMarkerRef(r.Config.Connector.EndMarker),
	}

	// Wrap the rendered content in a group offset by leftMargin so
	// the first node sits at the end of the start connector line.
	contentGroup := &Group{
		Transform: "translate(" + fmtFloat(leftMargin) + "," + fmtFloat(bannerHeight+padding) + ")",
		Children:  []SVGElement{rendered.Element},
	}

	// When BackgroundFill is set, prepend a full-viewBox rect so it
	// paints behind every other child. Width/height here are the final
	// SVG dimensions, already adjusted above for the banner and flags
	// add-ons, so the rect covers the entire visible surface.
	var children []SVGElement
	if r.Config.BackgroundFill != "" {
		children = append(children, &Rect{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
			Fill:   r.Config.BackgroundFill,
		})
	}
	children = append(children, startLine, endLine, contentGroup)

	// Add banner if present
	if bannerElement != nil {
		bannerGroup := &Group{
			Transform: "translate(" + fmtFloat(padding) + "," + fmtFloat(padding/2) + ")",
			Children:  []SVGElement{bannerElement},
		}
		children = append(children, bannerGroup)
	}

	// Add flags if present
	if flagsElement != nil {
		flagsGroup := &Group{
			Transform: "translate(" + fmtFloat(width-padding-flagsWidth+padding/2) + "," + fmtFloat(bannerHeight+padding) + ")",
			Children:  []SVGElement{flagsElement},
		}
		children = append(children, flagsGroup)
	}

	svg := &SVG{
		Width:    width,
		Height:   height,
		ViewBox:  "0 0 " + fmtFloat(width) + " " + fmtFloat(height),
		Defs:     r.getDefs(),
		Style:    r.getStyles(),
		Children: children,
	}

	return svg.Render()
}

// startMarkerRef returns the SVG marker reference string for a
// Connector.StartMarker setting, or an empty string if no marker is
// configured. Keeping this as a small helper means the render sites
// don't have to know which marker ids exist.
func startMarkerRef(kind string) string {
	switch kind {
	case "arrow":
		return "url(#start-arrow)"
	default:
		return ""
	}
}

// endMarkerRef returns the SVG marker reference string for a
// Connector.EndMarker setting, or an empty string if no marker is
// configured.
func endMarkerRef(kind string) string {
	switch kind {
	case "dot":
		return "url(#end-dot)"
	default:
		return ""
	}
}

// getDefs returns an SVG <defs> payload containing marker definitions
// for the configured connector terminators. The markers are colored
// with the connector color so they read as a single unit with the
// track lines they decorate.
func (r *Renderer) getDefs() string {
	color := r.Config.Connector.Color
	var b strings.Builder
	if r.Config.Connector.StartMarker == "arrow" {
		// The arrow points right (into the diagram). refX=0 places the
		// tip at the line's start; refY=3.5 centers it vertically.
		fmt.Fprintf(&b,
			`<marker id="start-arrow" markerWidth="10" markerHeight="7" refX="0" refY="3.5" orient="auto"><polygon points="0 0, 10 3.5, 0 7" fill="%s"/></marker>`,
			color)
	}
	if r.Config.Connector.EndMarker == "dot" {
		// refX=4 centers the dot on the line's end point.
		fmt.Fprintf(&b,
			`<marker id="end-dot" markerWidth="8" markerHeight="8" refX="4" refY="4"><circle cx="4" cy="4" r="3" fill="%s"/></marker>`,
			color)
	}
	return b.String()
}

// renderPatternOptions renders PCRE pattern start options as a banner.
// The banner text is a structural description regolith generates, so
// it uses the sans-serif label font family.
func (r *Renderer) renderPatternOptions(options []*parser.PatternOption) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding / 2

	// Build comma-separated label
	var parts []string
	for _, opt := range options {
		if opt.Value != "" {
			parts = append(parts, fmt.Sprintf("*%s=%s", opt.Name, opt.Value))
		} else {
			parts = append(parts, fmt.Sprintf("*%s", opt.Name))
		}
	}
	label := "Options: " + strings.Join(parts, ", ")

	textWidth := MeasureLabelText(label, cfg)
	width := textWidth + 2*padding
	height := cfg.FontSize + 2*padding

	rect := &Rect{
		X:           0,
		Y:           0,
		Width:       width,
		Height:      height,
		Rx:          cfg.CornerRadius,
		Ry:          cfg.CornerRadius,
		Fill:        "#e8e8e8",
		Stroke:      "#999",
		StrokeWidth: cfg.NodeStrokeWidth,
	}

	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.LabelFontSize/3,
		Content:    label,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Anchor:     "middle",
		Class:      "pattern-options-label",
	}

	group := &Group{
		Class:    "pattern-options",
		Children: []SVGElement{rect, textElem},
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// getStyles returns the CSS styles for the SVG.
//
// The stylesheet is generated from r.Config.NodeStyles so that a theme
// only needs to replace the map, not the render logic. Each category
// contributes three rules: rect fill, rect stroke (+ stroke-width), and
// text fill. Per-class text colors eliminate the old global "all text
// is black" rule, which lets dark-background nodes (anchors) pick their
// own readable foreground without a special case.
//
// Corner radii are NOT set here. Some SVG renderers (notably librsvg)
// don't honor CSS-set rx/ry; the render methods set them inline on
// each <rect> instead.
func (r *Renderer) getStyles() string {
	cfg := r.Config
	var b strings.Builder

	// Category rules — iterate in a stable, readable order rather
	// than whatever order range-over-map yields.
	categories := []string{
		"literal", "escape", "charset", "anchor", "any-character",
		"flags", "recursive-ref", "callout", "backtrack-control",
		"conditional", "comment",
	}
	strokeWidth := fmtFloat(cfg.NodeStrokeWidth)
	for _, class := range categories {
		style, ok := cfg.NodeStyles[class]
		if !ok {
			continue
		}
		// The comment class gets an extra stroke-dasharray so the box
		// reads as a comment bubble rather than a normal node.
		dashAttr := ""
		if class == "comment" {
			dashAttr = " stroke-dasharray: 4,2;"
		}
		fmt.Fprintf(&b,
			"\n\t\t.%s rect { fill: %s; stroke: %s; stroke-width: %s;%s }",
			class, style.Fill, style.Stroke, strokeWidth, dashAttr)
		fmt.Fprintf(&b,
			"\n\t\t.%s text { fill: %s; }",
			class, style.TextColor)
	}

	// Comment text keeps its italic treatment — it's prose inside a
	// code-shaped diagram, and the italic cue makes that obvious.
	b.WriteString("\n\t\t.comment text { font-style: italic; }")

	// Base text rule. FontFamily and FontSize are defaults for any
	// Text element that doesn't override them inline. Text fill is
	// deliberately NOT set globally — each category rule above sets
	// it per class, and TextColor on cfg is only a fallback for
	// elements outside any category.
	fmt.Fprintf(&b,
		"\n\t\ttext { font-family: %s; font-size: %spx; fill: %s; }",
		cfg.FontFamily, fmtFloat(cfg.FontSize), cfg.TextColor)

	// Structural labels (group names, charset header, flags header,
	// repeat labels) switch to the sans-serif label font. No italic
	// this time — the hierarchy is already carried by the font change.
	fmt.Fprintf(&b,
		"\n\t\t.subexp-label, .charset-label, .flags-label { font-family: %s; font-size: %spx; }",
		cfg.LabelFontFamily, fmtFloat(cfg.LabelFontSize))
	fmt.Fprintf(&b,
		"\n\t\t.repeat-label { fill: %s; font-family: %s; font-size: %spx; }",
		cfg.RepeatLabelColor, cfg.LabelFontFamily, fmtFloat(cfg.LabelFontSize))

	b.WriteString("\n\t")
	return b.String()
}

// renderNode dispatches to the appropriate render method based on node type.
// The result is passed through annotateNode, which overlays severity markers
// when an analysis report is active (nodeFindings is non-nil).
func (r *Renderer) renderNode(node parser.Node) RenderedNode {
	var rendered RenderedNode
	switch n := node.(type) {
	case *parser.Regexp:
		rendered = r.renderRegexp(n)
	case *parser.Match:
		rendered = r.renderMatch(n)
	case *parser.MatchFragment:
		rendered = r.renderMatchFragment(n)
	case *parser.Literal:
		rendered = r.renderLiteral(n)
	case *parser.Escape:
		rendered = r.renderEscape(n)
	case *parser.Anchor:
		rendered = r.renderAnchor(n)
	case *parser.AnyCharacter:
		rendered = r.renderAnyCharacter(n)
	case *parser.Charset:
		rendered = r.renderCharset(n)
	case *parser.Subexp:
		rendered = r.renderSubexp(n)
	case *parser.BackReference:
		rendered = r.renderBackReference(n)
	case *parser.UnicodePropertyEscape:
		rendered = r.renderUnicodePropertyEscape(n)
	case *parser.QuotedLiteral:
		rendered = r.renderQuotedLiteral(n)
	case *parser.Comment:
		rendered = r.renderComment(n)
	case *parser.InlineModifier:
		rendered = r.renderInlineModifier(n)
	case *parser.BalancedGroup:
		rendered = r.renderBalancedGroup(n)
	case *parser.Conditional:
		rendered = r.renderConditional(n)
	case *parser.RecursiveRef:
		rendered = r.renderRecursiveRef(n)
	case *parser.BranchReset:
		rendered = r.renderBranchReset(n)
	case *parser.BacktrackControl:
		rendered = r.renderBacktrackControl(n)
	case *parser.Callout:
		rendered = r.renderCallout(n)
	case *parser.CharsetIntersection:
		rendered = r.renderCharsetIntersection(n)
	case *parser.CharsetSubtraction:
		rendered = r.renderCharsetSubtraction(n)
	case *parser.CharsetStringDisjunction:
		rendered = r.renderCharsetStringDisjunction(n)
	default:
		rendered = r.renderStructuralLabel(fmt.Sprintf("<%s>", node.Type()), "unknown")
	}
	return r.annotateNode(node, rendered)
}

// cornerRadiusFor returns the effective corner radius for a node class.
// Most categories inherit the global Config.CornerRadius; anchors
// override to a larger radius so they render as full pills.
func (r *Renderer) cornerRadiusFor(class string) float64 {
	if override := r.Config.GetNodeStyle(class).CornerRadius; override > 0 {
		return override
	}
	return r.Config.CornerRadius
}

// renderLabel creates a labeled box whose text is **regex content** —
// escape sequences, back-references, anything that represents user-
// written regex syntax. Rendered in the monospace content font so it
// reads as code.
func (r *Renderer) renderLabel(text, class string) RenderedNode {
	cfg := r.Config
	textWidth := MeasureText(text, cfg)
	padding := cfg.Padding / 2

	width := textWidth + 2*padding
	height := cfg.FontSize + 2*padding
	radius := r.cornerRadiusFor(class)

	rect := &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     radius,
		Ry:     radius,
	}

	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.FontSize/3, // Approximate vertical centering
		Content:    text,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize,
		Anchor:     "middle",
	}

	group := &Group{
		Class:    class,
		Children: []SVGElement{rect, textElem},
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderStructuralLabel creates a labeled box whose text is a
// **description regolith generates** — anchor names, "any character",
// back-reference descriptions, condition words, group labels. Rendered
// in the sans-serif label font to distinguish it from the user's
// regex content at a glance.
//
// Box height stays aligned with content-font boxes so that a row of
// mixed label types reads as a uniform band. Only the font family and
// font size change.
//
// For nodes with an enlarged corner radius (pills like anchors), the
// horizontal padding is widened to the corner radius. Otherwise the
// text would extend into the rounded ends of the pill and appear to
// overflow the fill.
func (r *Renderer) renderStructuralLabel(text, class string) RenderedNode {
	cfg := r.Config
	textWidth := MeasureLabelText(text, cfg)
	radius := r.cornerRadiusFor(class)

	// For standard rectangular nodes, keep the existing tight padding.
	// For pills (anchors), widen to the corner radius so text stays
	// clear of the rounded ends.
	padding := cfg.Padding / 2
	if nodeRadius := cfg.GetNodeStyle(class).CornerRadius; nodeRadius > 0 && nodeRadius > padding {
		padding = nodeRadius
	}

	width := textWidth + 2*padding
	height := cfg.FontSize + 2*padding

	rect := &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     radius,
		Ry:     radius,
	}

	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.LabelFontSize/3,
		Content:    text,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Anchor:     "middle",
	}

	group := &Group{
		Class:    class,
		Children: []SVGElement{rect, textElem},
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderFlags renders regex flags (gimuy) as a labeled box
func (r *Renderer) renderFlags(flags string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	// Build flag descriptions
	var flagItems []string
	for _, f := range flags {
		switch f {
		case 'd':
			flagItems = append(flagItems, "hasIndices")
		case 'g':
			flagItems = append(flagItems, "global")
		case 'i':
			flagItems = append(flagItems, "ignore case")
		case 'm':
			flagItems = append(flagItems, "multiline")
		case 's':
			flagItems = append(flagItems, "dotAll")
		case 'u':
			flagItems = append(flagItems, "unicode")
		case 'y':
			flagItems = append(flagItems, "sticky")
		case 'v':
			flagItems = append(flagItems, "unicodeSets")
		}
	}

	label := "Flags:"

	// Calculate dimensions. Both the header and the flag item names
	// ("global", "ignore case", ...) are English descriptions regolith
	// generates, so both are measured against the sans-serif label
	// char-width.
	labelWidth := MeasureLabelText(label, cfg)
	maxItemWidth := 0.0
	for _, item := range flagItems {
		w := MeasureLabelText(item, cfg)
		if w > maxItemWidth {
			maxItemWidth = w
		}
	}

	contentWidth := maxItemWidth + 2*padding
	if labelWidth > contentWidth {
		contentWidth = labelWidth
	}

	labelHeight := cfg.FontSize + padding
	itemHeight := cfg.FontSize + padding/2
	contentHeight := float64(len(flagItems)) * itemHeight

	width := contentWidth + 2*padding
	height := labelHeight + contentHeight + padding

	var children []SVGElement

	// Background rect
	children = append(children, &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     cfg.CornerRadius,
		Ry:     cfg.CornerRadius,
	})

	// Header label
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Class:      "flags-label",
	})

	// Flag items
	y := labelHeight + cfg.FontSize
	for _, item := range flagItems {
		children = append(children, &Text{
			X:          width / 2,
			Y:          y,
			Content:    item,
			FontFamily: cfg.LabelFontFamily,
			FontSize:   cfg.LabelFontSize,
			Anchor:     "middle",
		})
		y += itemHeight
	}

	group := &Group{
		Class:    "flags",
		Children: children,
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderQuotedLabel creates a label with quotes around content (for literals)
func (r *Renderer) renderQuotedLabel(text, class string) RenderedNode {
	cfg := r.Config
	quotedText := `"` + text + `"`
	textWidth := MeasureText(quotedText, cfg)
	padding := cfg.Padding / 2

	width := textWidth + 2*padding
	height := cfg.FontSize + 2*padding
	radius := r.cornerRadiusFor(class)

	rect := &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     radius,
		Ry:     radius,
	}

	// Create text with styled quotes
	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.FontSize/3,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize,
		Anchor:     "middle",
		Spans: []*TSpan{
			{Content: `"`, Class: "quote"},
			{Content: text},
			{Content: `"`, Class: "quote"},
		},
	}

	group := &Group{
		Class:    class,
		Children: []SVGElement{rect, textElem},
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderLiteral renders a literal text node
func (r *Renderer) renderLiteral(lit *parser.Literal) RenderedNode {
	return r.renderQuotedLabel(lit.Text, "literal")
}

// renderEscape renders an escape sequence
func (r *Renderer) renderEscape(esc *parser.Escape) RenderedNode {
	return r.renderLabel(esc.Value, "escape")
}

// renderAnchor renders an anchor (^, $, \b, \B, \<, \>, \A, \Z, \z, \G)
func (r *Renderer) renderAnchor(anchor *parser.Anchor) RenderedNode {
	var label string
	switch anchor.AnchorType {
	case "start":
		label = "Start of line"
	case "end":
		label = "End of line"
	case "word_boundary":
		label = "Word boundary"
	case "non_word_boundary":
		label = "Non-word boundary"
	case "word_start":
		label = "Start of word"
	case "word_end":
		label = "End of word"
	case "string_start":
		label = "Start of input"
	case "string_end":
		label = "End of input"
	case "absolute_end":
		label = "Absolute end"
	case "end_of_previous_match":
		label = "End of previous match"
	case "grapheme_cluster_boundary":
		label = "Grapheme cluster boundary"
	default:
		label = anchor.AnchorType
	}
	return r.renderStructuralLabel(label, "anchor")
}

// renderAnyCharacter renders the . metacharacter
func (r *Renderer) renderAnyCharacter(_ *parser.AnyCharacter) RenderedNode {
	return r.renderStructuralLabel("any character", "any-character")
}

// renderBackReference renders a back-reference like \1 or \k<name>.
// The label is a description ("back reference #1"), not raw regex
// syntax, so it renders in the sans-serif structural font.
func (r *Renderer) renderBackReference(br *parser.BackReference) RenderedNode {
	var label string
	if br.Name != "" {
		label = fmt.Sprintf("back reference '%s'", br.Name)
	} else {
		label = fmt.Sprintf("back reference #%d", br.Number)
	}
	return r.renderStructuralLabel(label, "escape")
}

// renderUnicodePropertyEscape renders a Unicode property escape like
// \p{Letter} or \P{Number}. Like back-references, the label is a
// description ("Unicode Letter") and uses the structural font.
func (r *Renderer) renderUnicodePropertyEscape(upe *parser.UnicodePropertyEscape) RenderedNode {
	var label string
	if upe.Negated {
		label = fmt.Sprintf("NOT Unicode %s", upe.Property)
	} else {
		label = fmt.Sprintf("Unicode %s", upe.Property)
	}
	return r.renderStructuralLabel(label, "escape")
}

// renderQuotedLiteral renders a \Q...\E quoted literal sequence
func (r *Renderer) renderQuotedLiteral(ql *parser.QuotedLiteral) RenderedNode {
	return r.renderQuotedLabel(ql.Text, "literal")
}

// renderComment renders a (?#...) inline comment. Comment text is
// prose the user wrote in the regex — not regex syntax — so it reads
// more naturally in the sans-serif label font, kept italic via the
// comment CSS class.
func (r *Renderer) renderComment(comment *parser.Comment) RenderedNode {
	cfg := r.Config
	text := "# " + comment.Text
	textWidth := MeasureLabelText(text, cfg)
	padding := cfg.Padding / 2

	width := textWidth + 2*padding
	height := cfg.FontSize + 2*padding

	rect := &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     cfg.CornerRadius,
		Ry:     cfg.CornerRadius,
	}

	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.LabelFontSize/3,
		Content:    text,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Anchor:     "middle",
		Class:      "comment-text",
	}

	group := &Group{
		Class:    "comment",
		Children: []SVGElement{rect, textElem},
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderInlineModifier renders inline flag modifiers like (?i) or (?i:...)
func (r *Renderer) renderInlineModifier(im *parser.InlineModifier) RenderedNode {
	// Build the modifier label
	var label string
	if im.Enable != "" && im.Disable != "" {
		label = fmt.Sprintf("flags: +%s -%s", im.Enable, im.Disable)
	} else if im.Enable != "" {
		label = fmt.Sprintf("flags: +%s", im.Enable)
	} else if im.Disable != "" {
		label = fmt.Sprintf("flags: -%s", im.Disable)
	} else {
		label = "flags"
	}

	// If scoped (has Regexp), render as a group with the content
	if im.Regexp != nil {
		// Render the contained regexp
		content := r.renderRegexp(im.Regexp)
		return r.renderLabeledBoxWithContent(label, content, "flags")
	}

	// Global modifier - just render as a label
	return r.renderStructuralLabel(label, "flags")
}

// renderBalancedGroup renders a .NET balanced group (?<name-other>...) or (?<-other>...)
func (r *Renderer) renderBalancedGroup(bg *parser.BalancedGroup) RenderedNode {
	// Build the label based on whether it's capturing or non-capturing
	var label string
	if bg.Name != "" {
		// Capturing balanced group: (?<name-other>...)
		label = fmt.Sprintf("balanced group '%s' (pop '%s')", bg.Name, bg.OtherName)
	} else {
		// Non-capturing balanced group: (?<-other>...)
		label = fmt.Sprintf("balance (pop '%s')", bg.OtherName)
	}

	// Increment depth before rendering nested content
	r.subexpDepth++

	// Render the contained regexp
	content := r.renderRegexp(bg.Regexp)

	// Decrement depth after rendering
	r.subexpDepth--

	// Determine fill color based on depth (same logic as subexp)
	currentDepth := r.subexpDepth
	var fill string
	if currentDepth == 0 {
		fill = r.Config.SubexpFill
	} else if len(r.Config.SubexpColors) > 0 {
		colorIndex := (currentDepth - 1) % len(r.Config.SubexpColors)
		fill = r.Config.SubexpColors[colorIndex]
	} else {
		fill = r.Config.SubexpFill
	}

	return r.renderSubexpBox(label, content, fill)
}

// renderConditional renders a conditional pattern (?(cond)yes|no)
func (r *Renderer) renderConditional(cond *parser.Conditional) RenderedNode {
	cfg := r.Config

	// Build condition label
	var condLabel string
	switch c := cond.Condition.(type) {
	case *parser.BackReference:
		if c.Name != "" {
			condLabel = fmt.Sprintf("if '%s' matched", c.Name)
		} else if c.Number >= 0 {
			condLabel = fmt.Sprintf("if group %d matched", c.Number)
		} else {
			condLabel = fmt.Sprintf("if group %d matched", -c.Number)
		}
	case *parser.RecursiveRef:
		switch c.Target {
		case "R":
			condLabel = "if in recursion"
		case "DEFINE", "":
			condLabel = "DEFINE"
		default:
			condLabel = fmt.Sprintf("if in recursion to '%s'", c.Target)
		}
	case *parser.Literal:
		if c.Text == "DEFINE" {
			condLabel = "DEFINE"
		} else {
			condLabel = fmt.Sprintf("if %s", c.Text)
		}
	case *parser.Subexp:
		// Assertion as condition
		switch c.GroupType {
		case parser.GroupPositiveLookahead:
			condLabel = "if followed by..."
		case parser.GroupNegativeLookahead:
			condLabel = "if not followed by..."
		case parser.GroupPositiveLookbehind:
			condLabel = "if preceded by..."
		case parser.GroupNegativeLookbehind:
			condLabel = "if not preceded by..."
		default:
			condLabel = "if assertion"
		}
	default:
		condLabel = "if condition"
	}

	// Render the yes (true) branch
	yesContent := r.renderRegexp(cond.TrueMatch)
	yesLabel := r.renderStructuralLabel("then", "condition-label")

	// Combine yes label and content
	yesItems := []RenderedNode{yesLabel, yesContent}
	yesSpaced, yesBBox := SpaceHorizontally(yesItems, cfg.HorizontalGap)

	var children []SVGElement
	var totalWidth, totalHeight float64

	// Render yes branch elements
	yesGroup := &Group{Class: "condition-yes"}
	for _, item := range yesSpaced {
		yesGroup.Children = append(yesGroup.Children, item.Element)
	}

	if cond.FalseMatch != nil {
		// Has a no (false) branch
		noContent := r.renderRegexp(cond.FalseMatch)
		noLabel := r.renderStructuralLabel("else", "condition-label")

		// Combine no label and content
		noItems := []RenderedNode{noLabel, noContent}
		noSpaced, noBBox := SpaceHorizontally(noItems, cfg.HorizontalGap)

		noGroup := &Group{Class: "condition-no"}
		for _, item := range noSpaced {
			noGroup.Children = append(noGroup.Children, item.Element)
		}

		// Stack yes and no branches vertically
		verticalGap := cfg.VerticalGap
		totalHeight = yesBBox.Height + verticalGap + noBBox.Height
		totalWidth = max(yesBBox.Width, noBBox.Width)

		// Position yes branch
		yesGroup.Transform = "translate(" + fmtFloat((totalWidth-yesBBox.Width)/2) + ",0)"

		// Position no branch
		noGroup.Transform = "translate(" + fmtFloat((totalWidth-noBBox.Width)/2) + "," + fmtFloat(yesBBox.Height+verticalGap) + ")"

		children = append(children, yesGroup, noGroup)
	} else {
		// Only yes branch
		totalWidth = yesBBox.Width
		totalHeight = yesBBox.Height
		children = append(children, yesGroup)
	}

	// Create the content group
	contentGroup := &Group{Children: children}
	contentNode := RenderedNode{
		Element: contentGroup,
		BBox:    NewBoundingBox(0, 0, totalWidth, totalHeight),
	}

	// Wrap in a labeled box with the condition
	return r.renderLabeledBoxWithContent(condLabel, contentNode, "conditional")
}

// renderRecursiveRef renders a recursive pattern reference (?R), (?n), (?&name)
func (r *Renderer) renderRecursiveRef(ref *parser.RecursiveRef) RenderedNode {
	var label string
	switch ref.Target {
	case "R", "0":
		label = "recurse whole pattern"
	default:
		// Check if it's a number or name
		if len(ref.Target) > 0 {
			first := ref.Target[0]
			if first == '+' || first == '-' || (first >= '0' && first <= '9') {
				label = fmt.Sprintf("recurse to group %s", ref.Target)
			} else {
				label = fmt.Sprintf("recurse to '%s'", ref.Target)
			}
		} else {
			label = "recurse"
		}
	}

	return r.renderStructuralLabel(label, "recursive-ref")
}

// renderBranchReset renders a branch reset group (?|...)
func (r *Renderer) renderBranchReset(br *parser.BranchReset) RenderedNode {
	// Increment depth before rendering nested content
	r.subexpDepth++

	// Render the contained regexp
	content := r.renderRegexp(br.Regexp)

	// Decrement depth after rendering
	r.subexpDepth--

	// Determine fill color based on depth (same logic as subexp)
	currentDepth := r.subexpDepth
	var fill string
	if currentDepth == 0 {
		fill = r.Config.SubexpFill
	} else if len(r.Config.SubexpColors) > 0 {
		colorIndex := (currentDepth - 1) % len(r.Config.SubexpColors)
		fill = r.Config.SubexpColors[colorIndex]
	} else {
		fill = r.Config.SubexpFill
	}

	return r.renderSubexpBox("branch reset", content, fill)
}

// renderBacktrackControl renders a backtracking control verb (*FAIL), (*PRUNE), etc.
func (r *Renderer) renderBacktrackControl(bc *parser.BacktrackControl) RenderedNode {
	var label string
	switch bc.Verb {
	case "ACCEPT":
		label = "accept match"
	case "FAIL":
		label = "force fail"
	case "MARK":
		if bc.Arg != "" {
			label = fmt.Sprintf("mark '%s'", bc.Arg)
		} else {
			label = "mark"
		}
	case "COMMIT":
		label = "commit (no retry)"
	case "PRUNE":
		if bc.Arg != "" {
			label = fmt.Sprintf("prune '%s'", bc.Arg)
		} else {
			label = "prune"
		}
	case "SKIP":
		if bc.Arg != "" {
			label = fmt.Sprintf("skip to '%s'", bc.Arg)
		} else {
			label = "skip"
		}
	case "THEN":
		if bc.Arg != "" {
			label = fmt.Sprintf("then '%s'", bc.Arg)
		} else {
			label = "then (try next alt)"
		}
	default:
		if bc.Arg != "" {
			label = fmt.Sprintf("*%s:%s", bc.Verb, bc.Arg)
		} else {
			label = fmt.Sprintf("*%s", bc.Verb)
		}
	}

	return r.renderStructuralLabel(label, "backtrack-control")
}

// renderCallout renders a PCRE callout (?C), (?Cn), (?C"text")
func (r *Renderer) renderCallout(n *parser.Callout) RenderedNode {
	var label string
	if n.Number >= 0 {
		label = fmt.Sprintf("callout (%d)", n.Number)
	} else {
		label = fmt.Sprintf("callout \"%s\"", n.Text)
	}
	return r.renderStructuralLabel(label, "callout")
}

// renderMatch renders a sequence of fragments
func (r *Renderer) renderMatch(match *parser.Match) RenderedNode {
	if len(match.Fragments) == 0 {
		// Empty match - render as empty node
		return RenderedNode{
			Element: &Group{},
			BBox:    NewBoundingBox(0, 0, 0, 0),
		}
	}

	// Render all fragments
	items := make([]RenderedNode, len(match.Fragments))
	for i, frag := range match.Fragments {
		items[i] = r.renderMatchFragment(frag)
	}

	// Space horizontally
	spacedItems, totalBBox := SpaceHorizontally(items, r.Config.HorizontalGap)

	// Create connector path between items
	var children []SVGElement

	if len(spacedItems) > 1 {
		pb := NewPathBuilder()
		pb.MoveTo(spacedItems[0].BBox.AnchorRight, totalBBox.AnchorY)

		for i := 1; i < len(spacedItems); i++ {
			pb.LineTo(spacedItems[i].BBox.AnchorLeft, totalBBox.AnchorY)
			if i < len(spacedItems)-1 {
				pb.MoveTo(spacedItems[i].BBox.AnchorRight, totalBBox.AnchorY)
			}
		}

		connectorPath := &Path{
			D:           pb.String(),
			Stroke:      r.Config.Connector.Color,
			StrokeWidth: r.Config.Connector.StrokeWidth,
		}
		children = append(children, connectorPath)
	}

	// Add all rendered items
	for _, item := range spacedItems {
		children = append(children, item.Element)
	}

	group := &Group{
		Class:    "match",
		Children: children,
	}

	return RenderedNode{
		Element: group,
		BBox:    totalBBox,
	}
}

// renderMatchFragment renders a fragment (content with optional repeat)
func (r *Renderer) renderMatchFragment(frag *parser.MatchFragment) RenderedNode {
	content := r.renderNode(frag.Content)

	// Findings may target the MatchFragment node itself (for nested-quantifier,
	// redundant-group, etc.), so the result is passed through annotateNode.
	var result RenderedNode
	if frag.Repeat == nil {
		result = content
	} else {
		result = r.renderWithRepeat(content, frag.Repeat)
	}
	return r.annotateNode(frag, result)
}

// renderWithRepeat adds skip/loop paths for quantifiers
func (r *Renderer) renderWithRepeat(content RenderedNode, repeat *parser.Repeat) RenderedNode {
	cfg := r.Config
	curveRadius := 10.0

	hasSkip := repeat.Min == 0 // Optional: can skip content
	hasLoop := repeat.Max != 1 // Can repeat: show loop

	// Calculate extra space needed for skip/loop
	skipHeight := 0.0
	loopHeight := 0.0
	if hasSkip {
		skipHeight = curveRadius * 2
	}
	if hasLoop {
		loopHeight = curveRadius * 2
	}

	// Adjust content position
	contentOffsetY := skipHeight
	contentOffsetX := curveRadius

	// Calculate new bounding box
	width := content.BBox.Width + 2*curveRadius
	height := content.BBox.Height + skipHeight + loopHeight
	anchorY := contentOffsetY + content.BBox.AnchorY

	var children []SVGElement

	// Create skip path (above content)
	if hasSkip {
		skipPath := NewPathBuilder()
		skipPath.MoveTo(0, anchorY)
		skipPath.QuadraticTo(0, anchorY-curveRadius, curveRadius, anchorY-curveRadius)
		skipPath.HorizontalTo(width - curveRadius)
		skipPath.QuadraticTo(width, anchorY-curveRadius, width, anchorY)

		children = append(children, &Path{
			D:           skipPath.String(),
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
			Class:       "skip-path",
		})
	}

	// Create loop path (below content)
	if hasLoop {
		loopY := contentOffsetY + content.BBox.Height + curveRadius

		loopPath := NewPathBuilder()
		loopPath.MoveTo(width, anchorY)
		loopPath.QuadraticTo(width, loopY, width-curveRadius, loopY)
		loopPath.HorizontalTo(curveRadius)
		loopPath.QuadraticTo(0, loopY, 0, anchorY)

		children = append(children, &Path{
			D:           loopPath.String(),
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
			Class:       "loop-path",
		})

		// Add arrow on loop to indicate direction
		arrowX := width / 2
		arrowY := loopY
		arrowSize := 5.0

		if repeat.Greedy {
			// Arrow pointing left (greedy - tries to match more first)
			children = append(children, &Path{
				D: "M " + fmtFloat(arrowX+arrowSize) + " " + fmtFloat(arrowY-arrowSize) +
					" L " + fmtFloat(arrowX) + " " + fmtFloat(arrowY) +
					" L " + fmtFloat(arrowX+arrowSize) + " " + fmtFloat(arrowY+arrowSize),
				Stroke:      cfg.Connector.Color,
				StrokeWidth: cfg.Connector.StrokeWidth,
			})
		} else {
			// Arrow pointing right (non-greedy)
			children = append(children, &Path{
				D: "M " + fmtFloat(arrowX-arrowSize) + " " + fmtFloat(arrowY-arrowSize) +
					" L " + fmtFloat(arrowX) + " " + fmtFloat(arrowY) +
					" L " + fmtFloat(arrowX-arrowSize) + " " + fmtFloat(arrowY+arrowSize),
				Stroke:      cfg.Connector.Color,
				StrokeWidth: cfg.Connector.StrokeWidth,
			})
		}

		// Add repeat label. The label ("1+ times", "2 to 5 times") is
		// a structural description and uses the sans-serif label font
		// — the CSS class also recolors it to the connector gray.
		label := r.getRepeatLabel(repeat)
		if label != "" {
			children = append(children, &Text{
				X:          width / 2,
				Y:          loopY + cfg.FontSize,
				Content:    label,
				FontFamily: cfg.LabelFontFamily,
				FontSize:   cfg.LabelFontSize,
				Anchor:     "middle",
				Class:      "repeat-label",
			})
			height += cfg.FontSize
		}
	}

	// Add content
	contentGroup := &Group{
		Transform: "translate(" + fmtFloat(contentOffsetX) + "," + fmtFloat(contentOffsetY) + ")",
		Children:  []SVGElement{content.Element},
	}
	children = append(children, contentGroup)

	// Add connector lines from edges to content
	if hasSkip || hasLoop {
		// Left connector
		children = append(children, &Line{
			X1:          0,
			Y1:          anchorY,
			X2:          contentOffsetX,
			Y2:          anchorY,
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
		})
		// Right connector
		children = append(children, &Line{
			X1:          contentOffsetX + content.BBox.Width,
			Y1:          anchorY,
			X2:          width,
			Y2:          anchorY,
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
		})
	}

	group := &Group{
		Class:    "repeat",
		Children: children,
	}

	return RenderedNode{
		Element: group,
		BBox: BoundingBox{
			X:           0,
			Y:           0,
			Width:       width,
			Height:      height,
			AnchorLeft:  0,
			AnchorRight: width,
			AnchorY:     anchorY,
		},
	}
}

// getRepeatLabel returns the label for a repeat quantifier
func (r *Renderer) getRepeatLabel(repeat *parser.Repeat) string {
	var label string
	if repeat.Min == repeat.Max {
		if repeat.Min == 1 {
			label = ""
		} else {
			label = fmt.Sprintf("%d times", repeat.Min)
		}
	} else if repeat.Max == -1 {
		switch repeat.Min {
		case 0:
			label = "" // * quantifier - no label needed
		case 1:
			label = "" // + quantifier - no label needed
		default:
			label = fmt.Sprintf("%d+ times", repeat.Min)
		}
	} else {
		label = fmt.Sprintf("%d to %d times", repeat.Min, repeat.Max)
	}

	// Add possessive indicator
	if repeat.Possessive && label != "" {
		label += " (possessive)"
	} else if repeat.Possessive {
		label = "possessive"
	}

	return label
}

// renderRegexp renders alternation
func (r *Renderer) renderRegexp(regexp *parser.Regexp) RenderedNode {
	if len(regexp.Matches) == 0 {
		return RenderedNode{
			Element: &Group{},
			BBox:    NewBoundingBox(0, 0, 0, 0),
		}
	}

	if len(regexp.Matches) == 1 {
		// No alternation, just render the match
		return r.renderMatch(regexp.Matches[0])
	}

	// Render all alternatives
	items := make([]RenderedNode, len(regexp.Matches))
	for i, match := range regexp.Matches {
		items[i] = r.renderMatch(match)
	}

	// Space vertically
	spacedItems, totalBBox := SpaceVertically(items, r.Config.VerticalGap*2)

	cfg := r.Config
	curveRadius := 10.0
	connectorWidth := 20.0

	// Adjust for connector space
	width := totalBBox.Width + 2*connectorWidth
	height := totalBBox.Height
	anchorY := height / 2

	var children []SVGElement

	// Create connector paths
	for _, item := range spacedItems {
		itemAnchorY := item.BBox.AnchorY
		// Use actual anchor positions to account for centering by SpaceVertically
		itemLeftX := connectorWidth + item.BBox.AnchorLeft
		itemRightX := connectorWidth + item.BBox.AnchorRight

		// Use an effective radius that won't cause path reversal when branches are close
		gap := math.Abs(itemAnchorY - anchorY)
		effectiveRadius := math.Min(curveRadius, gap/2)

		// Left connector curve
		leftPath := NewPathBuilder()
		leftPath.MoveTo(0, anchorY)
		if itemAnchorY < anchorY {
			leftPath.QuadraticTo(curveRadius, anchorY, curveRadius, anchorY-effectiveRadius)
			leftPath.VerticalTo(itemAnchorY + effectiveRadius)
			leftPath.QuadraticTo(curveRadius, itemAnchorY, itemLeftX, itemAnchorY)
		} else if itemAnchorY > anchorY {
			leftPath.QuadraticTo(curveRadius, anchorY, curveRadius, anchorY+effectiveRadius)
			leftPath.VerticalTo(itemAnchorY - effectiveRadius)
			leftPath.QuadraticTo(curveRadius, itemAnchorY, itemLeftX, itemAnchorY)
		} else {
			leftPath.HorizontalTo(itemLeftX)
		}

		children = append(children, &Path{
			D:           leftPath.String(),
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
		})

		// Right connector curve
		rightPath := NewPathBuilder()
		rightPath.MoveTo(itemRightX, itemAnchorY)
		if itemAnchorY < anchorY {
			rightPath.QuadraticTo(width-curveRadius, itemAnchorY, width-curveRadius, itemAnchorY+effectiveRadius)
			rightPath.VerticalTo(anchorY - effectiveRadius)
			rightPath.QuadraticTo(width-curveRadius, anchorY, width, anchorY)
		} else if itemAnchorY > anchorY {
			rightPath.QuadraticTo(width-curveRadius, itemAnchorY, width-curveRadius, itemAnchorY-effectiveRadius)
			rightPath.VerticalTo(anchorY + effectiveRadius)
			rightPath.QuadraticTo(width-curveRadius, anchorY, width, anchorY)
		} else {
			rightPath.HorizontalTo(width)
		}

		children = append(children, &Path{
			D:           rightPath.String(),
			Stroke:      cfg.Connector.Color,
			StrokeWidth: cfg.Connector.StrokeWidth,
		})
	}

	// Add all rendered items with offset
	for _, item := range spacedItems {
		itemGroup := &Group{
			Transform: "translate(" + fmtFloat(connectorWidth) + ",0)",
			Children:  []SVGElement{item.Element},
		}
		children = append(children, itemGroup)
	}

	group := &Group{
		Class:    "regexp",
		Children: children,
	}

	return RenderedNode{
		Element: group,
		BBox: BoundingBox{
			X:           0,
			Y:           0,
			Width:       width,
			Height:      height,
			AnchorLeft:  0,
			AnchorRight: width,
			AnchorY:     anchorY,
		},
	}
}

// renderCharset renders a character class
func (r *Renderer) renderCharset(charset *parser.Charset) RenderedNode {
	if charset.SetExpression != nil {
		return r.renderCharsetSetExpression(charset)
	}

	// Render charset items
	var itemTexts []string
	for _, item := range charset.Items {
		itemTexts = append(itemTexts, r.charsetItemText(item))
	}

	label := "One of:"
	if charset.Inverted {
		label = "None of:"
	}

	return r.renderLabeledBox(label, itemTexts, "charset")
}

// charsetItemText returns the display text for a single charset item
func (r *Renderer) charsetItemText(item parser.CharsetItem) string {
	switch it := item.(type) {
	case *parser.CharsetLiteral:
		return fmt.Sprintf(`"%s"`, it.Text)
	case *parser.CharsetRange:
		return fmt.Sprintf(`"%s" - "%s"`, it.First, it.Last)
	case *parser.Escape:
		return it.Value
	case *parser.POSIXClass:
		return r.getPOSIXClassLabel(it)
	case *parser.Charset:
		return r.charsetOperandText(it)
	case *parser.UnicodePropertyEscape:
		return r.charsetOperandText(it)
	case *parser.CharsetStringDisjunction:
		return r.charsetOperandText(it)
	default:
		return fmt.Sprintf("<%s>", item.Type())
	}
}

// renderCharsetSetExpression renders a charset that uses v-mode set operations
func (r *Renderer) renderCharsetSetExpression(charset *parser.Charset) RenderedNode {
	var texts []string
	switch expr := charset.SetExpression.(type) {
	case *parser.CharsetIntersection:
		texts = r.charsetOperandTexts(expr.Operands)
		label := "Intersection:"
		if charset.Inverted {
			label = "NOT Intersection:"
		}
		return r.renderLabeledBox(label, texts, "charset")
	case *parser.CharsetSubtraction:
		texts = r.charsetOperandTexts(expr.Operands)
		label := "Subtraction:"
		if charset.Inverted {
			label = "NOT Subtraction:"
		}
		return r.renderLabeledBox(label, texts, "charset")
	default:
		return r.renderStructuralLabel("<set-expression>", "charset")
	}
}

// renderCharsetIntersection renders a CharsetIntersection node
func (r *Renderer) renderCharsetIntersection(node *parser.CharsetIntersection) RenderedNode {
	texts := r.charsetOperandTexts(node.Operands)
	return r.renderLabeledBox("Intersection:", texts, "charset")
}

// renderCharsetSubtraction renders a CharsetSubtraction node
func (r *Renderer) renderCharsetSubtraction(node *parser.CharsetSubtraction) RenderedNode {
	texts := r.charsetOperandTexts(node.Operands)
	return r.renderLabeledBox("Subtraction:", texts, "charset")
}

// renderCharsetStringDisjunction renders a \q{abc|def} string disjunction
func (r *Renderer) renderCharsetStringDisjunction(node *parser.CharsetStringDisjunction) RenderedNode {
	var items []string
	for _, s := range node.Strings {
		if s == "" {
			items = append(items, "(empty)")
		} else {
			items = append(items, fmt.Sprintf(`"%s"`, s))
		}
	}
	return r.renderLabeledBox("String:", items, "charset")
}

// charsetOperandTexts returns display strings for a slice of operand Nodes
func (r *Renderer) charsetOperandTexts(operands []parser.Node) []string {
	var texts []string
	for _, op := range operands {
		texts = append(texts, r.charsetOperandText(op))
	}
	return texts
}

// charsetOperandText converts an operand Node to a display string
func (r *Renderer) charsetOperandText(node parser.Node) string {
	switch n := node.(type) {
	case *parser.Charset:
		var inner string
		if n.SetExpression != nil {
			switch expr := n.SetExpression.(type) {
			case *parser.CharsetIntersection:
				inner = strings.Join(r.charsetOperandTexts(expr.Operands), " && ")
			case *parser.CharsetSubtraction:
				inner = strings.Join(r.charsetOperandTexts(expr.Operands), " -- ")
			}
		} else {
			var parts []string
			for _, item := range n.Items {
				parts = append(parts, r.charsetItemText(item))
			}
			inner = strings.Join(parts, ", ")
		}
		if n.Inverted {
			return "[^" + inner + "]"
		}
		return "[" + inner + "]"
	case *parser.UnicodePropertyEscape:
		if n.Negated {
			return fmt.Sprintf(`\P{%s}`, n.Property)
		}
		return fmt.Sprintf(`\p{%s}`, n.Property)
	case *parser.Escape:
		return n.Value
	case *parser.CharsetStringDisjunction:
		return fmt.Sprintf(`\q{%s}`, strings.Join(n.Strings, "|"))
	default:
		return fmt.Sprintf("<%s>", node.Type())
	}
}

// getPOSIXClassLabel returns a human-readable label for a POSIX character class
func (r *Renderer) getPOSIXClassLabel(pc *parser.POSIXClass) string {
	labels := map[string]string{
		"alnum":  "alphanumeric",
		"alpha":  "alphabetic",
		"blank":  "blank (space/tab)",
		"cntrl":  "control character",
		"digit":  "digit",
		"graph":  "visible character",
		"lower":  "lowercase",
		"print":  "printable",
		"punct":  "punctuation",
		"space":  "whitespace",
		"upper":  "uppercase",
		"xdigit": "hex digit",
	}

	label, ok := labels[pc.Name]
	if !ok {
		label = pc.Name
	}

	if pc.Negated {
		return "NOT " + label
	}
	return label
}

// renderSubexp renders a subexpression group
func (r *Renderer) renderSubexp(subexp *parser.Subexp) RenderedNode {
	// Get label based on group type
	var label string
	switch subexp.GroupType {
	case "capture":
		label = fmt.Sprintf("group #%d", subexp.Number)
	case "named_capture":
		label = fmt.Sprintf("group #%d '%s'", subexp.Number, subexp.Name)
	case "non_capture":
		label = "non-capturing group"
	case "positive_lookahead":
		label = "positive lookahead"
	case "negative_lookahead":
		label = "negative lookahead"
	case "positive_lookbehind":
		label = "positive lookbehind"
	case "negative_lookbehind":
		label = "negative lookbehind"
	case "non_atomic_positive_lookahead":
		label = "non-atomic lookahead"
	case "non_atomic_positive_lookbehind":
		label = "non-atomic lookbehind"
	case "script_run":
		label = "script run"
	case "atomic_script_run":
		label = "atomic script run"
	case "atomic":
		label = "atomic group"
	default:
		label = subexp.GroupType
	}

	// Determine fill color based on depth
	// Depth 0 (outermost) = transparent, depth 1+ = cycle through colors
	currentDepth := r.subexpDepth
	var fill string
	if currentDepth == 0 {
		fill = r.Config.SubexpFill // "none" by default
	} else if len(r.Config.SubexpColors) > 0 {
		// Cycle through colors for nested subexps (depth 1, 2, 3...)
		colorIndex := (currentDepth - 1) % len(r.Config.SubexpColors)
		fill = r.Config.SubexpColors[colorIndex]
	} else {
		fill = r.Config.SubexpFill
	}

	// Increment depth before rendering nested content
	r.subexpDepth++

	// Render the contained regexp
	content := r.renderRegexp(subexp.Regexp)

	// Decrement depth after rendering
	r.subexpDepth--

	return r.renderSubexpBox(label, content, fill)
}

// renderLabeledBox creates a labeled box with text items (for charset).
// The header (e.g. "One of:") is a structural label and uses the
// sans-serif label font, while each item ("a", "a" - "z") is regex
// content and stays in the monospace content font.
func (r *Renderer) renderLabeledBox(label string, items []string, class string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	// Calculate dimensions. Header measured as label text, items
	// measured as content text.
	labelWidth := MeasureLabelText(label, cfg)
	maxItemWidth := 0.0
	for _, item := range items {
		w := MeasureText(item, cfg)
		if w > maxItemWidth {
			maxItemWidth = w
		}
	}

	contentWidth := maxItemWidth + 2*padding
	if labelWidth > contentWidth {
		contentWidth = labelWidth
	}

	labelHeight := cfg.FontSize + padding
	itemHeight := cfg.FontSize + padding/2
	contentHeight := float64(len(items)) * itemHeight

	width := contentWidth + 2*padding
	height := labelHeight + contentHeight + padding

	var children []SVGElement

	// Background rect
	children = append(children, &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     cfg.CornerRadius,
		Ry:     cfg.CornerRadius,
	})

	// Header (structural label)
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Class:      class + "-label",
	})

	// Items (regex content)
	y := labelHeight + cfg.FontSize
	for _, item := range items {
		children = append(children, &Text{
			X:          width / 2,
			Y:          y,
			Content:    item,
			FontFamily: cfg.FontFamily,
			FontSize:   cfg.FontSize,
			Anchor:     "middle",
		})
		y += itemHeight
	}

	group := &Group{
		Class:    class,
		Children: children,
	}

	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, width, height),
	}
}

// renderSubexpBox creates a subexpression box with depth-based fill color.
// The subexp label ("group #1", "lookahead", etc.) is a structural
// label and uses the sans-serif label font.
func (r *Renderer) renderSubexpBox(label string, content RenderedNode, fill string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	labelWidth := MeasureLabelText(label, cfg)
	labelHeight := cfg.FontSize + padding

	contentWidth := content.BBox.Width
	if labelWidth > contentWidth {
		contentWidth = labelWidth
	}

	width := contentWidth + 2*padding
	height := labelHeight + content.BBox.Height + padding

	var children []SVGElement

	// Background rect with explicit fill and stroke. The subexp border
	// uses NodeStrokeWidth so it visually matches other node borders,
	// rather than pulling the connector stroke width.
	children = append(children, &Rect{
		X:           0,
		Y:           0,
		Width:       width,
		Height:      height,
		Rx:          cfg.CornerRadius,
		Ry:          cfg.CornerRadius,
		Fill:        fill,
		Stroke:      cfg.SubexpStroke,
		StrokeWidth: cfg.NodeStrokeWidth,
	})

	// Label (structural — group name / kind)
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Class:      "subexp-label",
	})

	// Content centered
	contentX := (width - content.BBox.Width) / 2
	contentY := labelHeight

	contentGroup := &Group{
		Transform: "translate(" + fmtFloat(contentX) + "," + fmtFloat(contentY) + ")",
		Children:  []SVGElement{content.Element},
	}
	children = append(children, contentGroup)

	group := &Group{
		Class:    "subexp",
		Children: children,
	}

	// Calculate anchor Y relative to content
	anchorY := contentY + content.BBox.AnchorY

	return RenderedNode{
		Element: group,
		BBox: BoundingBox{
			X:           0,
			Y:           0,
			Width:       width,
			Height:      height,
			AnchorLeft:  0,
			AnchorRight: width,
			AnchorY:     anchorY,
		},
	}
}

// renderLabeledBoxWithContent creates a labeled box containing rendered
// content. Used by scoped inline modifiers, conditionals, and similar
// constructs where the header is a structural description and the
// body is another rendered subdiagram.
func (r *Renderer) renderLabeledBoxWithContent(label string, content RenderedNode, class string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	labelWidth := MeasureLabelText(label, cfg)
	labelHeight := cfg.FontSize + padding

	contentWidth := content.BBox.Width
	if labelWidth > contentWidth {
		contentWidth = labelWidth
	}

	width := contentWidth + 2*padding
	height := labelHeight + content.BBox.Height + padding

	var children []SVGElement

	// Background rect
	children = append(children, &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     cfg.CornerRadius,
		Ry:     cfg.CornerRadius,
	})

	// Header (structural label)
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.LabelFontSize,
		Class:      class + "-label",
	})

	// Content centered
	contentX := (width - content.BBox.Width) / 2
	contentY := labelHeight

	contentGroup := &Group{
		Transform: "translate(" + fmtFloat(contentX) + "," + fmtFloat(contentY) + ")",
		Children:  []SVGElement{content.Element},
	}
	children = append(children, contentGroup)

	group := &Group{
		Class:    class,
		Children: children,
	}

	// Calculate anchor Y relative to content
	anchorY := contentY + content.BBox.AnchorY

	return RenderedNode{
		Element: group,
		BBox: BoundingBox{
			X:           0,
			Y:           0,
			Width:       width,
			Height:      height,
			AnchorLeft:  0,
			AnchorRight: width,
			AnchorY:     anchorY,
		},
	}
}
