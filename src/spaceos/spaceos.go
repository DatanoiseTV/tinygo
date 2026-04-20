// Package spaceos exposes the SpaceOS kernel's peripheral APIs to Go
// programs built with the `spaceos-amd64` TinyGo target. Subpackages
// cover specific subsystems:
//
//   - spaceos/net:    TCP and UDP through the lwIP stack
//   - spaceos/nic:    ethernet controller state and counters
//   - spaceos/cpu:    CPUID-derived info
//   - spaceos/lpt:    IEEE-1284 parallel port GPIO
//   - spaceos/serial: COM1 UART
//
// Go code runs inside a FreeRTOS task whose entry point is the Go
// main.main. The kernel is what started it, so there's no graceful
// program-exit semantics — returning from main just ends the task.
package spaceos
