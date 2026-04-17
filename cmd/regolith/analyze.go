package main

// ================================================================================
// analyze subcommand
// ================================================================================

import (
	"fmt"
	"io"
	"time"

	"github.com/muesli/termenv"
	flag "github.com/spf13/pflag"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/flavor"
	"github.com/0x4d5352/regolith/internal/output"
	"github.com/0x4d5352/regolith/internal/renderer"
)

// runAnalyze implements the `regolith analyze` subcommand. It parses the
// pattern with the selected flavor, runs static analysis, optionally
// benchmarks runtime performance, and outputs results in text, JSON, or
// annotated SVG format.
func runAnalyze(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("regolith analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var common commonFlags
	common.Register(fs, commonDefaults{Format: "text", Output: ""})

	var style svgStyleFlags
	style.Register(fs)

	benchmark := fs.Bool("benchmark", false, "Enable runtime benchmarking")
	timeout := fs.Duration("timeout", 5*time.Second, "Per-input timeout for benchmarking")
	corpus := fs.StringSlice("corpus", []string{"all"},
		"Corpus types: prose, json, yaml, repeated, random, all")
	// TODO: --corpus-file support for custom corpus input
	sizes := fs.IntSlice("sizes", []int{10, 100, 1000, 10000, 100000},
		"Input sizes for benchmarking")
	severity := fs.String("severity", "info",
		"Minimum severity: info, warning, error, critical")

	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "regolith analyze - Analyze regex performance\n\n")
		_, _ = fmt.Fprintf(stderr, "Usage:\n")
		_, _ = fmt.Fprintf(stderr, "  regolith analyze [flags] <pattern>\n\n")
		_, _ = fmt.Fprintf(stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	profile := output.ResolveColorProfile(common.Color)
	co := termenv.NewOutput(stderr, termenv.WithProfile(profile))
	stdoutCo := termenv.NewOutput(stdout, termenv.WithProfile(profile))

	f, ok := flavor.Get(common.Flavor)
	if !ok {
		_, _ = fmt.Fprintf(stderr, "Error: unknown flavor '%s'\n", common.Flavor)
		return fmt.Errorf("unknown flavor: %s", common.Flavor)
	}

	pattern, err := getInput(fs.Args(), stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		fs.Usage()
		return err
	}

	parsedAST, err := f.Parse(pattern)
	if err != nil {
		displayParseError(stderr, pattern, err, co)
		return fmt.Errorf("parse error: %w", err)
	}

	report := analyzer.Analyze(parsedAST, pattern, f.Name(), f.SupportedFeatures())

	if *benchmark {
		corpusTypes := resolveCorpusTypes(*corpus)
		engine, isFallback := analyzer.DetectEngine(f.Name())
		opts := &analyzer.BenchmarkOptions{
			Pattern: pattern,
			Timeout: *timeout,
			Corpus:  corpusTypes,
			Sizes:   *sizes,
		}
		summary := analyzer.RunBenchmark(engine, opts)
		summary.IsFallback = isFallback
		report.BenchmarkSummary = summary
	}

	minSev := parseSeverity(*severity)
	report.Findings = filterBySeverity(report.Findings, minSev)

	switch common.Format {
	case "text":
		// Text mode: ANSI to stdout, Markdown to file. The termenv
		// output wrapping differs by destination — stdoutCo for the
		// terminal branch so it auto-strips colors when piped, `co`
		// for the file branch (the markdown renderer ignores it).
		toFile := common.Output != ""
		var text string
		if toFile {
			text = output.RenderAnalysisText(report, true, co)
		} else {
			text = output.RenderAnalysisText(report, false, stdoutCo)
		}
		return writeTextOrStdout(text, common.Output, stdout, co)

	case "json":
		jsonStr, err := output.RenderAnalysisJSON(report)
		if err != nil {
			return fmt.Errorf("json render: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, jsonStr)

	case "svg":
		return renderAndWriteSVG(fs, &common, &style, stdout, stderr, co,
			func(r *renderer.Renderer) string { return r.RenderAnnotated(parsedAST, report) })

	default:
		_, _ = fmt.Fprintf(stderr, "Error: unknown format %q\nAvailable: json, svg, text\n", common.Format)
		return fmt.Errorf("unknown format: %s", common.Format)
	}

	return nil
}

// resolveCorpusTypes expands "all" to the full list of built-in corpus types.
func resolveCorpusTypes(requested []string) []string {
	for _, r := range requested {
		if r == "all" {
			return analyzer.CorpusTypes()
		}
	}
	return requested
}

// parseSeverity converts a severity flag string to the corresponding
// analyzer.Severity constant.
func parseSeverity(s string) analyzer.Severity {
	switch s {
	case "warning":
		return analyzer.SeverityWarning
	case "error":
		return analyzer.SeverityError
	case "critical":
		return analyzer.SeverityCritical
	default:
		return analyzer.SeverityInfo
	}
}

// filterBySeverity returns only findings at or above the minimum severity.
func filterBySeverity(findings []*analyzer.Finding, min analyzer.Severity) []*analyzer.Finding {
	var filtered []*analyzer.Finding
	for _, f := range findings {
		if f.Severity >= min {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
