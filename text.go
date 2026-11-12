package ui

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
)

// DrawText draws s with its baseline at (x, y), like tinyfont.WriteLine, but
// without allocating: WriteLine builds a colour slice and boxes a rotated
// display on every call (~2 heap objects per label per frame), which keeps
// the garbage collector busy on small heaps. '\n' starts a new line. With a
// regular tinyfont.Font no allocation happens; tinyfont's const1bit fonts
// allocate per rune inside their own GetGlyph (outside TinyGUI).
func DrawText(d drivers.Displayer, f tinyfont.Fonter, x, y int16, s string, c color.RGBA) {
	x0 := x
	for _, r := range s {
		if r == tinyfont.LineFeed || r == tinyfont.CarriageReturn {
			x = x0
			y += int16(f.GetYAdvance())
			continue
		}
		g := f.GetGlyph(r)
		g.Draw(d, x, y, c)
		x += int16(g.Info().XAdvance)
	}
}

// TextWidth returns the advance width of s in f (tinyfont.LineWidth's outer
// width), without allocating.
func TextWidth(f tinyfont.Fonter, s string) int16 {
	var w int16
	for _, r := range s {
		w += int16(f.GetGlyph(r).Info().XAdvance)
	}
	return w
}
