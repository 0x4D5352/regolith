package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// Catppuccin palette.
// Source: https://raw.githubusercontent.com/catppuccin/palette/main/palette.json
// (MIT License, Catppuccin Organization). Hex values are verbatim from
// palette.json v1.8.0. Four flavors: Latte (light) plus Frappé,
// Macchiato, and Mocha (dark, progressively deeper contrast).

// catppuccinPalette is the subset of Catppuccin colors this renderer
// uses. All four flavors have the same set of named colors; only the
// hex values differ. Keeping them in a struct lets applyCatppuccin
// handle all four flavors with one shared mapping function.
type catppuccinPalette struct {
	// Base surfaces — background, node fills, panel lifts.
	base     string
	mantle   string
	crust    string
	surface0 string
	surface1 string
	surface2 string

	// Overlays — used for muted roles (connectors, comments, subexp
	// outer stroke) that should read quieter than the accent colors.
	overlay0 string
	overlay1 string
	overlay2 string

	// Text tiers — body text, subtle text, subtler text.
	text     string
	subtext0 string
	subtext1 string

	// Accent colors — these carry the category cue in the node borders
	// and text. Catppuccin's accent set is luminance-balanced so the
	// same mapping works in both light and dark variants.
	red      string
	maroon   string
	peach    string
	yellow   string
	green    string
	teal     string
	sky      string
	sapphire string
	blue     string
	lavender string
	mauve    string
	pink     string
}

var (
	catppuccinLatte = catppuccinPalette{
		base: "#eff1f5", mantle: "#e6e9ef", crust: "#dce0e8",
		surface0: "#ccd0da", surface1: "#bcc0cc", surface2: "#acb0be",
		overlay0: "#9ca0b0", overlay1: "#8c8fa1", overlay2: "#7c7f93",
		text: "#4c4f69", subtext0: "#6c6f85", subtext1: "#5c5f77",
		red: "#d20f39", maroon: "#e64553", peach: "#fe640b", yellow: "#df8e1d",
		green: "#40a02b", teal: "#179299", sky: "#04a5e5", sapphire: "#209fb5",
		blue: "#1e66f5", lavender: "#7287fd", mauve: "#8839ef", pink: "#ea76cb",
	}

	catppuccinFrappe = catppuccinPalette{
		base: "#303446", mantle: "#292c3c", crust: "#232634",
		surface0: "#414559", surface1: "#51576d", surface2: "#626880",
		overlay0: "#737994", overlay1: "#838ba7", overlay2: "#949cbb",
		text: "#c6d0f5", subtext0: "#a5adce", subtext1: "#b5bfe2",
		red: "#e78284", maroon: "#ea999c", peach: "#ef9f76", yellow: "#e5c890",
		green: "#a6d189", teal: "#81c8be", sky: "#99d1db", sapphire: "#85c1dc",
		blue: "#8caaee", lavender: "#babbf1", mauve: "#ca9ee6", pink: "#f4b8e4",
	}

	catppuccinMacchiato = catppuccinPalette{
		base: "#24273a", mantle: "#1e2030", crust: "#181926",
		surface0: "#363a4f", surface1: "#494d64", surface2: "#5b6078",
		overlay0: "#6e738d", overlay1: "#8087a2", overlay2: "#939ab7",
		text: "#cad3f5", subtext0: "#a5adcb", subtext1: "#b8c0e0",
		red: "#ed8796", maroon: "#ee99a0", peach: "#f5a97f", yellow: "#eed49f",
		green: "#a6da95", teal: "#8bd5ca", sky: "#91d7e3", sapphire: "#7dc4e4",
		blue: "#8aadf4", lavender: "#b7bdf8", mauve: "#c6a0f6", pink: "#f5bde6",
	}

	catppuccinMocha = catppuccinPalette{
		base: "#1e1e2e", mantle: "#181825", crust: "#11111b",
		surface0: "#313244", surface1: "#45475a", surface2: "#585b70",
		overlay0: "#6c7086", overlay1: "#7f849c", overlay2: "#9399b2",
		text: "#cdd6f4", subtext0: "#a6adc8", subtext1: "#bac2de",
		red: "#f38ba8", maroon: "#eba0ac", peach: "#fab387", yellow: "#f9e2af",
		green: "#a6e3a1", teal: "#94e2d5", sky: "#89dceb", sapphire: "#74c7ec",
		blue: "#89b4fa", lavender: "#b4befe", mauve: "#cba6f7", pink: "#f5c2e7",
	}
)

// applyCatppuccin rewrites cfg's color fields from a Catppuccin palette.
// All four flavors share this mapping — the only difference is the
// palette struct passed in. The mapping assigns each node category a
// stable semantic slot (literal→red, charset→yellow, ...) so readers
// who learn the cues in one flavor transfer them to the others.
func applyCatppuccin(c *renderer.Config, p catppuccinPalette) {
	// Background / fallback text.
	c.BackgroundColor = p.base
	c.TextColor = p.text

	// Node palette. Fill is surface0 everywhere — a lifted panel that
	// reads clearly against base in both light and dark variants.
	// Stroke carries the category cue; TextColor uses the neutral
	// `text` tier for reliable legibility on the panel fill, matching
	// Catppuccin's own usage in their editor themes.
	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: p.surface0, Stroke: p.red, TextColor: p.text},
		"charset":           {Fill: p.surface0, Stroke: p.yellow, TextColor: p.text},
		"escape":            {Fill: p.surface0, Stroke: p.green, TextColor: p.text},
		"anchor":            {Fill: p.overlay1, Stroke: p.overlay2, TextColor: p.base, CornerRadius: 14},
		"any-character":     {Fill: p.surface0, Stroke: p.blue, TextColor: p.text},
		"flags":             {Fill: p.surface0, Stroke: p.sapphire, TextColor: p.text},
		"recursive-ref":     {Fill: p.surface0, Stroke: p.mauve, TextColor: p.text},
		"callout":           {Fill: p.surface0, Stroke: p.peach, TextColor: p.text},
		"backtrack-control": {Fill: p.surface0, Stroke: p.maroon, TextColor: p.text},
		"conditional":       {Fill: p.surface0, Stroke: p.sky, TextColor: p.text},
		"comment":           {Fill: p.surface0, Stroke: p.overlay0, TextColor: p.subtext0},
	}

	// Subexpressions: transparent outer box (so contents dominate) with
	// a muted overlay stroke. Depth cycling walks the theme's rainbow
	// (blue → green → yellow → pink → lavender) so nesting reads as a
	// color gradient regardless of palette.
	c.SubexpFill = "none"
	c.SubexpStroke = p.overlay0
	c.SubexpColors = []string{p.blue, p.green, p.yellow, p.pink, p.lavender}

	// Repeat labels and connectors sit below the accent colors in the
	// visual hierarchy — subtext0 / overlay1 keep them present but
	// quiet.
	c.RepeatLabelColor = p.subtext0
	c.Connector.Color = p.overlay1
}

func init() {
	Register(&paletteTheme{
		name:        "catppuccin-latte",
		description: "Catppuccin Latte — the light flavor",
		apply:       func(c *renderer.Config) { applyCatppuccin(c, catppuccinLatte) },
	})
	Register(&paletteTheme{
		name:        "catppuccin-frappe",
		description: "Catppuccin Frappé — muted dark",
		apply:       func(c *renderer.Config) { applyCatppuccin(c, catppuccinFrappe) },
	})
	Register(&paletteTheme{
		name:        "catppuccin-macchiato",
		description: "Catppuccin Macchiato — medium dark",
		apply:       func(c *renderer.Config) { applyCatppuccin(c, catppuccinMacchiato) },
	})
	Register(&paletteTheme{
		name:        "catppuccin-mocha",
		description: "Catppuccin Mocha — deep dark, warm",
		apply:       func(c *renderer.Config) { applyCatppuccin(c, catppuccinMocha) },
	})
}
