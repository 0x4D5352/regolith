// Package theme provides curated color palettes for the regolith
// railroad diagram renderer. A theme rewrites only the color-bearing
// fields of renderer.Config — node fills/strokes/text colors, the
// subexpression depth cycle, connector color, background, and repeat
// label color. It deliberately does NOT touch dimensions, typography,
// stroke widths, connector markers, or the analysis annotation severity
// colors: those belong to the shape/information-architecture layer of
// the diagram and stay stable across themes so readers can transfer
// the cues from one theme to another.
//
// Themes register themselves in init() and are looked up by name via
// Get. The CLI wires this in under --theme. Adding a new theme is
// purely additive: drop a new file in this package, call Register from
// init(), and extend the tests' expected-themes list.
package theme

import (
	"sort"
	"sync"

	"github.com/0x4d5352/regolith/internal/renderer"
)

// Theme is the contract every color theme implements. Name is the
// identifier the user passes to --theme. Description is surfaced in
// --help listings. Apply rewrites the color-bearing fields of cfg in
// place so a caller can: load the default config, apply a theme, then
// layer further CLI overrides on top.
type Theme interface {
	Name() string
	Description() string
	Apply(cfg *renderer.Config)
}

// registry holds the themes registered via init(). Protected with an
// RWMutex to mirror the flavor registry — parallel tests may call Get
// concurrently.
var (
	registry     = make(map[string]Theme)
	registryLock sync.RWMutex
)

// Register adds a theme to the registry. Called from init() in each
// theme file. Overwrites an existing entry with the same name; duplicate
// registrations within the same binary are a programming error but are
// not punished here because tests rebuild the registry by re-running
// init() on reload.
func Register(t Theme) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[t.Name()] = t
}

// Get returns the theme registered under name, or (nil, false) if no
// theme exists. Follows the same convention as flavor.Get so callers
// can use a consistent idiom at the CLI layer.
func Get(name string) (Theme, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	t, ok := registry[name]
	return t, ok
}

// List returns the registered theme names in sorted order. Used by
// --help, by the unknown-theme error message, and by tests that iterate
// over every theme.
func List() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// All returns every registered theme in sorted-by-name order. Used by
// tests and by golden generators that need access to Theme values
// (not just names).
func All() []Theme {
	registryLock.RLock()
	defer registryLock.RUnlock()
	themes := make([]Theme, 0, len(registry))
	for _, t := range registry {
		themes = append(themes, t)
	}
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name() < themes[j].Name()
	})
	return themes
}

// paletteTheme is the common implementation every registered theme
// uses. A theme is fully described by its name, a short description,
// and a closure that rewrites the color fields of a Config. Keeping
// them as closures lets each theme file be a single palette-constant
// block plus a thin init() — no per-theme struct boilerplate.
type paletteTheme struct {
	name        string
	description string
	apply       func(*renderer.Config)
}

func (t *paletteTheme) Name() string             { return t.name }
func (t *paletteTheme) Description() string      { return t.description }
func (t *paletteTheme) Apply(c *renderer.Config) { t.apply(c) }
