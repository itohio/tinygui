package widget

import (
	"image/color"
	"math"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
)

const defaultVolumeBars = 12
const defaultDimFactor = 0.4

// VolumeGauge renders a segmented horizontal gauge similar to the SurroundAmp volume widget.
type VolumeGauge[T Number] struct {
	HorizontalGauge[T]
	bars      uint8
	dimFactor float32
}

// NewVolumeGauge constructs a segmented volume gauge. When bars is zero a default of 12 is used.
func NewVolumeGauge[T Number](width, height uint16, value *T, min, max T, bars uint8, fg, bg color.RGBA) *VolumeGauge[T] {
	if bars == 0 {
		bars = defaultVolumeBars
	}
	g := &VolumeGauge[T]{
		HorizontalGauge: *NewHorizontalGauge(width, height, value, min, max, fg, bg),
		bars:            bars,
		dimFactor:       defaultDimFactor,
	}
	return g
}

// SolidGauge renders a solid horizontal gauge without inner padding.
type SolidGauge[T Number] struct {
	HorizontalGauge[T]
}

// NewSolidGauge constructs a solid gauge that fills the background then paints the active region.
func NewSolidGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *SolidGauge[T] {
	return &SolidGauge[T]{HorizontalGauge: *NewHorizontalGauge(width, height, value, min, max, fg, bg)}
}

// Draw renders the segmented volume gauge with coloured bars transitioning from green to red.
func (g *VolumeGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Value == nil {
		return
	}

	x, y := ctx.DisplayPos()
	width, height := int16(g.Width), int16(g.Height)
	bars := int(g.bars)
	if bars <= 0 {
		bars = defaultVolumeBars
	}

	barWidth := width / int16(bars)
	if barWidth <= 0 {
		barWidth = 1
	}

	gapWidth := int16(3)
	if gapWidth >= barWidth {
		gapWidth = max16(1, barWidth-1)
	}

	norm := normalisedValue(g.Value, g.Min, g.Max)
	filled := int16(float32(width) * norm)

	if !isZeroColor(g.Background) {
		fillRect(d, x, y, width, height, g.Background)
	}

	for i := 0; i < bars; i++ {
		startX := x + int16(i)*barWidth
		segWidth := barWidth - gapWidth
		if segWidth < 1 {
			segWidth = barWidth
		}

		ratio := float32(i)
		den := float32(max(1, bars-1))
		color := g.barColor(ratio / den)

		threshold := int16(math.Round(float64(width) * float64(i+1) / float64(bars)))
		if filled < threshold {
			color = dimColor(color, g.dimFactor)
		}

		fillWidth := segWidth
		if startX+fillWidth > x+width {
			fillWidth = x + width - startX
		}
		if fillWidth > 0 {
			fillRect(d, startX, y, fillWidth, height, color)
		}

		if gapWidth > 0 {
			gapX := startX + segWidth
			if gapX < x+width {
				fillRect(d, gapX, y, min16(gapWidth, x+width-gapX), height, g.Background)
			}
		}
	}
}

// Draw renders the solid gauge without inner padding.
func (g *SolidGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Value == nil {
		return
	}

	x, y := ctx.DisplayPos()
	width, height := int16(g.Width), int16(g.Height)

	if !isZeroColor(g.Background) {
		fillRect(d, x, y, width, height, g.Background)
	}

	fillWidth := int16(float32(width) * normalisedValue(g.Value, g.Min, g.Max))
	if fillWidth <= 0 {
		return
	}
	if fillWidth > width {
		fillWidth = width
	}

	fillRect(d, x, y, fillWidth, height, g.Foreground)
}

func (g *VolumeGauge[T]) barColor(ratio float32) color.RGBA {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	mid := float32(2) / 3
	if ratio <= mid {
		t := ratio / mid
		return lerpColor(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 255, 0, 255}, t)
	}
	t := (ratio - mid) / (1 - mid)
	return lerpColor(color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, t)
}

func lerpColor(a, b color.RGBA, t float32) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t),
		G: uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t),
		B: uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t),
		A: uint8(float32(a.A) + (float32(b.A)-float32(a.A))*t),
	}
}

func dimColor(c color.RGBA, factor float32) color.RGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return color.RGBA{
		R: uint8(float32(c.R) * factor),
		G: uint8(float32(c.G) * factor),
		B: uint8(float32(c.B) * factor),
		A: c.A,
	}
}

func fillRect(d drivers.Displayer, x, y, w, h int16, c color.RGBA) {
	if rect, ok := d.(ui.RectangleDisplayer); ok {
		_ = rect.FillRectangle(x, y, w, h, c)
		return
	}

	for dx := int16(0); dx < w; dx++ {
		for dy := int16(0); dy < h; dy++ {
			d.SetPixel(x+dx, y+dy, c)
		}
	}
}

func normalisedValue[T Number](value *T, min, max T) float32 {
	if value == nil {
		return 0
	}
	curr := float32(*value)
	minF := float32(min)
	maxF := float32(max)
	if maxF <= minF {
		return 0
	}
	norm := (curr - minF) / (maxF - minF)
	if norm < 0 {
		norm = 0
	}
	if norm > 1 {
		norm = 1
	}
	return norm
}

func min16(a, b int16) int16 {
	if a < b {
		return a
	}
	return b
}

func max16(a, b int16) int16 {
	if a > b {
		return a
	}
	return b
}
