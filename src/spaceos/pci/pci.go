// Package pci enumerates and pokes PCI devices.
package pci

import "unsafe"

type Device struct {
	Bus, Dev, Fn  uint8
	Revision      uint8
	Vendor, DevID uint16
	SubVendor     uint16
	SubID         uint16
	Class         uint8
	Subclass      uint8
	ProgIF        uint8
	HeaderType    uint8
	IRQLine       uint8
	IRQPin        uint8
	BAR           [6]uint32
}

type raw struct {
	Bus, Dev, Fn  uint8
	Revision      uint8
	Vendor, DevID uint16
	SubVendor     uint16
	SubID         uint16
	Class         uint8
	Subclass      uint8
	ProgIF        uint8
	HeaderType    uint8
	IRQLine       uint8
	IRQPin        uint8
	_             [2]uint8
	BAR           [6]uint32
}

//export spaceos_pci_count
func cCount() uintptr

//export spaceos_pci_get
func cGet(i uintptr, out *raw) bool

//export spaceos_pci_cfg_read32
func CfgRead32(bus, dev, fn, off uint8) uint32

//export spaceos_pci_cfg_write32
func CfgWrite32(bus, dev, fn, off uint8, v uint32)

// Devices returns a snapshot of all enumerated PCI devices. Returns
// nil on branches without pci.c linked.
func Devices() []Device {
	n := cCount()
	if n == 0 {
		return nil
	}
	out := make([]Device, 0, n)
	var r raw
	for i := uintptr(0); i < n; i++ {
		if !cGet(i, &r) {
			continue
		}
		out = append(out, Device{
			Bus: r.Bus, Dev: r.Dev, Fn: r.Fn,
			Revision: r.Revision,
			Vendor:   r.Vendor, DevID: r.DevID,
			SubVendor: r.SubVendor, SubID: r.SubID,
			Class: r.Class, Subclass: r.Subclass, ProgIF: r.ProgIF,
			HeaderType: r.HeaderType,
			IRQLine:    r.IRQLine, IRQPin: r.IRQPin,
			BAR: r.BAR,
		})
	}
	return out
}

var _ = unsafe.Sizeof(raw{})
