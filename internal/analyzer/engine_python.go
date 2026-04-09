package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PythonEngine executes regex patterns via Python's re module.
// For PCRE patterns, it attempts to use the regex module which
// has better PCRE compatibility.
type PythonEngine struct {
	UsePCRE bool
}

func (e *PythonEngine) Name() string { return "python3" }

func (e *PythonEngine) Run(pattern, input string, timeout time.Duration) (time.Duration, error) {
	module := "re"
	if e.UsePCRE {
		module = "regex"
	}

	script := fmt.Sprintf(`
import %s as re_mod
import time
import sys

pattern = %s
text = %s
start = time.perf_counter_ns()
re_mod.search(pattern, text)
end = time.perf_counter_ns()
sys.stdout.write(str(end - start))
`, module, strconv.Quote(pattern), strconv.Quote(input))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-c", script)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return timeout, fmt.Errorf("python3 timeout after %v", timeout)
		}
		return 0, fmt.Errorf("python3 exec: %w", err)
	}

	ns, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("python3 output parse: %w", err)
	}

	return time.Duration(ns), nil
}
