package renderer

import (
	"fmt"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/parser"
)

// ================================================================================
// Annotated SVG Rendering
// ================================================================================

// RenderAnnotated produces an SVG railroad diagram with analysis overlay
// annotations. If the report has no findings, it delegates to the normal
// Render method so that the output is identical to an unannotated diagram.
//
// When findings are present, each AST node that triggered a finding gets a
// dashed severity border and a circle badge. A legend summarizing all
// findings is appended below the diagram.
func (r *Renderer) RenderAnnotated(root *parser.Regexp, report *analyzer.AnalysisReport) string {
	if len(report.Findings) == 0 {
		return r.Render(root)
	}

	// Build the mapping from AST node pointers to their worst-severity finding.
	r.nodeFindings = buildNodeFindingMap(report.Findings)

	// Render the diagram. Because nodeFindings is non-nil, annotateNode will
	// add overlays to any node that has a finding.
	rendered := r.renderRegexp(root)

	// Clear the map so subsequent Render calls are unaffected.
	defer func() { r.nodeFindings = nil }()

	padding := r.Config.Padding
	leftMargin := contentLeftMargin(padding)
	rightMargin := contentRightMargin(padding)

	// Render the legend below the diagram.
	legend := r.renderLegend(report.Findings)

	// diagramWidth is the width the diagram itself needs (with marker
	// margins); totalWidth also accounts for the legend, which may be
	// wider than the diagram. These two widths are tracked separately
	// so the end connector / dot marker can terminate at the diagram's
	// right edge instead of stretching across the legend area.
	legendGap := padding
	diagramWidth := rendered.BBox.Width + leftMargin + rightMargin
	legendWidth := legend.BBox.Width + 2*padding
	totalWidth := diagramWidth
	if legendWidth > totalWidth {
		totalWidth = legendWidth
	}

	diagramHeight := rendered.BBox.Height + 2*padding
	totalHeight := diagramHeight + legendGap + legend.BBox.Height + padding

	// Check for flags and render them.
	var flagsElement SVGElement
	var flagsRendered RenderedNode
	flagsWidth := 0.0
	if root.Flags != "" {
		flagsRendered = r.renderFlags(root.Flags)
		flagsElement = flagsRendered.Element
		flagsWidth = flagsRendered.BBox.Width + padding
		diagramWidth += flagsWidth
		if diagramWidth > totalWidth {
			totalWidth = diagramWidth
		}
		if flagsRendered.BBox.Height+2*padding > diagramHeight {
			diagramHeight = flagsRendered.BBox.Height + 2*padding
			totalHeight = diagramHeight + legendGap + legend.BBox.Height + padding
		}
	}

	// Check for pattern start options (PCRE).
	var bannerHeight float64
	var bannerElement SVGElement
	if len(root.Options) > 0 {
		bannerRendered := r.renderPatternOptions(root.Options)
		bannerElement = bannerRendered.Element
		bannerHeight = bannerRendered.BBox.Height + padding/2

		bannerWidth := bannerRendered.BBox.Width + 2*padding
		if bannerWidth > totalWidth {
			totalWidth = bannerWidth
		}
		totalHeight += bannerHeight
	}

	// Start/end connector lines. The end line is computed from
	// diagramWidth (not totalWidth) so the end dot stays at the visual
	// end of the diagram even when the legend below is wider.
	anchorY := bannerHeight + padding + rendered.BBox.AnchorY
	startX := padding / 2
	contentEndX := diagramWidth - rightMargin - flagsWidth
	endLineLength := float64(visibleConnectorWidth + endDotRadius)

	startLine := &Line{
		X1: startX, Y1: anchorY,
		X2: leftMargin, Y2: anchorY,
		Stroke:      r.Config.Connector.Color,
		StrokeWidth: r.Config.Connector.StrokeWidth,
		MarkerStart: startMarkerRef(r.Config.Connector.StartMarker),
	}
	endLine := &Line{
		X1: contentEndX, Y1: anchorY,
		X2: contentEndX + endLineLength, Y2: anchorY,
		Stroke:      r.Config.Connector.Color,
		StrokeWidth: r.Config.Connector.StrokeWidth,
		MarkerEnd:   endMarkerRef(r.Config.Connector.EndMarker),
	}

	contentGroup := &Group{
		Transform: "translate(" + fmtFloat(leftMargin) + "," + fmtFloat(bannerHeight+padding) + ")",
		Children:  []SVGElement{rendered.Element},
	}

	children := []SVGElement{
		startLine,
		endLine,
		contentGroup,
	}

	if bannerElement != nil {
		children = append(children, &Group{
			Transform: "translate(" + fmtFloat(padding) + "," + fmtFloat(padding/2) + ")",
			Children:  []SVGElement{bannerElement},
		})
	}

	if flagsElement != nil {
		// Flags box sits against the right edge of the diagram
		// (not the total width) so it stays aligned with the
		// railroad diagram when the legend is wider.
		children = append(children, &Group{
			Transform: "translate(" + fmtFloat(diagramWidth-padding-flagsWidth+padding/2) + "," + fmtFloat(bannerHeight+padding) + ")",
			Children:  []SVGElement{flagsElement},
		})
	}

	// Place legend below the diagram.
	legendY := bannerHeight + diagramHeight + legendGap
	children = append(children, &Group{
		Transform: "translate(" + fmtFloat(padding) + "," + fmtFloat(legendY) + ")",
		Children:  []SVGElement{legend.Element},
		Class:     "analysis-legend",
	})

	svg := &SVG{
		Width:    totalWidth,
		Height:   totalHeight,
		ViewBox:  "0 0 " + fmtFloat(totalWidth) + " " + fmtFloat(totalHeight),
		Defs:     r.getDefs(),
		Style:    r.getStyles() + r.getAnnotationStyles(),
		Children: children,
	}

	return svg.Render()
}

// ================================================================================
// Node-to-Finding Mapping
// ================================================================================

// buildNodeFindingMap creates a mapping from AST node pointers to the
// worst-severity finding that targets each node. When multiple findings
// target the same node, only the highest-severity one is kept — that
// controls the border color and badge displayed on the annotated diagram.
//
// Because parser.Node and ast.Node are the same interface type via type
// aliases, the map lookup works with pointer identity across packages.
func buildNodeFindingMap(findings []*analyzer.Finding) map[parser.Node]*analyzer.Finding {
	m := make(map[parser.Node]*analyzer.Finding)
	for _, f := range findings {
		if f.Node == nil {
			continue
		}
		existing, ok := m[f.Node]
		if !ok || f.Severity > existing.Severity {
			m[f.Node] = f
		}
	}
	return m
}

// ================================================================================
// Annotation Overlay
// ================================================================================

// annotateNode wraps a rendered node with a dashed severity border and a
// circle badge when the corresponding AST node has a finding. This method
// is called from renderNode and renderMatchFragment for every node, but it
// is a deliberate no-op when r.nodeFindings is nil (normal rendering) or
// when the node has no associated finding.
func (r *Renderer) annotateNode(node parser.Node, rendered RenderedNode) RenderedNode {
	if r.nodeFindings == nil {
		return rendered
	}
	finding, ok := r.nodeFindings[node]
	if !ok {
		return rendered
	}

	cfg := r.Config
	annotPadding := cfg.Padding / 2
	badgeRadius := 8.0

	// Dashed border rect surrounding the rendered element.
	borderRect := &Rect{
		X:               rendered.BBox.X - annotPadding,
		Y:               rendered.BBox.Y - annotPadding,
		Width:           rendered.BBox.Width + 2*annotPadding,
		Height:          rendered.BBox.Height + 2*annotPadding,
		Rx:              cfg.CornerRadius,
		Ry:              cfg.CornerRadius,
		Fill:            "none",
		Stroke:          severityBorderColor(finding.Severity, cfg),
		StrokeWidth:     2,
		StrokeDashArray: "6,3",
		Class:           "analysis-border",
	}

	// Circle badge in the top-right corner of the border.
	badgeCx := rendered.BBox.X + rendered.BBox.Width + annotPadding - badgeRadius/2
	badgeCy := rendered.BBox.Y - annotPadding + badgeRadius/2
	badge := &Circle{
		Cx:    badgeCx,
		Cy:    badgeCy,
		R:     badgeRadius,
		Fill:  severityBadgeColor(finding.Severity, cfg),
		Class: "analysis-badge",
	}

	// Single-character label inside the badge (!, ⚠, i).
	badgeLabel := &Text{
		X:       badgeCx,
		Y:       badgeCy + 4,
		Content: severityBadgeChar(finding.Severity),
		Anchor:  "middle",
		Fill:    "#fff",
		Class:   "analysis-badge-label",
	}

	// Tooltip with the finding title.
	tooltip := &Title{Content: finding.Title}

	group := &Group{
		Children: []SVGElement{rendered.Element, borderRect, badge, badgeLabel, tooltip},
	}

	// The bounding box grows to include the annotation padding and badge.
	newBBox := BoundingBox{
		X:           rendered.BBox.X - annotPadding,
		Y:           rendered.BBox.Y - annotPadding,
		Width:       rendered.BBox.Width + 2*annotPadding,
		Height:      rendered.BBox.Height + 2*annotPadding,
		AnchorLeft:  rendered.BBox.AnchorLeft,
		AnchorRight: rendered.BBox.AnchorRight,
		AnchorY:     rendered.BBox.AnchorY,
	}

	return RenderedNode{Element: group, BBox: newBBox}
}

// ================================================================================
// Severity Helpers
// ================================================================================

// severityBorderColor returns the dashed-border stroke color for the given
// severity level. Critical and error share the error color; warning and info
// each have their own.
func severityBorderColor(sev analyzer.Severity, cfg *Config) string {
	switch sev {
	case analyzer.SeverityCritical, analyzer.SeverityError:
		return cfg.ErrorBorderColor
	case analyzer.SeverityWarning:
		return cfg.WarningBorderColor
	default:
		return cfg.InfoBorderColor
	}
}

// severityBadgeColor returns the fill color for the circle badge.
func severityBadgeColor(sev analyzer.Severity, cfg *Config) string {
	switch sev {
	case analyzer.SeverityCritical, analyzer.SeverityError:
		return cfg.ErrorBadgeColor
	case analyzer.SeverityWarning:
		return cfg.WarningBadgeColor
	default:
		return cfg.InfoBadgeColor
	}
}

// severityBadgeChar returns the single character shown inside the badge
// circle: "!" for error/critical, a warning symbol for warning, "i" for info.
func severityBadgeChar(sev analyzer.Severity) string {
	switch sev {
	case analyzer.SeverityCritical, analyzer.SeverityError:
		return "!"
	case analyzer.SeverityWarning:
		return "\u26a0" // ⚠
	default:
		return "i"
	}
}

// ================================================================================
// Legend
// ================================================================================

// renderLegend produces the SVG elements for the findings legend that is
// placed below the railroad diagram. Each finding gets a row with a colored
// severity indicator and the finding title and description.
//
// The legend is descriptive prose, not regex content, so every text
// element uses the sans-serif label font. The header stays at
// Config.FontSize (13) to give it slight prominence over the 11px
// finding body text.
func (r *Renderer) renderLegend(findings []*analyzer.Finding) RenderedNode {
	cfg := r.Config
	lineHeight := cfg.FontSize + 6
	indicatorR := 5.0
	leftMargin := 20.0
	y := cfg.FontSize // initial baseline

	var children []SVGElement

	// Legend header.
	children = append(children, &Text{
		X:          0,
		Y:          y,
		Content:    "Analysis Findings",
		FontFamily: cfg.LabelFontFamily,
		FontSize:   cfg.FontSize,
		Fill:       cfg.TextColor,
		Class:      "analysis-legend-title",
	})
	y += lineHeight

	maxWidth := MeasureLabelText("Analysis Findings", cfg)

	for _, f := range findings {
		// Colored circle indicator.
		children = append(children, &Circle{
			Cx:   indicatorR,
			Cy:   y - indicatorR + 2,
			R:    indicatorR,
			Fill: severityBadgeColor(f.Severity, cfg),
		})

		// Severity + title.
		label := fmt.Sprintf("[%s] %s", f.Severity, f.Title)
		children = append(children, &Text{
			X:          leftMargin,
			Y:          y,
			Content:    label,
			FontFamily: cfg.LabelFontFamily,
			FontSize:   cfg.FontSize - 1,
			Fill:       cfg.TextColor,
		})
		labelW := leftMargin + MeasureLabelText(label, cfg)
		if labelW > maxWidth {
			maxWidth = labelW
		}
		y += lineHeight

		// Description (indented, slightly smaller).
		if f.Description != "" {
			children = append(children, &Text{
				X:          leftMargin,
				Y:          y,
				Content:    f.Description,
				FontFamily: cfg.LabelFontFamily,
				FontSize:   cfg.LabelFontSize,
				Fill:       "#666",
			})
			descW := leftMargin + MeasureLabelText(f.Description, cfg)
			if descW > maxWidth {
				maxWidth = descW
			}
			y += lineHeight
		}

		// Suggestion (indented, italic via class).
		if f.Suggestion != "" {
			children = append(children, &Text{
				X:          leftMargin,
				Y:          y,
				Content:    f.Suggestion,
				FontFamily: cfg.LabelFontFamily,
				FontSize:   cfg.LabelFontSize,
				Fill:       "#888",
				Class:      "analysis-suggestion",
			})
			suggW := leftMargin + MeasureLabelText(f.Suggestion, cfg)
			if suggW > maxWidth {
				maxWidth = suggW
			}
			y += lineHeight
		}
	}

	totalHeight := y
	group := &Group{Children: children}
	return RenderedNode{
		Element: group,
		BBox:    NewBoundingBox(0, 0, maxWidth, totalHeight),
	}
}

// ================================================================================
// Annotation CSS
// ================================================================================

// getAnnotationStyles returns additional CSS rules for annotation elements.
// These are appended to the base stylesheet only when RenderAnnotated is used.
func (r *Renderer) getAnnotationStyles() string {
	return fmt.Sprintf(`
		.analysis-border { pointer-events: none; }
		.analysis-badge-label { font-size: %spx; font-weight: bold; pointer-events: none; }
		.analysis-legend-title { font-weight: bold; }
		.analysis-suggestion { font-style: italic; }
	`, fmtFloat(r.Config.FontSize-3))
}
