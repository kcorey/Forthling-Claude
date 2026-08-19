# Forthling-Claude

A small, fast Forth written in Go. One binary, three ways to run it, and no
third-party dependencies.

```
forth                            interactive REPL
forth FILE.fth [args...]         run a Forth script
forth -e "2 3 + ."               run code from the command line
forth -bundle FILE.fth [-o OUT]  write a standalone binary that runs FILE.fth
```

## Why it exists

This repo is one arm of an experiment: can a coding agent build a *real*
language, not a toy, from a single informal brief?

The same job was given to Qwen3.8 running under OpenCode, first on a Mac and
then on a PC. Both attempts hit repeated failures across two days and still
have not finished the job. This version was built with Claude Code (Opus): it
took about an hour and worked after a couple of prompts. The
[prompt that started it](#the-prompt) is below, so you can run the comparison
on your own setup.

Forth is a good test subject. It is the smallest interesting language that is
still genuinely useful: a stack, a dictionary, and a compiler you can extend
from inside the language. Getting one right means getting an inner interpreter,
a compiler, terminal I/O and a test suite all working together. The result here
is built around three goals.

**Small enough to read.** About 2,500 lines of Go and one page of Forth. The
whole language, the compiler, the terminal handling and the bundler fit in a
handful of files you can read in an afternoon.

**Fast enough to use.** Colon definitions compile to a flat token-threaded code
array, so there are no allocations and no dictionary lookups on the hot path. A
200,000-element prime sieve runs in about 15 ms; a Mandelbrot frame of 2,340
pixels at 240 iterations takes about 75 ms.

**Shippable.** `-bundle` turns a script into one self-contained executable you
can hand to somebody. No runtime to install, no Go toolchain needed at bundling
time, no interpreter to ship alongside it.

The `samples/` directory is the point of the exercise: a colour Mandelbrot
explorer, a real-time lunar lander, a perfect-play tic-tac-toe, Conway's Life,
Hammurabi, snake, and a prime sieve, all written in the language itself.

## The prompt

This is the brief, near enough as it was given. Paste it into a coding agent in
an empty directory and see what comes back.

> In this directory, build a version of Forth written in Go.
>
> It should have three modes. Run `forth` by itself and you get a REPL. Run
> `forth` with a filename and it executes that Forth script. Run
> `forth -bundle <filename>` and it bundles that script into a Forth binary, so
> that running the binary executes the script. That way I can share a single
> binary with the new code in it.
>
> I want a samples directory with several good demos: a Mandelbrot viewer that
> lets you zoom and rotate, tic-tac-toe, Hammurabi, that sort of thing. Enough
> to show the language off.
>
> Keep the language fairly minimal but make it fast. It is an optimisation
> problem: enough Forth words to carry the samples, but still a smallish Forth.
>
> This is a language, so I want an extensive unit test suite: every word in the
> language tested and validated as working.

The lunar lander, the momentum arrows, the window-filling layout and the
break-up animation came from a handful of follow-up prompts after that.

## Build and test

```sh
go build -o forth .      # or: make
go test ./...            # unit tests, the Forth-level suite, and sample smoke tests
make bundles             # a standalone binary per sample, into ./dist
```

Requires Go 1.24 or newer. There is nothing else to install: `go.mod` has no
dependencies.

## The three modes

**REPL.** `./forth` gives you a prompt. `WORDS` lists the dictionary, `BYE`
leaves. Definitions can span lines (the prompt changes to `...`), and control
structures work straight from the prompt, not just inside definitions:

```
> 10 0 do i . loop
0 1 2 3 4 5 6 7 8 9  ok
> : hypot ( fx fy -- fh ) fsq fswap fsq f+ fsqrt ;  ok
> 3.0 4.0 hypot f.
5  ok
```

An error prints a message, clears the stacks, abandons any half-finished
definition, and hands the prompt back. You cannot wedge the REPL.

**Script.** `./forth samples/sieve.fth` interprets a file and exits. Arguments
after the filename reach the program through `ARGC` and `ARG`:

```forth
argc 0> if 0 arg type cr then
```

The exit status is 0 on success, 1 if the script died with an error (the error
goes to stderr), 2 for a command line mistake.

**Bundle.** `./forth -bundle samples/lander.fth -o lander` writes `lander`: a
copy of the interpreter with the script appended plus a 16-byte trailer.

```
[ interpreter bytes ][ script bytes ][ length uint64 LE ][ "FTHBUNDL" ]
```

At startup every binary reads its own tail; if the magic is there it runs the
embedded script instead of starting the REPL. Bundling is two file copies, so
it takes milliseconds and needs no Go toolchain on the machine doing it. The
result is about 2.7 MB and runs anywhere the interpreter itself would.

## The language

64-bit integer cells with a **separate floating-point stack** (a subset of the
ANS FLOATING word set), byte-addressed data space, floored division, and a
case-insensitive dictionary. 255 words in all: about 210 Go primitives plus
`kernel.fth`, which is embedded in the binary and adds the rest in Forth.
`WORDS` prints the lot.

| Group | Words |
|---|---|
| Stack | `dup ?dup drop swap over nip tuck rot -rot 2dup 2drop 2swap 2over 3dup pick roll depth .s` |
| Return stack | `>r r> r@ 2>r 2r> rdepth` |
| Arithmetic | `+ - * / mod /mod */ negate abs min max 1+ 1- 2* 2/ lshift rshift and or xor invert sq 0max` |
| Comparison | `= <> < > <= >= u< u> 0= 0<> 0< 0> 0<= 0>= within between` |
| Memory | `@ ! c@ c! +! f@ f! , c, f, here allot align cells cell+ chars char+ move fill erase count pad on off incr decr ? array [] barray b[]` |
| Floats | `f+ f- f* f/ fnegate fabs fsqrt fsin fcos ftan fatan fatan2 fexp fln flog f** fmin fmax fmod floor fround f< f> f= f<= f>= f0< f0= fdup fdrop fswap fover fnip frot fdepth s>f f>s f. fe. f.r f.s f2* f2/ fsq pi e deg>rad` |
| Control | `if else then begin until while repeat again do ?do loop +loop leave unloop i j exit recurse case of endof endcase` |
| Defining | `: ; immediate create variable fvariable constant fconstant does> 2constant ' ['] execute postpone compile, literal fliteral [ ]` |
| Strings & I/O | `." s" c" .( type emit cr space spaces . u. .r words key key? accept word number char [char] bl str= tab star stars dots` |
| Terminal | `page at-xy term-size raw-on raw-off cursor-on cursor-off fg bg normal bright esc[ flush` |
| System | `base decimal hex ms ticks random randomize seed abort abort" bye quit include included evaluate argc arg` |

Number literals honour `BASE` and accept the `$` (hex), `#` (decimal), `%`
(binary) and `'c'` (character) prefixes. A token containing `.`, `e` or `E`
that parses as a float (`1.5`, `-2.0e3`, `1e-4`) goes on the **float** stack,
not the data stack.

A few words are extensions rather than standard Forth: `term-size` returns the
window as `( -- cols rows )`, `ticks` is milliseconds since start, `random` and
`seed` drive a xorshift generator, `raw-on`/`raw-off` switch the terminal to
character-at-a-time input, and `fg`/`bg` take 256-colour palette indices.

## Samples

| File | What it shows |
|---|---|
| `samples/lander.fth` | real-time lunar lander in colour, sized to the terminal window: gravity, a flickering exhaust plume, momentum arrows, scored landing pads, and a break-up animation when you get it wrong. `lander.fth auto` flies itself. |
| `samples/mandelbrot.fth` | colour Mandelbrot explorer: pan, zoom, **rotate**, iteration control. Floats and raw-mode keys. |
| `samples/tictactoe.fth` | perfect-play minimax opponent. Recursion and arrays. |
| `samples/hammurabi.fth` | the 1968 resource-management game. Line input, parsing and randomness. |
| `samples/life.fth` | Conway's Life on a torus with a Gosper glider gun. |
| `samples/snake.fth` | arcade loop driven by non-blocking `KEY?`. |
| `samples/sieve.fth` | prime sieve, also the speed benchmark. |

```sh
./forth samples/lander.fth            # a/d steer, w or space burns, q quits
./forth samples/lander.fth auto       # watch the autopilot land it
./forth samples/mandelbrot.fth        # wasd pan, +/- zoom, [ ] rotate, q quit
./forth samples/life.fth 200
./forth -bundle samples/snake.fth -o snake && ./snake
```

## Portability

The interpreter is pure Go with no cgo and no assembly, so it is not tied to
Apple silicon or to macOS. It was developed on an M-series MacBook Pro and
cross-compiles cleanly for Intel Macs and everything else:

```sh
GOOS=darwin  GOARCH=amd64 go build -o forth-intel-mac .
GOOS=linux   GOARCH=amd64 go build -o forth-linux .
GOOS=windows GOARCH=amd64 go build -o forth.exe .
```

`darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `linux/386`,
`windows/amd64` and `freebsd/amd64` are all verified to build.

What differs by platform:

- **Intel Macs and Apple silicon behave identically.** Nothing in the code
  depends on the word size or the byte order of the host: cells are always
  64-bit and data space is always little-endian, so `@` and `c@` see the same
  bytes everywhere, including in a 32-bit build.
- **macOS and Linux** get full raw-mode keyboard input and window-size
  detection. FreeBSD and anything else falls back to line-buffered input and a
  default 80x24.
- **Windows** compiles and runs the REPL, scripts and bundles, but `raw-on` is
  a no-op, so `key` and `key?` wait for Enter. The real-time samples (lander,
  snake, mandelbrot) will not play properly there. Colour needs a terminal with
  VT sequences enabled, which Windows Terminal has by default.

## Things to watch out for

**A bundle is platform-specific.** `-bundle` copies the *running* interpreter,
so a bundle made on an Apple silicon Mac is an arm64 macOS executable. To ship
a script to Linux, cross-build the interpreter for Linux and bundle from there
(or on that machine). There is no cross-bundling flag.

**A bundled binary cannot re-bundle.** It runs its payload instead of parsing
`-bundle`, because its arguments belong to the bundled program. Always bundle
from the plain `forth` binary.

**macOS code signing.** Appending the payload invalidates the ad-hoc signature
Go's linker applies, so `codesign -v` reports "failed strict validation" for a
bundle. It still runs locally, and re-signing does not fix the report. If you
send one to somebody else the recipient may have to clear the quarantine
attribute (`xattr -d com.apple.quarantine ./thing`), and bundles cannot be
notarised as they stand.

**Do not post-process a bundle.** Stripping, compressing (UPX) or otherwise
rewriting the file can drop the trailing payload, and the binary will silently
revert to being a plain REPL.

**Errors clear both stacks.** Recovery is deliberately brutal: on any error the
data stack, float stack and return stack are emptied and compilation state is
reset. That keeps the REPL usable, but you lose whatever you were holding.

**Defining words inside interpret-time control flow does not work.** `10 0 do
… loop` at the prompt is compiled into a scratch definition and run at the
closing `loop`, which is why it works at all; but `1 if 5 constant x then`
fails, because `x` is compiled rather than executed when the line is read. This
matches standard Forth. Keep defining words at the top level, or inside a
definition of their own.

**Redefinition is last-wins, and it is not retroactive.** `: t 1 ; : u t ; : t
2 ;` leaves `u` calling the *first* `t`, because `u` compiled a reference to
it. Inside a redefinition the old word is still visible (`: t t 2 + ;` works),
which is how you extend a word.

**Fixed stack sizes.** 4,096 data cells, 4,096 return cells, 1,024 float cells.
Runaway recursion gives a clean "return stack overflow" error rather than a
crash, but it does end the run.

**No `forget`, no `see`, no double-cell arithmetic.** The word set is
deliberately minimal. `*/` covers the common double-precision intermediate
case; there is no `d+`, `m*` or `um/mod`.

**Real-time samples need a real terminal.** `key?` returns false forever when
stdin is a pipe, and `term-size` falls back to `$COLUMNS`/`$LINES` and then to
80x24. That is what makes the samples testable in CI, but it also means piping
input into them will not do what you expect.

**`ms` sleeps and flushes.** Output is buffered for speed; input words and `ms`
flush it for you, but if you write a tight animation loop with no input and no
delay, call `flush` yourself.

## Implementation

- `vm.go`: stacks, data space, dictionary, inner interpreter
- `compile.go`: outer interpreter, number parsing, all compiling words
- `prims_core.go` / `prims_float.go` / `prims_io.go`: primitives
- `bundle.go`: the self-append bundler
- `term_*.go`: raw mode via `termios` and window size via `TIOCGWINSZ`,
  build-tagged per platform
- `kernel.fth`: the Forth-level part of the system, `go:embed`ed
- `repl.go`, `dict.go`, `main.go`: prompt, dictionary, mode dispatch

Colon definitions compile into one flat `[]int64` code array; each cell is an
execution token, and the inner interpreter is a switch over three word kinds
(primitive, colon, data). Literals and branch targets are inline cells.

Errors (`undefined word`, stack underflow, bad address, `ABORT"`) unwind
through a Go panic and are recovered at the outer interpreter, which resets the
machine and returns an ordinary Go `error`.

## Tests

`go test ./...` runs four layers:

- `forth_test.go`: every word group, error cases, number parsing, recovery
- `tests/*.fth` with `tests/harness.fth`: 176 assertions in an ANS-style
  `T{ … -> … }T` suite (`FT{ … F-> … }FT` for floats), so the language is
  tested in itself
- `modes_test.go`: REPL, script, `-e` and bundle modes end to end, plus a smoke
  test for every sample
- `bench_test.go`: sieve, recursive fib and a float loop

Benchmarks on an M4 Pro: a sieve to 100,000 in 5.8 ms, `fib(22)` in 1.4 ms, and
1,000 iterations of the Mandelbrot inner loop in 100 µs.

## Licence

MIT. See [LICENSE](LICENSE).
