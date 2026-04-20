// Package ata provides raw block I/O against the kernel's ATA/IDE driver.
package ata

import "errors"

const SectorBytes = 512

type Drive struct {
	Index          int
	Present        bool
	LBA48          bool
	BytesPerSector uint16
	TotalSectors   uint64
	Model          string
	Serial         string
	Firmware       string
}

type raw struct {
	Present        uint8
	LBA48          uint8
	BytesPerSector uint16
	TotalSectors   uint64
	Model          [41]byte
	Serial         [21]byte
	Firmware       [9]byte
}

//export spaceos_ata_info
func cInfo(drive int32, out *raw) bool

//export spaceos_ata_read
func cRead(drive int32, lba uint64, n uint32, buf *byte) int64

//export spaceos_ata_write
func cWrite(drive int32, lba uint64, n uint32, buf *byte) int64

func cstring(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// Open returns the drive at index (0..3). ok=false when the slot is empty.
func Open(index int) (Drive, bool) {
	var r raw
	if !cInfo(int32(index), &r) || r.Present == 0 {
		return Drive{}, false
	}
	return Drive{
		Index:          index,
		Present:        true,
		LBA48:          r.LBA48 != 0,
		BytesPerSector: r.BytesPerSector,
		TotalSectors:   r.TotalSectors,
		Model:          cstring(r.Model[:]),
		Serial:         cstring(r.Serial[:]),
		Firmware:       cstring(r.Firmware[:]),
	}, true
}

var ErrIO = errors.New("spaceos/ata: i/o error")

// Read fills buf with n sectors starting at LBA. len(buf) must be n*512.
func (d Drive) Read(lba uint64, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf)%SectorBytes != 0 {
		return 0, errors.New("spaceos/ata: buffer size must be a multiple of 512")
	}
	n := uint32(len(buf) / SectorBytes)
	got := cRead(int32(d.Index), lba, n, &buf[0])
	if got < 0 {
		return 0, ErrIO
	}
	return int(got) * SectorBytes, nil
}

// Write pushes len(buf)/512 sectors to LBA.
func (d Drive) Write(lba uint64, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf)%SectorBytes != 0 {
		return 0, errors.New("spaceos/ata: buffer size must be a multiple of 512")
	}
	n := uint32(len(buf) / SectorBytes)
	got := cWrite(int32(d.Index), lba, n, &buf[0])
	if got < 0 {
		return 0, ErrIO
	}
	return int(got) * SectorBytes, nil
}
