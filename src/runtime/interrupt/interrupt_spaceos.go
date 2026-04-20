//go:build spaceos

package interrupt

// On SpaceOS, Go runs inside a single FreeRTOS task with scheduler=none, so
// there are no Go-level goroutines or Go interrupts to disable. Any real
// critical section must be entered via taskENTER_CRITICAL on the kernel
// side. These stubs let the runtime's lockAtomics/unlockAtomics compile.

type State uintptr

func Disable() (state State) {
	return 0
}

func Restore(state State) {}

func In() bool {
	return false
}
