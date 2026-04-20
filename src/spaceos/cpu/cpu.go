// Package cpu exposes a CPUID-derived snapshot captured at boot.
package cpu

type Info struct {
	Vendor       string
	Brand        string
	Family       uint32
	Model        uint32
	Stepping     uint32
	LogicalCPUs  uint32
	BaseMHz      uint32
	MaxMHz       uint32
	TSCHz        uint64

	HasTSC, HasSSE2, HasAVX, HasAVX2         bool
	HasRDTSCP, HasInvariantTSC, HasX2APIC    bool
	InHypervisor                              bool
}

type raw struct {
	Vendor       [16]byte
	Brand        [52]byte
	Family       uint32
	Model        uint32
	Stepping     uint32
	LogicalCPUs  uint32
	BaseMHz      uint32
	MaxMHz       uint32
	TSCHz        uint64
	Flags        uint32
	_pad         uint32
}

const (
	flagTSC        = 1 << 0
	flagSSE2       = 1 << 1
	flagAVX        = 1 << 2
	flagAVX2       = 1 << 3
	flagRDTSCP     = 1 << 4
	flagINVTSC     = 1 << 5
	flagX2APIC     = 1 << 6
	flagHypervisor = 1 << 7
)

//export spaceos_cpuinfo
func cCPUInfo(out *raw)

func cstr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

func Current() Info {
	var r raw
	cCPUInfo(&r)
	return Info{
		Vendor:          cstr(r.Vendor[:]),
		Brand:           cstr(r.Brand[:]),
		Family:          r.Family,
		Model:           r.Model,
		Stepping:        r.Stepping,
		LogicalCPUs:     r.LogicalCPUs,
		BaseMHz:         r.BaseMHz,
		MaxMHz:          r.MaxMHz,
		TSCHz:           r.TSCHz,
		HasTSC:          r.Flags&flagTSC != 0,
		HasSSE2:         r.Flags&flagSSE2 != 0,
		HasAVX:          r.Flags&flagAVX != 0,
		HasAVX2:         r.Flags&flagAVX2 != 0,
		HasRDTSCP:       r.Flags&flagRDTSCP != 0,
		HasInvariantTSC: r.Flags&flagINVTSC != 0,
		HasX2APIC:       r.Flags&flagX2APIC != 0,
		InHypervisor:    r.Flags&flagHypervisor != 0,
	}
}
