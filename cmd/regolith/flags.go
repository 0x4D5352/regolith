package main

// Shared flag plumbing for the regolith CLI.
//
// Both the main render command and the `analyze` subcommand need the same
// core flags (flavor, format, output, color, theme, padding, font size,
// line width) plus the same SVG styling knobs. Historically each command
// registered these inline, which produced drift — different defaults,
// missing flags on one side, inconsistent help text.
//
// We keep two small structs here instead of reaching for a full CLI
// framework like cobra. For N=2 subcommands and ~500 lines of glue, a
// framework is not worth the ~20 extra transitive dependencies it would
// bring. If a third subcommand lands, revisit.

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
	flag "github.com/spf13/pflag"

	"github.com/0x4d5352/regolith/internal/renderer"
	"github.com/0x4d5352/regolith/internal/renderer/theme"
)

// commonFlags captures every flag the two commands share. Each field is
// bound to the FlagSet passed to Register, so the caller can read the
// resolved values directly off the struct after fs.Parse.
type commonFlags struct {
	Flavor    string
	Format    string
	Output    string
	Color     string
	Theme     string
	Padding   float64
	FontSize  float64
	LineWidth float64
}

// commonDefaults lets each command choose slightly different defaults at
// registration time. Phase 1 of the refactor preserves the historical
// defaults (main → "svg", analyze → "text"); phase 3 flips both to "text".
type commonDefaults struct {
	Format string
	Output string
}

// Register binds every shared flag onto fs using the supplied defaults.
// All other fields (flavor, color, theme, padding, etc.) use the same
// values across both commands.
func (c *commonFlags) Register(fs *flag.FlagSet, d commonDefaults) {
	fs.StringVarP(&c.Flavor, "flavor", "f", "javascript",
		"Regex flavor (javascript, java, dotnet, pcre, posix-bre, posix-ere, gnugrep, gnugrep-bre, gnugrep-ere)")
	fs.StringVar(&c.Format, "format", d.Format, "Output format: text, json, svg")
	fs.StringVarP(&c.Output, "output", "o", d.Output, "Output file path")
	fs.StringVar(&c.Color, "color", "auto", "Color output: auto, always, never")
	fs.StringVar(&c.Theme, "theme", "", "Color theme (e.g. light, dark, catppuccin-mocha, gruvbox-dark)")
	fs.Float64VarP(&c.Padding, "padding", "p", 10, "Padding around diagram")
	fs.Float64Var(&c.FontSize, "font-size", 13, "Font size in pixels")
	fs.Float64Var(&c.LineWidth, "line-width", 1.5, "Stroke width for connectors and loops")
}

// svgStyleFlags captures every SVG-specific color/fill override. These
// used to live only on the main command, but analyze also produces SVG
// (via `--format svg`) and silently ignored them. Promoting them to a
// shared struct closes that gap — analyze now honors --literal-fill and
// friends when rendering annotated SVG.
type svgStyleFlags struct {
	TextColor      string
	LineColor      string
	LiteralFill    string
	CharsetFill    string
	EscapeFill     string
	AnchorFill     string
	SubexpFill     string
	BackgroundFill string
}

// Register binds every SVG style flag onto fs. Defaults mirror the
// hard-coded values from the original main command; they only land in
// cfg when the user explicitly sets the flag (see Apply), so a selected
// theme is never clobbered by the defaults listed here.
func (s *svgStyleFlags) Register(fs *flag.FlagSet) {
	fs.StringVar(&s.TextColor, "text-color", "#000",
		"Fallback text color for elements outside any category")
	fs.StringVar(&s.LineColor, "line-color", "#64748b",
		"Connector / loop line color")
	fs.StringVar(&s.LiteralFill, "literal-fill", "#fee2e2",
		"Literal box fill color")
	fs.StringVar(&s.CharsetFill, "charset-fill", "#f5f0e1",
		"Character set box fill color")
	fs.StringVar(&s.EscapeFill, "escape-fill", "#ecfccb",
		"Escape sequence box fill color")
	fs.StringVar(&s.AnchorFill, "anchor-fill", "#334155",
		"Anchor box fill color")
	fs.StringVar(&s.SubexpFill, "subexp-fill", "none",
		"Outermost subexpression box fill color (nested groups use cycling colors)")
	fs.StringVar(&s.BackgroundFill, "background-fill", "",
		"Solid background fill color (hex or CSS name; 'theme' uses the active theme's background; default: off)")
}

// Apply layers the SVG style overrides onto cfg. Only flags the user
// actually changed land — unchanged flags are left alone so a selected
// theme keeps its palette.
func (s *svgStyleFlags) Apply(fs *flag.FlagSet, cfg *renderer.Config) {
	if fs.Changed("line-color") {
		cfg.Connector.Color = s.LineColor
	}
	if fs.Changed("text-color") {
		cfg.TextColor = s.TextColor
	}
	if fs.Changed("literal-fill") {
		patchNodeFill(cfg, "literal", s.LiteralFill)
	}
	if fs.Changed("charset-fill") {
		patchNodeFill(cfg, "charset", s.CharsetFill)
	}
	if fs.Changed("escape-fill") {
		patchNodeFill(cfg, "escape", s.EscapeFill)
	}
	if fs.Changed("anchor-fill") {
		patchNodeFill(cfg, "anchor", s.AnchorFill)
	}
	if fs.Changed("subexp-fill") {
		cfg.SubexpFill = s.SubexpFill
	}
	if fs.Changed("background-fill") {
		// The 'theme' sentinel opts into whatever background the
		// currently selected theme already wrote to cfg.BackgroundColor.
		// applyTheme runs before svgStyleFlags.Apply in buildSVGConfig,
		// so by the time we read it the theme has already spoken.
		if s.BackgroundFill == "theme" {
			cfg.BackgroundFill = cfg.BackgroundColor
		} else {
			cfg.BackgroundFill = s.BackgroundFill
		}
	}
}

// buildSVGConfig produces a fully-configured renderer.Config from the
// shared common and style flags. The layering order matters: defaults →
// theme → explicit overrides. A theme replaces color fields wholesale;
// the --literal-fill / --line-color / etc. flags then tint specific
// categories without rebuilding the whole palette.
func buildSVGConfig(fs *flag.FlagSet, common *commonFlags, style *svgStyleFlags) (*renderer.Config, error) {
	cfg := renderer.DefaultConfig()
	if err := applyTheme(cfg, common.Theme); err != nil {
		return nil, err
	}
	cfg.Padding = common.Padding
	cfg.FontSize = common.FontSize
	cfg.CharWidth = common.FontSize * 0.6
	cfg.Connector.StrokeWidth = common.LineWidth
	style.Apply(fs, cfg)
	return cfg, nil
}

// applyTheme resolves a theme name and applies it to cfg. An empty
// string is a no-op: DefaultConfig()'s built-in palette (which matches
// the registered "light" theme byte-for-byte) is used as-is. Any
// non-empty name must resolve via the theme registry — there is no
// "default" alias, so a future config-file or env-var layer can
// promote a real theme to install-wide default without fighting a
// shim.
func applyTheme(cfg *renderer.Config, name string) error {
	if name == "" {
		return nil
	}
	t, ok := theme.Get(name)
	if !ok {
		return fmt.Errorf("unknown theme %q (available: %s)",
			name, strings.Join(theme.List(), ", "))
	}
	t.Apply(cfg)
	return nil
}

// patchNodeFill overrides just the Fill field on a single category's
// NodeStyle, leaving the stroke and text color from the underlying
// theme in place. Used by --literal-fill / --charset-fill / etc. so
// users can tint a single category without providing a whole bundle.
func patchNodeFill(cfg *renderer.Config, class, fill string) {
	s := cfg.GetNodeStyle(class)
	s.Fill = fill
	cfg.NodeStyles[class] = s
}

// requireOutputForSVG fails when the caller picked --format svg but
// didn't supply --output. SVG blobs are multi-kilobyte; dumping them
// to a terminal would be worse than a clear error.
func requireOutputForSVG(format, output string) error {
	if format == "svg" && output == "" {
		return fmt.Errorf("svg format requires --output/-o (e.g., -o diagram.svg)")
	}
	return nil
}

// writeOutputFile writes data to path and prints a colorized confirmation
// to stdout. Used by every command path that produces a file (SVG render,
// markdown from --format text -o, etc).
func writeOutputFile(path string, data []byte, stdout io.Writer, co *termenv.Output) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	_, _ = fmt.Fprintln(stdout, co.String("Wrote "+path).Foreground(termenv.ANSIColor(2)).String())
	return nil
}
