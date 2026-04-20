// Package lpt drives the LPT1 parallel port as GPIO.
//
// 8 data pins (D0..D7, outputs), 4 control pins (outputs with inverted
// wire polarity on 3 of them), and 5 status pins (inputs).
package lpt

//export spaceos_lpt_present
func cPresent() bool

//export spaceos_lpt_read_data
func cReadData() uint8

//export spaceos_lpt_write_data
func cWriteData(v uint8)

//export spaceos_lpt_read_status
func cReadStatus() uint8

//export spaceos_lpt_read_control
func cReadControl() uint8

//export spaceos_lpt_write_control
func cWriteControl(v uint8)

//export spaceos_lpt_set_bit
func cSetBit(bit uint32, v bool)

// Status bits (as seen after wire-polarity correction in the kernel).
const (
	StatusBusy   = 1 << 7
	StatusAck    = 1 << 6
	StatusPaper  = 1 << 5
	StatusSelect = 1 << 4
	StatusError  = 1 << 3
)

// Control bits.
const (
	CtrlStrobe = 1 << 0
	CtrlAutoFD = 1 << 1
	CtrlInit   = 1 << 2
	CtrlSelIn  = 1 << 3
	CtrlIRQEn  = 1 << 4
)

func Present() bool               { return cPresent() }
func ReadData() byte              { return byte(cReadData()) }
func WriteData(v byte)            { cWriteData(v) }
func ReadStatus() byte            { return byte(cReadStatus()) }
func ReadControl() byte           { return byte(cReadControl()) }
func WriteControl(v byte)         { cWriteControl(v) }
func SetDataPin(bit int, v bool)  { cSetBit(uint32(bit), v) }
