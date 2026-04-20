// Package fs is a Go-flavoured wrapper over the kernel's FatFs glue.
//
//	f, err := fs.Open("/boot/config.txt", fs.Read)
//	defer f.Close()
//	var buf [128]byte
//	n, _ := f.Read(buf[:])
package fs

import (
	"errors"
	"io"
)

// FatFs f_open mode bits.
const (
	Read         uint32 = 0x01
	Write        uint32 = 0x02
	OpenExisting uint32 = 0x00
	CreateNew    uint32 = 0x04
	CreateAlways uint32 = 0x08
	OpenAlways   uint32 = 0x10
	OpenAppend   uint32 = 0x30
)

var (
	ErrOpen   = errors.New("spaceos/fs: open failed")
	ErrIO     = errors.New("spaceos/fs: i/o error")
	ErrClosed = errors.New("spaceos/fs: file is closed")
)

//export spaceos_fs_open
func cOpen(path *byte, mode uint32) int32

//export spaceos_fs_read
func cRead(h int32, buf *byte, cap uint32) int32

//export spaceos_fs_write
func cWrite(h int32, buf *byte, n uint32) int32

//export spaceos_fs_seek
func cSeek(h int32, off uint64) int32

//export spaceos_fs_close
func cClose(h int32)

//export spaceos_fs_unlink
func cUnlink(path *byte) int32

type File struct {
	h int32
}

func cpath(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}

// Open returns a file handle. Mode is a bitmask; see constants above.
func Open(path string, mode uint32) (*File, error) {
	h := cOpen(cpath(path), mode)
	if h <= 0 {
		return nil, ErrOpen
	}
	return &File{h: h}, nil
}

func (f *File) Read(p []byte) (int, error) {
	if f.h == 0 {
		return 0, ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := cRead(f.h, &p[0], uint32(len(p)))
	if n < 0 {
		return 0, ErrIO
	}
	if n == 0 {
		return 0, io.EOF
	}
	return int(n), nil
}

func (f *File) Write(p []byte) (int, error) {
	if f.h == 0 {
		return 0, ErrClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := cWrite(f.h, &p[0], uint32(len(p)))
	if n < 0 {
		return 0, ErrIO
	}
	return int(n), nil
}

// Seek sets the absolute file offset. FatFs doesn't distinguish
// whence, so callers compute the final offset themselves.
func (f *File) Seek(abs uint64) error {
	if f.h == 0 {
		return ErrClosed
	}
	if cSeek(f.h, abs) != 0 {
		return ErrIO
	}
	return nil
}

func (f *File) Close() error {
	if f.h == 0 {
		return nil
	}
	h := f.h
	f.h = 0
	cClose(h)
	return nil
}

// Remove deletes a file.
func Remove(path string) error {
	if cUnlink(cpath(path)) != 0 {
		return ErrIO
	}
	return nil
}
