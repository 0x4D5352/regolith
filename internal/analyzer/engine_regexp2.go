package analyzer

import (
	"fmt"
	"time"

	"github.com/dlclark/regexp2"
)

// Regexp2Engine uses the dlclark/regexp2 library as a backtracking regex
// engine. It serves as the universal fallback when the real engine for a
// flavor is not available.
type Regexp2Engine struct{}

func (e *Regexp2Engine) Name() string { return "regexp2" }

func (e *Regexp2Engine) Run(pattern, input string, timeout time.Duration) (time.Duration, error) {
	re, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return 0, fmt.Errorf("regexp2 compile: %w", err)
	}
	re.MatchTimeout = timeout

	start := time.Now()
	_, err = re.MatchString(input)
	elapsed := time.Since(start)

	if err != nil {
		return elapsed, fmt.Errorf("regexp2 timeout after %v: %w", elapsed, err)
	}

	return elapsed, nil
}
