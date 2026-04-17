package analyzer

import "time"

// GrepEngine executes regex patterns via GNU grep for POSIX-flavor
// benchmarking.
type GrepEngine struct {
	UseBRE bool
}

func (e *GrepEngine) Name() string { return "grep" }

func (e *GrepEngine) Run(pattern, input string, timeout time.Duration) (time.Duration, error) {
	args := []string{"-c"}
	if !e.UseBRE {
		args = append(args, "-E")
	}
	args = append(args, pattern)
	return runWallClockCommand("grep", args, input, timeout)
}
