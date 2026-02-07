package renderer

import (
	"fmt"
	"strings"

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

	// Create start and end connectors
	startX := padding / 2
	endX := width - padding/2
	anchorY := bannerHeight + padding + rendered.BBox.AnchorY

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
		Transform: fmt.Sprintf("translate(%g,%g)", padding, bannerHeight+padding),
		Children:  []SVGElement{rendered.Element},
	}

	children := []SVGElement{
		startLine,
		endLine,
		contentGroup,
	}

	// Add banner if present
	if bannerElement != nil {
		bannerGroup := &Group{
			Transform: fmt.Sprintf("translate(%g,%g)", padding, padding/2),
			Children:  []SVGElement{bannerElement},
		}
		children = append(children, bannerGroup)
	}

	// Add flags if present
	if flagsElement != nil {
		flagsGroup := &Group{
			Transform: fmt.Sprintf("translate(%g,%g)", width-padding-flagsWidth+padding/2, bannerHeight+padding),
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

// renderPatternOptions renders PCRE pattern start options as a banner
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

	textWidth := MeasureText(label, cfg)
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
		StrokeWidth: cfg.LineWidth,
	}

	textElem := &Text{
		X:          width / 2,
		Y:          height/2 + cfg.FontSize/3,
		Content:    label,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2,
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

// getStyles returns the CSS styles for the SVG
func (r *Renderer) getStyles() string {
	return fmt.Sprintf(`
		.literal rect { fill: %s; }
		.escape rect { fill: %s; }
		.charset rect { fill: %s; }
		.anchor rect { fill: %s; }
		.any-character rect { fill: %s; }
		.flags rect { fill: %s; }
		.comment rect { fill: #e8e8e8; stroke: #999; stroke-dasharray: 4,2; }
		.comment text { fill: #666; font-style: italic; }
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
	case *parser.QuotedLiteral:
		return r.renderQuotedLiteral(n)
	case *parser.Comment:
		return r.renderComment(n)
	case *parser.InlineModifier:
		return r.renderInlineModifier(n)
	case *parser.BalancedGroup:
		return r.renderBalancedGroup(n)
	case *parser.Conditional:
		return r.renderConditional(n)
	case *parser.RecursiveRef:
		return r.renderRecursiveRef(n)
	case *parser.BranchReset:
		return r.renderBranchReset(n)
	case *parser.BacktrackControl:
		return r.renderBacktrackControl(n)
	case *parser.Callout:
		return r.renderCallout(n)
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

// renderQuotedLiteral renders a \Q...\E quoted literal sequence
func (r *Renderer) renderQuotedLiteral(ql *parser.QuotedLiteral) RenderedNode {
	return r.renderQuotedLabel(ql.Text, "literal")
}

// renderComment renders a (?#...) inline comment
func (r *Renderer) renderComment(comment *parser.Comment) RenderedNode {
	cfg := r.Config
	text := "# " + comment.Text
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
		Y:          height/2 + cfg.FontSize/3,
		Content:    text,
		FontFamily: cfg.FontFamily,
		FontSize:   cfg.FontSize - 2, // Slightly smaller for comments
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
	return r.renderLabel(label, "flags")
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
		if c.Target == "R" {
			condLabel = "if in recursion"
		} else if c.Target == "DEFINE" || c.Target == "" {
			condLabel = "DEFINE"
		} else {
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
	yesLabel := r.renderLabel("then", "condition-label")

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
		noLabel := r.renderLabel("else", "condition-label")

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
		yesGroup.Transform = fmt.Sprintf("translate(%.2f,0)", (totalWidth-yesBBox.Width)/2)

		// Position no branch
		noGroup.Transform = fmt.Sprintf("translate(%.2f,%.2f)", (totalWidth-noBBox.Width)/2, yesBBox.Height+verticalGap)

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

	return r.renderLabel(label, "recursive-ref")
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

	return r.renderLabel(label, "backtrack-control")
}

// renderCallout renders a PCRE callout (?C), (?Cn), (?C"text")
func (r *Renderer) renderCallout(n *parser.Callout) RenderedNode {
	var label string
	if n.Number >= 0 {
		label = fmt.Sprintf("callout (%d)", n.Number)
	} else {
		label = fmt.Sprintf("callout \"%s\"", n.Text)
	}
	return r.renderLabel(label, "callout")
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
	var label string
	if repeat.Min == repeat.Max {
		if repeat.Min == 1 {
			label = ""
		} else {
			label = fmt.Sprintf("%d times", repeat.Min)
		}
	} else if repeat.Max == -1 {
		if repeat.Min == 0 {
			label = "" // * quantifier - no label needed
		} else if repeat.Min == 1 {
			label = "" // + quantifier - no label needed
		} else {
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
