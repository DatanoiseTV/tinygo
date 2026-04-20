// Package nic exposes the bound ethernet controller's state.
package nic

type Info struct {
	Name          string
	MAC           [6]byte
	LinkUp        bool
	FullDuplex    bool
	LinkSpeedMbps uint16
	TxFrames      uint64
	TxBytes       uint64
	RxFrames      uint64
	RxBytes       uint64
	IRQs          uint64
}

type raw struct {
	MAC            [6]byte
	LinkUp         uint8
	FullDuplex     uint8
	LinkSpeedMbps  uint16
	_              uint16
	TxFrames       uint64
	TxBytes        uint64
	RxFrames       uint64
	RxBytes        uint64
	IRQs           uint64
	Name           uintptr // C string; we ignore for now — kernel owns it.
}

//export spaceos_nic_info
func cNICInfo(out *raw) bool

// Current returns a snapshot of the bound NIC, or ok=false if none is bound.
func Current() (Info, bool) {
	var r raw
	if !cNICInfo(&r) {
		return Info{}, false
	}
	return Info{
		MAC:           r.MAC,
		LinkUp:        r.LinkUp != 0,
		FullDuplex:    r.FullDuplex != 0,
		LinkSpeedMbps: r.LinkSpeedMbps,
		TxFrames:      r.TxFrames,
		TxBytes:       r.TxBytes,
		RxFrames:      r.RxFrames,
		RxBytes:       r.RxBytes,
		IRQs:          r.IRQs,
	}, true
}
