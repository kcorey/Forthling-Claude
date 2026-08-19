package main

import "strings"

// lookup finds a word by name, case-insensitively, ignoring hidden entries.
func (v *VM) lookup(name string) (int32, bool) {
	xt, ok := v.dict[strings.ToLower(name)]
	if !ok {
		return 0, false
	}
	if v.words[xt].hidden {
		return 0, false
	}
	return xt, true
}

func (v *VM) addWord(w Word) int32 {
	xt := int32(len(v.words))
	v.words = append(v.words, w)
	v.dict[strings.ToLower(w.name)] = xt
	return xt
}

// defPrim defines a primitive implemented in Go.
func (v *VM) defPrim(name string, fn func(*VM)) int32 {
	return v.addWord(Word{name: name, kind: kPrim, prim: fn})
}

// defImm defines an immediate primitive (a compiler directive).
func (v *VM) defImm(name string, fn func(*VM)) int32 {
	return v.addWord(Word{name: name, kind: kPrim, prim: fn, imm: true})
}

// create makes a data-field word (VARIABLE / CREATE / CONSTANT base).
func (v *VM) create(name string, data int64) int32 {
	return v.addWord(Word{name: name, kind: kCreate, data: data})
}

// words returns every visible word name in definition order.
func (v *VM) wordNames() []string {
	out := make([]string, 0, len(v.words))
	seen := make(map[string]bool, len(v.words))
	for i := len(v.words) - 1; i >= 0; i-- {
		w := &v.words[i]
		if w.hidden || seen[w.name] {
			continue
		}
		// runtime helpers such as (lit) and (branch) are implementation detail
		if len(w.name) > 1 && w.name[0] == '(' && w.name[len(w.name)-1] == ')' {
			continue
		}
		seen[w.name] = true
		out = append(out, w.name)
	}
	// reverse into definition order
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
