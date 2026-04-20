// Package input exposes PS/2 keyboard and mouse events.
//
// On branches without the kernel's ps2kbd.c/ps2mouse.c linked in, the
// queues simply never fill, so ReadKey / ReadMouse block forever or
// return false on timeout.
package input

//export spaceos_kbd_read
func cKbdRead(out *uint8, timeoutMs int32) int32

//export spaceos_mouse_read
func cMouseRead(dx, dy *int16, buttons *uint8, timeoutMs int32) int32

// ReadKey returns one scancode-set-1 byte. ok=false on timeout or when
// no PS/2 driver is linked.
func ReadKey(timeoutMs int) (scan byte, ok bool) {
	var v uint8
	if cKbdRead(&v, int32(timeoutMs)) == 1 {
		return v, true
	}
	return 0, false
}

type MouseEvent struct {
	DX, DY  int16
	Buttons uint8 // bit0=left, bit1=right, bit2=middle
}

func ReadMouse(timeoutMs int) (MouseEvent, bool) {
	var e MouseEvent
	if cMouseRead(&e.DX, &e.DY, &e.Buttons, int32(timeoutMs)) == 1 {
		return e, true
	}
	return MouseEvent{}, false
}
