package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GrepEngine executes regex patterns via GNU grep for POSIX-flavor
// benchmarking.
type GrepEngine struct {
	UseBRE bool
}

func (e *GrepEngine) Name() string { return "grep" }

func (e *GrepEngine) Run(pattern, input string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{"-c"}
	if !e.UseBRE {
		args = append(args, "-E")
	}
	args = append(args, pattern)

	cmd := exec.CommandContext(ctx, "grep", args...)
	cmd.Stdin = strings.NewReader(input)

	start := time.Now()
	_ = cmd.Run()
	elapsed := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return timeout, fmt.Errorf("grep timeout after %v", timeout)
	}

	return elapsed, nil
}
