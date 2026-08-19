package main

import "fmt"

const banner = `Forthling-Claude 1.0 - a small fast Forth in Go
Type WORDS to list the dictionary, BYE to leave.`

// repl runs the interactive read-eval-print loop.
func (v *VM) repl() int {
	fmt.Fprintln(v.out, banner)
	for {
		if v.state() != 0 {
			v.out.WriteString("... ")
		} else {
			v.out.WriteString("> ")
		}
		v.out.Flush()
		line, ok := v.readLine()
		if !ok {
			v.out.WriteString("\n")
			v.out.Flush()
			return 0
		}
		err := v.Interpret(line, "repl")
		if err != nil {
			if b, isBye := err.(byeError); isBye {
				v.out.Flush()
				return b.code
			}
			fmt.Fprintf(v.out, "\n?? %v\n", err)
		} else if v.state() == 0 {
			v.out.WriteString(" ok\n")
		}
		v.out.Flush()
	}
}
