package analyzer

import (
	"fmt"
	"strconv"
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

	return runTimedScript("node", []string{"-e", script}, timeout)
}
