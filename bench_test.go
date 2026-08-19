package main

import (
	"io"
	"testing"
)

func benchVM(b *testing.B, setup string) *VM {
	b.Helper()
	v := NewVM()
	v.SetOutput(io.Discard)
	if err := v.Interpret(kernelSource, "kernel.fth"); err != nil {
		b.Fatal(err)
	}
	if setup != "" {
		if err := v.Interpret(setup, "setup"); err != nil {
			b.Fatal(err)
		}
	}
	return v
}

// BenchmarkSieve measures the inner interpreter on a memory-heavy workload.
func BenchmarkSieve(b *testing.B) {
	v := benchVM(b, `
100000 constant limit
create flags limit allot
variable p
: mark dup p ! dup * dup limit < if limit swap ?do 0 flags i + c! p @ +loop else drop then ;
: sieve flags limit 1 fill limit 2 do flags i + c@ if i mark then loop ;
`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Interpret("sieve", "bench"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFib measures call and return overhead.
func BenchmarkFib(b *testing.B) {
	v := benchVM(b, `: fib dup 2 < if exit then dup 1- recurse swap 2 - recurse + ;`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Interpret("22 fib drop", "bench"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFloatLoop measures the float stack in a Mandelbrot-shaped loop.
func BenchmarkFloatLoop(b *testing.B) {
	v := benchVM(b, `
fvariable zx fvariable zy fvariable t
: iterate 0.0 zx f! 0.0 zy f!
  1000 0 do
    zx f@ fsq zy f@ fsq f- -0.5 f+ t f!
    zx f@ zy f@ f* f2* 0.5 f+ zy f!
    t f@ zx f!
  loop ;
`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Interpret("iterate", "bench"); err != nil {
			b.Fatal(err)
		}
	}
}
