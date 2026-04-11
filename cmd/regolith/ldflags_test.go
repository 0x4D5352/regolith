package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestLdflagsDoNotAffectSVGOutput verifies that building with
// -ldflags "-X main.version=..." produces identical SVG output
// to a plain build. This guards against a regression where ldflags
// unexpectedly broke SVG rendering (GitHub issue #1).
func TestLdflagsDoNotAffectSVGOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: test shells out to go build")
	}

	// Locate the module root so `go build` resolves the package.
	modDir, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err != nil {
		t.Fatalf("finding module root: %v", err)
	}
	root := string(modDir[:len(modDir)-1]) // trim trailing newline
	pkg := filepath.Join(root, "cmd", "regolith")

	dir := t.TempDir()
	plainBin := filepath.Join(dir, "regolith-plain")
	ldflagBin := filepath.Join(dir, "regolith-ldflag")

	// Build without ldflags.
	build := exec.Command("go", "build", "-o", plainBin, pkg)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("plain build failed: %v\n%s", err, out)
	}

	// Build with ldflags that set main.version.
	build = exec.Command("go", "build",
		"-ldflags", "-X main.version=ldflags-test",
		"-o", ldflagBin, pkg)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("ldflag build failed: %v\n%s", err, out)
	}

	// Run both binaries with the same pattern and compare SVG output.
	pattern := `(?i)\p{Alpha}+\d{2,}`
	plainSVG := filepath.Join(dir, "plain.svg")
	ldflagSVG := filepath.Join(dir, "ldflag.svg")

	run := exec.Command(plainBin, "--format", "svg", "-f", "java", "-o", plainSVG, pattern)
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("plain binary run failed: %v\n%s", err, out)
	}

	run = exec.Command(ldflagBin, "--format", "svg", "-f", "java", "-o", ldflagSVG, pattern)
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("ldflag binary run failed: %v\n%s", err, out)
	}

	plain, err := os.ReadFile(plainSVG)
	if err != nil {
		t.Fatalf("reading plain SVG: %v", err)
	}
	ldflag, err := os.ReadFile(ldflagSVG)
	if err != nil {
		t.Fatalf("reading ldflag SVG: %v", err)
	}

	if string(plain) != string(ldflag) {
		t.Error("SVG output differs between plain and ldflag builds — ldflags affect rendering")
	}
}
