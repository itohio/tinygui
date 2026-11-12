// Package framebuffer is an in-memory RGB565 display for TinyGUI examples.
//
// It implements every interface TinyGUI type-asserts for (ui.Displayer), clips
// all writes to its bounds and hands the finished pixels to a flush callback on
// Display. Boards push them to their panel; the host simulator encodes them as
// PNG. Drawing into RAM and pushing once removes the clear-then-draw flicker of
// unbuffered SPI/parallel panels.
//
// The buffer may cover only a horizontal band of the screen (NewBand): the app
// draws the whole UI once per band and every write outside the band is clipped.
// Two 120-row bands of a 320×240 screen need 76.8 KB instead of 153.6 KB.
package framebuffer

import "image/color"

// RGB565 is a W×H band of native-endian RGB565 pixels, row-major, covering
// screen rows Y0 … Y0+H-1 of a W×ScreenH screen.
type RGB565 struct {
	W, H    int16 // buffer size
	ScreenH int16 // full screen height reported by Size
	Y0      int16 // first screen row held in the buffer
	Pix     []uint16
	flush   func(f *RGB565) error
}

// New wraps pix (len ≥ w*h) as a full w×h frame. flush may be nil.
func New(w, h int16, pix []uint16, flush func(f *RGB565) error) *RGB565 {
	return NewBand(w, h, h, pix, flush)
}

// NewBand wraps pix (len ≥ w*bandH) as a band of a w×screenH screen. Select the
// band with SetY0 before drawing it.
func NewBand(w, screenH, bandH int16, pix []uint16, flush func(f *RGB565) error) *RGB565 {
	return &RGB565{W: w, H: bandH, ScreenH: screenH, Pix: pix[:int(w)*int(bandH)], flush: flush}
}

// SetY0 selects which screen rows the buffer holds next.
func (f *RGB565) SetY0(y0 int16) { f.Y0 = y0 }

// Bands is the number of bands that cover the screen.
func (f *RGB565) Bands() int { return (int(f.ScreenH) + int(f.H) - 1) / int(f.H) }

// To565 converts an RGBA colour to RGB565 (alpha ignored).
func To565(c color.RGBA) uint16 {
	return uint16(c.R&0xF8)<<8 | uint16(c.G&0xFC)<<3 | uint16(c.B)>>3
}

// From565 expands an RGB565 pixel to RGBA.
func From565(p uint16) color.RGBA {
	r := uint8(p>>11) & 0x1F
	g := uint8(p>>5) & 0x3F
	b := uint8(p) & 0x1F
	return color.RGBA{R: r<<3 | r>>2, G: g<<2 | g>>4, B: b<<3 | b>>2, A: 255}
}

// Size reports the full screen, not the band.
func (f *RGB565) Size() (int16, int16) { return f.W, f.ScreenH }

func (f *RGB565) SetPixel(x, y int16, c color.RGBA) {
	y -= f.Y0
	if x < 0 || y < 0 || x >= f.W || y >= f.H {
		return
	}
	f.Pix[int(y)*int(f.W)+int(x)] = To565(c)
}

// Display hands the band to the flush callback.
func (f *RGB565) Display() error {
	if f.flush == nil {
		return nil
	}
	return f.flush(f)
}

// clip converts a screen rectangle to buffer coordinates and intersects it with
// the band; ok is false when nothing is left.
func (f *RGB565) clip(x, y, w, h int16) (int, int, int, int, bool) {
	x0, y0 := int(x), int(y)-int(f.Y0)
	x1, y1 := x0+int(w), y0+int(h)
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > int(f.W) {
		x1 = int(f.W)
	}
	if y1 > int(f.H) {
		y1 = int(f.H)
	}
	return x0, y0, x1, y1, x0 < x1 && y0 < y1
}

func (f *RGB565) FillRectangle(x, y, w, h int16, c color.RGBA) error {
	x0, y0, x1, y1, ok := f.clip(x, y, w, h)
	if !ok {
		return nil
	}
	p := To565(c)
	stride := int(f.W)
	for row := y0; row < y1; row++ {
		line := f.Pix[row*stride+x0 : row*stride+x1]
		for i := range line {
			line[i] = p
		}
	}
	return nil
}

func (f *RGB565) FillRectangleWithBuffer(x, y, w, h int16, buf []color.RGBA) error {
	if int(w)*int(h) > len(buf) {
		return nil
	}
	for j := int16(0); j < h; j++ {
		for i := int16(0); i < w; i++ {
			f.SetPixel(x+i, y+j, buf[int(j)*int(w)+int(i)])
		}
	}
	return nil
}

// FillScreen fills the band (the part of the screen the buffer holds).
func (f *RGB565) FillScreen(c color.RGBA) {
	p := To565(c)
	for i := range f.Pix {
		f.Pix[i] = p
	}
}

func (f *RGB565) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	_ = f.FillRectangle(x0, y, x1-x0+1, 1, c)
}

func (f *RGB565) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	_ = f.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawRGBBitmap copies native RGB565 pixels; pixels equal to 0 are drawn too.
func (f *RGB565) DrawRGBBitmap(x, y int16, data []uint16, w, h int16) error {
	if int(w)*int(h) > len(data) {
		return nil
	}
	x0, y0, x1, y1, ok := f.clip(x, y, w, h)
	if !ok {
		return nil
	}
	stride := int(f.W)
	top := int(y) - int(f.Y0) // bitmap row 0 in buffer coordinates
	for row := y0; row < y1; row++ {
		src := data[(row-top)*int(w)+(x0-int(x)):]
		copy(f.Pix[row*stride+x0:row*stride+x1], src[:x1-x0])
	}
	return nil
}

// DrawRGBBitmap8 draws one byte per pixel as RGB332 (TinyGUI's Bitmap8 passes
// w*h bytes; tinygo drivers interpret the same method as raw RGB565 bytes).
func (f *RGB565) DrawRGBBitmap8(x, y int16, data []uint8, w, h int16) error {
	if int(w)*int(h) > len(data) {
		return nil
	}
	for j := int16(0); j < h; j++ {
		for i := int16(0); i < w; i++ {
			v := data[int(j)*int(w)+int(i)]
			r := v >> 5
			g := (v >> 2) & 7
			b := v & 3
			f.SetPixel(x+i, y+j, color.RGBA{R: r<<5 | r<<2 | r>>1, G: g<<5 | g<<2 | g>>1, B: b * 85, A: 255})
		}
	}
	return nil
}
