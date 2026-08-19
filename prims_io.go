package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var startTime = time.Now()

// ---- input plumbing: one reader goroutine feeds every input word ----

func (v *VM) startInput() {
	if v.keyCh != nil {
		return
	}
	ch := make(chan byte, 8192)
	v.keyCh = ch
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := os.Stdin.Read(buf)
			for i := 0; i < n; i++ {
				ch <- buf[i]
			}
			if err != nil {
				close(ch)
				return
			}
		}
	}()
}

// getKey blocks for one byte; ok is false at end of input.
func (v *VM) getKey() (byte, bool) {
	if v.keyPeek >= 0 {
		b := byte(v.keyPeek)
		v.keyPeek = -1
		return b, true
	}
	v.out.Flush()
	v.startInput()
	b, ok := <-v.keyCh
	return b, ok
}

// keyAvail reports whether a byte can be read without blocking.
func (v *VM) keyAvail() bool {
	if v.keyPeek >= 0 {
		return true
	}
	v.out.Flush()
	v.startInput()
	select {
	case b, ok := <-v.keyCh:
		if !ok {
			return false
		}
		v.keyPeek = int(b)
		return true
	default:
		return false
	}
}

// readLine reads one line (without the newline); ok is false at end of input.
func (v *VM) readLine() (string, bool) {
	var sb strings.Builder
	for {
		b, ok := v.getKey()
		if !ok {
			if sb.Len() == 0 {
				return "", false
			}
			return sb.String(), true
		}
		switch b {
		case '\n':
			return sb.String(), true
		case '\r':
			// swallow; the \n (if any) ends the line
		case 8, 127:
			s := sb.String()
			if len(s) > 0 {
				sb.Reset()
				sb.WriteString(s[:len(s)-1])
			}
		default:
			sb.WriteByte(b)
		}
	}
}

// ---- number formatting ----

func formatInt(n, base int64) string {
	if base == 10 {
		return strconv.FormatInt(n, 10)
	}
	if base < 2 || base > 36 {
		throw("invalid base %d", base)
	}
	neg := n < 0
	u := uint64(n)
	if neg {
		u = uint64(-n)
	}
	if u == 0 {
		return "0"
	}
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	var buf [70]byte
	i := len(buf)
	for u > 0 {
		i--
		buf[i] = digits[u%uint64(base)]
		u /= uint64(base)
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func installIO(v *VM) {
	v.defPrim(".", func(v *VM) {
		v.out.WriteString(formatInt(v.pop(), v.numBase()))
		v.out.WriteByte(' ')
	})
	v.defPrim("u.", func(v *VM) {
		n := uint64(v.pop())
		base := uint64(v.numBase())
		const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
		if n == 0 {
			v.out.WriteString("0 ")
			return
		}
		var buf [70]byte
		i := len(buf)
		for n > 0 {
			i--
			buf[i] = digits[n%base]
			n /= base
		}
		v.out.Write(buf[i:])
		v.out.WriteByte(' ')
	})
	v.defPrim(".r", func(v *VM) {
		w := int(v.pop())
		s := formatInt(v.pop(), v.numBase())
		for len(s) < w {
			s = " " + s
		}
		v.out.WriteString(s)
	})
	v.defPrim("emit", func(v *VM) { v.out.WriteByte(byte(v.pop())) })
	v.defPrim("cr", func(v *VM) { v.out.WriteByte('\n') })
	v.defPrim("space", func(v *VM) { v.out.WriteByte(' ') })
	v.defPrim("spaces", func(v *VM) {
		n := v.pop()
		for i := int64(0); i < n; i++ {
			v.out.WriteByte(' ')
		}
	})
	v.defPrim("type", func(v *VM) {
		n := v.pop()
		a := v.pop()
		if n <= 0 {
			return
		}
		v.checkAddr(a, n)
		v.out.Write(v.mem[a : a+n])
	})
	v.defPrim("flush", func(v *VM) { v.out.Flush() })
	v.defPrim(".s", func(v *VM) {
		base := v.numBase()
		v.out.WriteString("<" + strconv.Itoa(v.dsp) + "> ")
		for i := 0; i < v.dsp; i++ {
			v.out.WriteString(formatInt(v.ds[i], base))
			v.out.WriteByte(' ')
		}
	})

	// ---- input ----
	v.defPrim("key", func(v *VM) {
		b, ok := v.getKey()
		if !ok {
			v.push(-1)
			return
		}
		v.push(int64(b))
	})
	v.defPrim("key?", func(v *VM) { v.pushBool(v.keyAvail()) })
	v.defPrim("accept", func(v *VM) {
		max := v.pop()
		a := v.pop()
		line, ok := v.readLine()
		if !ok {
			v.push(0)
			return
		}
		if int64(len(line)) > max {
			line = line[:max]
		}
		v.checkAddr(a, int64(len(line)))
		copy(v.mem[a:], line)
		v.push(int64(len(line)))
	})
	v.defPrim("word", func(v *VM) {
		delim := byte(v.pop())
		var s string
		if delim == ' ' {
			s, _ = v.nextWord()
		} else {
			src := v.top()
			for src.pos < len(src.text) && src.text[src.pos] == delim {
				src.pos++
			}
			s = v.parseTo(delim)
		}
		if len(s) > 255 {
			s = s[:255]
		}
		a := v.addrPad
		v.storeByte(a, byte(len(s)))
		copy(v.mem[a+1:], s)
		v.push(a)
	})
	v.defPrim("number", func(v *VM) { // ( addr len -- n flag )
		n := v.pop()
		a := v.pop()
		val, ok := parseInt(strings.TrimSpace(v.str(a, n)), v.numBase())
		v.push(val)
		v.pushBool(ok)
	})

	// ---- terminal ----
	v.defPrim("page", func(v *VM) { v.out.WriteString("\x1b[2J\x1b[H") })
	v.defPrim("at-xy", func(v *VM) {
		y := v.pop()
		x := v.pop()
		fmt.Fprintf(v.out, "\x1b[%d;%dH", y+1, x+1)
	})
	v.defPrim("raw-on", func(v *VM) {
		v.out.Flush()
		if err := setRaw(true); err == nil {
			v.raw = true
		}
	})
	v.defPrim("raw-off", func(v *VM) {
		v.out.Flush()
		if err := setRaw(false); err == nil {
			v.raw = false
		}
	})
	v.defPrim("cursor-off", func(v *VM) { v.out.WriteString("\x1b[?25l") })
	v.defPrim("term-size", func(v *VM) { // ( -- cols rows )
		cols, rows := terminalSize()
		v.push(int64(cols))
		v.push(int64(rows))
	})
	v.defPrim("cursor-on", func(v *VM) { v.out.WriteString("\x1b[?25h") })

	// ---- time and randomness ----
	v.defPrim("ms", func(v *VM) {
		n := v.pop()
		if n > 0 {
			v.out.Flush()
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	})
	v.defPrim("ticks", func(v *VM) { v.push(int64(time.Since(startTime) / time.Millisecond)) })
	v.defPrim("randomize", func(v *VM) { v.rng = uint64(time.Now().UnixNano()) | 1 })
	v.defPrim("random", func(v *VM) { // ( n -- 0..n-1 )
		n := v.pop()
		if n <= 0 {
			throw("random needs a positive limit")
		}
		v.push(int64(v.nextRand() % uint64(n)))
	})
	v.defPrim("seed", func(v *VM) { v.rng = uint64(v.pop()) | 1 })

	// ---- system ----
	v.defPrim("bye", func(v *VM) { panic(byeError{code: 0}) })
	v.defPrim("abort", func(v *VM) { throw("aborted") })
	v.defPrim("quit", func(v *VM) { panic(byeError{code: 0}) })
	v.defPrim("words", func(v *VM) {
		col := 0
		for _, n := range v.wordNames() {
			if col+len(n)+1 > 78 {
				v.out.WriteByte('\n')
				col = 0
			}
			v.out.WriteString(n)
			v.out.WriteByte(' ')
			col += len(n) + 1
		}
		v.out.WriteByte('\n')
	})
	v.defPrim("argc", func(v *VM) { v.push(int64(len(v.args))) })
	v.defPrim("arg", func(v *VM) { // ( n -- addr len )
		n := v.pop()
		if n < 0 || int(n) >= len(v.args) {
			v.push(v.addrPad)
			v.push(0)
			return
		}
		s := v.args[n]
		v.push(v.transient(s))
		v.push(int64(len(s)))
	})
	v.defPrim("depth!", func(v *VM) { v.dsp = 0 })
}

// terminalSize reports the window size, falling back to COLUMNS/LINES and
// finally to 80x24 when stdout is not a terminal (a pipe, or a test).
func terminalSize() (int, int) {
	if cols, rows, ok := termSize(); ok {
		return cols, rows
	}
	cols, rows := 80, 24
	if n, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && n > 0 {
		cols = n
	}
	if n, err := strconv.Atoi(os.Getenv("LINES")); err == nil && n > 0 {
		rows = n
	}
	return cols, rows
}

func (v *VM) nextRand() uint64 {
	x := v.rng
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	v.rng = x
	return x
}

func (v *VM) includeFile(name string) {
	data, err := os.ReadFile(name)
	if err != nil {
		throw("cannot read %s: %v", name, err)
	}
	v.interpretText(string(data), name)
}

func installPrims(v *VM) {
	installCore(v)
	installFloat(v)
	installIO(v)
	installCompiler(v)
}
