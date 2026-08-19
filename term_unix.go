//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var savedTermios syscall.Termios
var haveSaved bool

func ioctlTermios(fd uintptr, req uintptr, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return errno
	}
	return nil
}

// setRaw switches the terminal between raw (character-at-a-time, no echo) and
// its original mode. It is a no-op when stdin is not a terminal.
func setRaw(on bool) error {
	fd := os.Stdin.Fd()
	if !on {
		if !haveSaved {
			return nil
		}
		t := savedTermios
		return ioctlTermios(fd, ioctlSet, &t)
	}
	var t syscall.Termios
	if err := ioctlTermios(fd, ioctlGet, &t); err != nil {
		return err
	}
	if !haveSaved {
		savedTermios = t
		haveSaved = true
	}
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN
	t.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0
	return ioctlTermios(fd, ioctlSet, &t)
}

// restoreTerminal is called on exit and on signals.
func restoreTerminal() {
	if haveSaved {
		t := savedTermios
		_ = ioctlTermios(os.Stdin.Fd(), ioctlSet, &t)
		os.Stdout.WriteString("\x1b[?25h")
	}
}
