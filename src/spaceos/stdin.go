package spaceos

import _ "runtime"

//export spaceos_stdin_read
func cStdinRead(buf *byte, cap uintptr, timeoutMs int32) int32

// ReadInput reads input fed into the Go task by the kernel's CLI or
// telnet driver. Returns 0 on timeout, -1 if no stdin pipe exists.
// timeoutMs<0 means block.
func ReadInput(p []byte, timeoutMs int) int {
	if len(p) == 0 {
		return 0
	}
	return int(cStdinRead(&p[0], uintptr(len(p)), int32(timeoutMs)))
}
