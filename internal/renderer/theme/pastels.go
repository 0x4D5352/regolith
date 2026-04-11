package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// Pastels: a hand-tuned low-chroma palette for when the default theme
// reads too loud. The aesthetic is closer to an illustrated notebook
// than a terminal colorscheme — fills carry most of the category cue
// (light tinted panels, one hue per category), with muted strokes
// reinforcing them. Not derived from an upstream project.
//
// The light variant sits on a warm off-white; the dark variant uses a
// muted plum background so the pastels stay readable without glowing.
// Both avoid fully-saturated accents on purpose: the point of this
// theme is that nothing screams.

type pastelPalette struct {
	bg, fg, fill           string
	anchorFill, anchorFg   string
	mutedStroke, mutedText string

	// Each "slot" is a tint + stroke pair. Slots are named by the
	// category they're assigned to below — keeping the palette fields
	// tied to semantic roles makes the mapping easier to read than a
	// color-name palette would.
	literalFill, literalStroke     string
	charsetFill, charsetStroke     string
	escapeFill, escapeStroke       string
	blueFill, blueStroke           string
	recursiveFill, recursiveStroke string
	calloutFill, calloutStroke     string
}

var (
	pastelsLight = pastelPalette{
		bg:          "#faf5ef",
		fg:          "#4a4458",
		fill:        "#fdfaf3",
		anchorFill:  "#d8d0e3",
		anchorFg:    "#4a4458",
		mutedStroke: "#b8b0c4",
		mutedText:   "#6d6580",

		literalFill:     "#fde4e4",
		literalStroke:   "#d88a8a",
		charsetFill:     "#fdf2d5",
		charsetStroke:   "#c9a66b",
		escapeFill:      "#e0f0d9",
		escapeStroke:    "#8bb878",
		blueFill:        "#d9eaf7",
		blueStroke:      "#7ea8d1",
		recursiveFill:   "#e8dbf5",
		recursiveStroke: "#a890c8",
		calloutFill:     "#fce4d0",
		calloutStroke:   "#d8a478",
	}

	pastelsDark = pastelPalette{
		bg:          "#2a2438",
		fg:          "#e8e4f0",
		fill:        "#3f3549",
		anchorFill:  "#5a4f6b",
		anchorFg:    "#e8e4f0",
		mutedStroke: "#7a6d8c",
		mutedText:   "#b8acc8",

		literalFill:     "#3f3549",
		literalStroke:   "#e8a5a5",
		charsetFill:     "#3f3549",
		charsetStroke:   "#ecd6a0",
		escapeFill:      "#3f3549",
		escapeStroke:    "#b5d5a8",
		blueFill:        "#3f3549",
		blueStroke:      "#a8c9e0",
		recursiveFill:   "#3f3549",
		recursiveStroke: "#c2a8d8",
		calloutFill:     "#3f3549",
		calloutStroke:   "#e8c0a0",
	}
)

func applyPastels(c *renderer.Config, p pastelPalette) {
	c.BackgroundColor = p.bg
	c.TextColor = p.fg

	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: p.literalFill, Stroke: p.literalStroke, TextColor: p.fg},
		"charset":           {Fill: p.charsetFill, Stroke: p.charsetStroke, TextColor: p.fg},
		"escape":            {Fill: p.escapeFill, Stroke: p.escapeStroke, TextColor: p.fg},
		"anchor":            {Fill: p.anchorFill, Stroke: p.mutedStroke, TextColor: p.anchorFg, CornerRadius: 14},
		"any-character":     {Fill: p.blueFill, Stroke: p.blueStroke, TextColor: p.fg},
		"flags":             {Fill: p.blueFill, Stroke: p.blueStroke, TextColor: p.fg},
		"recursive-ref":     {Fill: p.recursiveFill, Stroke: p.recursiveStroke, TextColor: p.fg},
		"callout":           {Fill: p.calloutFill, Stroke: p.calloutStroke, TextColor: p.fg},
		"backtrack-control": {Fill: p.literalFill, Stroke: p.literalStroke, TextColor: p.fg},
		"conditional":       {Fill: p.escapeFill, Stroke: p.escapeStroke, TextColor: p.fg},
		"comment":           {Fill: p.fill, Stroke: p.mutedStroke, TextColor: p.mutedText},
	}

	c.SubexpFill = "none"
	c.SubexpStroke = p.mutedStroke
	c.SubexpColors = []string{
		p.blueStroke,
		p.escapeStroke,
		p.charsetStroke,
		p.recursiveStroke,
		p.calloutStroke,
	}

	c.RepeatLabelColor = p.mutedText
	c.Connector.Color = p.mutedStroke
}

func init() {
	Register(&paletteTheme{
		name:        "pastels-light",
		description: "Soft pastels on a warm off-white background",
		apply:       func(c *renderer.Config) { applyPastels(c, pastelsLight) },
	})
	Register(&paletteTheme{
		name:        "pastels-dark",
		description: "Soft pastels on a muted plum background",
		apply:       func(c *renderer.Config) { applyPastels(c, pastelsDark) },
	})
}
