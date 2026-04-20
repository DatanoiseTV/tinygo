// Package fb gives Go direct access to SpaceOS's linear framebuffer.
//
// Drawing happens into a back-buffer (VFB) in RAM. Call Flush or
// FlushAll to push dirty regions to the visible framebuffer in a
// single tight memcpy.
//
//	s, ok := fb.Surface()
//	if !ok { return } // no framebuffer (legacy VGA-text boot, etc.)
//
//	// Write RGB pixels anywhere you like.
//	s.SetPixel(10, 20, fb.RGB(0xff, 0x80, 0x00))
//
//	// Bulk fill by writing to the underlying byte slice directly.
//	for y := 0; y < 100; y++ {
//	    row := s.Row(y)
//	    for x := 0; x < 100*4; x += 4 {
//	        row[x+0] = 0xff   // B
//	        row[x+1] = 0x00   // G
//	        row[x+2] = 0xff   // R
//	    }
//	}
//	s.FlushAll()
//
// Byte order is BGRA (little-endian 0xAARRGGBB), which matches every
// BIOS/GRUB framebuffer we've run into on x86.
package fb

import "unsafe"

type rawInfo struct {
	Base     uintptr
	Pitch    uint32
	Cols     uint32
	Rows     uint32
	WidthPx  uint32
	HeightPx uint32
	BPP      uint32
	_pad     uint32
}

//export spaceos_fb_info
func cFBInfo(out *rawInfo) bool

//export spaceos_fb_flush
func cFBFlush(x, y, w, h uint32)

//export spaceos_fb_flush_all
func cFBFlushAll()

// Surface is a handle to the VFB.
type Surface struct {
	pixels   []byte // memory-mapped slice aliasing the real back-buffer
	pitch    uint32
	width    uint32
	height   uint32
	bytesPP  uint32
}

// RGB packs r, g, b into a framebuffer pixel (BGRA 32bpp).
func RGB(r, g, b byte) uint32 {
	return uint32(b) | uint32(g)<<8 | uint32(r)<<16 | 0xff000000
}

// Surface returns the current framebuffer back-buffer, or ok=false if
// no framebuffer is available (VGA-text-only boot, or fb not linked).
func SurfaceOpen() (*Surface, bool) {
	var r rawInfo
	if !cFBInfo(&r) {
		return nil, false
	}
	if r.Base == 0 || r.Pitch == 0 {
		return nil, false
	}
	n := int(r.Pitch) * int(r.HeightPx)
	pixels := unsafe.Slice((*byte)(unsafe.Pointer(r.Base)), n)
	bpp := r.BPP
	if bpp == 0 {
		bpp = 32
	}
	return &Surface{
		pixels:  pixels,
		pitch:   r.Pitch,
		width:   r.WidthPx,
		height:  r.HeightPx,
		bytesPP: bpp / 8,
	}, true
}

func (s *Surface) Width() int  { return int(s.width) }
func (s *Surface) Height() int { return int(s.height) }
func (s *Surface) Pitch() int  { return int(s.pitch) }
func (s *Surface) BytesPerPixel() int { return int(s.bytesPP) }

// Pixels returns the raw back-buffer bytes. Layout: pitch * height,
// BGRA at bpp=32. Modifications are only visible after Flush.
func (s *Surface) Pixels() []byte { return s.pixels }

// Row returns a slice for pixel row y.
func (s *Surface) Row(y int) []byte {
	off := y * int(s.pitch)
	return s.pixels[off : off+int(s.pitch)]
}

// SetPixel writes a 32bpp BGRA pixel. No bounds check in the hot path
// beyond the slice bounds check Go does for us.
func (s *Surface) SetPixel(x, y int, px uint32) {
	if uint32(x) >= s.width || uint32(y) >= s.height {
		return
	}
	off := y*int(s.pitch) + x*int(s.bytesPP)
	if s.bytesPP == 4 {
		*(*uint32)(unsafe.Pointer(&s.pixels[off])) = px
	} else { // 24 bpp
		s.pixels[off+0] = byte(px)
		s.pixels[off+1] = byte(px >> 8)
		s.pixels[off+2] = byte(px >> 16)
	}
}

// Pixel reads a 32bpp BGRA pixel.
func (s *Surface) Pixel(x, y int) uint32 {
	if uint32(x) >= s.width || uint32(y) >= s.height {
		return 0
	}
	off := y*int(s.pitch) + x*int(s.bytesPP)
	if s.bytesPP == 4 {
		return *(*uint32)(unsafe.Pointer(&s.pixels[off]))
	}
	return uint32(s.pixels[off]) |
		uint32(s.pixels[off+1])<<8 |
		uint32(s.pixels[off+2])<<16
}

// FillRect paints (w,h) pixels starting at (x,y) in px. Clips to the
// surface. Does not flush.
func (s *Surface) FillRect(x, y, w, h int, px uint32) {
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x >= int(s.width) || y >= int(s.height) || w <= 0 || h <= 0 {
		return
	}
	if x+w > int(s.width) {
		w = int(s.width) - x
	}
	if y+h > int(s.height) {
		h = int(s.height) - y
	}
	bpp := int(s.bytesPP)
	pitch := int(s.pitch)
	for row := y; row < y+h; row++ {
		off := row*pitch + x*bpp
		if bpp == 4 {
			end := off + w*4
			for i := off; i < end; i += 4 {
				*(*uint32)(unsafe.Pointer(&s.pixels[i])) = px
			}
		} else {
			for i := 0; i < w; i++ {
				p := off + i*3
				s.pixels[p+0] = byte(px)
				s.pixels[p+1] = byte(px >> 8)
				s.pixels[p+2] = byte(px >> 16)
			}
		}
	}
}

// Clear fills the entire surface with px.
func (s *Surface) Clear(px uint32) {
	s.FillRect(0, 0, int(s.width), int(s.height), px)
}

// Flush pushes a dirty rect from the back-buffer to the visible fb.
func (s *Surface) Flush(x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	cFBFlush(uint32(x), uint32(y), uint32(w), uint32(h))
}

// FlushAll pushes the entire back-buffer.
func (s *Surface) FlushAll() { cFBFlushAll() }
