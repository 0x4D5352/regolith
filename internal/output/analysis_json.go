package output

import (
	"encoding/json"
	"fmt"

	"github.com/0x4d5352/regolith/internal/analyzer"
)

// analysisDocument is the top-level JSON envelope for an AnalysisReport.
type analysisDocument struct {
	Pattern   string         `json:"pattern"`
	Flavor    string         `json:"flavor"`
	Findings  []findingJSON  `json:"findings"`
	Benchmark *benchmarkJSON `json:"benchmark,omitempty"`
}

// findingJSON is the stable consumer-facing representation of a single
// analysis finding. Severity is serialized as a string ("info", "warning",
// "error", "critical") so consumers don't depend on our internal iota order.
type findingJSON struct {
	ID          string       `json:"id"`
	Category    string       `json:"category"`
	Severity    string       `json:"severity"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Suggestion  string       `json:"suggestion,omitempty"`
	Runtime     *runtimeJSON `json:"runtime,omitempty"`
}

// runtimeJSON captures empirical timing data collected during optional
// benchmark execution. Durations are expressed in microseconds (integers)
// to avoid floating-point variance across platforms.
type runtimeJSON struct {
	Engine      string         `json:"engine"`
	TimedOut    bool           `json:"timedOut"`
	Durations   map[string]int `json:"durations"`
	ScalingHint string         `json:"scalingHint"`
}

// benchmarkJSON holds the full benchmark corpus summary. Each corpus entry
// maps a string-encoded input size to a duration in microseconds.
type benchmarkJSON struct {
	Engine     string                    `json:"engine"`
	IsFallback bool                      `json:"isFallback"`
	Corpus     map[string]map[string]int `json:"corpus"`
}

// RenderAnalysisJSON serializes an AnalysisReport to pretty-printed JSON.
// Each finding is translated to a stable schema that uses string-typed
// severity values rather than the internal iota, so the output format is
// not tied to the declaration order of our constants.
func RenderAnalysisJSON(report *analyzer.AnalysisReport) (string, error) {
	doc := analysisDocument{
		Pattern:  report.Pattern,
		Flavor:   report.Flavor,
		Findings: make([]findingJSON, len(report.Findings)),
	}

	for i, f := range report.Findings {
		fj := findingJSON{
			ID:          f.ID,
			Category:    string(f.Category),
			Severity:    f.Severity.String(),
			Title:       f.Title,
			Description: f.Description,
			Suggestion:  f.Suggestion,
		}
		if f.Runtime != nil {
			fj.Runtime = convertRuntimeResult(f.Runtime)
		}
		doc.Findings[i] = fj
	}

	if report.BenchmarkSummary != nil {
		doc.Benchmark = convertBenchmarkSummary(report.BenchmarkSummary)
	}

	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(b), nil
}

func convertRuntimeResult(r *analyzer.RuntimeResult) *runtimeJSON {
	durations := make(map[string]int, len(r.Durations))
	for size, dur := range r.Durations {
		durations[fmt.Sprintf("%d", size)] = int(dur.Microseconds())
	}
	return &runtimeJSON{
		Engine:      r.Engine,
		TimedOut:    r.TimedOut,
		Durations:   durations,
		ScalingHint: r.ScalingHint,
	}
}

func convertBenchmarkSummary(s *analyzer.BenchmarkSummary) *benchmarkJSON {
	corpus := make(map[string]map[string]int, len(s.Corpus))
	for name, results := range s.Corpus {
		sizes := make(map[string]int, len(results))
		for size, dur := range results {
			sizes[fmt.Sprintf("%d", size)] = int(dur.Microseconds())
		}
		corpus[name] = sizes
	}
	return &benchmarkJSON{
		Engine:     s.Engine,
		IsFallback: s.IsFallback,
		Corpus:     corpus,
	}
}
