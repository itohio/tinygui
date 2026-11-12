package app

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	ui "github.com/itohio/tinygui"
)

// Screen geometry of the Tufty 2040 panel (landscape).
const (
	ScreenW = 320
	ScreenH = 240

	tabsH   = 22
	footerH = 14
	bodyH   = ScreenH - tabsH - footerH
)

var (
	colBG     = color.RGBA{10, 12, 18, 255}
	colPanel  = color.RGBA{26, 30, 42, 255}
	colTrack  = color.RGBA{40, 44, 58, 255}
	colFG     = color.RGBA{228, 230, 236, 255}
	colMuted  = color.RGBA{130, 136, 152, 255}
	colGood   = color.RGBA{80, 200, 120, 255}
	colWarn   = color.RGBA{240, 180, 40, 255}
	colBad    = color.RGBA{230, 70, 70, 255}
	colAccent = accents[0]

	accentNames = []string{"Teal", "Amber", "Pink", "Lime"}
	accents     = []color.RGBA{
		{0, 190, 200, 255},
		{250, 170, 30, 255},
		{240, 90, 170, 255},
		{150, 220, 40, 255},
	}
)

// Fonts. TinyGUI widgets take any tinyfont.Fonter; the title font is the body
// font scaled ×2 by a Fonter adapter, so no second font table lands in RAM.
var (
	fontBody  tinyfont.Fonter = &proggy.TinySZ8pt7b
	fontSmall tinyfont.Fonter = &tinyfont.TomThumb
	fontTitle                 = newScaledFont(fontBody, 2)
)

// scaledFont draws every glyph of base at an integer scale without allocating:
// it hands out one reusable glyph and one reusable scaling displayer.
type scaledFont struct {
	base  tinyfont.Fonter
	scale int16
	glyph scaledGlyph
}

type scaledGlyph struct {
	inner tinyfont.Glypher
	scale int16
	disp  scaleDisplay
}

type scaleDisplay struct {
	d      drivers.Displayer
	ox, oy int16
	scale  int16
}

func newScaledFont(base tinyfont.Fonter, scale int16) *scaledFont {
	return &scaledFont{base: base, scale: scale}
}

func (f *scaledFont) GetGlyph(r rune) tinyfont.Glypher {
	f.glyph.inner = f.base.GetGlyph(r)
	f.glyph.scale = f.scale
	return &f.glyph
}

func (f *scaledFont) GetYAdvance() uint8 { return f.base.GetYAdvance() * uint8(f.scale) }

func (g *scaledGlyph) Info() tinyfont.GlyphInfo {
	i := g.inner.Info()
	s := uint8(g.scale)
	i.Width *= s
	i.Height *= s
	i.XAdvance *= s
	i.XOffset *= int8(s)
	i.YOffset *= int8(s)
	return i
}

func (g *scaledGlyph) Draw(d drivers.Displayer, x, y int16, c color.RGBA) {
	g.disp = scaleDisplay{d: d, ox: x, oy: y, scale: g.scale}
	g.inner.Draw(&g.disp, 0, 0, c)
}

func (s *scaleDisplay) Size() (int16, int16) { return s.d.Size() }
func (s *scaleDisplay) Display() error       { return nil }
func (s *scaleDisplay) SetPixel(x, y int16, c color.RGBA) {
	fillRect(s.d, s.ox+x*s.scale, s.oy+y*s.scale, s.scale, s.scale, c)
}

// fillRect uses the displayer's rectangle fast path when it has one.
func fillRect(d drivers.Displayer, x, y, w, h int16, c color.RGBA) {
	if r, ok := d.(ui.RectangleDisplayer); ok {
		_ = r.FillRectangle(x, y, w, h, c)
		return
	}
	for j := int16(0); j < h; j++ {
		for i := int16(0); i < w; i++ {
			d.SetPixel(x+i, y+j, c)
		}
	}
}

// strokeRect draws a 1 px rectangle outline.
func strokeRect(d drivers.Displayer, x, y, w, h int16, c color.RGBA) {
	fillRect(d, x, y, w, 1, c)
	fillRect(d, x, y+h-1, w, 1, c)
	fillRect(d, x, y, 1, h, c)
	fillRect(d, x+w-1, y, 1, h, c)
}

func textWidth(f tinyfont.Fonter, s string) int16 { return ui.TextWidth(f, s) }
