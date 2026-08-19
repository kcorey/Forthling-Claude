package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildForth compiles the interpreter once per test run.
func buildForth(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping build-based test in short mode")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "forth")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return bin
}

func TestScriptMode(t *testing.T) {
	bin := buildForth(t)
	script := filepath.Join(t.TempDir(), "prog.fth")
	if err := os.WriteFile(script, []byte(`: main 2 3 + . ." done" cr ; main`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, script).CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if got := string(out); got != "5 done\n" {
		t.Errorf("script mode gave %q", got)
	}
}

func TestScriptModeArguments(t *testing.T) {
	bin := buildForth(t)
	script := filepath.Join(t.TempDir(), "args.fth")
	if err := os.WriteFile(script, []byte(`argc . 0 arg type cr`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, script, "hello", "world").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if got := string(out); got != "2 hello\n" {
		t.Errorf("argument passing gave %q", got)
	}
}

func TestScriptModeError(t *testing.T) {
	bin := buildForth(t)
	script := filepath.Join(t.TempDir(), "bad.fth")
	if err := os.WriteFile(script, []byte(`1 nosuchword`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, script).CombinedOutput()
	if err == nil {
		t.Fatalf("expected a non-zero exit, output %q", out)
	}
	if !strings.Contains(string(out), "undefined word") {
		t.Errorf("error output was %q", out)
	}
}

func TestEvalMode(t *testing.T) {
	bin := buildForth(t)
	out, err := exec.Command(bin, "-e", "7 6 * .").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if got := string(out); got != "42 " {
		t.Errorf("-e gave %q", got)
	}
}

func TestReplMode(t *testing.T) {
	bin := buildForth(t)
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("1 2 + .\nbye\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	got := string(out)
	if !strings.Contains(got, "Forthling-Claude") {
		t.Errorf("no banner in %q", got)
	}
	if !strings.Contains(got, "3 ") || !strings.Contains(got, "ok") {
		t.Errorf("repl output was %q", got)
	}
}

func TestReplRecoversFromErrors(t *testing.T) {
	bin := buildForth(t)
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("nosuchword\n1 2 + .\nbye\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	got := string(out)
	if !strings.Contains(got, "undefined word") || !strings.Contains(got, "3 ") {
		t.Errorf("repl did not recover: %q", got)
	}
}

func TestBundleMode(t *testing.T) {
	bin := buildForth(t)
	dir := t.TempDir()
	script := filepath.Join(dir, "greet.fth")
	body := `: main ." bundled says " argc . cr ; main`
	if err := os.WriteFile(script, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	outBin := filepath.Join(dir, "greet")

	if out, err := exec.Command(bin, "-bundle", script, "-o", outBin).CombinedOutput(); err != nil {
		t.Fatalf("bundle failed: %v\n%s", err, out)
	}

	// The bundled binary runs the script instead of the REPL.
	out, err := exec.Command(outBin, "a", "b", "c").CombinedOutput()
	if err != nil {
		t.Fatalf("bundled binary failed: %v\n%s", err, out)
	}
	if got := string(out); got != "bundled says 3 \n" {
		t.Errorf("bundled binary gave %q", got)
	}

	// The payload is discoverable and matches the script.
	p, err := readPayloadFrom(outBin)
	if err != nil || p == nil {
		t.Fatalf("payload not found: %v", err)
	}
	if string(p.script) != body {
		t.Errorf("payload mismatch: %q", p.script)
	}

	// Bundling from an already-bundled binary replaces the payload instead of
	// nesting it, so the interpreter part stays exactly one copy long.
	body2 := `." second" cr`
	outBin2 := filepath.Join(dir, "second")
	if err := writeBundleFrom(outBin, []byte(body2), outBin2); err != nil {
		t.Fatalf("re-bundle failed: %v", err)
	}
	out, err = exec.Command(outBin2).CombinedOutput()
	if err != nil {
		t.Fatalf("second bundle failed: %v\n%s", err, out)
	}
	if got := string(out); got != "second\n" {
		t.Errorf("second bundle gave %q", got)
	}
	p2, err := readPayloadFrom(outBin2)
	if err != nil || p2 == nil {
		t.Fatalf("second payload not found: %v", err)
	}
	if p2.coreLen != p.coreLen {
		t.Errorf("interpreter section grew: %d -> %d", p.coreLen, p2.coreLen)
	}

	// The interpreter itself still has no payload.
	if p, err := readPayloadFrom(bin); err != nil || p != nil {
		t.Errorf("plain interpreter reports a payload: %v %v", p, err)
	}
}

func TestBundleDetectionIgnoresOtherFiles(t *testing.T) {
	f := filepath.Join(t.TempDir(), "random.bin")
	if err := os.WriteFile(f, []byte("just some bytes, definitely not a bundle"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := readPayloadFrom(f)
	if err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Error("found a payload where there is none")
	}
}

// TestSamplesRun exercises every sample end to end.
func TestSamplesRun(t *testing.T) {
	bin := buildForth(t)
	cases := []struct {
		file     string
		args     []string
		stdin    string
		contains string
	}{
		{"samples/sieve.fth", nil, "", "primes found"},
		{"samples/life.fth", []string{"3"}, "", "generation 3"},
		{"samples/mandelbrot.fth", nil, "q", "48;5;"},
		{"samples/snake.fth", nil, "q", "game over"},
		{"samples/tictactoe.fth", nil, "5\n1\nq\n", "thinking"},
		{"samples/hammurabi.fth", nil, "q\n", "HAMMURABI"},
		{"samples/lander.fth", nil, "q", "mission aborted"},
		{"samples/lander.fth", []string{"auto"}, "", "TOUCHDOWN"},
	}
	for _, c := range cases {
		name := filepath.Base(c.file)
		if len(c.args) > 0 {
			name += "-" + c.args[0]
		}
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(bin, append([]string{c.file}, c.args...)...)
			cmd.Stdin = strings.NewReader(c.stdin)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", c.file, err, truncate(string(out)))
			}
			if !strings.Contains(string(out), c.contains) {
				t.Errorf("%s output missing %q:\n%s", c.file, c.contains, truncate(string(out)))
			}
		})
	}
}

func truncate(s string) string {
	if len(s) > 2000 {
		return s[:2000] + "..."
	}
	return s
}
