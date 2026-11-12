package ui

import (
	"image/color"
	"testing"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

// pixelSet records which pixels were set to which colour.
type pixelSet struct{ px map[[2]int16]color.RGBA }

func newPixelSet() *pixelSet                          { return &pixelSet{px: map[[2]int16]color.RGBA{}} }
func (p *pixelSet) Size() (int16, int16)              { return 320, 240 }
func (p *pixelSet) Display() error                    { return nil }
func (p *pixelSet) SetPixel(x, y int16, c color.RGBA) { p.px[[2]int16{x, y}] = c }

// DrawText must paint exactly what tinyfont.WriteLine paints, including line
// breaks and runes the font lacks, and must not allocate.
func TestDrawTextMatchesTinyfont(t *testing.T) {
	const s = "Hello, TinyGUI 123 %\nsecond line – ▲"
	c := color.RGBA{10, 200, 30, 255}
	for _, f := range []tinyfont.Fonter{&tinyfont.TomThumb, &proggy.TinySZ8pt7b} {
		want, got := newPixelSet(), newPixelSet()
		tinyfont.WriteLine(want, f, 7, 20, s, c)
		DrawText(got, f, 7, 20, s, c)
		if len(got.px) == 0 || len(got.px) != len(want.px) {
			t.Fatalf("pixel count: got %d, want %d", len(got.px), len(want.px))
		}
		for k, v := range want.px {
			if got.px[k] != v {
				t.Fatalf("pixel %v: got %v, want %v", k, got.px[k], v)
			}
		}
		_, ww := tinyfont.LineWidth(f, "Hello, TinyGUI")
		if gw := TextWidth(f, "Hello, TinyGUI"); gw != int16(ww) {
			t.Fatalf("TextWidth = %d, tinyfont.LineWidth = %d", gw, ww)
		}
	}

	d := newPixelSet()
	if n := testing.AllocsPerRun(50, func() { DrawText(d, &tinyfont.TomThumb, 0, 10, "no garbage", c) }); n != 0 {
		t.Fatalf("DrawText allocates %.0f times per call", n)
	}
}
