package renderer

// NodeStyle bundles the colors for a rendered node category. One entry
// lives in Config.NodeStyles per node type ("literal", "charset", ...).
// CornerRadius is optional; when zero, callers fall back to
// Config.CornerRadius. This keeps the theming contract narrow —
// replacing a theme is a matter of replacing the NodeStyles map.
type NodeStyle struct {
	Fill         string
	Stroke       string
	TextColor    string
	CornerRadius float64 // 0 = inherit Config.CornerRadius
}

// ConnectorStyle groups the look of the "railroad track" (connector
// lines between nodes, loop/skip curves, start/end terminators).
// Keeping these in their own struct means a theme can tune the
// trackwork independently of the nodes.
type ConnectorStyle struct {
	Color       string
	StrokeWidth float64
	StartMarker string // "arrow" | "none"
	EndMarker   string // "dot" | "none"
}

// Config holds all styling and dimension configuration
type Config struct {
	// ================================================================
	// Dimensions
	// ================================================================
	Padding       float64
	HorizontalGap float64
	VerticalGap   float64
	CornerRadius  float64

	// ================================================================
	// Typography
	// ================================================================
	// Regex-content text (literals, escape labels, charset items) uses
	// the monospace family — it is code, and should read as code.
	FontFamily string
	FontSize   float64
	CharWidth  float64 // Approximate character width for content text

	// Structural labels (anchor descriptions, "one of" headers, repeat
	// labels, group names) use a sans-serif family. The contrast with
	// the monospace content creates a visual hierarchy between "what
	// the regex says" and "what regolith says about it".
	LabelFontFamily string
	LabelFontSize   float64
	LabelCharWidth  float64

	// ================================================================
	// Global stroke / background
	// ================================================================
	// BackgroundColor is theme-advisory metadata: each theme sets it to
	// the background color it was designed against. It is not rendered
	// directly — the renderer only emits a background <rect> when
	// BackgroundFill is non-empty. Keeping these two fields separate
	// means a theme can suggest a background color without forcing every
	// rendered SVG (including historical golden files) to suddenly grow
	// an opaque backdrop.
	BackgroundColor string
	// BackgroundFill, when non-empty, causes the renderer to inject a
	// <rect> filling the entire viewBox as the first child of the root
	// <svg>. Set by the --background-fill CLI flag; themes leave it
	// alone.
	BackgroundFill  string
	TextColor       string  // Fallback for text without a category color
	NodeStrokeWidth float64 // Default stroke width for node borders

	// ================================================================
	// Node palette
	// ================================================================
	// NodeStyles is keyed by the CSS class name used for each node type
	// ("literal", "charset", "escape", "anchor", "any-character",
	// "flags", "recursive-ref", "callout", "backtrack-control",
	// "conditional", "comment"). A theme feature (see issue #5) will
	// ship by replacing this map wholesale.
	NodeStyles map[string]NodeStyle

	// Subexpression styling is depth-cycled and does not fit the
	// category-keyed map. It stays as flat fields for now.
	SubexpFill   string   // Used for outermost subexp (depth 0)
	SubexpStroke string   // Stroke color for subexp boxes
	SubexpColors []string // Colors cycled through for nested depths (1+)

	// RepeatLabelColor is the color of the "1+ times" style labels
	// below repeat loops. Defaulted to the connector color so loops
	// and their labels read as one unit, but kept as its own field so
	// a theme could override independently.
	RepeatLabelColor string

	// ================================================================
	// Connectors
	// ================================================================
	Connector ConnectorStyle

	// ================================================================
	// Analysis annotation colors (used by annotated SVG output)
	// ================================================================
	// These are severity-driven, not category-driven, and stay
	// unchanged by themes that only swap NodeStyles.
	ErrorBorderColor   string
	WarningBorderColor string
	InfoBorderColor    string
	ErrorBadgeColor    string
	WarningBadgeColor  string
	InfoBadgeColor     string
}

// GetNodeStyle returns the style bundle for a node class, falling back
// to a neutral gray default if the class is not registered. This lets
// the renderer treat unknown categories gracefully rather than panicking
// on a missing map entry.
func (c *Config) GetNodeStyle(class string) NodeStyle {
	if style, ok := c.NodeStyles[class]; ok {
		return style
	}
	return NodeStyle{
		Fill:      "#f3f4f6",
		Stroke:    "#9ca3af",
		TextColor: "#374151",
	}
}

// DefaultConfig returns the default styling configuration — the refreshed
// style shipped with the visual refresh (issue #2).
func DefaultConfig() *Config {
	return &Config{
		// Dimensions. Spacing stayed constant across the refresh; only
		// corner radius changed (3 -> 8) for the rounder silhouette.
		Padding:       10,
		HorizontalGap: 10,
		VerticalGap:   5,
		CornerRadius:  8,

		// Typography. Content font is a smidge smaller (14 -> 13) to
		// read closer in weight to the new sans-serif label font.
		// CharWidth is recalibrated for 13px monospace (~0.6 * size).
		FontFamily:      "monospace",
		FontSize:        13,
		CharWidth:       7.8,
		LabelFontFamily: "system-ui, -apple-system, sans-serif",
		LabelFontSize:   11,
		// LabelCharWidth is an overestimate for system-ui at 11px
		// so measured boxes comfortably fit the actual rendered text
		// even for wide-glyph-heavy strings (analysis findings,
		// long descriptions). System-ui averages around 6.5-7 per
		// glyph but capitals, "m", "w", and digits push the
		// effective average closer to 8 for English prose.
		LabelCharWidth: 8.0,

		// Background / baseline text / node stroke
		BackgroundColor: "transparent",
		TextColor:       "#000",
		NodeStrokeWidth: 1.5,

		// ============================================================
		// Node palette
		// ============================================================
		// Each category is a light fill + a strong colored border + a
		// dark tint of the category hue for text. The combination gives
		// both a clear category cue and readable contrast. Anchors are
		// the exception: dark slate background with pale text, because
		// position assertions read more naturally as "stop marker".
		NodeStyles: map[string]NodeStyle{
			"literal":           {Fill: "#fee2e2", Stroke: "#ef4444", TextColor: "#991b1b"},
			"charset":           {Fill: "#f5f0e1", Stroke: "#a39e8a", TextColor: "#57534e"},
			"escape":            {Fill: "#ecfccb", Stroke: "#84cc16", TextColor: "#365314"},
			"anchor":            {Fill: "#334155", Stroke: "#1e293b", TextColor: "#e2e8f0", CornerRadius: 14},
			"any-character":     {Fill: "#dbeafe", Stroke: "#3b82f6", TextColor: "#1e3a5f"},
			"flags":             {Fill: "#dbeafe", Stroke: "#3b82f6", TextColor: "#1e3a5f"},
			"recursive-ref":     {Fill: "#ede9fe", Stroke: "#8b5cf6", TextColor: "#4c1d95"},
			"callout":           {Fill: "#fff7ed", Stroke: "#f97316", TextColor: "#7c2d12"},
			"backtrack-control": {Fill: "#fee2e2", Stroke: "#ef4444", TextColor: "#991b1b"},
			"conditional":       {Fill: "#e0f2fe", Stroke: "#0ea5e9", TextColor: "#0c4a6e"},
			"comment":           {Fill: "#f3f4f6", Stroke: "#9ca3af", TextColor: "#6b7280"},
		},

		// Subexpressions get a transparent outer box (so nested content
		// is what catches the eye) and a cycling palette for nested
		// depths. These values match the prior palette — they've held
		// up well for accessibility and color-blindness, and the
		// refresh didn't touch them.
		SubexpFill:   "none",
		SubexpStroke: "#908c83",
		SubexpColors: []string{
			"#cce5ff", // Light blue
			"#d4edda", // Light green
			"#fff3cd", // Light yellow
			"#f8d7da", // Light pink
			"#e2d5f0", // Light lavender
		},
		RepeatLabelColor: "#64748b", // matches Connector.Color by default

		// ============================================================
		// Connectors
		// ============================================================
		// A mid slate gray is visible enough to give the diagram
		// structure but quiet enough that the colorful nodes dominate
		// visually. Arrow/dot terminators give the eye clear anchor
		// points at the ends of each diagram.
		Connector: ConnectorStyle{
			Color:       "#64748b",
			StrokeWidth: 1.5,
			StartMarker: "arrow",
			EndMarker:   "dot",
		},

		// Analysis annotation colors — unchanged by the visual refresh.
		// Error/warning/info severities have their own established
		// color language that the refresh deliberately doesn't touch.
		ErrorBorderColor:   "#e53e3e",
		WarningBorderColor: "#dd6b20",
		InfoBorderColor:    "#3182ce",
		ErrorBadgeColor:    "#e53e3e",
		WarningBadgeColor:  "#dd6b20",
		InfoBadgeColor:     "#3182ce",
	}
}
