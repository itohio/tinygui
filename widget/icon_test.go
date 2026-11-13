package widget

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
)

// bitmapRecorder keeps the last RGB565 bitmap blit and counts blits.
type bitmapRecorder struct {
	mockDisplay
	blits  int
	x, y   int16
	w, h   int16
	pixels []uint16
}

func (r *bitmapRecorder) DrawRGBBitmap(x, y int16, data []uint16, w, h int16) error {
	r.blits++
	r.x, r.y, r.w, r.h = x, y, w, h
	r.pixels = append(r.pixels[:0], data[:int(w)*int(h)]...)
	return nil
}

func (r *bitmapRecorder) DrawRGBBitmap8(int16, int16, []uint8, int16, int16) error { return nil }

func testPNG(t *testing.T, w, h int, c color.RGBA) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, c)
		}
	}
	img.SetRGBA(w-1, h-1, color.RGBA{0, 0, 255, 255}) // one distinct corner pixel
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return b.String()
}

// An Icon shows its PNG as one RGB565 blit per draw, at the context position,
// and re-decodes only when the image changes.
func TestIconDrawsDecodedPixels(t *testing.T) {
	red := testPNG(t, 4, 3, color.RGBA{255, 0, 0, 255})
	green := testPNG(t, 4, 3, color.RGBA{0, 255, 0, 255})
	icon := NewIcon(4, 3, func() string { return red })
	rec := &bitmapRecorder{}
	ctx := ui.NewContext(rec, 20, 20, 5, 6)

	icon.Draw(&ctx)
	icon.Draw(&ctx)

	check.Equal(t, 2, rec.blits, "one blit per draw")
	check.Equal(t, [4]int16{5, 6, 4, 3}, [4]int16{rec.x, rec.y, rec.w, rec.h})
	check.Equal(t, uint16(0xF800), rec.pixels[0], "red in RGB565")
	check.Equal(t, uint16(0x001F), rec.pixels[len(rec.pixels)-1], "blue corner")
	check.Equal(t, 1, icon.decodes, "decoded once for two draws")

	icon.SetImage(green)
	icon.Draw(&ctx)
	check.Equal(t, uint16(0x07E0), rec.pixels[0], "green after SetImage")
	check.Equal(t, 2, icon.decodes)
}
