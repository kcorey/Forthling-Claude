//go:build !darwin && !linux

package main

// Platforms without a termios implementation stay line-buffered. Samples still
// run; they just need Enter after each keystroke.
func setRaw(on bool) error { return nil }

func restoreTerminal() {}
