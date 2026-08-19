package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestForthSuite runs every tests/*.fth file through the harness and fails if
// the Forth-level suite reports any error.
func TestForthSuite(t *testing.T) {
	files, err := filepath.Glob("tests/*.fth")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 2 {
		t.Fatalf("expected several test files, found %v", files)
	}
	for _, f := range files {
		if filepath.Base(f) == "harness.fth" {
			continue
		}
		t.Run(filepath.Base(f), func(t *testing.T) {
			v, buf := newTestVM(t)
			harness, err := os.ReadFile("tests/harness.fth")
			if err != nil {
				t.Fatal(err)
			}
			if err := v.Interpret(string(harness), "harness.fth"); err != nil {
				t.Fatalf("harness failed to load: %v", err)
			}
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if err := v.Interpret(string(src), f); err != nil {
				v.out.Flush()
				t.Fatalf("%s: %v\noutput:\n%s", f, err, buf.String())
			}
			if err := v.Interpret("report", f); err != nil {
				t.Fatal(err)
			}
			v.out.Flush()
			out := buf.String()
			if strings.Contains(out, "FAILED") || strings.Contains(out, "errors: 0") == false {
				t.Errorf("%s reported failures:\n%s", f, out)
			}
			if v.dsp != 0 {
				t.Errorf("%s left %d items on the stack", f, v.dsp)
			}
			if v.fsp != 0 {
				t.Errorf("%s left %d items on the float stack", f, v.fsp)
			}
		})
	}
}
