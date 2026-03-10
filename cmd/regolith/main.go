package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/0x4d5352/regolith/internal/flavor"
	"github.com/0x4d5352/regolith/internal/output"
	"github.com/0x4d5352/regolith/internal/renderer"
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
	fs := flag.NewFlagSet("regolith", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Basic flags
	outputFile := fs.StringP("output", "o", "regex.svg", "Output file path")
	showVersion := fs.BoolP("version", "v", false, "Show version")
	flavorName := fs.StringP("flavor", "f", "javascript", "Regex flavor (javascript, java, dotnet, pcre, posix-bre, posix-ere, gnugrep, gnugrep-bre, gnugrep-ere)")
	formatName := fs.String("format", "svg", "Output format: svg, json, markdown")
	unescapeFlag := fs.BoolP("unescape", "u", false, `Apply string literal unescaping before parsing (e.g., \\ becomes \)`)

	// Dimension flags
	padding := fs.Float64P("padding", "p", 10, "Padding around diagram")
	fontSize := fs.Float64("font-size", 14, "Font size in pixels")
	lineWidth := fs.Float64("line-width", 2, "Stroke width for lines")

	// Color flags
	textColor := fs.String("text-color", "#000", "Text color")
	lineColor := fs.String("line-color", "#000", "Line/stroke color")
	literalFill := fs.String("literal-fill", "#ff6b6b", "Literal box fill color")
	charsetFill := fs.String("charset-fill", "#cbcbba", "Character set box fill color")
	escapeFill := fs.String("escape-fill", "#bada55", "Escape sequence box fill color")
	anchorFill := fs.String("anchor-fill", "#6b6659", "Anchor box fill color")
	subexpFill := fs.String("subexp-fill", "none", "Outermost subexpression box fill color (nested groups use cycling colors)")

	// Custom usage message
	fs.Usage = func() {
		fmt.Fprintf(stderr, "regolith - Visualize regular expressions as SVG diagrams\n\n")
		fmt.Fprintf(stderr, "Usage:\n")
		fmt.Fprintf(stderr, "  regolith [flags] <pattern>\n")
		fmt.Fprintf(stderr, "  echo 'pattern' | regolith [flags]\n\n")
		fmt.Fprintf(stderr, "Arguments:\n")
		fmt.Fprintf(stderr, "  pattern    Regular expression to visualize (reads from stdin if omitted)\n\n")
		fmt.Fprintf(stderr, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nAvailable flavors:\n")
		for _, name := range flavor.List() {
			f, _ := flavor.Get(name)
			fmt.Fprintf(stderr, "  %-12s %s\n", name, f.Description())
		}
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  regolith 'a|b|c'\n")
		fmt.Fprintf(stderr, "  regolith -o output.svg '[a-z]+'\n")
		fmt.Fprintf(stderr, "  regolith --flavor javascript '/pattern/gi'\n")
		fmt.Fprintf(stderr, "  regolith --literal-fill '#ff0000' 'hello'\n")
		fmt.Fprintf(stderr, "  echo '^hello$' | regolith\n")
		fmt.Fprintf(stderr, "  regolith -f java -u '\\\\d+\\\\.\\\\d+'\n")
		fmt.Fprintf(stderr, "\n  regolith --format json 'foo([a-z]+)' | jq .\n")
		fmt.Fprintf(stderr, "  regolith --format markdown '^hello$' | glow -\n")
		fmt.Fprintf(stderr, "  echo '[a-z]+' | regolith --format json\n")
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
		fmt.Fprintf(stdout, "regolith version %s\n", version)
		return nil
	}

	// Get the flavor
	f, ok := flavor.Get(*flavorName)
	if !ok {
		fmt.Fprintf(stderr, "Error: unknown flavor '%s'\n", *flavorName)
		fmt.Fprintf(stderr, "Available flavors: %s\n", strings.Join(flavor.List(), ", "))
		return fmt.Errorf("unknown flavor: %s", *flavorName)
	}

	// Get input pattern
	pattern, err := getInput(fs.Args(), stdin)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		fs.Usage()
		return err
	}

	// Apply string literal unescaping if requested, or warn about double escapes
	if *unescapeFlag {
		pattern = unescape.JavaStringLiteral(pattern)
	} else if (*flavorName == "java" || *flavorName == "dotnet") && unescape.ContainsDoubleEscapes(pattern) {
		fmt.Fprintf(stderr, "Note: Pattern contains '\\\\' sequences. If copied from source code, use --unescape to apply string literal unescaping.\n")
	}

	// Parse the pattern using the selected flavor
	parsedAST, err := f.Parse(pattern)
	if err != nil {
		displayParseError(stderr, pattern, err)
		return fmt.Errorf("parse error: %w", err)
	}

	switch *formatName {
	case "svg":
		// Build config from flags
		cfg := renderer.DefaultConfig()
		cfg.Padding = *padding
		cfg.FontSize = *fontSize
		cfg.CharWidth = *fontSize * 0.6 // Approximate monospace character width
		cfg.LineWidth = *lineWidth
		cfg.TextColor = *textColor
		cfg.LineColor = *lineColor
		cfg.LiteralFill = *literalFill
		cfg.CharsetFill = *charsetFill
		cfg.EscapeFill = *escapeFill
		cfg.AnchorFill = *anchorFill
		cfg.SubexpFill = *subexpFill

		r := renderer.New(cfg)
		svg := r.Render(parsedAST)

		err = os.WriteFile(*outputFile, []byte(svg), 0644)
		if err != nil {
			fmt.Fprintf(stderr, "Error writing output file: %v\n", err)
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(stdout, "Wrote %s\n", *outputFile)

	case "json":
		out, err := output.RenderJSON(parsedAST, pattern, f.Name())
		if err != nil {
			fmt.Fprintf(stderr, "Error rendering JSON: %v\n", err)
			return fmt.Errorf("json render: %w", err)
		}
		fmt.Fprintln(stdout, out)

	case "markdown":
		out := output.RenderMarkdown(parsedAST, pattern, f.Name())
		fmt.Fprint(stdout, out)

	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\nAvailable: svg, json, markdown\n", *formatName)
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
func displayParseError(w io.Writer, pattern string, err error) {
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

	fmt.Fprintf(w, "Error parsing pattern:\n\n")
	fmt.Fprintf(w, "  %s\n", pattern)

	// Show position indicator if we have column info
	if col > 0 && col <= len(pattern) {
		fmt.Fprintf(w, "  %s^\n", strings.Repeat(" ", col-1))
	}

	if msg != "" {
		fmt.Fprintf(w, "\n%s\n", msg)
	} else {
		fmt.Fprintf(w, "\n%s\n", errStr)
	}
}
