package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// newTestVM builds a VM with kernel.fth loaded and output captured.
func newTestVM(t *testing.T) (*VM, *bytes.Buffer) {
	t.Helper()
	v := NewVM()
	buf := &bytes.Buffer{}
	v.SetOutput(buf)
	if err := v.Interpret(kernelSource, "kernel.fth"); err != nil {
		t.Fatalf("kernel.fth failed to load: %v", err)
	}
	return v, buf
}

// eval runs src and returns the data stack and everything printed.
func eval(t *testing.T, src string) ([]int64, string) {
	t.Helper()
	v, buf := newTestVM(t)
	if err := v.Interpret(src, "test"); err != nil {
		t.Fatalf("%s\n  error: %v", src, err)
	}
	v.out.Flush()
	return append([]int64(nil), v.ds[:v.dsp]...), buf.String()
}

// evalErr runs src expecting a failure, and returns the error message.
func evalErr(t *testing.T, src string) string {
	t.Helper()
	v, _ := newTestVM(t)
	err := v.Interpret(src, "test")
	if err == nil {
		t.Fatalf("expected an error from %q, got none", src)
	}
	return err.Error()
}

// evalF runs src and returns the float stack.
func evalF(t *testing.T, src string) []float64 {
	t.Helper()
	v, _ := newTestVM(t)
	if err := v.Interpret(src, "test"); err != nil {
		t.Fatalf("%s\n  error: %v", src, err)
	}
	v.out.Flush()
	return append([]float64(nil), v.fs[:v.fsp]...)
}

type stackCase struct {
	src  string
	want []int64
}

func runStackCases(t *testing.T, cases []stackCase) {
	t.Helper()
	for _, c := range cases {
		got, _ := eval(t, c.src)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%-40s got %v want %v", c.src, got, c.want)
		}
	}
}

func TestStackWords(t *testing.T) {
	runStackCases(t, []stackCase{
		{"1 2 3", []int64{1, 2, 3}},
		{"1 dup", []int64{1, 1}},
		{"0 ?dup", []int64{0}},
		{"5 ?dup", []int64{5, 5}},
		{"1 2 drop", []int64{1}},
		{"1 2 swap", []int64{2, 1}},
		{"1 2 over", []int64{1, 2, 1}},
		{"1 2 nip", []int64{2}},
		{"1 2 tuck", []int64{2, 1, 2}},
		{"1 2 3 rot", []int64{2, 3, 1}},
		{"1 2 3 -rot", []int64{3, 1, 2}},
		{"1 2 2dup", []int64{1, 2, 1, 2}},
		{"1 2 3 4 2drop", []int64{1, 2}},
		{"1 2 3 4 2swap", []int64{3, 4, 1, 2}},
		{"1 2 3 4 2over", []int64{1, 2, 3, 4, 1, 2}},
		{"1 2 3 0 pick", []int64{1, 2, 3, 3}},
		{"1 2 3 2 pick", []int64{1, 2, 3, 1}},
		{"1 2 3 2 roll", []int64{2, 3, 1}},
		{"1 2 3 depth", []int64{1, 2, 3, 3}},
		{"1 2 3 3dup", []int64{1, 2, 3, 1, 2, 3}},
	})
}

func TestReturnStackWords(t *testing.T) {
	runStackCases(t, []stackCase{
		{": t 5 >r 1 r> ; t", []int64{1, 5}},
		{": t 7 >r r@ r> ; t", []int64{7, 7}},
		{": t 1 2 2>r 9 2r> ; t", []int64{9, 1, 2}},
	})
}

func TestArithmetic(t *testing.T) {
	runStackCases(t, []stackCase{
		{"2 3 +", []int64{5}},
		{"10 3 -", []int64{7}},
		{"6 7 *", []int64{42}},
		{"20 5 /", []int64{4}},
		{"-7 2 /", []int64{-4}},  // floored division
		{"-7 2 mod", []int64{1}}, // floored modulo
		{"17 5 /mod", []int64{2, 3}},
		{"10 3 2 */", []int64{15}},
		{"5 negate", []int64{-5}},
		{"-9 abs", []int64{9}},
		{"3 7 min", []int64{3}},
		{"3 7 max", []int64{7}},
		{"5 1+", []int64{6}},
		{"5 1-", []int64{4}},
		{"5 2*", []int64{10}},
		{"5 2/", []int64{2}},
		{"1 4 lshift", []int64{16}},
		{"16 4 rshift", []int64{1}},
		{"12 10 and", []int64{8}},
		{"12 10 or", []int64{14}},
		{"12 10 xor", []int64{6}},
		{"0 invert", []int64{-1}},
		{"4 sq", []int64{16}},
		{"-3 0max", []int64{0}},
	})
}

func TestComparison(t *testing.T) {
	runStackCases(t, []stackCase{
		{"3 3 =", []int64{-1}},
		{"3 4 =", []int64{0}},
		{"3 4 <>", []int64{-1}},
		{"3 4 <", []int64{-1}},
		{"4 3 >", []int64{-1}},
		{"3 3 <=", []int64{-1}},
		{"3 3 >=", []int64{-1}},
		{"1 -1 u<", []int64{-1}},
		{"-1 1 u>", []int64{-1}},
		{"0 0=", []int64{-1}},
		{"1 0<>", []int64{-1}},
		{"-1 0<", []int64{-1}},
		{"1 0>", []int64{-1}},
		{"true", []int64{-1}},
		{"false", []int64{0}},
		{"5 1 10 within", []int64{-1}},
		{"10 1 10 within", []int64{0}},
		{"10 1 10 between", []int64{-1}},
	})
}

func TestMemory(t *testing.T) {
	runStackCases(t, []stackCase{
		{"variable v 42 v ! v @", []int64{42}},
		{"variable v 1 v ! 5 v +! v @", []int64{6}},
		{"create b 3 c, 4 c, b c@ b 1+ c@", []int64{3, 4}},
		{"create c 7 , 8 , c @ c cell+ @", []int64{7, 8}},
		{"here 8 allot here swap -", []int64{8}},
		{"3 cells", []int64{24}},
		{"1 cell+", []int64{9}},
		{"5 chars", []int64{5}},
		{"5 char+", []int64{6}},
		{"variable v v on v @", []int64{-1}},
		{"variable v v on v off v @", []int64{0}},
		{"variable v 1 v ! v incr v @", []int64{2}},
		{"variable v 1 v ! v decr v @", []int64{0}},
	})

	// counted string address check, separate because the address is dynamic
	v, buf := newTestVM(t)
	if err := v.Interpret(`create s 3 c, char A c, char B c, char C c, s count type`, "t"); err != nil {
		t.Fatal(err)
	}
	v.out.Flush()
	if buf.String() != "ABC" {
		t.Errorf("count/type gave %q", buf.String())
	}
}

func TestMemoryBlockOps(t *testing.T) {
	v, buf := newTestVM(t)
	src := `
create buf 16 allot
buf 16 char x fill
buf 5 type cr
buf 16 erase
buf c@ .
s" hello" pad swap move
pad 5 type
`
	if err := v.Interpret(src, "t"); err != nil {
		t.Fatal(err)
	}
	v.out.Flush()
	want := "xxxxx\n0 hello"
	if buf.String() != want {
		t.Errorf("got %q want %q", buf.String(), want)
	}
}

func TestControlFlow(t *testing.T) {
	runStackCases(t, []stackCase{
		{": t 1 if 42 then ; t", []int64{42}},
		{": t 0 if 42 then ; t", nil},
		{": t 0 if 1 else 2 then ; t", []int64{2}},
		{": t 0 begin 1+ dup 5 = until ; t", []int64{5}},
		{": t 0 begin dup 5 < while 1+ repeat ; t", []int64{5}},
		{": t 0 5 0 do i + loop ; t", []int64{10}},
		{": t 0 10 0 do i + 2 +loop ; t", []int64{20}},
		{": t 0 10 0 do 1+ dup 3 = if leave then loop ; t", []int64{3}},
		{": t 0 5 5 ?do 1+ loop ; t", []int64{0}},
		{": t 0 5 0 ?do 1+ loop ; t", []int64{5}},
		{": t 0 3 0 do 3 0 do i j + + loop loop ; t", []int64{18}},
		{": t 2 case 1 of 10 endof 2 of 20 endof 99 swap endcase ; t", []int64{20}},
		{": t 7 case 1 of 10 endof 2 of 20 endof 99 swap endcase ; t", []int64{99}},
		{": t 1 exit 2 ; t", []int64{1}},
		// interpret-time control flow (implicit definitions)
		{"0 5 0 do i + loop", []int64{10}},
		{"1 if 42 then", []int64{42}},
		{"0 begin 1+ dup 3 = until", []int64{3}},
	})
}

func TestRecursion(t *testing.T) {
	runStackCases(t, []stackCase{
		{": fact dup 1 > if dup 1- recurse * then ; 5 fact", []int64{120}},
		{": fib dup 2 < if exit then dup 1- recurse swap 2 - recurse + ; 10 fib", []int64{55}},
	})
}

func TestDefiningWords(t *testing.T) {
	runStackCases(t, []stackCase{
		{"7 constant seven seven seven +", []int64{14}},
		{"variable v v @", []int64{0}},
		{": mk create , does> @ ; 5 mk five five", []int64{5}},
		{": mk create , does> @ 2 * ; 5 mk ten ten", []int64{10}},
		{"3 array a 9 1 a [] ! 1 a [] @", []int64{9}},
		{"1 2 2constant pair pair", []int64{1, 2}},
		{"5 ' dup execute", []int64{5, 5}},
		{": t ['] + execute ; 2 3 t", []int64{5}},
		{": add1 1 + ; : t postpone add1 ; immediate", nil},
	})
}

func TestStrings(t *testing.T) {
	cases := []struct{ src, want string }{
		{`." hello"`, "hello"},
		{`: t ." hi" ; t t`, "hihi"},
		{`s" abc" type`, "abc"},
		{`: t s" xyz" type ; t`, "xyz"},
		{`c" hey" count type`, "hey"},
		{`.( direct)`, "direct"},
		{`char A emit`, "A"},
		{`: t [char] B emit ; t`, "B"},
		{`bl emit`, " "},
		{`65 emit 66 emit`, "AB"},
		{`s" ab" s" ab" str= .`, "-1 "},
		{`s" ab" s" ac" str= .`, "0 "},
		{`s" ab" s" abc" str= .`, "0 "},
	}
	for _, c := range cases {
		_, out := eval(t, c.src)
		if out != c.want {
			t.Errorf("%-30s got %q want %q", c.src, out, c.want)
		}
	}
}

func TestNumberOutput(t *testing.T) {
	cases := []struct{ src, want string }{
		{"42 .", "42 "},
		{"-42 .", "-42 "},
		{"255 hex . decimal", "ff "},
		{"$ff .", "255 "},
		{"%1010 .", "10 "},
		{"#99 .", "99 "},
		{"-1 u.", "18446744073709551615 "},
		{"7 5 .r", "    7"},
		{"1 2 .s", "<2> 1 2 "},
		{"3 spaces", "   "},
		{"cr", "\n"},
		{"space", " "},
		{"3 stars", "***"},
		{"decimal 255 hex . decimal 255 .", "ff 255 "},
	}
	for _, c := range cases {
		_, out := eval(t, c.src)
		if out != c.want {
			t.Errorf("%-30s got %q want %q", c.src, out, c.want)
		}
	}
}

func TestFloats(t *testing.T) {
	cases := []struct {
		src  string
		want []float64
	}{
		{"1.5 2.5 f+", []float64{4}},
		{"1.5 0.5 f-", []float64{1}},
		{"3.0 4.0 f*", []float64{12}},
		{"9.0 3.0 f/", []float64{3}},
		{"2.0 fnegate", []float64{-2}},
		{"-2.5 fabs", []float64{2.5}},
		{"16.0 fsqrt", []float64{4}},
		{"2.0 3.0 f**", []float64{8}},
		{"1.0 2.0 fmin", []float64{1}},
		{"1.0 2.0 fmax", []float64{2}},
		{"3.7 floor", []float64{3}},
		{"3.5 fround", []float64{4}},
		{"7 s>f", []float64{7}},
		{"1.0 fdup", []float64{1, 1}},
		{"1.0 2.0 fswap", []float64{2, 1}},
		{"1.0 2.0 fover", []float64{1, 2, 1}},
		{"1.0 2.0 3.0 frot", []float64{2, 3, 1}},
		{"1.0 2.0 fdrop", []float64{1}},
		{"1.0 2.0 fnip", []float64{2}},
		{"2.0 f2*", []float64{4}},
		{"5.0 f2/", []float64{2.5}},
		{"3.0 fsq", []float64{9}},
		{"1e3", []float64{1000}},
		{"-2.5e-1", []float64{-0.25}},
	}
	for _, c := range cases {
		got := evalF(t, c.src)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%-24s got %v want %v", c.src, got, c.want)
		}
	}

	intCases := []stackCase{
		{"1.0 2.0 f<", []int64{-1}},
		{"1.0 2.0 f>", []int64{0}},
		{"1.0 1.0 f=", []int64{-1}},
		{"1.0 1.0 f<=", []int64{-1}},
		{"1.0 1.0 f>=", []int64{-1}},
		{"-1.0 f0<", []int64{-1}},
		{"0.0 f0=", []int64{-1}},
		{"3.9 f>s", []int64{3}},
		{"1.0 2.0 fdepth", []int64{2}},
	}
	runStackCases(t, intCases)

	if _, out := eval(t, "3.25 f."); out != "3.25 " {
		t.Errorf("f. gave %q", out)
	}
	if _, out := eval(t, "1.5 8 2 f.r"); out != "    1.50" {
		t.Errorf("f.r gave %q", out)
	}
	if _, out := eval(t, "pi f."); !strings.HasPrefix(out, "3.14159") {
		t.Errorf("pi gave %q", out)
	}
	runStackCases(t, []stackCase{{"variable fv 2.5 fv f! fv f@ 2.5 f=", []int64{-1}}})
}

func TestFloatVariableRoundTrip(t *testing.T) {
	got := evalF(t, "variable fv 2.5 fv f! fv f@")
	if len(got) != 1 || got[0] != 2.5 {
		t.Errorf("f@/f! round trip gave %v", got)
	}
}

func TestBaseAndState(t *testing.T) {
	runStackCases(t, []stackCase{
		{"base @", []int64{10}},
		{"hex base @ decimal", []int64{16}},
		{"state @", []int64{0}},
	})
}

func TestErrors(t *testing.T) {
	cases := []struct{ src, contains string }{
		{"drop", "underflow"},
		{"1 0 /", "division by zero"},
		{"1 0 mod", "division by zero"},
		{"nosuchword", "undefined word"},
		{": t 1 ", ""}, // unterminated definition is tolerated until used
		{"1.0 0.0 f/", "float division by zero"},
		{"-1 @", "address out of range"},
		{": t ; ;", "only valid inside a definition"},
		{"then", "without matching opener"},
		{": t 1 if ;", "unclosed control structure"},
		{"0 random", "positive limit"},
		{"' nosuchword", "undefined word"},
		{`: t 1 abort" boom" ; t`, "boom"},
	}
	for _, c := range cases {
		if c.contains == "" {
			continue
		}
		msg := evalErr(t, c.src)
		if !strings.Contains(msg, c.contains) {
			t.Errorf("%-30s error %q does not contain %q", c.src, msg, c.contains)
		}
	}
}

func TestErrorRecovery(t *testing.T) {
	v, buf := newTestVM(t)
	if err := v.Interpret("1 2 nosuchword", "t"); err == nil {
		t.Fatal("expected error")
	}
	if v.dsp != 0 {
		t.Errorf("stack not cleared after error: %d items", v.dsp)
	}
	if v.state() != 0 {
		t.Error("state not reset after error")
	}
	if err := v.Interpret("7 .", "t"); err != nil {
		t.Fatalf("VM unusable after error: %v", err)
	}
	v.out.Flush()
	if !strings.HasSuffix(buf.String(), "7 ") {
		t.Errorf("got %q", buf.String())
	}
}

func TestRedefinition(t *testing.T) {
	runStackCases(t, []stackCase{
		{": t 1 ; : t 2 ; t", []int64{2}},
		{": t 1 ; : u t ; : t 2 ; u", []int64{1}}, // old definition still bound
		{": t 1 ; : t t 2 + ; t", []int64{3}},     // a redefinition may use its predecessor
	})
}

func TestAbortDoesNotCorruptDictionary(t *testing.T) {
	v, _ := newTestVM(t)
	if err := v.Interpret(": broken 1 nosuchword 2 ;", "t"); err == nil {
		t.Fatal("expected error")
	}
	if _, ok := v.lookup("broken"); ok {
		t.Error("failed definition should not be visible")
	}
	if err := v.Interpret(": fine 5 ; ", "t"); err != nil {
		t.Fatalf("could not define after failure: %v", err)
	}
	out, _ := eval(t, ": fine 5 ; fine")
	if len(out) != 1 || out[0] != 5 {
		t.Errorf("got %v", out)
	}
}

func TestEvaluateAndInclude(t *testing.T) {
	runStackCases(t, []stackCase{
		{`s" 2 3 +" evaluate`, []int64{5}},
	})
}

func TestRandom(t *testing.T) {
	v, _ := newTestVM(t)
	if err := v.Interpret("100 0 do 6 random 1+ loop", "t"); err != nil {
		t.Fatal(err)
	}
	if v.dsp != 100 {
		t.Fatalf("expected 100 rolls, got %d", v.dsp)
	}
	seen := map[int64]bool{}
	for _, n := range v.ds[:v.dsp] {
		if n < 1 || n > 6 {
			t.Fatalf("roll out of range: %d", n)
		}
		seen[n] = true
	}
	if len(seen) < 4 {
		t.Errorf("random looks degenerate, only saw %d distinct values", len(seen))
	}
}

func TestNumberParsing(t *testing.T) {
	cases := []struct {
		tok  string
		base int64
		want int64
		ok   bool
	}{
		{"0", 10, 0, true},
		{"123", 10, 123, true},
		{"-123", 10, -123, true},
		{"+7", 10, 7, true},
		{"ff", 16, 255, true},
		{"FF", 16, 255, true},
		{"$ff", 10, 255, true},
		{"%101", 10, 5, true},
		{"#42", 16, 42, true},
		{"'a'", 10, 97, true},
		{"12x", 10, 0, false},
		{"", 10, 0, false},
		{"-", 10, 0, false},
		{"dup", 10, 0, false},
	}
	for _, c := range cases {
		got, ok := parseInt(c.tok, c.base)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("parseInt(%q, %d) = %d,%v want %d,%v", c.tok, c.base, got, ok, c.want, c.ok)
		}
	}

	fcases := []struct {
		tok  string
		want float64
		ok   bool
	}{
		{"1.5", 1.5, true},
		{"-2.0e3", -2000, true},
		{"1e-4", 0.0001, true},
		{"5", 0, false},
		{"dup", 0, false},
		{"e", 0, false},
	}
	for _, c := range fcases {
		got, ok := parseFloat(c.tok, 10)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("parseFloat(%q) = %v,%v want %v,%v", c.tok, got, ok, c.want, c.ok)
		}
	}
}

func TestFlooredDivision(t *testing.T) {
	cases := []struct{ a, b, q, m int64 }{
		{7, 2, 3, 1},
		{-7, 2, -4, 1},
		{7, -2, -4, -1},
		{-7, -2, 3, -1},
	}
	for _, c := range cases {
		if q := fdiv(c.a, c.b); q != c.q {
			t.Errorf("fdiv(%d,%d)=%d want %d", c.a, c.b, q, c.q)
		}
		if m := fmod(c.a, c.b); m != c.m {
			t.Errorf("fmod(%d,%d)=%d want %d", c.a, c.b, m, c.m)
		}
	}
}

func TestTerminalSize(t *testing.T) {
	// Not a terminal under test, so this exercises the COLUMNS/LINES fallback.
	t.Setenv("COLUMNS", "123")
	t.Setenv("LINES", "45")
	got, _ := eval(t, "term-size")
	if len(got) != 2 || got[0] != 123 || got[1] != 45 {
		t.Errorf("term-size gave %v, want [123 45]", got)
	}
}

func TestWordsListing(t *testing.T) {
	_, out := eval(t, "words")
	for _, w := range []string{"dup", "swap", ":", "if", "f+", "str="} {
		if !strings.Contains(out, w) {
			t.Errorf("WORDS output missing %q", w)
		}
	}
}

func TestBye(t *testing.T) {
	v, _ := newTestVM(t)
	err := v.Interpret("1 2 bye 3", "t")
	if _, ok := err.(byeError); !ok {
		t.Fatalf("bye gave %v", err)
	}
}
