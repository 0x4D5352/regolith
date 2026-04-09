package analyzer

import (
	"math"
	"sort"
	"time"
)

// BenchmarkOptions configures a benchmark run.
type BenchmarkOptions struct {
	Pattern string
	Timeout time.Duration
	Corpus  []string // Corpus type names
	Sizes   []int    // Input sizes to test
}

// RunBenchmark executes the pattern against each corpus type at
// escalating sizes. If a smaller size times out, larger sizes for
// that corpus are skipped. Timed-out entries are stored as negative
// durations (the negation of the timeout value).
func RunBenchmark(engine Engine, opts *BenchmarkOptions) *BenchmarkSummary {
	summary := &BenchmarkSummary{
		Engine: engine.Name(),
		Corpus: make(map[string]map[int]time.Duration),
	}

	// Sort sizes ascending so we can break early on timeout without
	// skipping smaller inputs that might still complete.
	sizes := make([]int, len(opts.Sizes))
	copy(sizes, opts.Sizes)
	sort.Ints(sizes)

	for _, corpusType := range opts.Corpus {
		results := make(map[int]time.Duration)
		summary.Corpus[corpusType] = results

		for _, size := range sizes {
			input := GenerateCorpus(corpusType, size)
			// Append a non-matching character to force the engine to
			// exhaust all paths for backtracking patterns.
			testInput := input + "!"

			dur, err := engine.Run(opts.Pattern, testInput, opts.Timeout)
			if err != nil {
				// Record a negative sentinel so callers can detect which
				// sizes timed out without losing the timeout value itself.
				results[size] = -opts.Timeout
				break
			}
			results[size] = dur
		}
	}

	return summary
}

// ClassifyScaling determines the scaling behavior from a set of
// size -> duration measurements.
// Returns "linear", "quadratic", "exponential", or "unknown".
func ClassifyScaling(durations map[int]time.Duration) string {
	if len(durations) < 2 {
		return "unknown"
	}

	type measurement struct {
		size int
		dur  time.Duration
	}

	// Collect only positive (non-timed-out) measurements for ratio analysis.
	var points []measurement
	for size, dur := range durations {
		if dur > 0 {
			points = append(points, measurement{size, dur})
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].size < points[j].size })

	// If we have fewer than two successful measurements, check whether any
	// timed-out entries exist; a timeout in a two-point set strongly implies
	// exponential growth.
	if len(points) < 2 {
		for _, dur := range durations {
			if dur < 0 {
				return "exponential"
			}
		}
		return "unknown"
	}

	// Compute the ratio of duration growth to size growth across consecutive
	// measurement pairs. A ratio near 1 suggests linear scaling; a ratio
	// near the size-step multiplier suggests quadratic; much larger implies
	// exponential.
	var ratios []float64
	for i := 1; i < len(points); i++ {
		sizeRatio := float64(points[i].size) / float64(points[i-1].size)
		durRatio := float64(points[i].dur) / float64(points[i-1].dur)
		if sizeRatio > 0 && durRatio > 0 {
			ratios = append(ratios, durRatio/sizeRatio)
		}
	}

	if len(ratios) == 0 {
		return "unknown"
	}

	avgRatio := 0.0
	for _, r := range ratios {
		avgRatio += r
	}
	avgRatio /= float64(len(ratios))

	// A negative duration anywhere in the input map means at least one size
	// timed out, which is the strongest signal for exponential behavior and
	// overrides the ratio classification.
	for _, dur := range durations {
		if dur < 0 {
			return "exponential"
		}
	}

	switch {
	case math.IsInf(avgRatio, 0) || math.IsNaN(avgRatio):
		return "unknown"
	case avgRatio < 3:
		return "linear"
	case avgRatio < 50:
		return "quadratic"
	default:
		return "exponential"
	}
}
