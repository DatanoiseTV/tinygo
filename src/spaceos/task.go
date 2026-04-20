package spaceos

// Task spawns f on a new FreeRTOS task. Use this as a surrogate for
// `go f()` — with scheduler=none, real goroutines aren't supported.
//
//	spaceos.Task("probe", 0, 0, func() {
//	    for {
//	        check()
//	        time.Sleep(time.Second)
//	    }
//	})
//
// stackWords of 0 defaults to 8*configMINIMAL_STACK_SIZE; priority of
// 0 defaults to 2. Returns false if the kernel is out of heap or we've
// exhausted the 128-slot registry (happens after 128 spawns in one
// boot, which would be a bug in the caller).
//
// Caveat: the registry isn't concurrency-safe — serialise Task() calls
// externally if multiple tasks might spawn at once.
func Task(name string, stackWords, priority uint32, f func()) bool {
	// Monotonic allocation — we never reuse a slot, so the in-flight
	// task always sees its own entry. Wraps at registrySize; since
	// tasks never come back to free slots, a long-running program must
	// stay under 128 total Task() calls.
	if nextID >= registrySize {
		return false
	}
	id := nextID
	nextID++
	registry[id] = f

	nb := make([]byte, len(name)+1)
	copy(nb, name)

	if !cTaskSpawn(uintptr(id), &nb[0], stackWords, priority) {
		registry[id] = nil
		return false
	}
	return true
}

const registrySize = 128

var (
	registry [registrySize]func()
	nextID   uint32
)

//export spaceos_task_spawn
func cTaskSpawn(id uintptr, name *byte, stackWords, priority uint32) bool

//export spaceos_run_task_id
func runTaskID(id uintptr) {
	if id >= registrySize {
		return
	}
	f := registry[id]
	if f != nil {
		f()
	}
}
