// Package analyzer performs static analysis and optional runtime benchmarking
// of regular expressions to identify performance hazards, redundancies,
// correctness issues, and optimization opportunities.
package analyzer

import (
	"time"

	"github.com/0x4d5352/regolith/internal/ast"
)

// Severity indicates how critical a finding is, from informational hints
// to confirmed catastrophic performance issues.
type Severity int

const (
	SeverityInfo     Severity = iota // Style suggestions, minor redundancies
	SeverityWarning                  // Potential issues worth reviewing
	SeverityError                    // Likely performance or correctness problems
	SeverityCritical                 // Confirmed catastrophic behavior (e.g., exponential runtime)
)

// String returns the human-readable name for a severity level.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Category groups findings by the kind of issue they represent.
type Category string

const (
	CategoryBacktracking Category = "backtracking"
	CategoryRedundancy   Category = "redundancy"
	CategoryPerformance  Category = "performance"
	CategoryCorrectness  Category = "correctness"
)

// Finding represents a single issue detected by static analysis or
// confirmed by runtime benchmarking. Each finding is tied to the
// specific AST node that triggered it.
type Finding struct {
	ID          string         // Rule identifier, e.g. "adjacent-unbounded"
	Category    Category       // Which family this finding belongs to
	Severity    Severity       // How critical the issue is
	Title       string         // Short label for display (SVG badge, text header)
	Description string         // Detailed explanation of the problem
	Suggestion  string         // Optional recommended fix
	Node        ast.Node       // The AST node this finding applies to
	Runtime     *RuntimeResult // Non-nil when runtime benchmarking confirmed/denied
}

// RuntimeResult holds the outcome of running the pattern against test
// inputs at escalating sizes for a single finding.
type RuntimeResult struct {
	Engine      string                // Which engine was used, e.g. "node", "regexp2"
	TimedOut    bool                  // Whether any input size hit the timeout
	Durations   map[int]time.Duration // Input size (chars) -> measured duration
	ScalingHint string                // Derived: "linear", "quadratic", "exponential"
}

// BenchmarkSummary holds the full benchmark results across all corpus
// types. Per-finding RuntimeResult is derived from the worst-performing
// corpus for the pattern region that finding covers.
type BenchmarkSummary struct {
	Engine     string                           // Engine name
	IsFallback bool                             // True if using regexp2 fallback
	Corpus     map[string]map[int]time.Duration // corpus name -> size -> duration
}

// AnalysisReport is the top-level result of analyzing a regex pattern.
type AnalysisReport struct {
	Pattern          string            // Original pattern string
	Flavor           string            // Flavor name used for parsing
	Findings         []*Finding        // All findings, ordered by position in pattern
	BenchmarkSummary *BenchmarkSummary // Nil when --benchmark was not passed
}
