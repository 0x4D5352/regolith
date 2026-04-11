package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/muesli/termenv"
	flag "github.com/spf13/pflag"

	"github.com/0x4d5352/regolith/internal/analyzer"
	"github.com/0x4d5352/regolith/internal/flavor"
	"github.com/0x4d5352/regolith/internal/output"
	"github.com/0x4d5352/regolith/internal/renderer"
	"github.com/0x4d5352/regolith/internal/renderer/theme"
	"github.com/0x4d5352/regolith/internal/unescape"

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

var (
	version = "0.2.0"
)

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

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// Route to the analyze subcommand before pflag parsing, since it has
	// its own FlagSet with different defaults (e.g., --format defaults to
	// "text" instead of "svg").
	if len(args) > 1 && args[1] == "analyze" {
		return runAnalyze(args, stdin, stdout, stderr)
	}

	fs := flag.NewFlagSet("regolith", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Basic flags
	outputFile := fs.StringP("output", "o", "regex.svg", "Output file path")
	showVersion := fs.BoolP("version", "v", false, "Show version")
	flavorName := fs.StringP("flavor", "f", "javascript", "Regex flavor (javascript, java, dotnet, pcre, posix-bre, posix-ere, gnugrep, gnugrep-bre, gnugrep-ere)")
	formatName := fs.String("format", "svg", "Output format: svg, json, markdown")
	unescapeFlag := fs.BoolP("unescape", "u", false, `Apply string literal unescaping before parsing (e.g., \\ becomes \)`)
	colorMode := fs.String("color", "auto", "Color output: auto, always, never")
	themeName := fs.String("theme", "", "Color theme (e.g. catppuccin-mocha, gruvbox-dark, default)")

	// Dimension flags
	padding := fs.Float64P("padding", "p", 10, "Padding around diagram")
	fontSize := fs.Float64("font-size", 13, "Font size in pixels")
	lineWidth := fs.Float64("line-width", 1.5, "Stroke width for connectors and loops")

	// Color flags. These layer on top of the default config (or the
	// selected --theme). Defaults match the built-in palette, but the
	// overrides are only applied when the user explicitly sets them —
	// otherwise a --theme selection would be silently clobbered by the
	// hard-coded defaults below. See applyColorFlags.
	textColor := fs.String("text-color", "#000", "Fallback text color for elements outside any category")
	lineColor := fs.String("line-color", "#64748b", "Connector / loop line color")
	literalFill := fs.String("literal-fill", "#fee2e2", "Literal box fill color")
	charsetFill := fs.String("charset-fill", "#f5f0e1", "Character set box fill color")
	escapeFill := fs.String("escape-fill", "#ecfccb", "Escape sequence box fill color")
	anchorFill := fs.String("anchor-fill", "#334155", "Anchor box fill color")
	subexpFill := fs.String("subexp-fill", "none", "Outermost subexpression box fill color (nested groups use cycling colors)")

	// Custom usage message
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "regolith - Visualize regular expressions as SVG diagrams\n\n")
		_, _ = fmt.Fprintf(stderr, "Usage:\n")
		_, _ = fmt.Fprintf(stderr, "  regolith [flags] <pattern>\n")
		_, _ = fmt.Fprintf(stderr, "  echo 'pattern' | regolith [flags]\n\n")
		_, _ = fmt.Fprintf(stderr, "Arguments:\n")
		_, _ = fmt.Fprintf(stderr, "  pattern    Regular expression to visualize (reads from stdin if omitted)\n\n")
		_, _ = fmt.Fprintf(stderr, "Flags:\n")
		fs.PrintDefaults()
		_, _ = fmt.Fprintf(stderr, "\nAvailable flavors:\n")
		for _, name := range flavor.List() {
			f, _ := flavor.Get(name)
			_, _ = fmt.Fprintf(stderr, "  %-12s %s\n", name, f.Description())
		}
		_, _ = fmt.Fprintf(stderr, "\nAvailable themes:\n")
		_, _ = fmt.Fprintf(stderr, "  %-22s %s\n", "default", "Built-in palette (refreshed visual style)")
		for _, name := range theme.List() {
			t, _ := theme.Get(name)
			_, _ = fmt.Fprintf(stderr, "  %-22s %s\n", name, t.Description())
		}
		_, _ = fmt.Fprintf(stderr, "\nExamples:\n")
		_, _ = fmt.Fprintf(stderr, "  regolith 'a|b|c'\n")
		_, _ = fmt.Fprintf(stderr, "  regolith -o output.svg '[a-z]+'\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --flavor javascript '/pattern/gi'\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --literal-fill '#ff0000' 'hello'\n")
		_, _ = fmt.Fprintf(stderr, "  echo '^hello$' | regolith\n")
		_, _ = fmt.Fprintf(stderr, "  regolith -f java -u '\\\\d+\\\\.\\\\d+'\n")
		_, _ = fmt.Fprintf(stderr, "\n  regolith --format json 'foo([a-z]+)' | jq .\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --format markdown '^hello$' | glow -\n")
		_, _ = fmt.Fprintf(stderr, "  echo '[a-z]+' | regolith --format json\n")
	}

	err := fs.Parse(args[1:])
	if errors.Is(err, flag.ErrHelp) {
		// FlagSet already printed default usage; print our custom usage instead
		return nil
	}
	if err != nil {
		return err
	}

	if *showVersion {
		_, _ = fmt.Fprintf(stdout, "regolith version %s\n", version)
		return nil
	}

	profile := output.ResolveColorProfile(*colorMode)
	co := termenv.NewOutput(stderr, termenv.WithProfile(profile))

	// Get the flavor
	f, ok := flavor.Get(*flavorName)
	if !ok {
		_, _ = fmt.Fprintf(stderr, "Error: unknown flavor '%s'\n", *flavorName)
		_, _ = fmt.Fprintf(stderr, "Available flavors: %s\n", strings.Join(flavor.List(), ", "))
		return fmt.Errorf("unknown flavor: %s", *flavorName)
	}

	// Get input pattern
	pattern, err := getInput(fs.Args(), stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		fs.Usage()
		return err
	}

	// Apply string literal unescaping if requested, or warn about double escapes
	if *unescapeFlag {
		pattern = unescape.JavaStringLiteral(pattern)
	} else if (*flavorName == "java" || *flavorName == "dotnet") && unescape.ContainsDoubleEscapes(pattern) {
		_, _ = fmt.Fprintf(stderr, "Note: Pattern contains '\\\\' sequences. If copied from source code, use --unescape to apply string literal unescaping.\n")
	}

	// Parse the pattern using the selected flavor
	parsedAST, err := f.Parse(pattern)
	if err != nil {
		displayParseError(stderr, pattern, err, co)
		return fmt.Errorf("parse error: %w", err)
	}

	switch *formatName {
	case "svg":
		// Build config: defaults → theme → explicit CLI overrides.
		// Order matters. A theme replaces the color fields wholesale;
		// the --literal-fill / --line-color / etc. flags then layer on
		// top so a user can tint one category without rebuilding the
		// whole palette. Only flags the user actually set are applied;
		// otherwise the hard-coded defaults would silently clobber the
		// theme's choices.
		cfg := renderer.DefaultConfig()
		if err := applyTheme(cfg, *themeName); err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
			return err
		}
		cfg.Padding = *padding
		cfg.FontSize = *fontSize
		cfg.CharWidth = *fontSize * 0.6
		cfg.Connector.StrokeWidth = *lineWidth
		if fs.Changed("line-color") {
			cfg.Connector.Color = *lineColor
		}
		if fs.Changed("text-color") {
			cfg.TextColor = *textColor
		}
		if fs.Changed("literal-fill") {
			patchNodeFill(cfg, "literal", *literalFill)
		}
		if fs.Changed("charset-fill") {
			patchNodeFill(cfg, "charset", *charsetFill)
		}
		if fs.Changed("escape-fill") {
			patchNodeFill(cfg, "escape", *escapeFill)
		}
		if fs.Changed("anchor-fill") {
			patchNodeFill(cfg, "anchor", *anchorFill)
		}
		if fs.Changed("subexp-fill") {
			cfg.SubexpFill = *subexpFill
		}

		r := renderer.New(cfg)
		svg := r.Render(parsedAST)

		err = os.WriteFile(*outputFile, []byte(svg), 0644)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error writing output file: %v\n", err)
			return fmt.Errorf("writing output: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, co.String("Wrote "+*outputFile).Foreground(termenv.ANSIColor(2)).String())

	case "json":
		out, err := output.RenderJSON(parsedAST, pattern, f.Name())
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error rendering JSON: %v\n", err)
			return fmt.Errorf("json render: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, out)

	case "markdown":
		out := output.RenderMarkdown(parsedAST, pattern, f.Name())
		_, _ = fmt.Fprint(stdout, out)

	default:
		_, _ = fmt.Fprintf(stderr, "Error: unknown format %q\nAvailable: svg, json, markdown\n", *formatName)
		return fmt.Errorf("unknown format: %s", *formatName)
	}

	return nil
}

// getInput retrieves the regex pattern from CLI args or stdin
func getInput(args []string, stdin io.Reader) (string, error) {
	// Check for argument (args take priority over stdin)
	if len(args) > 0 {
		return args[0], nil
	}

	// Check for piped stdin
	if stdin != nil {
		input, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return strings.TrimSpace(string(input)), nil
	}

	return "", fmt.Errorf("no pattern provided")
}

// displayParseError shows a parse error with position indicator
func displayParseError(w io.Writer, pattern string, err error, co *termenv.Output) {
	errStr := err.Error()

	// Try to extract position from pigeon error format
	// Format: "parse error: 1:5 (4): ..."
	var line, col int
	var msg string

	// Check for pigeon error format
	if strings.Contains(errStr, "parse error:") {
		// Try to parse position
		_, parseErr := fmt.Sscanf(errStr, "parse error: %d:%d", &line, &col)
		if parseErr == nil {
			// Extract message after position
			idx := strings.Index(errStr, ":")
			if idx != -1 {
				idx = strings.Index(errStr[idx+1:], ":")
				if idx != -1 {
					// Find the actual message after position info
					remaining := errStr[strings.Index(errStr, "parse error:")+len("parse error:"):]
					parts := strings.SplitN(remaining, ":", 2)
					if len(parts) > 1 {
						msg = strings.TrimSpace(parts[1])
						// Clean up the message
						if strings.Contains(msg, "):") {
							msg = strings.TrimSpace(strings.SplitN(msg, "):", 2)[1])
						}
					}
				}
			}
		}
	}

	header := co.String("Error parsing pattern:").Bold().Foreground(termenv.ANSIColor(1)).String()
	_, _ = fmt.Fprintf(w, "%s\n\n", header)
	_, _ = fmt.Fprintf(w, "  %s\n", pattern)

	// Show position indicator if we have column info
	if col > 0 && col <= len(pattern) {
		caret := co.String("^").Bold().Foreground(termenv.ANSIColor(1)).String()
		_, _ = fmt.Fprintf(w, "  %s%s\n", strings.Repeat(" ", col-1), caret)
	}

	if msg != "" {
		_, _ = fmt.Fprintf(w, "\n%s\n", msg)
	} else {
		_, _ = fmt.Fprintf(w, "\n%s\n", errStr)
	}
}

// ================================================================================
// analyze subcommand
// ================================================================================

// runAnalyze implements the "regolith analyze" subcommand. It parses the
// pattern with the selected flavor, runs static analysis, optionally
// benchmarks runtime performance, and outputs results in text, JSON, or
// annotated SVG format.
func runAnalyze(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("regolith analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	flavorName := fs.StringP("flavor", "f", "javascript", "Regex flavor")
	formatName := fs.String("format", "text", "Output format: text, json, svg")
	outputFile := fs.StringP("output", "o", "", "Output file path (required for svg)")
	benchmark := fs.Bool("benchmark", false, "Enable runtime benchmarking")
	timeout := fs.Duration("timeout", 5*time.Second, "Per-input timeout for benchmarking")
	corpus := fs.StringSlice("corpus", []string{"all"}, "Corpus types: prose, json, yaml, repeated, random, all")
	// TODO: --corpus-file support for custom corpus input
	// corpusFile := fs.String("corpus-file", "", "Path to custom corpus file")
	sizes := fs.IntSlice("sizes", []int{10, 100, 1000, 10000, 100000}, "Input sizes for benchmarking")
	severity := fs.String("severity", "info", "Minimum severity: info, warning, error, critical")
	colorMode := fs.String("color", "auto", "Color output: auto, always, never")

	padding := fs.Float64P("padding", "p", 10, "Padding around diagram")
	fontSize := fs.Float64("font-size", 13, "Font size in pixels")
	lineWidth := fs.Float64("line-width", 1.5, "Stroke width for connectors and loops")
	themeName := fs.String("theme", "", "Color theme for svg output (e.g. catppuccin-mocha, gruvbox-dark, default)")

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

	profile := output.ResolveColorProfile(*colorMode)
	co := termenv.NewOutput(stderr, termenv.WithProfile(profile))

	f, ok := flavor.Get(*flavorName)
	if !ok {
		_, _ = fmt.Fprintf(stderr, "Error: unknown flavor '%s'\n", *flavorName)
		return fmt.Errorf("unknown flavor: %s", *flavorName)
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

	switch *formatName {
	case "text":
		markdown := *outputFile != ""
		text := output.RenderAnalysisText(report, markdown, co)
		if *outputFile != "" {
			if err := os.WriteFile(*outputFile, []byte(text), 0644); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			_, _ = fmt.Fprintln(stdout, co.String("Wrote "+*outputFile).Foreground(termenv.ANSIColor(2)).String())
		} else {
			_, _ = fmt.Fprint(stdout, text)
		}

	case "json":
		jsonStr, err := output.RenderAnalysisJSON(report)
		if err != nil {
			return fmt.Errorf("json render: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, jsonStr)

	case "svg":
		if *outputFile == "" {
			_, _ = fmt.Fprintf(stderr, "Error: --output is required for svg format\n")
			return fmt.Errorf("--output required for svg")
		}
		cfg := renderer.DefaultConfig()
		if err := applyTheme(cfg, *themeName); err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
			return err
		}
		cfg.Padding = *padding
		cfg.FontSize = *fontSize
		cfg.CharWidth = *fontSize * 0.6
		cfg.Connector.StrokeWidth = *lineWidth
		r := renderer.New(cfg)
		svg := r.RenderAnnotated(parsedAST, report)
		if err := os.WriteFile(*outputFile, []byte(svg), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, co.String("Wrote "+*outputFile).Foreground(termenv.ANSIColor(2)).String())

	default:
		_, _ = fmt.Fprintf(stderr, "Error: unknown format %q\n", *formatName)
		return fmt.Errorf("unknown format: %s", *formatName)
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

// applyTheme resolves a theme name and applies it to cfg. An empty
// string or the literal "default" are both treated as no-ops so the
// user can type `--theme default` explicitly to opt out of any theme
// selection that might come from a future config file, alias, or
// environment variable. An unknown theme produces an error listing
// every registered theme so the user can pick a valid one without
// leaving the terminal.
func applyTheme(cfg *renderer.Config, name string) error {
	if name == "" || name == "default" {
		return nil
	}
	t, ok := theme.Get(name)
	if !ok {
		return fmt.Errorf("unknown theme %q (available: default, %s)",
			name, strings.Join(theme.List(), ", "))
	}
	t.Apply(cfg)
	return nil
}

// patchNodeFill overrides just the Fill field on a single category's
// NodeStyle, leaving the stroke and text color from the underlying
// theme in place. Used by the --literal-fill / --charset-fill /
// --escape-fill / --anchor-fill flags so users can tint a single
// category without having to provide a whole color bundle.
func patchNodeFill(cfg *renderer.Config, class, fill string) {
	style := cfg.GetNodeStyle(class)
	style.Fill = fill
	cfg.NodeStyles[class] = style
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
