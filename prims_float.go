package main

import (
	"math"
	"strconv"
)

func toBits(f float64) int64   { return int64(math.Float64bits(f)) }
func fromBits(n int64) float64 { return math.Float64frombits(uint64(n)) }

func installFloat(v *VM) {
	bin := func(name string, fn func(a, b float64) float64) {
		v.defPrim(name, func(v *VM) {
			b := v.fpop()
			a := v.fpop()
			v.fpush(fn(a, b))
		})
	}
	un := func(name string, fn func(a float64) float64) {
		v.defPrim(name, func(v *VM) { v.fpush(fn(v.fpop())) })
	}
	fcmp := func(name string, fn func(a, b float64) bool) {
		v.defPrim(name, func(v *VM) {
			b := v.fpop()
			a := v.fpop()
			v.pushBool(fn(a, b))
		})
	}

	bin("f+", func(a, b float64) float64 { return a + b })
	bin("f-", func(a, b float64) float64 { return a - b })
	bin("f*", func(a, b float64) float64 { return a * b })
	bin("f/", func(a, b float64) float64 {
		if b == 0 {
			throw("float division by zero")
		}
		return a / b
	})
	bin("fmin", math.Min)
	bin("fmax", math.Max)
	bin("f**", math.Pow)
	bin("fatan2", math.Atan2)
	bin("fmod", math.Mod)

	un("fnegate", func(a float64) float64 { return -a })
	un("fabs", math.Abs)
	un("fsqrt", math.Sqrt)
	un("fsin", math.Sin)
	un("fcos", math.Cos)
	un("ftan", math.Tan)
	un("fatan", math.Atan)
	un("fexp", math.Exp)
	un("fln", func(a float64) float64 {
		if a <= 0 {
			throw("fln of non-positive number")
		}
		return math.Log(a)
	})
	un("flog", func(a float64) float64 {
		if a <= 0 {
			throw("flog of non-positive number")
		}
		return math.Log10(a)
	})
	un("floor", math.Floor)
	un("fround", math.Round)

	fcmp("f<", func(a, b float64) bool { return a < b })
	fcmp("f>", func(a, b float64) bool { return a > b })
	fcmp("f=", func(a, b float64) bool { return a == b })
	fcmp("f<=", func(a, b float64) bool { return a <= b })
	fcmp("f>=", func(a, b float64) bool { return a >= b })
	v.defPrim("f0<", func(v *VM) { v.pushBool(v.fpop() < 0) })
	v.defPrim("f0=", func(v *VM) { v.pushBool(v.fpop() == 0) })

	v.defPrim("fdup", func(v *VM) { v.fneed(1); v.fpush(v.fs[v.fsp-1]) })
	v.defPrim("fdrop", func(v *VM) { v.fpop() })
	v.defPrim("fswap", func(v *VM) {
		v.fneed(2)
		v.fs[v.fsp-1], v.fs[v.fsp-2] = v.fs[v.fsp-2], v.fs[v.fsp-1]
	})
	v.defPrim("fover", func(v *VM) { v.fneed(2); v.fpush(v.fs[v.fsp-2]) })
	v.defPrim("fnip", func(v *VM) {
		v.fneed(2)
		v.fs[v.fsp-2] = v.fs[v.fsp-1]
		v.fsp--
	})
	v.defPrim("frot", func(v *VM) {
		v.fneed(3)
		a, b, c := v.fs[v.fsp-3], v.fs[v.fsp-2], v.fs[v.fsp-1]
		v.fs[v.fsp-3], v.fs[v.fsp-2], v.fs[v.fsp-1] = b, c, a
	})
	v.defPrim("fdepth", func(v *VM) { v.push(int64(v.fsp)) })

	v.defPrim("s>f", func(v *VM) { v.fpush(float64(v.pop())) })
	v.defPrim("f>s", func(v *VM) { v.push(int64(v.fpop())) })

	v.defPrim("f.", func(v *VM) {
		v.out.WriteString(strconv.FormatFloat(v.fpop(), 'f', -1, 64))
		v.out.WriteByte(' ')
	})
	v.defPrim("fe.", func(v *VM) {
		v.out.WriteString(strconv.FormatFloat(v.fpop(), 'e', -1, 64))
		v.out.WriteByte(' ')
	})
	v.defPrim("f.r", func(v *VM) {
		places := int(v.pop())
		width := int(v.pop())
		s := strconv.FormatFloat(v.fpop(), 'f', places, 64)
		for len(s) < width {
			s = " " + s
		}
		v.out.WriteString(s)
	})
	v.defPrim("f.s", func(v *VM) {
		v.out.WriteString("<f")
		v.out.WriteString(strconv.Itoa(v.fsp))
		v.out.WriteString("> ")
		for i := 0; i < v.fsp; i++ {
			v.out.WriteString(strconv.FormatFloat(v.fs[i], 'f', -1, 64))
			v.out.WriteByte(' ')
		}
	})
	v.defPrim("pi", func(v *VM) { v.fpush(math.Pi) })
	v.defPrim("e", func(v *VM) { v.fpush(math.E) })
}
