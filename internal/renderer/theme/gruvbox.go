package theme

import "github.com/0x4d5352/regolith/internal/renderer"

// Gruvbox palette.
// Source: https://github.com/morhetz/gruvbox/blob/master/colors/gruvbox.vim
// (MIT License, Pavel "morhetz" Pertsev). Hex values are verbatim from
// the vim colorscheme at medium contrast. Dark and light share the
// same accent families; the convention is bright_* accents on the
// dark background and faded_* accents on the light background, so
// each variant reads saturated against its own field.

// gruvboxPalette bundles the subset of gruvbox colors the renderer
// uses. The base/fg/accent fields are already variant-specific — we
// don't parameterize a base struct like Catppuccin because gruvbox
// conventionally uses *different* accent families per variant
// (bright_* vs faded_*), not just different surfaces.
type gruvboxPalette struct {
	bg     string // main background
	bg0    string // darker surface (panel lift on light, base on dark)
	bg1    string // mid surface
	fg     string // primary text
	fg2    string // secondary text (labels, repeat tags)
	gray   string // shared neutral — same #928374 in both variants
	red    string
	green  string
	yellow string
	blue   string
	purple string
	aqua   string
	orange string
}

var (
	// gruvbox-dark — medium contrast, the canonical retro groove look.
	// bg=dark0, fg=light1, accents from bright_*.
	gruvboxDark = gruvboxPalette{
		bg:     "#282828", // dark0
		bg0:    "#3c3836", // dark1 (slight lift above base)
		bg1:    "#504945", // dark2
		fg:     "#ebdbb2", // light1
		fg2:    "#a89984", // light4 — muted label text
		gray:   "#928374",
		red:    "#fb4934",
		green:  "#b8bb26",
		yellow: "#fabd2f",
		blue:   "#83a598",
		purple: "#d3869b",
		aqua:   "#8ec07c",
		orange: "#fe8019",
	}

	// gruvbox-light — paper-and-ink aesthetic. bg=light0, fg=dark1,
	// accents from faded_* so they read saturated on the cream field
	// rather than washed out.
	gruvboxLight = gruvboxPalette{
		bg:     "#fbf1c7", // light0
		bg0:    "#ebdbb2", // light1 (slight press-down from base)
		bg1:    "#d5c4a1", // light2
		fg:     "#3c3836", // dark1
		fg2:    "#7c6f64", // dark4 — muted label text
		gray:   "#928374",
		red:    "#9d0006", // faded_red
		green:  "#79740e", // faded_green
		yellow: "#b57614", // faded_yellow
		blue:   "#076678", // faded_blue
		purple: "#8f3f71", // faded_purple
		aqua:   "#427b58", // faded_aqua
		orange: "#af3a03", // faded_orange
	}
)

// applyGruvbox rewrites cfg's colors from a gruvbox palette. The
// category-to-color mapping matches the semantic table used by every
// theme in this package, with one deviation: gruvbox's flags and
// any-character share blue (already the case in the default theme),
// and backtrack-control shares red with literal — gruvbox's warm set
// has only one red hue to draw from.
func applyGruvbox(c *renderer.Config, p gruvboxPalette) {
	c.BackgroundColor = p.bg
	c.TextColor = p.fg

	// bg0 is a subtle lift (dark) or press-down (light) relative to
	// the main background — it gives every node panel a visible
	// boundary without relying solely on the stroke.
	c.NodeStyles = map[string]renderer.NodeStyle{
		"literal":           {Fill: p.bg0, Stroke: p.red, TextColor: p.fg},
		"charset":           {Fill: p.bg0, Stroke: p.yellow, TextColor: p.fg},
		"escape":            {Fill: p.bg0, Stroke: p.green, TextColor: p.fg},
		"anchor":            {Fill: p.bg1, Stroke: p.fg2, TextColor: p.bg, CornerRadius: 14},
		"any-character":     {Fill: p.bg0, Stroke: p.blue, TextColor: p.fg},
		"flags":             {Fill: p.bg0, Stroke: p.blue, TextColor: p.fg},
		"recursive-ref":     {Fill: p.bg0, Stroke: p.purple, TextColor: p.fg},
		"callout":           {Fill: p.bg0, Stroke: p.orange, TextColor: p.fg},
		"backtrack-control": {Fill: p.bg0, Stroke: p.red, TextColor: p.fg},
		"conditional":       {Fill: p.bg0, Stroke: p.aqua, TextColor: p.fg},
		"comment":           {Fill: p.bg0, Stroke: p.gray, TextColor: p.fg2},
	}

	c.SubexpFill = "none"
	c.SubexpStroke = p.gray
	c.SubexpColors = []string{p.blue, p.green, p.yellow, p.purple, p.orange}

	c.RepeatLabelColor = p.fg2
	c.Connector.Color = p.gray
}

func init() {
	Register(&paletteTheme{
		name:        "gruvbox-dark",
		description: "Gruvbox Dark — retro groove, warm dark background",
		apply:       func(c *renderer.Config) { applyGruvbox(c, gruvboxDark) },
	})
	Register(&paletteTheme{
		name:        "gruvbox-light",
		description: "Gruvbox Light — retro groove, cream paper background",
		apply:       func(c *renderer.Config) { applyGruvbox(c, gruvboxLight) },
	})
}
