package cpu

//export spaceos_rdtsc
func RDTSC() uint64

//export spaceos_rdtscp
func cRdtscp(aux *uint32) uint64

// RDTSCP returns the TSC and the AUX register contents (CPU ID on most
// platforms).
func RDTSCP() (tsc uint64, aux uint32) {
	tsc = cRdtscp(&aux)
	return
}

// CPUIDRegs holds EAX, EBX, ECX, EDX from a CPUID invocation.
type CPUIDRegs struct{ EAX, EBX, ECX, EDX uint32 }

//export spaceos_cpuid
func cCpuid(leaf, sub uint32, regs *[4]uint32)

func CPUID(leaf, sub uint32) CPUIDRegs {
	var r [4]uint32
	cCpuid(leaf, sub, &r)
	return CPUIDRegs{r[0], r[1], r[2], r[3]}
}

//export spaceos_rdmsr
func RDMSR(msr uint32) uint64

//export spaceos_wrmsr
func WRMSR(msr uint32, val uint64)

//export spaceos_rdrand
func cRdrand(out *uint64) bool

// Rand pulls one 64-bit value from RDRAND. ok=false means retry
// exhaustion (rare on real silicon, somewhat common under QEMU without
// -cpu host).
func Rand() (v uint64, ok bool) {
	ok = cRdrand(&v)
	return
}
