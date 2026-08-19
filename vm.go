package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	cellSize   = 8
	dataStackN = 4096
	retStackN  = 4096
	fltStackN  = 1024
	initialMem = 1 << 20
)

type kind uint8

const (
	kPrim kind = iota
	kColon
	kCreate
)

// Word is a dictionary entry. Its index in VM.words is its execution token (xt).
type Word struct {
	name   string
	imm    bool
	hidden bool
	kind   kind
	prim   func(*VM)
	body   int64 // kColon: start index in VM.code
	data   int64 // kCreate: data-field address
	does   int64 // kCreate: code index of does> behaviour, 0 = none
}

// forthError is what every runtime fault panics with; the outer interpreter recovers it.
type forthError struct{ msg string }

func (e forthError) Error() string { return e.msg }

func throw(format string, a ...any) {
	panic(forthError{msg: fmt.Sprintf(format, a...)})
}

// byeError unwinds the interpreter all the way out on BYE.
type byeError struct{ code int }

func (e byeError) Error() string { return "bye" }

type ctrl struct {
	kind   int   // control-stack marker, see cs* constants in compile.go
	addr   int64 // patch or target address
	leaves []int64
}

type source struct {
	text string
	pos  int
	name string
}

// VM is a complete Forth machine: stacks, data space, code space and dictionary.
type VM struct {
	ds  []int64
	dsp int
	rs  []int64
	rsp int
	fs  []float64
	fsp int

	mem  []byte
	here int64

	code  []int64
	words []Word
	dict  map[string]int32

	ip int64

	// Well-known variable addresses inside mem.
	addrState int64
	addrBase  int64
	addrPad   int64
	addrTrans int64 // transient string buffers
	transSlot int

	out *bufio.Writer
	in  *bufio.Reader

	src       []source // source stack (nested INCLUDE / EVALUATE)
	cs        []ctrl   // compiler control stack
	tempDef   bool     // an implicit (interpret-time) definition is open
	tempStart int64    // where that definition starts in code space
	curr      int32    // word currently being compiled, -1 when none
	rbase     int      // return-stack floor of the innermost execute()

	rng  uint64
	args []string

	raw     bool
	keyCh   chan byte
	keyPeek int // -1 when empty
}

func NewVM() *VM {
	v := &VM{
		ds:      make([]int64, dataStackN),
		rs:      make([]int64, retStackN),
		fs:      make([]float64, fltStackN),
		mem:     make([]byte, initialMem),
		code:    make([]int64, 0, 8192),
		words:   make([]Word, 0, 512),
		dict:    make(map[string]int32, 512),
		out:     bufio.NewWriterSize(os.Stdout, 32*1024),
		in:      bufio.NewReader(os.Stdin),
		curr:    -1,
		rng:     0x9E3779B97F4A7C15,
		keyPeek: -1,
	}
	v.here = cellSize // leave address 0 unused so 0 is never a valid address
	v.addrState = v.alloc(cellSize)
	v.addrBase = v.alloc(cellSize)
	v.addrPad = v.alloc(256)
	v.addrTrans = v.alloc(4 * 256)
	v.storeCell(v.addrBase, 10)
	v.storeCell(v.addrState, 0)
	v.code = append(v.code, 0) // code index 0 is never a valid body
	installPrims(v)
	return v
}

// SetOutput redirects all Forth output (used by the test suite).
func (v *VM) SetOutput(w io.Writer) { v.out = bufio.NewWriter(w) }

func (v *VM) alloc(n int64) int64 {
	a := v.here
	v.growTo(a + n)
	v.here = a + n
	return a
}

func (v *VM) growTo(n int64) {
	if n <= int64(len(v.mem)) {
		return
	}
	sz := int64(len(v.mem))
	for sz < n {
		sz *= 2
	}
	nm := make([]byte, sz)
	copy(nm, v.mem)
	v.mem = nm
}

func (v *VM) align() {
	if r := v.here % cellSize; r != 0 {
		v.alloc(cellSize - r)
	}
}

// ---- memory ----

func (v *VM) checkAddr(a, n int64) {
	if a < 0 || a+n > int64(len(v.mem)) {
		throw("address out of range: %d", a)
	}
}

func (v *VM) fetchCell(a int64) int64 {
	v.checkAddr(a, cellSize)
	return int64(binary.LittleEndian.Uint64(v.mem[a:]))
}

func (v *VM) storeCell(a, val int64) {
	v.checkAddr(a, cellSize)
	binary.LittleEndian.PutUint64(v.mem[a:], uint64(val))
}

func (v *VM) fetchByte(a int64) byte {
	v.checkAddr(a, 1)
	return v.mem[a]
}

func (v *VM) storeByte(a int64, b byte) {
	v.checkAddr(a, 1)
	v.mem[a] = b
}

func (v *VM) state() int64   { return v.fetchCell(v.addrState) }
func (v *VM) numBase() int64 { return v.fetchCell(v.addrBase) }

// storeString copies s into fresh data space and returns its address.
func (v *VM) storeString(s string) int64 {
	a := v.alloc(int64(len(s)) + 1)
	copy(v.mem[a:], s)
	v.mem[a+int64(len(s))] = 0
	v.align()
	return a
}

// transient copies s into one of four rotating buffers (for interpret-time S").
func (v *VM) transient(s string) int64 {
	if len(s) > 255 {
		s = s[:255]
	}
	a := v.addrTrans + int64(v.transSlot)*256
	v.transSlot = (v.transSlot + 1) % 4
	copy(v.mem[a:], s)
	return a
}

func (v *VM) str(a, n int64) string {
	if n < 0 {
		throw("negative string length")
	}
	v.checkAddr(a, n)
	return string(v.mem[a : a+n])
}

// ---- stacks ----

func (v *VM) push(n int64) {
	if v.dsp >= len(v.ds) {
		throw("data stack overflow")
	}
	v.ds[v.dsp] = n
	v.dsp++
}

func (v *VM) pop() int64 {
	if v.dsp == 0 {
		throw("data stack underflow")
	}
	v.dsp--
	return v.ds[v.dsp]
}

func (v *VM) need(n int) {
	if v.dsp < n {
		throw("data stack underflow")
	}
}

func (v *VM) rpush(n int64) {
	if v.rsp >= len(v.rs) {
		throw("return stack overflow")
	}
	v.rs[v.rsp] = n
	v.rsp++
}

func (v *VM) rpop() int64 {
	if v.rsp == 0 {
		throw("return stack underflow")
	}
	v.rsp--
	return v.rs[v.rsp]
}

func (v *VM) fpush(f float64) {
	if v.fsp >= len(v.fs) {
		throw("float stack overflow")
	}
	v.fs[v.fsp] = f
	v.fsp++
}

func (v *VM) fpop() float64 {
	if v.fsp == 0 {
		throw("float stack underflow")
	}
	v.fsp--
	return v.fs[v.fsp]
}

func (v *VM) fneed(n int) {
	if v.fsp < n {
		throw("float stack underflow")
	}
}

func (v *VM) pushBool(b bool) {
	if b {
		v.push(-1)
	} else {
		v.push(0)
	}
}

// ---- inner interpreter ----

// execute runs the word xt to completion.
func (v *VM) execute(xt int32) {
	if int(xt) >= len(v.words) || xt < 0 {
		throw("invalid execution token %d", xt)
	}
	w := &v.words[xt]
	switch w.kind {
	case kPrim:
		w.prim(v)
		return
	case kCreate:
		v.push(w.data)
		if w.does == 0 {
			return
		}
		v.run(w.does)
	default:
		v.run(w.body)
	}
}

// run is the inner interpreter: it threads code starting at ip until the
// return stack unwinds past the level it started at.
func (v *VM) run(start int64) {
	savedIP := v.ip
	savedBase := v.rbase
	v.rbase = v.rsp
	v.ip = start
	defer func() {
		v.ip = savedIP
		v.rbase = savedBase
	}()
	for v.ip >= 0 {
		xt := int32(v.code[v.ip])
		v.ip++
		w := &v.words[xt]
		switch w.kind {
		case kPrim:
			w.prim(v)
		case kColon:
			v.rpush(v.ip)
			v.ip = w.body
		default: // kCreate
			v.push(w.data)
			if w.does != 0 {
				v.rpush(v.ip)
				v.ip = w.does
			}
		}
	}
}

// ---- code space ----

func (v *VM) compileCell(n int64) int64 {
	a := int64(len(v.code))
	v.code = append(v.code, n)
	return a
}

func (v *VM) compileXT(name string) {
	xt, ok := v.lookup(name)
	if !ok {
		throw("internal: missing word %q", name)
	}
	v.compileCell(int64(xt))
}

func (v *VM) compileLit(n int64) {
	v.compileXT("(lit)")
	v.compileCell(n)
}

func (v *VM) compileFLit(f float64) {
	v.compileXT("(flit)")
	v.compileCell(int64(math.Float64bits(f)))
}
