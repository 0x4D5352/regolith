package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/muesli/termenv"
	flag "github.com/spf13/pflag"

	"github.com/0x4d5352/regolith/internal/flavor"
	"github.com/0x4d5352/regolith/internal/output"
	"github.com/0x4d5352/regolith/internal/renderer"
	"github.com/0x4d5352/regolith/internal/renderer/theme"
	"github.com/0x4d5352/regolith/internal/unescape"
)

// runRender implements the main `regolith` command — parse a regex and
// emit it as SVG, JSON, or Markdown. Phase 1 of the refactor preserves
// the historical default (svg to regex.svg); phase 3 flips the default
// to text/ANSI on stdout.
func runRender(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("regolith", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var common commonFlags
	common.Register(fs, commonDefaults{Format: "text", Output: ""})

	var style svgStyleFlags
	style.Register(fs)

	showVersion := fs.BoolP("version", "v", false, "Show version")
	unescapeFlag := fs.BoolP("unescape", "u", false,
		`Apply string literal unescaping before parsing (e.g., \\ becomes \)`)

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
		for _, name := range theme.List() {
			t, _ := theme.Get(name)
			_, _ = fmt.Fprintf(stderr, "  %-22s %s\n", name, t.Description())
		}
		_, _ = fmt.Fprintf(stderr, "\nOutput:\n")
		_, _ = fmt.Fprintf(stderr, "  Default format is 'text': an ANSI-colored AST walk on stdout.\n")
		_, _ = fmt.Fprintf(stderr, "  Redirecting text to a file via -o writes Markdown instead.\n")
		_, _ = fmt.Fprintf(stderr, "  The 'svg' format requires -o with a destination filename.\n")
		_, _ = fmt.Fprintf(stderr, "\nExamples:\n")
		_, _ = fmt.Fprintf(stderr, "  regolith 'a|b|c'                              # ANSI walk on stdout\n")
		_, _ = fmt.Fprintf(stderr, "  regolith 'a|b|c' -o outline.md                # Markdown to file\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --format svg -o diagram.svg '[a-z]+' # SVG diagram to file\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --flavor javascript '/pattern/gi'\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --format svg --literal-fill '#ff0000' -o out.svg 'hello'\n")
		_, _ = fmt.Fprintf(stderr, "  echo '^hello$' | regolith\n")
		_, _ = fmt.Fprintf(stderr, "  regolith -f java -u '\\\\d+\\\\.\\\\d+'\n")
		_, _ = fmt.Fprintf(stderr, "  regolith --format json 'foo([a-z]+)' | jq .\n")
		_, _ = fmt.Fprintf(stderr, "  echo '[a-z]+' | regolith --format json\n")
	}

	err := fs.Parse(args[1:])
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	if *showVersion {
		_, _ = fmt.Fprintf(stdout, "regolith version %s\n", version)
		return nil
	}

	profile := output.ResolveColorProfile(common.Color)
	// Two termenv outputs so stdout-bound content and stderr-bound
	// status messages each get the auto-detected profile for their
	// own writer. Piping stdout to a file correctly yields plain
	// text on stdout while leaving stderr colored if that's still
	// a TTY.
	co := termenv.NewOutput(stderr, termenv.WithProfile(profile))
	stdoutCo := termenv.NewOutput(stdout, termenv.WithProfile(profile))

	f, ok := flavor.Get(common.Flavor)
	if !ok {
		_, _ = fmt.Fprintf(stderr, "Error: unknown flavor '%s'\n", common.Flavor)
		_, _ = fmt.Fprintf(stderr, "Available flavors: %s\n", strings.Join(flavor.List(), ", "))
		return fmt.Errorf("unknown flavor: %s", common.Flavor)
	}

	pattern, err := getInput(fs.Args(), stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		fs.Usage()
		return err
	}

	if *unescapeFlag {
		pattern = unescape.JavaStringLiteral(pattern)
	} else if (common.Flavor == "java" || common.Flavor == "dotnet") && unescape.ContainsDoubleEscapes(pattern) {
		_, _ = fmt.Fprintf(stderr, "Note: Pattern contains '\\\\' sequences. If copied from source code, use --unescape to apply string literal unescaping.\n")
	}

	parsedAST, err := f.Parse(pattern)
	if err != nil {
		displayParseError(stderr, pattern, err, co)
		return fmt.Errorf("parse error: %w", err)
	}

	switch common.Format {
	case "text":
		// Text format has two personalities: ANSI on stdout (default)
		// and Markdown when redirected to a file via -o. This mirrors
		// the convention established by `regolith analyze`, keeping
		// both commands predictable.
		toFile := common.Output != ""
		text := output.RenderText(parsedAST, pattern, f.Name(), toFile, stdoutCo)
		if toFile {
			return writeOutputFile(common.Output, []byte(text), stdout, co)
		}
		_, _ = fmt.Fprint(stdout, text)

	case "svg":
		if err := requireOutputForSVG(common.Format, common.Output); err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
			return err
		}
		cfg, err := buildSVGConfig(fs, &common, &style)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
			return err
		}
		r := renderer.New(cfg)
		svg := r.Render(parsedAST)
		return writeOutputFile(common.Output, []byte(svg), stdout, co)

	case "json":
		out, err := output.RenderJSON(parsedAST, pattern, f.Name())
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error rendering JSON: %v\n", err)
			return fmt.Errorf("json render: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, out)

	default:
		_, _ = fmt.Fprintf(stderr, "Error: unknown format %q\nAvailable: json, svg, text\n", common.Format)
		return fmt.Errorf("unknown format: %s", common.Format)
	}

	return nil
}

// getInput retrieves the regex pattern from CLI args or stdin.
// Args take priority; stdin is only consulted when no pattern was given.
func getInput(args []string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if stdin != nil {
		input, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return strings.TrimSpace(string(input)), nil
	}
	return "", fmt.Errorf("no pattern provided")
}

// displayParseError shows a parse error with a caret pointing at the
// offending column when the pigeon error text has usable position
// information.
func displayParseError(w io.Writer, pattern string, err error, co *termenv.Output) {
	errStr := err.Error()

	var line, col int
	var msg string

	if strings.Contains(errStr, "parse error:") {
		_, parseErr := fmt.Sscanf(errStr, "parse error: %d:%d", &line, &col)
		if parseErr == nil {
			idx := strings.Index(errStr, ":")
			if idx != -1 {
				idx = strings.Index(errStr[idx+1:], ":")
				if idx != -1 {
					remaining := errStr[strings.Index(errStr, "parse error:")+len("parse error:"):]
					parts := strings.SplitN(remaining, ":", 2)
					if len(parts) > 1 {
						msg = strings.TrimSpace(parts[1])
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
