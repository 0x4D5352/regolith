package main

import (
	"io"
	"os"

	// Import flavors to register them via init()
	_ "github.com/0x4d5352/regolith/internal/flavor/dotnet"
	_ "github.com/0x4d5352/regolith/internal/flavor/gnugrep_bre"
	_ "github.com/0x4d5352/regolith/internal/flavor/gnugrep_ere"
	_ "github.com/0x4d5352/regolith/internal/flavor/java"
	_ "github.com/0x4d5352/regolith/internal/flavor/javascript"
	_ "github.com/0x4d5352/regolith/internal/flavor/pcre"
	_ "github.com/0x4d5352/regolith/internal/flavor/posix_bre"
	_ "github.com/0x4d5352/regolith/internal/flavor/posix_ere"
)

var version = "0.3.0"

func main() {
	var stdin io.Reader
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		stdin = os.Stdin
	}
	if err := run(os.Args, stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

// run is the top-level dispatcher. The subcommand routing happens
// before pflag parsing because `regolith analyze` has its own FlagSet
// with a different default --format.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) > 1 && args[1] == "analyze" {
		return runAnalyze(args, stdin, stdout, stderr)
	}
	return runRender(args, stdin, stdout, stderr)
}
