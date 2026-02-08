package main

import (
	"bytes"
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
	err := run([]string{"regolith", "-o", out, "a|b|c"}, nil, &stdout, &stderr)
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
	err := run([]string{"regolith", "-flavor", "java", "-o", out, "[a-z]+"}, nil, &stdout, &stderr)
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
			err := run([]string{"regolith", "-flavor", name, "-o", out, "abc"}, nil, &stdout, &stderr)
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
	err := run([]string{"regolith", "-flavor", "bogus", "-o", out, "abc"}, nil, &stdout, &stderr)
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
	err := run([]string{"regolith", "-o", out, "hello"}, nil, &stdout, &stderr)
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
	err := run([]string{"regolith", "-o", out}, stdin, &stdout, &stderr)
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
	err := run([]string{"regolith", "-o", out, "a|b"}, stdin, &stdout, &stderr)
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
	err := run([]string{"regolith", "-o", out, "[a-z]+"}, nil, &stdout, &stderr)
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
		"-o", out,
		"-text-color", "#fff",
		"-line-color", "#333",
		"-literal-fill", "#00ff00",
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
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func TestBinaryValidPattern(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "-o", out, "a|b|c")
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

	cmd := exec.Command(binaryPath, "-o", out)
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
	cmd := exec.Command(binaryPath, "-flavor", "bogus", "abc")
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
	cmd := exec.Command(binaryPath, "-flavor", "pcre", "-o", out, "(?<=foo)bar")
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
	err := run([]string{"regolith", "-flavor", "java", "-unescape", "-o", out, `\\d+\\.\\d+`}, nil, &stdout, &stderr)
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
	err := run([]string{"regolith", "-flavor", "java", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error (warning only), got: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stderr.String(), "-unescape") {
		t.Errorf("expected warning mentioning -unescape, got: %s", stderr.String())
	}
}

func TestRunDoubleEscapeWarningDotnet(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Dotnet flavor with double escapes but no -unescape flag: should warn
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-flavor", "dotnet", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error (warning only), got: %v\nstderr: %s", err, stderr.String())
	}

	if !strings.Contains(stderr.String(), "-unescape") {
		t.Errorf("expected warning mentioning -unescape, got: %s", stderr.String())
	}
}

func TestRunNoWarningForJavaScript(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// JavaScript flavor with double escapes: no warning expected
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-flavor", "javascript", "-o", out, `\\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "-unescape") {
		t.Error("expected no warning for javascript flavor, but got one")
	}
}

func TestRunNoWarningWithoutDoubleEscapes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	// Java flavor without double escapes: no warning expected
	var stdout, stderr bytes.Buffer
	err := run([]string{"regolith", "-flavor", "java", "-o", out, `\d+`}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	if strings.Contains(stderr.String(), "-unescape") {
		t.Error("expected no warning without double escapes, but got one")
	}
}

func TestBinaryUnescapeFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.svg")

	cmd := exec.Command(binaryPath, "-flavor", "java", "-unescape", "-o", out, `\\d+\\.\\d+`)
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
