//go:build spaceos

package runtime

// Go runtime for SpaceOS (freertos-x86_64).
//
// The kernel is a C program. TinyGo produces a relocatable object file which
// the kernel's Makefile links into the final ELF. A FreeRTOS task created by
// the kernel calls spaceos_go_start, which runs package init and main.main.
//
// Time, sleep, console I/O, and abort are delegated to the kernel via extern
// C functions (spaceos_*). The heap, globals, and stack_top symbols
// referenced by baremetal.go come from src/kernel/tinygo_glue.c on the
// kernel side.

// Entry point invoked from a FreeRTOS task on the kernel side.
//
//export spaceos_go_start
func main() {
	preinit()
	run()
}

func preinit() {
	// Boot code already set up long mode, paging, and a stack for the
	// calling task; there's nothing to initialize from the Go side.
}

// Kernel-provided C functions (see src/kernel/tinygo_glue.c).
//
//export spaceos_putchar
func spaceos_putchar(c uint8)

//export spaceos_getchar
func spaceos_getchar() int32

//export spaceos_ticks_ns
func spaceos_ticks_ns() int64

//export spaceos_sleep_ns
func spaceos_sleep_ns(ns int64)

//export spaceos_abort
func spaceos_abort()

func putchar(c byte) {
	spaceos_putchar(uint8(c))
}

func getchar() byte {
	r := spaceos_getchar()
	if r < 0 {
		return 0
	}
	return byte(r)
}

func buffered() int {
	return 0
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func ticks() timeUnit {
	return timeUnit(spaceos_ticks_ns())
}

func sleepTicks(d timeUnit) {
	spaceos_sleep_ns(int64(d))
}

func abort() {
	spaceos_abort()
	for {
	}
}

func exit(code int) {
	abort()
}
