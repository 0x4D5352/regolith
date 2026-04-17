package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// runTimedScript executes an external engine whose script writes a
// single nanosecond timing integer to stdout (Node, Python), then
// parses that value and returns it as a time.Duration. Timeout errors
// are surfaced in a uniform shape so callers don't re-implement the
// DeadlineExceeded check each time.
//
// This helper collapses ~20 lines of identical boilerplate that used
// to live in engine_node.go and engine_python.go.
func runTimedScript(name string, args []string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return timeout, fmt.Errorf("%s timeout after %v", name, timeout)
		}
		return 0, fmt.Errorf("%s exec: %w", name, err)
	}

	ns, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s output parse: %w", name, err)
	}
	return time.Duration(ns), nil
}

// runWallClockCommand executes an engine that does not self-report
// timing (grep) and measures wall-clock time around the subprocess.
// A non-zero exit is ignored — grep returns 1 for "no match", which is
// a legitimate outcome of benchmarking a pattern against arbitrary
// input. Only the timeout is surfaced as an error.
func runWallClockCommand(name string, args []string, stdin string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	start := time.Now()
	_ = cmd.Run()
	elapsed := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return timeout, fmt.Errorf("%s timeout after %v", name, timeout)
	}
	return elapsed, nil
}
