package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// Colorblind-safe palette.
// Source: Wong, B. (2011), "Color blindness," Nature Methods 8, 441.
// https://www.nature.com/articles/nmeth.1618
//
// Wong's 8-color palette is the de-facto standard for colorblind-safe
// data visualization: the colors are chosen so that all pairwise
// combinations remain distinguishable under simulated protanopia,
// deuteranopia, and tritanopia. We use it as a universal palette
// (one theme that works for every form of colorblindness) rather
// than shipping per-type variants, since Wong's values already cover
// all three cases.
const (
	wongOrange     = "#E69F00"
	wongSkyBlue    = "#56B4E9"
	wongGreen      = "#009E73" // "bluish green"
	wongYellow     = "#F0E442"
	wongBlue       = "#0072B2"
	wongVermillion = "#D55E00"
	wongPurple     = "#CC79A7" // "reddish purple"
)

// applyColorblind assigns the Wong accents to semantic slots and
// swaps in a variant-specific background/text pair. The accent
// mapping is identical for dark and light — Wong's colors are
// luminance-tuned to work on either field — but the node fill uses
// the *opposite-lift* trick (slightly brighter than bg on dark;
// slightly dimmer than bg on light) so every panel has a visible
// outline even before the stroke is drawn. TextColor sticks to the
// neutral foreground to avoid low-contrast issues (yellow text on
// white would be unreadable), leaving the stroke as the sole
// category cue.
func applyColorblind(c *renderer.Config, bg, fg, fill, anchorFill, anchorText, mutedText, mutedStroke string) {
	c.BackgroundColor = bg
	c.TextColor = fg

	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: fill, Stroke: wongVermillion, TextColor: fg},
		"charset":           {Fill: fill, Stroke: wongYellow, TextColor: fg},
		"escape":            {Fill: fill, Stroke: wongGreen, TextColor: fg},
		"anchor":            {Fill: anchorFill, Stroke: mutedStroke, TextColor: anchorText, CornerRadius: 14},
		"any-character":     {Fill: fill, Stroke: wongBlue, TextColor: fg},
		"flags":             {Fill: fill, Stroke: wongSkyBlue, TextColor: fg},
		"recursive-ref":     {Fill: fill, Stroke: wongPurple, TextColor: fg},
		"callout":           {Fill: fill, Stroke: wongOrange, TextColor: fg},
		"backtrack-control": {Fill: fill, Stroke: wongVermillion, TextColor: fg},
		"conditional":       {Fill: fill, Stroke: wongSkyBlue, TextColor: fg},
		"comment":           {Fill: fill, Stroke: mutedStroke, TextColor: mutedText},
	}

	c.SubexpFill = "none"
	c.SubexpStroke = mutedStroke
	c.SubexpColors = []string{wongBlue, wongGreen, wongYellow, wongOrange, wongPurple}

	c.RepeatLabelColor = mutedText
	c.Connector.Color = mutedStroke
}

func init() {
	Register(&paletteTheme{
		name:        "colorblind-dark",
		description: "Colorblind-safe (Wong 2011) on a dark background",
		apply: func(c *renderer.Config) {
			applyColorblind(c,
				/* bg */ "#1a1a1a",
				/* fg */ "#f0f0f0",
				/* fill */ "#262626",
				/* anchorFill */ "#4a4a4a",
				/* anchorText */ "#f0f0f0",
				/* mutedText */ "#a0a0a0",
				/* mutedStroke */ "#666666",
			)
		},
	})
	Register(&paletteTheme{
		name:        "colorblind-light",
		description: "Colorblind-safe (Wong 2011) on a light background",
		apply: func(c *renderer.Config) {
			applyColorblind(c,
				/* bg */ "#ffffff",
				/* fg */ "#000000",
				/* fill */ "#f0f0f0",
				/* anchorFill */ "#333333",
				/* anchorText */ "#ffffff",
				/* mutedText */ "#555555",
				/* mutedStroke */ "#888888",
			)
		},
	})
}
