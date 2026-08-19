//go:build !darwin && !linux

package main

// Platforms without the TIOCGWINSZ ioctl always fall back to the default size.
func termSize() (cols, rows int, ok bool) { return 0, 0, false }
