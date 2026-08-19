package main

func installCore(v *VM) {
	// ---- stack ----
	v.defPrim("dup", func(v *VM) { v.need(1); v.push(v.ds[v.dsp-1]) })
	v.defPrim("?dup", func(v *VM) {
		v.need(1)
		if v.ds[v.dsp-1] != 0 {
			v.push(v.ds[v.dsp-1])
		}
	})
	v.defPrim("drop", func(v *VM) { v.pop() })
	v.defPrim("swap", func(v *VM) {
		v.need(2)
		v.ds[v.dsp-1], v.ds[v.dsp-2] = v.ds[v.dsp-2], v.ds[v.dsp-1]
	})
	v.defPrim("over", func(v *VM) { v.need(2); v.push(v.ds[v.dsp-2]) })
	v.defPrim("nip", func(v *VM) {
		v.need(2)
		v.ds[v.dsp-2] = v.ds[v.dsp-1]
		v.dsp--
	})
	v.defPrim("tuck", func(v *VM) {
		v.need(2)
		a, b := v.ds[v.dsp-2], v.ds[v.dsp-1]
		v.ds[v.dsp-2] = b
		v.ds[v.dsp-1] = a
		v.push(b)
	})
	v.defPrim("rot", func(v *VM) {
		v.need(3)
		a, b, c := v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1]
		v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1] = b, c, a
	})
	v.defPrim("-rot", func(v *VM) {
		v.need(3)
		a, b, c := v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1]
		v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1] = c, a, b
	})
	v.defPrim("2dup", func(v *VM) {
		v.need(2)
		a, b := v.ds[v.dsp-2], v.ds[v.dsp-1]
		v.push(a)
		v.push(b)
	})
	v.defPrim("2drop", func(v *VM) { v.pop(); v.pop() })
	v.defPrim("2swap", func(v *VM) {
		v.need(4)
		a, b, c, d := v.ds[v.dsp-4], v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1]
		v.ds[v.dsp-4], v.ds[v.dsp-3], v.ds[v.dsp-2], v.ds[v.dsp-1] = c, d, a, b
	})
	v.defPrim("2over", func(v *VM) {
		v.need(4)
		a, b := v.ds[v.dsp-4], v.ds[v.dsp-3]
		v.push(a)
		v.push(b)
	})
	v.defPrim("pick", func(v *VM) {
		n := int(v.pop())
		if n < 0 || n >= v.dsp {
			throw("pick out of range")
		}
		v.push(v.ds[v.dsp-1-n])
	})
	v.defPrim("roll", func(v *VM) {
		n := int(v.pop())
		if n < 0 || n >= v.dsp {
			throw("roll out of range")
		}
		x := v.ds[v.dsp-1-n]
		copy(v.ds[v.dsp-1-n:], v.ds[v.dsp-n:v.dsp])
		v.ds[v.dsp-1] = x
	})
	v.defPrim("depth", func(v *VM) { v.push(int64(v.dsp)) })

	// ---- return stack ----
	v.defPrim(">r", func(v *VM) { v.rpush(v.pop()) })
	v.defPrim("r>", func(v *VM) { v.push(v.rpop()) })
	v.defPrim("r@", func(v *VM) {
		if v.rsp == 0 {
			throw("return stack underflow")
		}
		v.push(v.rs[v.rsp-1])
	})
	v.defPrim("2>r", func(v *VM) {
		b := v.pop()
		a := v.pop()
		v.rpush(a)
		v.rpush(b)
	})
	v.defPrim("2r>", func(v *VM) {
		b := v.rpop()
		a := v.rpop()
		v.push(a)
		v.push(b)
	})
	v.defPrim("rdepth", func(v *VM) { v.push(int64(v.rsp)) })

	// ---- arithmetic ----
	v.defPrim("+", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] += v.ds[v.dsp] })
	v.defPrim("-", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] -= v.ds[v.dsp] })
	v.defPrim("*", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] *= v.ds[v.dsp] })
	v.defPrim("/", func(v *VM) {
		b := v.pop()
		a := v.pop()
		if b == 0 {
			throw("division by zero")
		}
		v.push(fdiv(a, b))
	})
	v.defPrim("mod", func(v *VM) {
		b := v.pop()
		a := v.pop()
		if b == 0 {
			throw("division by zero")
		}
		v.push(fmod(a, b))
	})
	v.defPrim("/mod", func(v *VM) {
		b := v.pop()
		a := v.pop()
		if b == 0 {
			throw("division by zero")
		}
		v.push(fmod(a, b))
		v.push(fdiv(a, b))
	})
	v.defPrim("*/", func(v *VM) {
		c := v.pop()
		b := v.pop()
		a := v.pop()
		if c == 0 {
			throw("division by zero")
		}
		v.push(int64((int64(a) * int64(b)) / c))
	})
	v.defPrim("negate", func(v *VM) { v.need(1); v.ds[v.dsp-1] = -v.ds[v.dsp-1] })
	v.defPrim("abs", func(v *VM) {
		v.need(1)
		if v.ds[v.dsp-1] < 0 {
			v.ds[v.dsp-1] = -v.ds[v.dsp-1]
		}
	})
	v.defPrim("min", func(v *VM) {
		b := v.pop()
		a := v.pop()
		if a < b {
			v.push(a)
		} else {
			v.push(b)
		}
	})
	v.defPrim("max", func(v *VM) {
		b := v.pop()
		a := v.pop()
		if a > b {
			v.push(a)
		} else {
			v.push(b)
		}
	})
	v.defPrim("1+", func(v *VM) { v.need(1); v.ds[v.dsp-1]++ })
	v.defPrim("1-", func(v *VM) { v.need(1); v.ds[v.dsp-1]-- })
	v.defPrim("2*", func(v *VM) { v.need(1); v.ds[v.dsp-1] <<= 1 })
	v.defPrim("2/", func(v *VM) { v.need(1); v.ds[v.dsp-1] >>= 1 })
	v.defPrim("lshift", func(v *VM) {
		n := uint64(v.pop())
		a := v.pop()
		v.push(int64(uint64(a) << (n & 63)))
	})
	v.defPrim("rshift", func(v *VM) {
		n := uint64(v.pop())
		a := v.pop()
		v.push(int64(uint64(a) >> (n & 63)))
	})
	v.defPrim("and", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] &= v.ds[v.dsp] })
	v.defPrim("or", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] |= v.ds[v.dsp] })
	v.defPrim("xor", func(v *VM) { v.need(2); v.dsp--; v.ds[v.dsp-1] ^= v.ds[v.dsp] })
	v.defPrim("invert", func(v *VM) { v.need(1); v.ds[v.dsp-1] = ^v.ds[v.dsp-1] })

	// ---- comparison ----
	cmp := func(name string, fn func(a, b int64) bool) {
		v.defPrim(name, func(v *VM) {
			b := v.pop()
			a := v.pop()
			v.pushBool(fn(a, b))
		})
	}
	cmp("=", func(a, b int64) bool { return a == b })
	cmp("<>", func(a, b int64) bool { return a != b })
	cmp("<", func(a, b int64) bool { return a < b })
	cmp(">", func(a, b int64) bool { return a > b })
	cmp("<=", func(a, b int64) bool { return a <= b })
	cmp(">=", func(a, b int64) bool { return a >= b })
	cmp("u<", func(a, b int64) bool { return uint64(a) < uint64(b) })
	cmp("u>", func(a, b int64) bool { return uint64(a) > uint64(b) })
	zcmp := func(name string, fn func(a int64) bool) {
		v.defPrim(name, func(v *VM) { v.pushBool(fn(v.pop())) })
	}
	zcmp("0=", func(a int64) bool { return a == 0 })
	zcmp("0<>", func(a int64) bool { return a != 0 })
	zcmp("0<", func(a int64) bool { return a < 0 })
	zcmp("0>", func(a int64) bool { return a > 0 })

	// ---- memory ----
	v.defPrim("@", func(v *VM) { v.push(v.fetchCell(v.pop())) })
	v.defPrim("!", func(v *VM) {
		a := v.pop()
		n := v.pop()
		v.storeCell(a, n)
	})
	v.defPrim("c@", func(v *VM) { v.push(int64(v.fetchByte(v.pop()))) })
	v.defPrim("c!", func(v *VM) {
		a := v.pop()
		n := v.pop()
		v.storeByte(a, byte(n))
	})
	v.defPrim("+!", func(v *VM) {
		a := v.pop()
		n := v.pop()
		v.storeCell(a, v.fetchCell(a)+n)
	})
	v.defPrim("f@", func(v *VM) { v.fpush(fromBits(v.fetchCell(v.pop()))) })
	v.defPrim("f!", func(v *VM) {
		a := v.pop()
		v.storeCell(a, toBits(v.fpop()))
	})
	v.defPrim("here", func(v *VM) { v.push(v.here) })
	v.defPrim("allot", func(v *VM) {
		n := v.pop()
		if n < 0 {
			if -n > v.here {
				throw("allot below start of data space")
			}
			v.here += n
			return
		}
		v.alloc(n)
	})
	v.defPrim(",", func(v *VM) {
		v.align()
		a := v.alloc(cellSize)
		v.storeCell(a, v.pop())
	})
	v.defPrim("c,", func(v *VM) {
		a := v.alloc(1)
		v.storeByte(a, byte(v.pop()))
	})
	v.defPrim("f,", func(v *VM) {
		v.align()
		a := v.alloc(cellSize)
		v.storeCell(a, toBits(v.fpop()))
	})
	v.defPrim("align", func(v *VM) { v.align() })
	v.defPrim("cells", func(v *VM) { v.need(1); v.ds[v.dsp-1] *= cellSize })
	v.defPrim("cell+", func(v *VM) { v.need(1); v.ds[v.dsp-1] += cellSize })
	v.defPrim("chars", func(v *VM) { v.need(1) })
	v.defPrim("char+", func(v *VM) { v.need(1); v.ds[v.dsp-1]++ })
	v.defPrim("floats", func(v *VM) { v.need(1); v.ds[v.dsp-1] *= cellSize })
	v.defPrim("float+", func(v *VM) { v.need(1); v.ds[v.dsp-1] += cellSize })
	v.defPrim("move", func(v *VM) {
		n := v.pop()
		dst := v.pop()
		src := v.pop()
		if n <= 0 {
			return
		}
		v.checkAddr(src, n)
		v.checkAddr(dst, n)
		copy(v.mem[dst:dst+n], v.mem[src:src+n])
	})
	v.defPrim("fill", func(v *VM) {
		c := byte(v.pop())
		n := v.pop()
		a := v.pop()
		if n <= 0 {
			return
		}
		v.checkAddr(a, n)
		for i := int64(0); i < n; i++ {
			v.mem[a+i] = c
		}
	})
	v.defPrim("erase", func(v *VM) {
		n := v.pop()
		a := v.pop()
		if n <= 0 {
			return
		}
		v.checkAddr(a, n)
		for i := int64(0); i < n; i++ {
			v.mem[a+i] = 0
		}
	})
	v.defPrim("count", func(v *VM) {
		a := v.pop()
		n := int64(v.fetchByte(a))
		v.push(a + 1)
		v.push(n)
	})
	v.defPrim("pad", func(v *VM) { v.push(v.addrPad) })
	v.defPrim("state", func(v *VM) { v.push(v.addrState) })
	v.defPrim("base", func(v *VM) { v.push(v.addrBase) })
	v.defPrim("decimal", func(v *VM) { v.storeCell(v.addrBase, 10) })
	v.defPrim("hex", func(v *VM) { v.storeCell(v.addrBase, 16) })

	// ---- constants ----
	v.defPrim("true", func(v *VM) { v.push(-1) })
	v.defPrim("false", func(v *VM) { v.push(0) })
	v.defPrim("bl", func(v *VM) { v.push(32) })
	v.defPrim("cell", func(v *VM) { v.push(cellSize) })
}

// fdiv and fmod implement floored division, as ANS Forth requires.
func fdiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func fmod(a, b int64) int64 {
	return a - fdiv(a, b)*b
}
