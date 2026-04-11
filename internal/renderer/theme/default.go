package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// The light and dark themes are regolith's own "house" palettes. The
// light variant is a literal transcription of the refreshed palette
// shipped via renderer.DefaultConfig() — it exists as a registered
// theme so users have a canonical name for "the built-in style" and
// so a future config-file / env-var resolver can promote a real
// theme as the install-wide default without fighting a no-op alias.
// The dark variant is the conceptual inverse: a slate-900 field with
// the same category hues reused as bright strokes, dark-tinted fills,
// and pale in-hue text. The anchor category is inverted to a pale
// pill so position assertions still read as "stop marker" rather
// than blending into the dark background.
//
// Both themes respect the theming contract: they only touch color
// fields of renderer.Config. Dimensions, typography, stroke widths,
// markers, and analysis annotation colors are left alone, exactly
// like every other theme in this package.

// palette collects the named colors a theme needs to paint. Every
// entry is a hex literal so TestApplyUsesValidColors passes without
// special cases. Fields are intentionally flat (rather than reused
// via a "muted" alias) so the light palette can be a byte-for-byte
// transcription of renderer.DefaultConfig() without coincidental
// overlap between comment, subexp, and connector colors.
type palette struct {
	bg, fg string

	subexpStroke string
	subexpCycle  []string

	connectorColor string
	repeatLabel    string

	commentFill, commentStroke, commentText string

	literalFill, literalStroke, literalText       string
	charsetFill, charsetStroke, charsetText       string
	escapeFill, escapeStroke, escapeText          string
	anchorFill, anchorStroke, anchorText          string
	blueFill, blueStroke, blueText                string
	recursiveFill, recursiveStroke, recursiveText string
	calloutFill, calloutStroke, calloutText       string
	conditionalFill, conditionalStroke            string
	conditionalText                               string
}

// lightPalette is the refreshed default palette, lifted verbatim from
// renderer.DefaultConfig() so light.svg renders byte-for-byte the same
// as default.svg (aside from BackgroundColor, which Apply sets to a
// real hex here so the theme-validation tests accept it).
var lightPalette = palette{
	bg: "#ffffff",
	fg: "#000000",

	subexpStroke: "#908c83",
	subexpCycle: []string{
		"#cce5ff",
		"#d4edda",
		"#fff3cd",
		"#f8d7da",
		"#e2d5f0",
	},

	connectorColor: "#64748b",
	repeatLabel:    "#64748b",

	commentFill:   "#f3f4f6",
	commentStroke: "#9ca3af",
	commentText:   "#6b7280",

	literalFill:   "#fee2e2",
	literalStroke: "#ef4444",
	literalText:   "#991b1b",

	charsetFill:   "#f5f0e1",
	charsetStroke: "#a39e8a",
	charsetText:   "#57534e",

	escapeFill:   "#ecfccb",
	escapeStroke: "#84cc16",
	escapeText:   "#365314",

	anchorFill:   "#334155",
	anchorStroke: "#1e293b",
	anchorText:   "#e2e8f0",

	blueFill:   "#dbeafe",
	blueStroke: "#3b82f6",
	blueText:   "#1e3a5f",

	recursiveFill:   "#ede9fe",
	recursiveStroke: "#8b5cf6",
	recursiveText:   "#4c1d95",

	calloutFill:   "#fff7ed",
	calloutStroke: "#f97316",
	calloutText:   "#7c2d12",

	conditionalFill:   "#e0f2fe",
	conditionalStroke: "#0ea5e9",
	conditionalText:   "#0c4a6e",
}

// darkPalette is the slate-900 counterpart. Accent strokes are the
// same hues as the light palette so readers who transfer between
// themes still recognize "red == literal, green == escape, blue ==
// any-character". Fills are deep dark-mode tints, text is a pale
// shade of the stroke hue, and anchors flip to a pale pill so the
// position-assertion cue does not disappear against the background.
var darkPalette = palette{
	bg: "#0f172a",
	fg: "#e2e8f0",

	subexpStroke: "#475569",
	subexpCycle: []string{
		"#1e3a5f",
		"#14532d",
		"#713f12",
		"#831843",
		"#4c1d95",
	},

	connectorColor: "#94a3b8",
	repeatLabel:    "#94a3b8",

	commentFill:   "#1f2937",
	commentStroke: "#6b7280",
	commentText:   "#9ca3af",

	literalFill:   "#3f1d1d",
	literalStroke: "#ef4444",
	literalText:   "#fecaca",

	charsetFill:   "#3d3a2a",
	charsetStroke: "#a39e8a",
	charsetText:   "#e7e5e4",

	escapeFill:   "#1a3e1f",
	escapeStroke: "#84cc16",
	escapeText:   "#d9f99d",

	anchorFill:   "#cbd5e1",
	anchorStroke: "#e2e8f0",
	anchorText:   "#0f172a",

	blueFill:   "#172554",
	blueStroke: "#3b82f6",
	blueText:   "#bfdbfe",

	recursiveFill:   "#2e1065",
	recursiveStroke: "#8b5cf6",
	recursiveText:   "#ddd6fe",

	calloutFill:   "#431407",
	calloutStroke: "#f97316",
	calloutText:   "#fed7aa",

	conditionalFill:   "#082f49",
	conditionalStroke: "#0ea5e9",
	conditionalText:   "#bae6fd",
}

func applyPalette(c *renderer.Config, p palette) {
	c.BackgroundColor = p.bg
	c.TextColor = p.fg

	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: p.literalFill, Stroke: p.literalStroke, TextColor: p.literalText},
		"charset":           {Fill: p.charsetFill, Stroke: p.charsetStroke, TextColor: p.charsetText},
		"escape":            {Fill: p.escapeFill, Stroke: p.escapeStroke, TextColor: p.escapeText},
		"anchor":            {Fill: p.anchorFill, Stroke: p.anchorStroke, TextColor: p.anchorText, CornerRadius: 14},
		"any-character":     {Fill: p.blueFill, Stroke: p.blueStroke, TextColor: p.blueText},
		"flags":             {Fill: p.blueFill, Stroke: p.blueStroke, TextColor: p.blueText},
		"recursive-ref":     {Fill: p.recursiveFill, Stroke: p.recursiveStroke, TextColor: p.recursiveText},
		"callout":           {Fill: p.calloutFill, Stroke: p.calloutStroke, TextColor: p.calloutText},
		"backtrack-control": {Fill: p.literalFill, Stroke: p.literalStroke, TextColor: p.literalText},
		"conditional":       {Fill: p.conditionalFill, Stroke: p.conditionalStroke, TextColor: p.conditionalText},
		"comment":           {Fill: p.commentFill, Stroke: p.commentStroke, TextColor: p.commentText},
	}

	c.SubexpFill = "none"
	c.SubexpStroke = p.subexpStroke
	c.SubexpColors = p.subexpCycle
	c.RepeatLabelColor = p.repeatLabel
	c.Connector.Color = p.connectorColor
}

func init() {
	Register(&paletteTheme{
		name:        "light",
		description: "Regolith's refreshed light palette (built-in default)",
		apply:       func(c *renderer.Config) { applyPalette(c, lightPalette) },
	})
	Register(&paletteTheme{
		name:        "dark",
		description: "Regolith's slate-900 dark palette with bright accent strokes",
		apply:       func(c *renderer.Config) { applyPalette(c, darkPalette) },
	})
}
