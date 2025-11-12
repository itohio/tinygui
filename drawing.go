package ui

import (
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
