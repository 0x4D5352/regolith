package theme

import (
	"regexp"
	"sort"
	"testing"

	"github.com/0x4d5352/regolith/internal/renderer"
)

// expectedThemes is the canonical list of themes this package ships.
// Keep it sorted alphabetically so it matches List() output directly.
// When adding a new theme, append it here — a divergence between this
// list and the registry is what TestAllThemesRegistered catches.
var expectedThemes = []string{
	"catppuccin-frappe",
	"catppuccin-latte",
	"catppuccin-macchiato",
	"catppuccin-mocha",
	"colorblind-dark",
	"colorblind-light",
	"gruvbox-dark",
	"gruvbox-light",
	"high-contrast-dark",
	"high-contrast-light",
	"pastels-dark",
	"pastels-light",
}

// expectedNodeCategories is every category the renderer currently
// knows about. Every theme must populate all of them — a missing key
// would silently fall back to a neutral gray via GetNodeStyle, which
// defeats the purpose of a curated palette and is almost always a
// bug. Keep this in sync with DefaultConfig's NodeStyles keys.
var expectedNodeCategories = []string{
	"literal",
	"charset",
	"escape",
	"anchor",
	"any-character",
	"flags",
	"recursive-ref",
	"callout",
	"backtrack-control",
	"conditional",
	"comment",
}

// hexColorRe matches a 6-digit hex color, optionally with a leading #.
// Themes may also use the literal "none" (for SubexpFill) so the
// validity check accepts that specifically.
var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func isColorValid(v string) bool {
	return v == "none" || hexColorRe.MatchString(v)
}

// TestAllThemesRegistered verifies every expected theme is registered
// and that the registry does not contain any extras. If this test
// fails, the expectedThemes list is out of sync with the init()
// registrations — either add the new theme here or remove the stale
// registration.
func TestAllThemesRegistered(t *testing.T) {
	got := List()
	sort.Strings(got) // List is already sorted; belt-and-braces.

	if len(got) != len(expectedThemes) {
		t.Fatalf("theme count mismatch: got %d (%v), want %d (%v)",
			len(got), got, len(expectedThemes), expectedThemes)
	}
	for i, name := range expectedThemes {
		if got[i] != name {
			t.Errorf("theme[%d]: got %q, want %q", i, got[i], name)
		}
	}
}

// TestGetResolvesEveryTheme makes sure every theme we claim to ship
// can actually be fetched by name. A Register() without a matching
// Get() would only manifest as a weird lookup failure at CLI time, so
// catching it at the unit-test level is cheap insurance.
func TestGetResolvesEveryTheme(t *testing.T) {
	for _, name := range expectedThemes {
		t.Run(name, func(t *testing.T) {
			th, ok := Get(name)
			if !ok {
				t.Fatalf("Get(%q): not registered", name)
			}
			if th.Name() != name {
				t.Errorf("Name(): got %q, want %q", th.Name(), name)
			}
			if th.Description() == "" {
				t.Errorf("Description(): empty")
			}
		})
	}
}

// TestApplyPreservesNonColorFields asserts the core theming
// invariant: Apply may only touch the color-bearing fields. Anything
// that belongs to the shape/typography/information-architecture layer
// must survive untouched so the visual language stays stable across
// themes. When this test fails, a theme is reaching further than it
// should — fix the theme, not the test.
func TestApplyPreservesNonColorFields(t *testing.T) {
	for _, name := range expectedThemes {
		t.Run(name, func(t *testing.T) {
			th, _ := Get(name)
			base := renderer.DefaultConfig()
			cfg := renderer.DefaultConfig()

			th.Apply(cfg)

			// Dimensions
			if cfg.Padding != base.Padding {
				t.Errorf("Padding changed: %v -> %v", base.Padding, cfg.Padding)
			}
			if cfg.HorizontalGap != base.HorizontalGap {
				t.Errorf("HorizontalGap changed: %v -> %v", base.HorizontalGap, cfg.HorizontalGap)
			}
			if cfg.VerticalGap != base.VerticalGap {
				t.Errorf("VerticalGap changed: %v -> %v", base.VerticalGap, cfg.VerticalGap)
			}
			if cfg.CornerRadius != base.CornerRadius {
				t.Errorf("CornerRadius changed: %v -> %v", base.CornerRadius, cfg.CornerRadius)
			}

			// Typography
			if cfg.FontFamily != base.FontFamily {
				t.Errorf("FontFamily changed: %q -> %q", base.FontFamily, cfg.FontFamily)
			}
			if cfg.FontSize != base.FontSize {
				t.Errorf("FontSize changed: %v -> %v", base.FontSize, cfg.FontSize)
			}
			if cfg.CharWidth != base.CharWidth {
				t.Errorf("CharWidth changed: %v -> %v", base.CharWidth, cfg.CharWidth)
			}
			if cfg.LabelFontFamily != base.LabelFontFamily {
				t.Errorf("LabelFontFamily changed: %q -> %q", base.LabelFontFamily, cfg.LabelFontFamily)
			}
			if cfg.LabelFontSize != base.LabelFontSize {
				t.Errorf("LabelFontSize changed: %v -> %v", base.LabelFontSize, cfg.LabelFontSize)
			}
			if cfg.LabelCharWidth != base.LabelCharWidth {
				t.Errorf("LabelCharWidth changed: %v -> %v", base.LabelCharWidth, cfg.LabelCharWidth)
			}

			// Stroke / shape
			if cfg.NodeStrokeWidth != base.NodeStrokeWidth {
				t.Errorf("NodeStrokeWidth changed: %v -> %v", base.NodeStrokeWidth, cfg.NodeStrokeWidth)
			}
			if cfg.Connector.StrokeWidth != base.Connector.StrokeWidth {
				t.Errorf("Connector.StrokeWidth changed: %v -> %v", base.Connector.StrokeWidth, cfg.Connector.StrokeWidth)
			}
			if cfg.Connector.StartMarker != base.Connector.StartMarker {
				t.Errorf("Connector.StartMarker changed: %q -> %q", base.Connector.StartMarker, cfg.Connector.StartMarker)
			}
			if cfg.Connector.EndMarker != base.Connector.EndMarker {
				t.Errorf("Connector.EndMarker changed: %q -> %q", base.Connector.EndMarker, cfg.Connector.EndMarker)
			}

			// Analysis annotation colors. Severity has its own stable
			// color language and themes do not override it.
			if cfg.ErrorBorderColor != base.ErrorBorderColor {
				t.Errorf("ErrorBorderColor changed: %q -> %q", base.ErrorBorderColor, cfg.ErrorBorderColor)
			}
			if cfg.WarningBorderColor != base.WarningBorderColor {
				t.Errorf("WarningBorderColor changed: %q -> %q", base.WarningBorderColor, cfg.WarningBorderColor)
			}
			if cfg.InfoBorderColor != base.InfoBorderColor {
				t.Errorf("InfoBorderColor changed: %q -> %q", base.InfoBorderColor, cfg.InfoBorderColor)
			}
			if cfg.ErrorBadgeColor != base.ErrorBadgeColor {
				t.Errorf("ErrorBadgeColor changed: %q -> %q", base.ErrorBadgeColor, cfg.ErrorBadgeColor)
			}
			if cfg.WarningBadgeColor != base.WarningBadgeColor {
				t.Errorf("WarningBadgeColor changed: %q -> %q", base.WarningBadgeColor, cfg.WarningBadgeColor)
			}
			if cfg.InfoBadgeColor != base.InfoBadgeColor {
				t.Errorf("InfoBadgeColor changed: %q -> %q", base.InfoBadgeColor, cfg.InfoBadgeColor)
			}
		})
	}
}

// TestApplyPopulatesEveryCategory guards against typos and drift in
// the NodeStyles map. Every theme must provide a non-empty entry for
// every category the renderer knows how to draw.
func TestApplyPopulatesEveryCategory(t *testing.T) {
	for _, name := range expectedThemes {
		t.Run(name, func(t *testing.T) {
			th, _ := Get(name)
			cfg := renderer.DefaultConfig()
			th.Apply(cfg)

			for _, cat := range expectedNodeCategories {
				style, ok := cfg.NodeStyles[cat]
				if !ok {
					t.Errorf("missing category %q", cat)
					continue
				}
				if style.Fill == "" || style.Stroke == "" || style.TextColor == "" {
					t.Errorf("category %q has empty field(s): %+v", cat, style)
				}
			}
		})
	}
}

// TestApplyPopulatesSubexpCycle checks that the depth-cycling color
// list has enough entries to match the default (5) and that every
// entry is a valid color. Short lists would cause nested groups past
// the cycle length to reuse the outermost color.
func TestApplyPopulatesSubexpCycle(t *testing.T) {
	for _, name := range expectedThemes {
		t.Run(name, func(t *testing.T) {
			th, _ := Get(name)
			cfg := renderer.DefaultConfig()
			th.Apply(cfg)

			if len(cfg.SubexpColors) < 5 {
				t.Errorf("SubexpColors has %d entries; want at least 5", len(cfg.SubexpColors))
			}
			for i, v := range cfg.SubexpColors {
				if !isColorValid(v) {
					t.Errorf("SubexpColors[%d] = %q: invalid hex color", i, v)
				}
			}
		})
	}
}

// TestApplyUsesValidColors catches typos in palette literals. The
// renderer will happily produce an SVG with bogus color values, but
// browsers will silently fall back to black — a very subtle failure
// mode that this test makes loud.
func TestApplyUsesValidColors(t *testing.T) {
	for _, name := range expectedThemes {
		t.Run(name, func(t *testing.T) {
			th, _ := Get(name)
			cfg := renderer.DefaultConfig()
			th.Apply(cfg)

			check := func(field, v string) {
				if !isColorValid(v) {
					t.Errorf("%s = %q: invalid color", field, v)
				}
			}

			check("BackgroundColor", cfg.BackgroundColor)
			check("TextColor", cfg.TextColor)
			check("SubexpFill", cfg.SubexpFill)
			check("SubexpStroke", cfg.SubexpStroke)
			check("RepeatLabelColor", cfg.RepeatLabelColor)
			check("Connector.Color", cfg.Connector.Color)

			for cat, s := range cfg.NodeStyles {
				check("NodeStyles["+cat+"].Fill", s.Fill)
				check("NodeStyles["+cat+"].Stroke", s.Stroke)
				check("NodeStyles["+cat+"].TextColor", s.TextColor)
			}
		})
	}
}

// TestRegistryRoundTrip exercises Register / Get / List on a freshly
// isolated registry to confirm the core primitives work without
// relying on the init()-populated themes. Mirrors flavor_test.go.
func TestRegistryRoundTrip(t *testing.T) {
	registryLock.Lock()
	saved := registry
	registry = make(map[string]Theme)
	registryLock.Unlock()
	defer func() {
		registryLock.Lock()
		registry = saved
		registryLock.Unlock()
	}()

	if got := List(); len(got) != 0 {
		t.Errorf("empty registry: List() = %v", got)
	}

	mock := &paletteTheme{
		name:        "mock",
		description: "test",
		apply:       func(*renderer.Config) {},
	}
	Register(mock)

	th, ok := Get("mock")
	if !ok {
		t.Fatalf("Get(mock): not registered")
	}
	if th.Name() != "mock" {
		t.Errorf("Name(): got %q, want %q", th.Name(), "mock")
	}

	if _, ok := Get("nonexistent"); ok {
		t.Errorf("Get(nonexistent): want false, got true")
	}

	if got := List(); len(got) != 1 || got[0] != "mock" {
		t.Errorf("List(): got %v, want [mock]", got)
	}
}
