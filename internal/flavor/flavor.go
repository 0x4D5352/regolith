// Package flavor defines the interface for regex flavor implementations
// and provides a registry for discovering available flavors.
package flavor

import (
	"sort"
	"sync"

	"github.com/0x4d5352/regolith/internal/ast"
)

// Flavor defines the interface for a regex flavor implementation.
// Each flavor provides parsing for a specific regex syntax variant.
type Flavor interface {
	// Name returns the flavor identifier (e.g., "javascript", "pcre")
	// This is used for CLI flag values and should be lowercase.
	Name() string

	// Description returns a human-readable description of the flavor.
	Description() string

	// Parse parses a regex pattern and returns an AST.
	// The pattern may include delimiters and flags (e.g., /pattern/flags)
	// depending on the flavor's syntax.
	Parse(pattern string) (*ast.Regexp, error)

	// SupportedFlags returns information about valid flags for this flavor.
	SupportedFlags() []FlagInfo

	// SupportedFeatures returns the feature capabilities of this flavor.
	SupportedFeatures() FeatureSet
}

// FlagInfo describes a regex flag.
type FlagInfo struct {
	Char        rune   // The flag character (e.g., 'i')
	Name        string // Human-readable name (e.g., "case-insensitive")
	Description string // Longer description of what the flag does
}

// FeatureSet describes what features a flavor supports.
// This can be used for documentation and for validating patterns.
type FeatureSet struct {
	Lookahead             bool // Supports (?=...) and (?!...)
	Lookbehind            bool // Supports (?<=...) and (?<!...)
	LookbehindUnlimited   bool // Lookbehind can have variable length (.NET only)
	NamedGroups           bool // Supports (?<name>...) or (?P<name>...)
	AtomicGroups          bool // Supports (?>...)
	PossessiveQuantifiers bool // Supports *+, ++, ?+, {n,m}+
	RecursivePatterns     bool // Supports (?R), (?1), (?&name)
	ConditionalPatterns   bool // Supports (?(cond)yes|no)
	UnicodeProperties     bool // Supports \p{...} and \P{...}
	POSIXClasses          bool // Supports [:alpha:], [:digit:], etc.
	BalancedGroups        bool // Supports (?<name-other>...) (.NET only)
	InlineModifiers       bool // Supports (?i), (?m), etc.
	Comments              bool // Supports (?#...) comments
	BranchReset           bool // Supports (?|...)
	BacktrackingControl   bool // Supports (*PRUNE), (*SKIP), etc.
}

// registry holds all registered flavors.
// It's protected by a mutex for safe concurrent access.
var (
	registry     = make(map[string]Flavor)
	registryLock sync.RWMutex
)

// Register adds a flavor to the registry.
// If a flavor with the same name is already registered, it is replaced.
// This function is typically called from init() functions in flavor packages.
func Register(f Flavor) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[f.Name()] = f
}

// Get retrieves a flavor by name.
// Returns nil, false if the flavor is not registered.
func Get(name string) (Flavor, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	f, ok := registry[name]
	return f, ok
}

// List returns all registered flavor names in sorted order.
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

// All returns all registered flavors as a map.
// The returned map is a copy, so modifications won't affect the registry.
func All() map[string]Flavor {
	registryLock.RLock()
	defer registryLock.RUnlock()
	result := make(map[string]Flavor, len(registry))
	for name, f := range registry {
		result[name] = f
	}
	return result
}

// Count returns the number of registered flavors.
func Count() int {
	registryLock.RLock()
	defer registryLock.RUnlock()
	return len(registry)
}
