// Package io provides x86 port and CPU primitives. Kernel-only — do
// not use in code that might ever run in user mode.
package io

//export spaceos_inb
func InB(port uint16) uint8

//export spaceos_inw
func InW(port uint16) uint16

//export spaceos_inl
func InL(port uint16) uint32

//export spaceos_outb
func OutB(port uint16, v uint8)

//export spaceos_outw
func OutW(port uint16, v uint16)

//export spaceos_outl
func OutL(port uint16, v uint32)
