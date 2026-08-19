package main

import (
	"math"
	"strconv"
	"strings"
)

// control-stack markers
const (
	csIf = iota
	csBegin
	csWhile
	csDo
	csCase
	csOf
)

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == 0
}

func (v *VM) top() *source {
	if len(v.src) == 0 {
		throw("no input source")
	}
	return &v.src[len(v.src)-1]
}

// nextWord returns the next whitespace-delimited token from the current source.
func (v *VM) nextWord() (string, bool) {
	s := v.top()
	for s.pos < len(s.text) && isSpace(s.text[s.pos]) {
		s.pos++
	}
	if s.pos >= len(s.text) {
		return "", false
	}
	start := s.pos
	for s.pos < len(s.text) && !isSpace(s.text[s.pos]) {
		s.pos++
	}
	return s.text[start:s.pos], true
}

// parseTo consumes text up to (and including) the delimiter, returning what came before.
func (v *VM) parseTo(delim byte) string {
	s := v.top()
	start := s.pos
	for s.pos < len(s.text) && s.text[s.pos] != delim {
		s.pos++
	}
	out := s.text[start:s.pos]
	if s.pos < len(s.text) {
		s.pos++ // consume delimiter
	}
	return out
}

// parseStringArg skips the single blank that delimits a string-parsing word
// from its text, then consumes up to delim.
func (v *VM) parseStringArg(delim byte) string {
	src := v.top()
	if src.pos < len(src.text) && (src.text[src.pos] == ' ' || src.text[src.pos] == '\t') {
		src.pos++
	}
	return v.parseTo(delim)
}

func (v *VM) skipLine() {
	s := v.top()
	for s.pos < len(s.text) && s.text[s.pos] != '\n' {
		s.pos++
	}
}

// ---- number parsing ----

func digitVal(c byte) int64 {
	switch {
	case c >= '0' && c <= '9':
		return int64(c - '0')
	case c >= 'a' && c <= 'z':
		return int64(c-'a') + 10
	case c >= 'A' && c <= 'Z':
		return int64(c-'A') + 10
	}
	return -1
}

// parseInt parses a Forth integer literal honouring BASE and the $ # % ' prefixes.
func parseInt(tok string, base int64) (int64, bool) {
	if tok == "" {
		return 0, false
	}
	if len(tok) == 3 && tok[0] == '\'' && tok[2] == '\'' {
		return int64(tok[1]), true
	}
	switch tok[0] {
	case '$':
		base, tok = 16, tok[1:]
	case '#':
		base, tok = 10, tok[1:]
	case '%':
		base, tok = 2, tok[1:]
	}
	neg := false
	if strings.HasPrefix(tok, "-") {
		neg, tok = true, tok[1:]
	} else if strings.HasPrefix(tok, "+") {
		tok = tok[1:]
	}
	if tok == "" || base < 2 || base > 36 {
		return 0, false
	}
	var n int64
	for i := 0; i < len(tok); i++ {
		d := digitVal(tok[i])
		if d < 0 || d >= base {
			return 0, false
		}
		n = n*base + d
	}
	if neg {
		n = -n
	}
	return n, true
}

// parseFloat recognises literals such as 1.5, -2.0e3, 1e-4 (base 10 only).
func parseFloat(tok string, base int64) (float64, bool) {
	if base != 10 || tok == "" {
		return 0, false
	}
	if !strings.ContainsAny(tok, ".eE") {
		return 0, false
	}
	// reject bare words that merely contain e/E
	f, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// ---- outer interpreter ----

// Interpret runs source text, recovering Forth errors into a Go error.
func (v *VM) Interpret(text, name string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case forthError:
				v.resetAfterError()
				err = e
			case byeError:
				err = e
			default:
				panic(r)
			}
		}
	}()
	v.interpretText(text, name)
	return nil
}

func (v *VM) resetAfterError() {
	v.dsp, v.rsp, v.fsp = 0, 0, 0
	v.storeCell(v.addrState, 0)
	if v.curr >= 0 {
		v.words[v.curr].hidden = true
	}
	v.curr = -1
	v.cs = v.cs[:0]
	if v.tempDef {
		v.code = v.code[:v.tempStart]
		v.tempDef = false
	}
	v.src = v.src[:0]
	v.out.Flush()
}

// interpretText pushes text as the current source and runs it to exhaustion.
func (v *VM) interpretText(text, name string) {
	v.src = append(v.src, source{text: text, name: name})
	defer func() { v.src = v.src[:len(v.src)-1] }()
	for {
		tok, ok := v.nextWord()
		if !ok {
			return
		}
		v.processToken(tok)
	}
}

func (v *VM) processToken(tok string) {
	if xt, ok := v.lookup(tok); ok {
		w := &v.words[xt]
		if v.state() != 0 && !w.imm {
			v.compileCell(int64(xt))
		} else {
			v.execute(xt)
		}
		return
	}
	if n, ok := parseInt(tok, v.numBase()); ok {
		if v.state() != 0 {
			v.compileLit(n)
		} else {
			v.push(n)
		}
		return
	}
	if f, ok := parseFloat(tok, v.numBase()); ok {
		if v.state() != 0 {
			v.compileFLit(f)
		} else {
			v.fpush(f)
		}
		return
	}
	throw("undefined word: %s", tok)
}

// ---- compiler words ----

func (v *VM) pushCtrl(c ctrl) { v.cs = append(v.cs, c) }

func (v *VM) popCtrl(want int, who string) ctrl {
	if len(v.cs) == 0 {
		throw("%s without matching opener", who)
	}
	c := v.cs[len(v.cs)-1]
	if c.kind != want {
		throw("%s does not match the open control structure", who)
	}
	v.cs = v.cs[:len(v.cs)-1]
	return c
}

// innermostDo finds the enclosing DO for LEAVE.
func (v *VM) innermostDo() *ctrl {
	for i := len(v.cs) - 1; i >= 0; i-- {
		if v.cs[i].kind == csDo {
			return &v.cs[i]
		}
	}
	throw("leave outside of do...loop")
	return nil
}

func (v *VM) compileBranch(name string) int64 {
	v.compileXT(name)
	return v.compileCell(0)
}

func (v *VM) patch(slot int64) { v.code[slot] = int64(len(v.code)) }

// beginTemp starts an implicit definition so that control structures can be
// used straight from the REPL or at the top level of a script.
func (v *VM) beginTemp() {
	if v.state() == 0 {
		v.tempDef = true
		v.tempStart = int64(len(v.code))
		v.storeCell(v.addrState, 1)
	}
}

// maybeRunTemp closes an implicit definition once its outermost control
// structure is complete, runs it, then reclaims the code space.
func (v *VM) maybeRunTemp() {
	if !v.tempDef || len(v.cs) != 0 {
		return
	}
	v.compileXT("exit")
	v.storeCell(v.addrState, 0)
	v.tempDef = false
	start := v.tempStart
	nwords := len(v.words)
	defer func() {
		// Only reclaim the scratch code if nothing permanent was defined while
		// it ran; a CONSTANT or : inside the structure lives in that space.
		if len(v.words) == nwords {
			v.code = v.code[:start]
		}
	}()
	v.run(start)
}

func (v *VM) mustCompiling(who string) {
	if v.state() == 0 {
		throw("%s is only valid inside a definition", who)
	}
}

func installCompiler(v *VM) {
	v.defImm(":", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw(": needs a name")
		}
		lower := strings.ToLower(name)
		prev, hadPrev := v.dict[lower]
		xt := v.addWord(Word{name: name, kind: kColon, body: int64(len(v.code)), hidden: true})
		// Until ; runs, the name still refers to the previous definition, so a
		// redefinition may build on the word it replaces.
		if hadPrev {
			v.dict[lower] = prev
		} else {
			delete(v.dict, lower)
		}
		v.curr = xt
		v.storeCell(v.addrState, 1)
	})

	v.defImm(";", func(v *VM) {
		v.mustCompiling(";")
		if len(v.cs) != 0 {
			throw("unclosed control structure at ;")
		}
		v.compileXT("exit")
		if v.curr >= 0 {
			v.words[v.curr].hidden = false
			// re-register: a redefinition must win the name lookup
			v.dict[strings.ToLower(v.words[v.curr].name)] = v.curr
		}
		v.curr = -1
		v.storeCell(v.addrState, 0)
	})

	v.defImm("immediate", func(v *VM) {
		if len(v.words) == 0 {
			throw("immediate with no definition")
		}
		v.words[len(v.words)-1].imm = true
	})

	v.defImm("recurse", func(v *VM) {
		v.mustCompiling("recurse")
		if v.curr < 0 {
			throw("recurse outside a definition")
		}
		v.compileCell(int64(v.curr))
	})

	v.defImm("[", func(v *VM) { v.storeCell(v.addrState, 0) })
	v.defImm("]", func(v *VM) { v.storeCell(v.addrState, 1) })

	v.defImm("(", func(v *VM) { v.parseTo(')') })
	v.defImm("\\", func(v *VM) { v.skipLine() })

	// --- conditionals ---
	v.defImm("if", func(v *VM) {
		v.beginTemp()
		v.mustCompiling("if")
		v.pushCtrl(ctrl{kind: csIf, addr: v.compileBranch("(0branch)")})
	})
	v.defImm("else", func(v *VM) {
		c := v.popCtrl(csIf, "else")
		slot := v.compileBranch("(branch)")
		v.patch(c.addr)
		v.pushCtrl(ctrl{kind: csIf, addr: slot})
	})
	v.defImm("then", func(v *VM) {
		c := v.popCtrl(csIf, "then")
		v.patch(c.addr)
		v.maybeRunTemp()
	})

	// --- indefinite loops ---
	v.defImm("begin", func(v *VM) {
		v.beginTemp()
		v.mustCompiling("begin")
		v.pushCtrl(ctrl{kind: csBegin, addr: int64(len(v.code))})
	})
	v.defImm("until", func(v *VM) {
		c := v.popCtrl(csBegin, "until")
		v.compileXT("(0branch)")
		v.compileCell(c.addr)
		v.maybeRunTemp()
	})
	v.defImm("again", func(v *VM) {
		c := v.popCtrl(csBegin, "again")
		v.compileXT("(branch)")
		v.compileCell(c.addr)
		v.maybeRunTemp()
	})
	v.defImm("while", func(v *VM) {
		if len(v.cs) == 0 || v.cs[len(v.cs)-1].kind != csBegin {
			throw("while without begin")
		}
		v.pushCtrl(ctrl{kind: csWhile, addr: v.compileBranch("(0branch)")})
	})
	v.defImm("repeat", func(v *VM) {
		w := v.popCtrl(csWhile, "repeat")
		b := v.popCtrl(csBegin, "repeat")
		v.compileXT("(branch)")
		v.compileCell(b.addr)
		v.patch(w.addr)
		v.maybeRunTemp()
	})

	// --- counted loops ---
	v.defImm("do", func(v *VM) {
		v.beginTemp()
		v.mustCompiling("do")
		v.compileXT("(do)")
		v.pushCtrl(ctrl{kind: csDo, addr: int64(len(v.code))})
	})
	v.defImm("?do", func(v *VM) {
		v.beginTemp()
		v.mustCompiling("?do")
		slot := v.compileBranch("(?do)")
		v.pushCtrl(ctrl{kind: csDo, addr: int64(len(v.code)), leaves: []int64{slot}})
	})
	endLoop := func(v *VM, prim, who string) {
		c := v.popCtrl(csDo, who)
		v.compileXT(prim)
		v.compileCell(c.addr)
		for _, s := range c.leaves {
			v.patch(s)
		}
		v.maybeRunTemp()
	}
	v.defImm("loop", func(v *VM) { endLoop(v, "(loop)", "loop") })
	v.defImm("+loop", func(v *VM) { endLoop(v, "(+loop)", "+loop") })
	v.defImm("leave", func(v *VM) {
		v.mustCompiling("leave")
		v.compileXT("unloop")
		slot := v.compileBranch("(branch)")
		d := v.innermostDo()
		d.leaves = append(d.leaves, slot)
	})

	// --- case ---
	v.defImm("case", func(v *VM) {
		v.beginTemp()
		v.mustCompiling("case")
		v.pushCtrl(ctrl{kind: csCase})
	})
	v.defImm("of", func(v *VM) {
		v.mustCompiling("of")
		v.compileXT("over")
		v.compileXT("=")
		slot := v.compileBranch("(0branch)")
		v.compileXT("drop")
		v.pushCtrl(ctrl{kind: csOf, addr: slot})
	})
	v.defImm("endof", func(v *VM) {
		c := v.popCtrl(csOf, "endof")
		slot := v.compileBranch("(branch)")
		v.patch(c.addr)
		if len(v.cs) == 0 || v.cs[len(v.cs)-1].kind != csCase {
			throw("endof without case")
		}
		cc := &v.cs[len(v.cs)-1]
		cc.leaves = append(cc.leaves, slot)
	})
	v.defImm("endcase", func(v *VM) {
		c := v.popCtrl(csCase, "endcase")
		v.compileXT("drop")
		for _, s := range c.leaves {
			v.patch(s)
		}
		v.maybeRunTemp()
	})

	// --- literals and execution tokens ---
	v.defImm("literal", func(v *VM) {
		v.mustCompiling("literal")
		v.compileLit(v.pop())
	})
	v.defImm("fliteral", func(v *VM) {
		v.mustCompiling("fliteral")
		v.compileFLit(v.fpop())
	})
	v.defPrim("'", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("' needs a name")
		}
		xt, found := v.lookup(name)
		if !found {
			throw("undefined word: %s", name)
		}
		v.push(int64(xt))
	})
	v.defImm("[']", func(v *VM) {
		v.mustCompiling("[']")
		name, ok := v.nextWord()
		if !ok {
			throw("['] needs a name")
		}
		xt, found := v.lookup(name)
		if !found {
			throw("undefined word: %s", name)
		}
		v.compileLit(int64(xt))
	})
	v.defImm("postpone", func(v *VM) {
		v.mustCompiling("postpone")
		name, ok := v.nextWord()
		if !ok {
			throw("postpone needs a name")
		}
		xt, found := v.lookup(name)
		if !found {
			throw("undefined word: %s", name)
		}
		if v.words[xt].imm {
			v.compileCell(int64(xt))
		} else {
			v.compileLit(int64(xt))
			v.compileXT("compile,")
		}
	})
	v.defPrim("compile,", func(v *VM) { v.compileCell(v.pop()) })
	v.defPrim("execute", func(v *VM) { v.execute(int32(v.pop())) })

	// --- character and string literals ---
	v.defPrim("char", func(v *VM) {
		tok, ok := v.nextWord()
		if !ok || tok == "" {
			throw("char needs a character")
		}
		v.push(int64(tok[0]))
	})
	v.defImm("[char]", func(v *VM) {
		tok, ok := v.nextWord()
		if !ok || tok == "" {
			throw("[char] needs a character")
		}
		if v.state() != 0 {
			v.compileLit(int64(tok[0]))
		} else {
			v.push(int64(tok[0]))
		}
	})
	v.defImm("s\"", func(v *VM) {
		s := v.parseStringArg('"')
		if v.state() != 0 {
			a := v.storeString(s)
			v.compileLit(a)
			v.compileLit(int64(len(s)))
		} else {
			v.push(v.transient(s))
			v.push(int64(len(s)))
		}
	})
	v.defImm("c\"", func(v *VM) {
		s := v.parseStringArg('"')
		if len(s) > 255 {
			throw("counted string too long")
		}
		if v.state() != 0 {
			a := v.storeString(string([]byte{byte(len(s))}) + s)
			v.compileLit(a)
		} else {
			v.push(v.transient(string(byte(len(s))) + s))
		}
	})
	v.defImm(".\"", func(v *VM) {
		s := v.parseStringArg('"')
		if v.state() != 0 {
			a := v.storeString(s)
			v.compileLit(a)
			v.compileLit(int64(len(s)))
			v.compileXT("type")
		} else {
			v.out.WriteString(s)
		}
	})
	v.defImm(".(", func(v *VM) {
		v.out.WriteString(v.parseStringArg(')'))
	})
	v.defImm("abort\"", func(v *VM) {
		v.mustCompiling("abort\"")
		s := v.parseStringArg('"')
		a := v.storeString(s)
		v.compileLit(a)
		v.compileLit(int64(len(s)))
		v.compileXT("(abort\")")
	})
	v.defPrim("(abort\")", func(v *VM) {
		n := v.pop()
		a := v.pop()
		if v.pop() != 0 {
			throw("%s", v.str(a, n))
		}
	})

	// --- defining words ---
	v.defPrim("create", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("create needs a name")
		}
		v.align()
		v.create(name, v.here)
	})
	v.defPrim("variable", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("variable needs a name")
		}
		v.align()
		a := v.alloc(cellSize)
		v.storeCell(a, 0)
		v.create(name, a)
	})
	v.defPrim("fvariable", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("fvariable needs a name")
		}
		v.align()
		a := v.alloc(cellSize)
		v.storeCell(a, 0)
		v.create(name, a)
	})
	v.defPrim("constant", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("constant needs a name")
		}
		n := v.pop()
		v.addWord(Word{name: name, kind: kColon, body: int64(len(v.code))})
		v.compileLit(n)
		v.compileXT("exit")
	})
	v.defPrim("fconstant", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("fconstant needs a name")
		}
		f := v.fpop()
		v.addWord(Word{name: name, kind: kColon, body: int64(len(v.code))})
		v.compileFLit(f)
		v.compileXT("exit")
	})
	v.defImm("does>", func(v *VM) {
		v.mustCompiling("does>")
		// end the defining word here; the rest becomes the runtime behaviour
		v.compileXT("(does>)")
		slot := v.compileCell(0)
		v.compileXT("exit")
		v.code[slot] = int64(len(v.code))
	})
	v.defPrim("(does>)", func(v *VM) {
		target := v.code[v.ip]
		v.ip++
		if len(v.words) == 0 {
			throw("does> with no word to modify")
		}
		w := &v.words[len(v.words)-1]
		if w.kind != kCreate {
			throw("does> requires a create'd word")
		}
		w.does = target
	})

	v.defPrim("include", func(v *VM) {
		name, ok := v.nextWord()
		if !ok {
			throw("include needs a filename")
		}
		v.includeFile(name)
	})
	v.defPrim("included", func(v *VM) {
		n := v.pop()
		a := v.pop()
		v.includeFile(v.str(a, n))
	})
	v.defPrim("evaluate", func(v *VM) {
		n := v.pop()
		a := v.pop()
		v.interpretText(v.str(a, n), "evaluate")
	})

	// literal runtime helpers
	v.defPrim("(lit)", func(v *VM) {
		v.push(v.code[v.ip])
		v.ip++
	})
	v.defPrim("(flit)", func(v *VM) {
		v.fpush(math.Float64frombits(uint64(v.code[v.ip])))
		v.ip++
	})
	v.defPrim("(branch)", func(v *VM) { v.ip = v.code[v.ip] })
	v.defPrim("(0branch)", func(v *VM) {
		if v.pop() == 0 {
			v.ip = v.code[v.ip]
		} else {
			v.ip++
		}
	})
	v.defPrim("exit", func(v *VM) {
		if v.rsp <= v.rbase {
			v.ip = -1
		} else {
			v.ip = v.rpop()
		}
	})
	v.defPrim("(do)", func(v *VM) {
		i := v.pop()
		l := v.pop()
		v.rpush(l)
		v.rpush(i)
	})
	v.defPrim("(?do)", func(v *VM) {
		i := v.pop()
		l := v.pop()
		if i == l {
			v.ip = v.code[v.ip]
			return
		}
		v.rpush(l)
		v.rpush(i)
		v.ip++
	})
	v.defPrim("(loop)", func(v *VM) {
		if v.rsp < 2 {
			throw("loop without do")
		}
		i := v.rs[v.rsp-1] + 1
		if i < v.rs[v.rsp-2] {
			v.rs[v.rsp-1] = i
			v.ip = v.code[v.ip]
			return
		}
		v.rsp -= 2
		v.ip++
	})
	v.defPrim("(+loop)", func(v *VM) {
		if v.rsp < 2 {
			throw("+loop without do")
		}
		n := v.pop()
		i := v.rs[v.rsp-1]
		l := v.rs[v.rsp-2]
		ni := i + n
		cont := (n >= 0 && ni < l) || (n < 0 && ni >= l)
		if cont {
			v.rs[v.rsp-1] = ni
			v.ip = v.code[v.ip]
			return
		}
		v.rsp -= 2
		v.ip++
	})
	v.defPrim("unloop", func(v *VM) {
		if v.rsp < 2 {
			throw("unloop without do")
		}
		v.rsp -= 2
	})
	v.defPrim("i", func(v *VM) {
		if v.rsp < 1 {
			throw("i outside of do...loop")
		}
		v.push(v.rs[v.rsp-1])
	})
	v.defPrim("j", func(v *VM) {
		if v.rsp < 3 {
			throw("j outside of nested do...loop")
		}
		v.push(v.rs[v.rsp-3])
	})
}
