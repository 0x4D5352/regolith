package analyzer

import (
	"fmt"
	"strconv"
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

	return runTimedScript("python3", []string{"-c", script}, timeout)
}
