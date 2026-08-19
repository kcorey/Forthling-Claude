package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// A bundled binary is the forth interpreter with a payload appended:
//
//	[ interpreter bytes ][ script bytes ][ length uint64 LE ][ "FTHBUNDL" ]
//
// The trailer is 16 bytes, so detection costs one seek and one 16-byte read.
const bundleMagic = "FTHBUNDL"

const trailerSize = 8 + len(bundleMagic)

// payloadInfo describes a bundle found in the file at path.
type payloadInfo struct {
	script []byte
	// coreLen is the size of the interpreter portion, i.e. where the payload starts.
	coreLen int64
}

// readPayloadFrom looks for a bundle trailer at the end of path.
func readPayloadFrom(path string) (*payloadInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size < int64(trailerSize) {
		return nil, nil
	}
	tail := make([]byte, trailerSize)
	if _, err := f.ReadAt(tail, size-int64(trailerSize)); err != nil {
		return nil, err
	}
	if string(tail[8:]) != bundleMagic {
		return nil, nil
	}
	n := int64(binary.LittleEndian.Uint64(tail[:8]))
	coreLen := size - int64(trailerSize) - n
	if n < 0 || coreLen < 0 {
		return nil, fmt.Errorf("corrupt bundle trailer")
	}
	script := make([]byte, n)
	if _, err := f.ReadAt(script, coreLen); err != nil {
		return nil, err
	}
	return &payloadInfo{script: script, coreLen: coreLen}, nil
}

// readPayload checks the running executable for an embedded script.
func readPayload() *payloadInfo {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	p, err := readPayloadFrom(exe)
	if err != nil {
		return nil
	}
	return p
}

// writeBundle copies the interpreter part of the running executable to out and
// appends script. Re-bundling an already-bundled binary replaces the payload
// instead of nesting it.
func writeBundle(script []byte, out string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return writeBundleFrom(exe, script, out)
}

// writeBundleFrom builds a bundle from the interpreter at exe. If exe is
// itself a bundle its payload is stripped first, so bundles never nest.
func writeBundleFrom(exe string, script []byte, out string) error {
	coreLen := int64(-1)
	if p, err := readPayloadFrom(exe); err == nil && p != nil {
		coreLen = p.coreLen
	}
	src, err := os.Open(exe)
	if err != nil {
		return err
	}
	defer src.Close()
	if coreLen < 0 {
		st, err := src.Stat()
		if err != nil {
			return err
		}
		coreLen = st.Size()
	}

	tmp := out + ".tmp"
	dst, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.CopyN(dst, src, coreLen); err != nil {
		dst.Close()
		os.Remove(tmp)
		return err
	}
	if _, err := dst.Write(script); err != nil {
		dst.Close()
		os.Remove(tmp)
		return err
	}
	trailer := make([]byte, trailerSize)
	binary.LittleEndian.PutUint64(trailer[:8], uint64(len(script)))
	copy(trailer[8:], bundleMagic)
	if _, err := dst.Write(trailer); err != nil {
		dst.Close()
		os.Remove(tmp)
		return err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, out)
}
