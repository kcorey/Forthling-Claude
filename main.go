package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed kernel.fth
var kernelSource string

const usage = `Forthling-Claude - a small fast Forth in Go

  forth                          start the interactive REPL
  forth FILE.fth [args...]       run a Forth script
  forth -e "CODE"                run Forth code from the command line
  forth -bundle FILE.fth [-o OUT]
                                 write a standalone binary that runs FILE.fth
                                 (default OUT is the script name without .fth)
`

func main() {
	os.Exit(run())
}

func run() int {
	installSignalHandler()
	defer restoreTerminal()

	// Mode 3b: this binary carries an embedded script.
	if p := readPayload(); p != nil {
		return runSource(string(p.script), "<bundled>", os.Args[1:])
	}

	args := os.Args[1:]
	if len(args) == 0 {
		return startREPL()
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	case "-e", "--eval":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "forth: -e needs code")
			return 2
		}
		return runSource(args[1], "-e", args[2:])
	case "-bundle", "--bundle":
		return doBundle(args[1:])
	}

	if strings.HasPrefix(args[0], "-") {
		fmt.Fprintf(os.Stderr, "forth: unknown option %s\n", args[0])
		return 2
	}

	src, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "forth: %v\n", err)
		return 1
	}
	return runSource(string(src), args[0], args[1:])
}

// newLoadedVM builds a VM with the kernel already interpreted.
func newLoadedVM(args []string) (*VM, error) {
	v := NewVM()
	v.args = args
	if err := v.Interpret(kernelSource, "kernel.fth"); err != nil {
		return nil, fmt.Errorf("kernel.fth: %w", err)
	}
	return v, nil
}

func runSource(src, name string, args []string) int {
	v, err := newLoadedVM(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer v.out.Flush()
	err = v.Interpret(src, name)
	v.out.Flush()
	if err != nil {
		if b, ok := err.(byeError); ok {
			return b.code
		}
		fmt.Fprintf(os.Stderr, "\nforth: %s: %v\n", name, err)
		return 1
	}
	return 0
}

func startREPL() int {
	v, err := newLoadedVM(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer v.out.Flush()
	return v.repl()
}

func doBundle(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "forth: -bundle needs a script")
		return 2
	}
	script := args[0]
	out := strings.TrimSuffix(filepath.Base(script), filepath.Ext(script))
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == "-o" || rest[i] == "--out" {
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, "forth: -o needs a filename")
				return 2
			}
			out = rest[i+1]
			i++
			continue
		}
		fmt.Fprintf(os.Stderr, "forth: unknown bundle option %s\n", rest[i])
		return 2
	}
	if out == "" {
		out = "a.out"
	}

	data, err := os.ReadFile(script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forth: %v\n", err)
		return 1
	}
	// Fail early on a script that does not even parse into words we know.
	if err := writeBundle(data, out); err != nil {
		fmt.Fprintf(os.Stderr, "forth: bundle: %v\n", err)
		return 1
	}
	fmt.Printf("wrote %s\n", out)
	return 0
}

func installSignalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		restoreTerminal()
		os.Exit(130)
	}()
}
