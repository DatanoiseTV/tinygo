// Package serial is a thin wrapper over COM1 (0x3F8).
package serial

//export spaceos_serial_write
func cWrite(buf *byte, n uintptr)

//export spaceos_serial_try_read
func cTryRead() int32

// Port is COM1. Write satisfies io.Writer, TryRead is non-blocking.
type Port struct{}

func (Port) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	cWrite(&p[0], uintptr(len(p)))
	return len(p), nil
}

// TryRead returns the next RX byte, or -1 if the FIFO is empty.
func (Port) TryRead() int {
	return int(cTryRead())
}
