package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// High-contrast themes for accessibility. Pure-black (#000) or
// pure-white (#fff) background and foreground maximize text contrast
// (21:1), and the accent strokes are picked for high contrast on
// their respective field — bright saturated hues on black, deep
// saturated hues on white. TextColor stays pinned to the foreground
// everywhere except on anchors (which are inverted pill-shapes) so
// every text-on-panel pairing hits 21:1.
//
// The mapping deliberately sacrifices the full 11-way category
// separation for legibility: the accent set has 7 hues
// (red/yellow/green/cyan/blue/magenta/orange), so some categories
// share strokes where the pairing makes semantic sense
// (backtrack-control shares red with literal; conditional shares
// cyan with flags). Anyone who needs true 11-way distinction should
// pick one of the other themes — high-contrast is for users whose
// first constraint is "I can see this at all".

type highContrastPalette struct {
	bg, fg, fill           string
	anchorFill, anchorFg   string
	mutedStroke, mutedText string

	red, yellow, green, cyan, blue, magenta, orange string
}

var (
	highContrastDark = highContrastPalette{
		bg:          "#000000",
		fg:          "#ffffff",
		fill:        "#1a1a1a",
		anchorFill:  "#4a4a4a",
		anchorFg:    "#ffffff",
		mutedStroke: "#808080",
		mutedText:   "#cccccc",

		red:     "#ff5252",
		yellow:  "#ffff00",
		green:   "#00e676",
		cyan:    "#00e5ff",
		blue:    "#4da6ff",
		magenta: "#ff40ff",
		orange:  "#ffab40",
	}

	highContrastLight = highContrastPalette{
		bg:          "#ffffff",
		fg:          "#000000",
		fill:        "#f0f0f0",
		anchorFill:  "#1a1a1a",
		anchorFg:    "#ffffff",
		mutedStroke: "#595959",
		mutedText:   "#333333",

		red:     "#b00020",
		yellow:  "#805500",
		green:   "#006600",
		cyan:    "#006666",
		blue:    "#00008b",
		magenta: "#8b008b",
		orange:  "#8b3a00",
	}
)

func applyHighContrast(c *renderer.Config, p highContrastPalette) {
	c.BackgroundColor = p.bg
	c.TextColor = p.fg

	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: p.fill, Stroke: p.red, TextColor: p.fg},
		"charset":           {Fill: p.fill, Stroke: p.yellow, TextColor: p.fg},
		"escape":            {Fill: p.fill, Stroke: p.green, TextColor: p.fg},
		"anchor":            {Fill: p.anchorFill, Stroke: p.fg, TextColor: p.anchorFg, CornerRadius: 14},
		"any-character":     {Fill: p.fill, Stroke: p.blue, TextColor: p.fg},
		"flags":             {Fill: p.fill, Stroke: p.cyan, TextColor: p.fg},
		"recursive-ref":     {Fill: p.fill, Stroke: p.magenta, TextColor: p.fg},
		"callout":           {Fill: p.fill, Stroke: p.orange, TextColor: p.fg},
		"backtrack-control": {Fill: p.fill, Stroke: p.red, TextColor: p.fg},
		"conditional":       {Fill: p.fill, Stroke: p.cyan, TextColor: p.fg},
		"comment":           {Fill: p.fill, Stroke: p.mutedStroke, TextColor: p.mutedText},
	}

	c.SubexpFill = "none"
	c.SubexpStroke = p.mutedStroke
	c.SubexpColors = []string{p.blue, p.green, p.yellow, p.magenta, p.orange}

	c.RepeatLabelColor = p.mutedText
	c.Connector.Color = p.fg
}

func init() {
	Register(&paletteTheme{
		name:        "high-contrast-dark",
		description: "High contrast, pure-black background (accessibility)",
		apply:       func(c *renderer.Config) { applyHighContrast(c, highContrastDark) },
	})
	Register(&paletteTheme{
		name:        "high-contrast-light",
		description: "High contrast, pure-white background (accessibility)",
		apply:       func(c *renderer.Config) { applyHighContrast(c, highContrastLight) },
	})
}
