package framebuffer

import (
	"image/color"
	"slices"
	"testing"
)

// Drawing the same scene band by band must give the same pixels as drawing it
// into one full-height buffer, including shapes that straddle band edges.
func TestBandsMatchFullFrame(t *testing.T) {
	const w, h, band = 16, 12, 5 // 3 bands, the last one partial
	scene := func(f *RGB565) {
		f.FillScreen(color.RGBA{1, 2, 3, 255})
		_ = f.FillRectangle(2, 3, 9, 6, color.RGBA{200, 0, 0, 255})
		bmp := make([]uint16, 4*7)
		for i := range bmp {
			bmp[i] = uint16(i + 1)
		}
		_ = f.DrawRGBBitmap(10, 2, bmp, 4, 7)
		f.SetPixel(0, 11, color.RGBA{0, 255, 0, 255})
		_ = f.FillRectangle(-3, -3, 5, 5, color.RGBA{0, 0, 255, 255}) // clipped corner
	}

	full := New(w, h, make([]uint16, w*h), nil)
	scene(full)

	var stitched []uint16
	b := NewBand(w, h, band, make([]uint16, w*band), func(f *RGB565) error {
		rows := min(int(f.H), int(f.ScreenH-f.Y0))
		stitched = append(stitched, f.Pix[:rows*int(f.W)]...)
		return nil
	})
	if b.Bands() != 3 {
		t.Fatalf("Bands() = %d, want 3", b.Bands())
	}
	for i := range b.Bands() {
		b.SetY0(int16(i * band))
		scene(b)
		_ = b.Display()
	}
	if !slices.Equal(stitched, full.Pix) {
		t.Fatalf("banded frame differs from full frame:\n got %v\nwant %v", stitched, full.Pix)
	}
}
