package spaceos

import "unsafe"

// Task spawns f on a new FreeRTOS task. Use this as a surrogate for
// `go f()` until real goroutines are supported.
//
//	spaceos.Task("probe", 0, 0, func() {
//	    for {
//	        check()
//	        time.Sleep(time.Second)
//	    }
//	})
//
// stackWords of 0 defaults to 8*configMINIMAL_STACK_SIZE; priority of
// 0 defaults to 2. Returns false if the kernel is out of heap.
//
// Caveat: call Task from the main Go task only, or coordinate calls
// externally — the internal closure-keep list isn't currently
// concurrency-safe.
func Task(name string, stackWords, priority uint32, f func()) bool {
	c := &taskCtx{fn: f}
	keeps = append(keeps, c)

	nb := make([]byte, len(name)+1)
	copy(nb, name)

	return cTaskSpawn(unsafe.Pointer(c), &nb[0], stackWords, priority)
}

type taskCtx struct {
	fn func()
}

// Keep a reference so the GC (when we get one) doesn't reclaim the
// closure before the task finishes with it.
var keeps = make([]*taskCtx, 0, 32)

//export spaceos_task_spawn
func cTaskSpawn(ctx unsafe.Pointer, name *byte, stackWords, priority uint32) bool

// Callback run from a new FreeRTOS task by the C trampoline.
//
//export spaceos_run_task_ctx
func runTaskCtx(p unsafe.Pointer) {
	c := (*taskCtx)(p)
	c.fn()
}
