package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// NodeEngine executes regex patterns via Node.js for accurate
// JavaScript-flavor benchmarking.
type NodeEngine struct{}

func (e *NodeEngine) Name() string { return "node" }

func (e *NodeEngine) Run(pattern, input string, timeout time.Duration) (time.Duration, error) {
	script := fmt.Sprintf(`
const pattern = %s;
const input = %s;
const re = new RegExp(pattern);
const start = process.hrtime.bigint();
re.test(input);
const end = process.hrtime.bigint();
process.stdout.write(String(end - start));
`, strconv.Quote(pattern), strconv.Quote(input))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return timeout, fmt.Errorf("node timeout after %v", timeout)
		}
		return 0, fmt.Errorf("node exec: %w", err)
	}

	ns, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("node output parse: %w", err)
	}

	return time.Duration(ns), nil
}
