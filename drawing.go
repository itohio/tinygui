package ui

import (
	"errors"
	"image/color"
	"strings"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/image/png"
)

type BitmapDisplayer interface {
	DrawRGBBitmap(x int16, y int16, data []uint16, w int16, h int16) error
	DrawRGBBitmap8(x int16, y int16, data []uint8, w int16, h int16) error
}

type LineDisplayer interface {
	DrawFastHLine(x0 int16, x1 int16, y int16, c color.RGBA)
	DrawFastVLine(x int16, y0 int16, y1 int16, c color.RGBA)
}

type RectangleDisplayer interface {
	FillRectangle(x int16, y int16, width int16, height int16, c color.RGBA) error
	FillRectangleWithBuffer(x int16, y int16, width int16, height int16, buffer []color.RGBA) error
	FillScreen(c color.RGBA)
}

// Clipper is implemented by displays that hold only part of the screen at a
// time, such as one band of a framebuffer drawn band by band. Widgets that
// draw many primitives can skip the ones entirely outside Clip; drawing
// outside it is still safe, only wasted work.
type Clipper interface {
	Clip() (x, y, w, h int16)
}

type Displayer interface {
	BitmapDisplayer
	LineDisplayer
	RectangleDisplayer
	drivers.Displayer
}

// HLine draws a horizontal line one pixel thick on any displayer.
func HLine(d drivers.Displayer, x, y, w int16, c color.RGBA) {
	for w > 0 {
		d.SetPixel(x, y, c)
		x++
		w--
	}
}

// VLine draws a vertical line one pixel thick on any displayer.
func VLine(d drivers.Displayer, x, y, h int16, c color.RGBA) {
	for h > 0 {
		d.SetPixel(x, y, c)
		y++
		h--
	}
}

var buffer [3 * 256]uint16

// NOTE: This part does not work with tinygo version 0.23.0 windows/amd64 (using go version go1.18 and LLVM version 14.0.0)
// The effect is that some memory overwrite occurs (serios BUG that should be isolated) and panics happen.
// However, it is mostly not needed due to the fact, that raw bytes of the icons are smaller than embedded PNG files.
// DrawPng decodes pngImage and streams pixels into the provided displayer.
func DrawPng(d BitmapDisplayer, x0, y0 int16, pngImage string) error {
	p := strings.NewReader(pngImage)
	png.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		// W, H := d.Size()
		// println(x0+x, y0+y, w, h, width, height, W, H, len(data))
		err := d.DrawRGBBitmap(x0+x, y0+y, data[:w*h], w, h)
		if err != nil {
			println("DrawRGBBitmap: " + err.Error())
		}
	})

	// w, h := d.Size()
	_, err := png.Decode(p)
	if err != nil {
		println("DrawPng Decode: " + err.Error())
	}
	// w, h = d.Size()
	return err
}

// DecodePng decodes pngImage into dst as a w×h RGB565 bitmap (row-major, the
// format DrawRGBBitmap takes). Pixels outside w×h are dropped. Decoding
// allocates the decoder's working state; decode once and keep dst.
func DecodePng(dst []uint16, w, h int16, pngImage string) error {
	if len(dst) < int(w)*int(h) {
		return errShortBuffer
	}
	png.SetCallback(buffer[:], func(data []uint16, x, y, cw, ch, width, height int16) {
		for row := int16(0); row < ch; row++ {
			dy := y + row
			if dy < 0 || dy >= h {
				continue
			}
			for col := int16(0); col < cw; col++ {
				if dx := x + col; dx >= 0 && dx < w {
					dst[int(dy)*int(w)+int(dx)] = data[int(row)*int(cw)+int(col)]
				}
			}
		}
	})
	_, err := png.Decode(strings.NewReader(pngImage))
	return err
}

var errShortBuffer = errors.New("tinygui: bitmap buffer shorter than w*h")
