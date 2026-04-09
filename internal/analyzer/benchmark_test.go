package analyzer

import (
	"testing"
	"time"
)

func TestRunBenchmark(t *testing.T) {
	opts := &BenchmarkOptions{
		Pattern: `hello`,
		Timeout: 5 * time.Second,
		Corpus:  []string{"repeated", "random"},
		Sizes:   []int{10, 100, 1000},
	}

	engine := &Regexp2Engine{}
	summary := RunBenchmark(engine, opts)

	if summary.Engine != "regexp2" {
		t.Errorf("expected engine regexp2, got %s", summary.Engine)
	}

	if len(summary.Corpus) != 2 {
		t.Errorf("expected 2 corpus results, got %d", len(summary.Corpus))
	}

	for _, corpus := range []string{"repeated", "random"} {
		results, ok := summary.Corpus[corpus]
		if !ok {
			t.Errorf("missing corpus %s", corpus)
			continue
		}
		if len(results) != 3 {
			t.Errorf("%s: expected 3 size results, got %d", corpus, len(results))
		}
	}
}

func TestClassifyScaling(t *testing.T) {
	tests := []struct {
		name      string
		durations map[int]time.Duration
		want      string
	}{
		{
			name: "linear",
			durations: map[int]time.Duration{
				100:  1 * time.Millisecond,
				1000: 10 * time.Millisecond,
			},
			want: "linear",
		},
		{
			name: "quadratic",
			durations: map[int]time.Duration{
				100:  1 * time.Millisecond,
				1000: 100 * time.Millisecond,
			},
			want: "quadratic",
		},
		{
			name: "single measurement",
			durations: map[int]time.Duration{
				100: 1 * time.Millisecond,
			},
			want: "unknown",
		},
		{
			name: "has timeout entry",
			durations: map[int]time.Duration{
				100:  1 * time.Millisecond,
				1000: -5 * time.Second,
			},
			want: "exponential",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyScaling(tc.durations)
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
