package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x4d5352/regolith/internal/flavor"
)

// ---------------------------------------------------------------------------
// run() function tests
// ---------------------------------------------------------------------------

func TestRunValidPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "a|b|c"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	if !strings.Contains(stdout.String(), "Wrote") {
		t.Errorf("expected stdout to contain 'Wrote', got: %s", stdout.String())
	}
}

func TestRunInvalidPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-o", out, "(?P<"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}

	if stderr.Len() == 0 {
		t.Error("expected stderr to contain error message")
	}
}

func TestRunFlavorFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "java", "-o", out, "[a-z]+"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with java flavor, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestRunAllFlavors(t *testing.T) {
	for _, name := range flavor.List() {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "out.svg")

			var stdout, stderr bytes.Buffer
			err := run([]string{"regolith", "--format", "svg", "--flavor", name, "-o", out, "abc"}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("flavor %s failed on basic pattern: %v\nstderr: %s", name, err, stderr.String())
			}
		})
	}
}

func TestRunUnknownFlavor(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--flavor", "bogus", "-o", out, "abc"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown flavor, got nil")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "unknown flavor") {
		t.Errorf("expected stderr to mention 'unknown flavor', got: %s", stderrStr)
	}
	if !strings.Contains(stderrStr, "Available flavors") {
		t.Errorf("expected stderr to list available flavors, got: %s", stderrStr)
	}
}

func TestRunOutputFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "custom-name.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "hello"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatalf("expected output at %s, file not found", out)
	}

	if !strings.Contains(stdout.String(), "custom-name.svg") {
		t.Errorf("expected stdout to reference output file name, got: %s", stdout.String())
	}
}

func TestRunStdinInput(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	stdin := strings.NewReader("a|b\n")
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error reading from stdin, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created from stdin input")
	}
}

func TestRunStdinAndArgs(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Provide both args and stdin; args should take priority
	stdin := strings.NewReader("x|y\n")
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "a|b"}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	// Verify the file was created (we can't easily distinguish which pattern
	// was used just from the SVG, but the key assertion is args win over stdin)
	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestRunNoInput(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-o", out}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when no input provided, got nil")
	}

	if !strings.Contains(stderr.String(), "no pattern provided") {
		t.Errorf("expected stderr to mention 'no pattern provided', got: %s", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-v"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error for -v, got: %v", err)
	}

	if !strings.Contains(stdout.String(), "regolith version") {
		t.Errorf("expected version string in stdout, got: %s", stdout.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-h"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error for -h, got: %v", err)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Usage:") {
		t.Errorf("expected usage in stderr, got: %s", stderrStr)
	}
	if !strings.Contains(stderrStr, "Available flavors:") {
		t.Errorf("expected flavor list in stderr, got: %s", stderrStr)
	}
}

func TestRunSVGContent(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "[a-z]+"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	svg := string(data)
	if !strings.Contains(svg, "<svg") {
		t.Error("output does not contain <svg tag")
	}
	if !strings.Contains(svg, "xmlns") {
		t.Error("output does not contain xmlns attribute")
	}
	if !strings.Contains(svg, "viewBox") {
		t.Error("output does not contain viewBox attribute")
	}
}

func TestRunCustomColors(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"regolith",
		"--format", "svg",
		"-o", out,
		"--text-color", "#fff",
		"--line-color", "#333",
		"--literal-fill", "#00ff00",
		"hello",
	}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error with custom colors: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created with custom colors")
	}
}

// ---------------------------------------------------------------------------
// os/exec binary tests
// ---------------------------------------------------------------------------

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "regolith-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmp, "regolith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(".", "") // current package directory
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n", err)
		_ = os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func TestBinaryValidPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "--format", "svg", "-o", out, "a|b|c")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\noutput: %s", err, output)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if !strings.Contains(string(data), "<svg") {
		t.Error("output file does not contain valid SVG")
	}
}

func TestBinaryInvalidPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "-o", out, "(?P<")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code for invalid pattern")
	}

	if stderr.Len() == 0 {
		t.Error("expected error output on stderr")
	}
}

func TestBinaryVersion(t *testing.T) {
	cmd := exec.Command(binaryPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0 for -v, got: %v", err)
	}

	if !strings.Contains(string(output), "regolith version") {
		t.Errorf("expected version string, got: %s", output)
	}
}

func TestBinaryHelp(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0 for -h, got: %v", err)
	}

	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage in stderr, got: %s", stderr.String())
	}
}

func TestBinaryStdin(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "--format", "svg", "-o", out)
	cmd.Stdin = strings.NewReader("a|b\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected exit 0 for stdin input, got: %v\noutput: %s", err, output)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("SVG file was not created from stdin input")
	}
}

func TestBinaryUnknownFlavor(t *testing.T) {
	cmd := exec.Command(binaryPath, "--flavor", "bogus", "abc")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for unknown flavor")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Available flavors") {
		t.Errorf("expected available flavors list in stderr, got: %s", stderrStr)
	}
}

func TestBinaryFlavorFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Use PCRE-specific lookbehind syntax
	cmd := exec.Command(binaryPath, "--format", "svg", "--flavor", "pcre", "-o", out, "(?<=foo)bar")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected exit 0 for pcre flavor, got: %v\noutput: %s", err, output)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("SVG file was not created with pcre flavor")
	}
}

// ---------------------------------------------------------------------------
// -unescape flag tests
// ---------------------------------------------------------------------------

func TestRunUnescapeFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Pattern with double escapes + -unescape flag: should produce SVG, no warning
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "java", "--unescape", "-o", out, `\\d+\\.\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with -unescape, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	if strings.Contains(stderr.String(), "Note:") {
		t.Error("expected no warning with -unescape flag, but got one")
	}
}

func TestRunDoubleEscapeWarningJava(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Java flavor with double escapes but no -unescape flag: should warn
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "java", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error (warning only), got: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stderr.String(), "--unescape") {
		t.Errorf("expected warning mentioning -unescape, got: %s", stderr.String())
	}
}

func TestRunDoubleEscapeWarningDotnet(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Dotnet flavor with double escapes but no -unescape flag: should warn
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "dotnet", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error (warning only), got: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stderr.String(), "--unescape") {
		t.Errorf("expected warning mentioning -unescape, got: %s", stderr.String())
	}
}

func TestRunNoWarningForJavaScript(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// JavaScript flavor with double escapes: no warning expected
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "javascript", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "--unescape") {
		t.Error("expected no warning for javascript flavor, but got one")
	}
}

func TestRunNoWarningWithoutDoubleEscapes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Java flavor without double escapes: no warning expected
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "java", "-o", out, `\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "--unescape") {
		t.Error("expected no warning without double escapes, but got one")
	}
}

// ---------------------------------------------------------------------------
// Interspersed flag tests (flags after positional args)
// ---------------------------------------------------------------------------

func TestRunFlagsAfterPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "a|b|c", "--output", out}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with flags after pattern, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatalf("expected output at %s, file not found", out)
	}
}

func TestRunMixedFlagPositions(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--flavor", "java", "[a-z]+", "--output", out}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with mixed flag positions, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created with mixed flag positions")
	}
}

func TestRunShortFlags(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "-f", "java", "[a-z]+"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with short flags, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created with short flags")
	}
}

func TestRunDoubleDashSeparator(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	// Pattern starts with dash; use -- to separate it from flags
	err := run([]string{"regolith", "--format", "svg", "--output", out, "--", "-abc"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with -- separator, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("output file was not created with -- separator")
	}
}

func TestRunColorFlagsAfterPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--output", out, "hello", "--literal-fill", "#ff0000"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error with color flags after pattern, got: %v\nstderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if !strings.Contains(string(data), "#ff0000") {
		t.Error("expected custom literal-fill color in SVG output")
	}
}

// ---------------------------------------------------------------------------
// --format flag tests
// ---------------------------------------------------------------------------

func TestRunFormatJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "json", "foo([a-z]+)"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !json.Valid([]byte(out)) {
		t.Fatalf("expected valid JSON, got: %s", out)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if doc["pattern"] != "foo([a-z]+)" {
		t.Errorf("expected pattern 'foo([a-z]+)', got: %v", doc["pattern"])
	}
	if doc["flavor"] != "javascript" {
		t.Errorf("expected flavor 'javascript', got: %v", doc["flavor"])
	}
	if doc["root"] == nil {
		t.Error("expected root node in JSON output")
	}
}

// TestRunTextToFileIsMarkdown verifies that the text format switches
// to Markdown output when redirected to a file via -o. This is the
// dual-mode behavior that replaced the old --format markdown.
func TestRunTextToFileIsMarkdown(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "outline.md")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "text", "-o", out, "^hello$"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "# Regex:") {
		t.Errorf("expected file to start with '# Regex:', got: %s", content[:min(50, len(content))])
	}
	if !strings.Contains(content, "**Flavor:**") {
		t.Error("expected file to contain '**Flavor:**'")
	}
}

// TestRunMarkdownFormatIsRemoved guards against accidentally reintroducing
// the legacy --format markdown option. Its role has been subsumed by
// --format text -o file.md (see TestRunTextToFileIsMarkdown).
func TestRunMarkdownFormatIsRemoved(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "markdown", "^hello$"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for removed markdown format, got nil")
	}
	if !strings.Contains(stderr.String(), "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %s", stderr.String())
	}
}

func TestRunFormatSVG(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "-o", out, "hello"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("expected SVG output file to be created")
	}
	if !strings.Contains(stdout.String(), "Wrote") {
		t.Errorf("expected stdout to contain 'Wrote', got: %s", stdout.String())
	}
}

func TestRunFormatUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "xml", "hello"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "unknown format") {
		t.Errorf("expected stderr to mention 'unknown format', got: %s", stderrStr)
	}
	if !strings.Contains(stderrStr, "Available: json, svg, text") {
		t.Errorf("expected stderr to list available formats, got: %s", stderrStr)
	}
}

// TestRunDefaultFormatIsText covers the standardized default behavior:
// bare `regolith <pattern>` prints a text walk to stdout and does not
// touch the filesystem. Previously the default was svg and the binary
// would quietly create `regex.svg` in the working directory.
func TestRunDefaultFormatIsText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--color", "never", "hello"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if out == "" {
		t.Fatal("expected text output on stdout, got empty")
	}
	if !strings.Contains(out, "Regex: hello") {
		t.Errorf("expected text banner 'Regex: hello', got: %s", out)
	}
	if _, err := os.Stat("regex.svg"); err == nil {
		t.Error("default should not write regex.svg to cwd anymore")
		_ = os.Remove("regex.svg")
	}
}

func TestRunFormatJSONNoFileCreated(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "json", "-o", out, "hello"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got: %v\nstderr: %s", err, stderr.String())
	}

	// JSON format should write to stdout, not create a file
	if _, err := os.Stat(out); err == nil {
		t.Error("expected no SVG file to be created when using --format json")
	}
}

// ---------------------------------------------------------------------------
// analyze subcommand tests
// ---------------------------------------------------------------------------

func TestAnalyzeSubcommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantOut string
	}{
		{
			name:    "basic text output",
			args:    []string{"regolith", "analyze", ".*.*=.*"},
			wantErr: false,
			wantOut: "adjacent-unbounded",
		},
		{
			name:    "json format",
			args:    []string{"regolith", "analyze", "--format", "json", ".*.*=.*"},
			wantErr: false,
			wantOut: `"id": "adjacent-unbounded"`,
		},
		{
			name:    "no pattern",
			args:    []string{"regolith", "analyze"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			err := run(tc.args, nil, &stdout, &stderr)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v\nstderr: %s", err, stderr.String())
			}
			if tc.wantOut != "" && !strings.Contains(stdout.String(), tc.wantOut) {
				t.Errorf("output missing %q\ngot: %s", tc.wantOut, stdout.String())
			}
		})
	}
}

func TestAnalyzeColorNever(t *testing.T) {
	var stdout, stderr strings.Builder
	err := run([]string{"regolith", "analyze", "--color", "never", ".*.*=.*"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stdout.String(), "\033[") {
		t.Error("expected no ANSI codes with --color never")
	}
	if !strings.Contains(stdout.String(), "ERRORS") {
		t.Error("expected ERRORS section in output")
	}
}

func TestAnalyzeColorAlways(t *testing.T) {
	var stdout, stderr strings.Builder
	err := run([]string{"regolith", "analyze", "--color", "always", ".*.*=.*"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "\033[") {
		t.Error("expected ANSI codes with --color always")
	}
}

func TestRunColorNever(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--color", "never", "-o", out, "a|b"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stdout.String(), "\033[") {
		t.Error("expected no ANSI codes with --color never")
	}
}

func TestRunColorAlways(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "--color", "always", "-o", out, "a|b"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	// "Wrote" message should have color.
	if !strings.Contains(stdout.String(), "\033[") {
		t.Error("expected ANSI codes with --color always")
	}
}

func TestRunInvalidPatternColorAlways(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--color", "always", "-o", out, "(?P<"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}

	if !strings.Contains(stderr.String(), "\033[") {
		t.Error("expected ANSI codes in error output with --color always")
	}
	if !strings.Contains(stderr.String(), "Error parsing pattern") {
		t.Error("expected error header text")
	}
}

func TestBinaryUnescapeFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "--format", "svg", "--flavor", "java", "--unescape", "-o", out, `\\d+\\.\\d+`)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("binary exited with error: %v\nstderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Fatal("SVG file was not created with -unescape flag")
	}

	if strings.Contains(stderr.String(), "Note:") {
		t.Error("expected no warning with -unescape flag")
	}
}

// ---------------------------------------------------------------------------
// New tests for standardized output behavior (text default, svg requires -o)
// ---------------------------------------------------------------------------

// TestRunTextToStdoutColorNever covers the default path — ANSI stripped
// when --color never is set, so assertions on walker content are
// deterministic.
func TestRunTextToStdoutColorNever(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "text", "--color", "never", "a|b"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes under --color never, got: %q", out)
	}
	if !strings.Contains(out, "Regex: a|b") {
		t.Errorf("expected banner 'Regex: a|b' in output, got: %s", out)
	}
	if !strings.Contains(out, "Alternation") {
		t.Errorf("expected walker to describe alternation, got: %s", out)
	}
}

// TestRunTextToStdoutColorAlways forces ANSI output so we can verify
// the walker's post-processing actually emits escape codes.
func TestRunTextToStdoutColorAlways(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "text", "--color", "always", "a|b"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[") {
		t.Errorf("expected ANSI escape codes with --color always, got: %s", stdout.String())
	}
}

// TestRunDefaultTextWithOutputWritesMarkdown verifies that even without
// --format text, the implicit default + -o routes to a Markdown file
// (since text is the default format).
func TestRunDefaultTextWithOutputWritesMarkdown(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "outline.md")

	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-o", out, "a|b"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.HasPrefix(string(data), "# Regex:") {
		t.Errorf("expected markdown header in file, got: %s", string(data)[:min(50, len(data))])
	}
}

// TestRunSVGRequiresOutput guards the behavior flip: svg no longer
// defaults to regex.svg, so --format svg without -o must error.
func TestRunSVGRequiresOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "--format", "svg", "hello"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when --format svg has no -o, got nil")
	}
	if !strings.Contains(stderr.String(), "--output") {
		t.Errorf("expected error to mention --output, got: %s", stderr.String())
	}
}

// TestAnalyzeSVGRequiresOutput confirms the shared requireOutputForSVG
// helper is wired into the analyze subcommand's svg path.
func TestAnalyzeSVGRequiresOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "analyze", "--format", "svg", "hello"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when analyze --format svg has no -o, got nil")
	}
	if !strings.Contains(stderr.String(), "--output") {
		t.Errorf("expected error to mention --output, got: %s", stderr.String())
	}
}

// TestAnalyzeSVGStyleFlags exercises the new capability promoted by
// moving svgStyleFlags to the shared struct: analyze's SVG output now
// honors --literal-fill (and its siblings) the same way the render
// command always did.
func TestAnalyzeSVGStyleFlags(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "annotated.svg")

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"regolith", "analyze",
		"--format", "svg",
		"-o", out,
		"--literal-fill", "#ff0000",
		"hello",
	}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read analyzed SVG: %v", err)
	}
	if !strings.Contains(string(data), "#ff0000") {
		t.Error("expected --literal-fill color in analyze SVG output")
	}
}
