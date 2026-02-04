package renderer

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/parser"
)

// Renderer handles rendering regex AST to SVG
type Renderer struct {
	Config      *Config
	subexpDepth int // Tracks nesting depth for subexpressions
}

// New creates a new Renderer with the given config
func New(cfg *Config) *Renderer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Renderer{Config: cfg}
}

// Render renders a regex AST to SVG
func (r *Renderer) Render(ast *parser.Regexp) string {
	rendered := r.renderRegexp(ast)

	// Add padding around the diagram
	padding := r.Config.Padding
	width := rendered.BBox.Width + 2*padding
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

	// Create start and end connectors
	startX := padding / 2
	endX := width - padding/2
	anchorY := padding + rendered.BBox.AnchorY

	startLine := &Line{
		X1:          startX,
		Y1:          anchorY,
		X2:          padding,
		Y2:          anchorY,
		Stroke:      r.Config.LineColor,
		StrokeWidth: r.Config.LineWidth,
	}

	endLine := &Line{
		X1:          width - padding - flagsWidth,
		Y1:          anchorY,
		X2:          endX - flagsWidth,
		Y2:          anchorY,
		Stroke:      r.Config.LineColor,
		StrokeWidth: r.Config.LineWidth,
	}

	// Wrap the rendered content in a group with padding offset
	contentGroup := &Group{
		Transform: fmt.Sprintf("translate(%g,%g)", padding, padding),
		Children:  []SVGElement{rendered.Element},
	}

	children := []SVGElement{
		startLine,
		endLine,
		contentGroup,
	}

	// Add flags if present
	if flagsElement != nil {
		flagsGroup := &Group{
			Transform: fmt.Sprintf("translate(%g,%g)", width-padding-flagsWidth+padding/2, padding),
			Children:  []SVGElement{flagsElement},
		}
		children = append(children, flagsGroup)
	}

	svg := &SVG{
		Width:   width,
		Height:  height,
		ViewBox: fmt.Sprintf("0 0 %g %g", width, height),
		Style:   r.getStyles(),
		Children: children,
	}

	return svg.Render()
}

// getStyles returns the CSS styles for the SVG
func (r *Renderer) getStyles() string {
	return fmt.Sprintf(`
		.literal rect { fill: %s; }
		.escape rect { fill: %s; }
		.charset rect { fill: %s; }
		.anchor rect { fill: %s; }
		.any-character rect { fill: %s; }
		.flags rect { fill: %s; }
		text { font-family: %s; font-size: %gpx; fill: %s; }
		.anchor text { fill: #fff; }
		.quote { fill: #000; }
		.subexp-label, .charset-label, .flags-label { font-size: %gpx; font-style: italic; }
		.repeat-label { fill: %s; font-size: %gpx; }
	`,
		r.Config.LiteralFill,
		r.Config.EscapeFill,
		r.Config.CharsetFill,
		r.Config.AnchorFill,
		r.Config.AnyCharFill,
		r.Config.FlagsFill,
		r.Config.FontFamily,
		r.Config.FontSize,
		r.Config.TextColor,
		r.Config.FontSize-2,
		r.Config.RepeatLabelColor,
		r.Config.FontSize-2,
	)
}

// renderNode dispatches to the appropriate render method based on node type
func (r *Renderer) renderNode(node parser.Node) RenderedNode {
	switch n := node.(type) {
	case *parser.Regexp:
		return r.renderRegexp(n)
	case *parser.Match:
		return r.renderMatch(n)
	case *parser.MatchFragment:
		return r.renderMatchFragment(n)
	case *parser.Literal:
		return r.renderLiteral(n)
	case *parser.Escape:
		return r.renderEscape(n)
	case *parser.Anchor:
		return r.renderAnchor(n)
	case *parser.AnyCharacter:
		return r.renderAnyCharacter(n)
	case *parser.Charset:
		return r.renderCharset(n)
	case *parser.Subexp:
		return r.renderSubexp(n)
	case *parser.BackReference:
		return r.renderBackReference(n)
	case *parser.UnicodePropertyEscape:
		return r.renderUnicodePropertyEscape(n)
	default:
		// Fallback: render as a simple label
		return r.renderLabel(fmt.Sprintf("<%s>", node.Type()), "unknown")
	}
}

// renderLabel creates a labeled box (used by many node types)
func (r *Renderer) renderLabel(text, class string) RenderedNode {
	cfg := r.Config
	textWidth := MeasureText(text, cfg)
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
		}
	}

	label := "Flags:"

	// Calculate dimensions
	labelWidth := MeasureText(label, cfg)
	maxItemWidth := 0.0
	for _, item := range flagItems {
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

	// Label
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2,
		Class:      "flags-label",
	})

	// Flag items
	y := labelHeight + cfg.FontSize
	for _, item := range flagItems {
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

	rect := &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Rx:     cfg.CornerRadius,
		Ry:     cfg.CornerRadius,
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

// renderAnchor renders an anchor (^, $, \b, \B)
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
	default:
		label = anchor.AnchorType
	}
	return r.renderLabel(label, "anchor")
}

// renderAnyCharacter renders the . metacharacter
func (r *Renderer) renderAnyCharacter(_ *parser.AnyCharacter) RenderedNode {
	return r.renderLabel("any character", "any-character")
}

// renderBackReference renders a back-reference like \1 or \k<name>
func (r *Renderer) renderBackReference(br *parser.BackReference) RenderedNode {
	var label string
	if br.Name != "" {
		label = fmt.Sprintf("back reference '%s'", br.Name)
	} else {
		label = fmt.Sprintf("back reference #%d", br.Number)
	}
	return r.renderLabel(label, "escape")
}

// renderUnicodePropertyEscape renders a Unicode property escape like \p{Letter} or \P{Number}
func (r *Renderer) renderUnicodePropertyEscape(upe *parser.UnicodePropertyEscape) RenderedNode {
	var label string
	if upe.Negated {
		label = fmt.Sprintf("NOT Unicode %s", upe.Property)
	} else {
		label = fmt.Sprintf("Unicode %s", upe.Property)
	}
	return r.renderLabel(label, "escape")
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
			Stroke:      r.Config.LineColor,
			StrokeWidth: r.Config.LineWidth,
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

	if frag.Repeat == nil {
		// No repeat, just return the content
		return content
	}

	// Add repeat visualization
	return r.renderWithRepeat(content, frag.Repeat)
}

// renderWithRepeat adds skip/loop paths for quantifiers
func (r *Renderer) renderWithRepeat(content RenderedNode, repeat *parser.Repeat) RenderedNode {
	cfg := r.Config
	curveRadius := 10.0

	hasSkip := repeat.Min == 0  // Optional: can skip content
	hasLoop := repeat.Max != 1  // Can repeat: show loop

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
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
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
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
			Class:       "loop-path",
		})

		// Add arrow on loop to indicate direction
		arrowX := width / 2
		arrowY := loopY
		arrowSize := 5.0

		if repeat.Greedy {
			// Arrow pointing left (greedy - tries to match more first)
			children = append(children, &Path{
				D: fmt.Sprintf("M %g %g L %g %g L %g %g",
					arrowX+arrowSize, arrowY-arrowSize,
					arrowX, arrowY,
					arrowX+arrowSize, arrowY+arrowSize),
				Stroke:      cfg.LineColor,
				StrokeWidth: cfg.LineWidth,
			})
		} else {
			// Arrow pointing right (non-greedy)
			children = append(children, &Path{
				D: fmt.Sprintf("M %g %g L %g %g L %g %g",
					arrowX-arrowSize, arrowY-arrowSize,
					arrowX, arrowY,
					arrowX-arrowSize, arrowY+arrowSize),
				Stroke:      cfg.LineColor,
				StrokeWidth: cfg.LineWidth,
			})
		}

		// Add repeat label
		label := r.getRepeatLabel(repeat)
		if label != "" {
			children = append(children, &Text{
				X:          width / 2,
				Y:          loopY + cfg.FontSize,
				Content:    label,
				FontFamily: cfg.FontFamily,
				FontSize:   cfg.FontSize - 2,
				Anchor:     "middle",
				Class:      "repeat-label",
			})
			height += cfg.FontSize
		}
	}

	// Add content
	contentGroup := &Group{
		Transform: fmt.Sprintf("translate(%g,%g)", contentOffsetX, contentOffsetY),
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
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
		})
		// Right connector
		children = append(children, &Line{
			X1:          contentOffsetX + content.BBox.Width,
			Y1:          anchorY,
			X2:          width,
			Y2:          anchorY,
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
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
	if repeat.Min == repeat.Max {
		if repeat.Min == 1 {
			return ""
		}
		return fmt.Sprintf("%d times", repeat.Min)
	}
	if repeat.Max == -1 {
		if repeat.Min == 0 {
			return "" // * quantifier - no label needed
		}
		if repeat.Min == 1 {
			return "" // + quantifier - no label needed
		}
		return fmt.Sprintf("%d+ times", repeat.Min)
	}
	return fmt.Sprintf("%d to %d times", repeat.Min, repeat.Max)
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
		itemX := connectorWidth

		// Left connector curve
		leftPath := NewPathBuilder()
		leftPath.MoveTo(0, anchorY)
		if itemAnchorY < anchorY {
			leftPath.QuadraticTo(curveRadius, anchorY, curveRadius, anchorY-curveRadius)
			leftPath.VerticalTo(itemAnchorY + curveRadius)
			leftPath.QuadraticTo(curveRadius, itemAnchorY, itemX, itemAnchorY)
		} else if itemAnchorY > anchorY {
			leftPath.QuadraticTo(curveRadius, anchorY, curveRadius, anchorY+curveRadius)
			leftPath.VerticalTo(itemAnchorY - curveRadius)
			leftPath.QuadraticTo(curveRadius, itemAnchorY, itemX, itemAnchorY)
		} else {
			leftPath.HorizontalTo(itemX)
		}

		children = append(children, &Path{
			D:           leftPath.String(),
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
		})

		// Right connector curve
		rightX := connectorWidth + item.BBox.Width
		rightPath := NewPathBuilder()
		rightPath.MoveTo(rightX, itemAnchorY)
		if itemAnchorY < anchorY {
			rightPath.QuadraticTo(width-curveRadius, itemAnchorY, width-curveRadius, itemAnchorY+curveRadius)
			rightPath.VerticalTo(anchorY - curveRadius)
			rightPath.QuadraticTo(width-curveRadius, anchorY, width, anchorY)
		} else if itemAnchorY > anchorY {
			rightPath.QuadraticTo(width-curveRadius, itemAnchorY, width-curveRadius, itemAnchorY-curveRadius)
			rightPath.VerticalTo(anchorY + curveRadius)
			rightPath.QuadraticTo(width-curveRadius, anchorY, width, anchorY)
		} else {
			rightPath.HorizontalTo(width)
		}

		children = append(children, &Path{
			D:           rightPath.String(),
			Stroke:      cfg.LineColor,
			StrokeWidth: cfg.LineWidth,
		})
	}

	// Add all rendered items with offset
	for _, item := range spacedItems {
		itemGroup := &Group{
			Transform: fmt.Sprintf("translate(%g,0)", connectorWidth),
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
	// Render charset items
	var itemTexts []string
	for _, item := range charset.Items {
		switch it := item.(type) {
		case *parser.CharsetLiteral:
			itemTexts = append(itemTexts, fmt.Sprintf(`"%s"`, it.Text))
		case *parser.CharsetRange:
			itemTexts = append(itemTexts, fmt.Sprintf(`"%s" - "%s"`, it.First, it.Last))
		case *parser.Escape:
			itemTexts = append(itemTexts, it.Value)
		case *parser.POSIXClass:
			itemTexts = append(itemTexts, r.getPOSIXClassLabel(it))
		}
	}

	label := "One of:"
	if charset.Inverted {
		label = "None of:"
	}

	return r.renderLabeledBox(label, itemTexts, "charset")
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

// renderLabeledBox creates a labeled box with text items (for charset)
func (r *Renderer) renderLabeledBox(label string, items []string, class string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	// Calculate dimensions
	labelWidth := MeasureText(label, cfg)
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

	// Label
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2,
		Class:      class + "-label",
	})

	// Items
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

// renderSubexpBox creates a subexpression box with depth-based fill color
func (r *Renderer) renderSubexpBox(label string, content RenderedNode, fill string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	labelWidth := MeasureText(label, cfg)
	labelHeight := cfg.FontSize + padding

	contentWidth := content.BBox.Width
	if labelWidth > contentWidth {
		contentWidth = labelWidth
	}

	width := contentWidth + 2*padding
	height := labelHeight + content.BBox.Height + padding

	var children []SVGElement

	// Background rect with explicit fill and stroke
	children = append(children, &Rect{
		X:           0,
		Y:           0,
		Width:       width,
		Height:      height,
		Rx:          cfg.CornerRadius,
		Ry:          cfg.CornerRadius,
		Fill:        fill,
		Stroke:      cfg.SubexpStroke,
		StrokeWidth: cfg.LineWidth,
	})

	// Label
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2,
		Class:      "subexp-label",
	})

	// Content centered
	contentX := (width - content.BBox.Width) / 2
	contentY := labelHeight

	contentGroup := &Group{
		Transform: fmt.Sprintf("translate(%g,%g)", contentX, contentY),
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

// renderLabeledBoxWithContent creates a labeled box containing rendered content
func (r *Renderer) renderLabeledBoxWithContent(label string, content RenderedNode, class string) RenderedNode {
	cfg := r.Config
	padding := cfg.Padding

	labelWidth := MeasureText(label, cfg)
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

	// Label
	children = append(children, &Text{
		X:          padding,
		Y:          cfg.FontSize,
		Content:    label,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2,
		Class:      class + "-label",
	})

	// Content centered
	contentX := (width - content.BBox.Width) / 2
	contentY := labelHeight

	contentGroup := &Group{
		Transform: fmt.Sprintf("translate(%g,%g)", contentX, contentY),
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
