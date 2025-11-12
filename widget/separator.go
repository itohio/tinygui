package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// Separator renders a thin line either horizontally or vertically.
type Separator struct {
	ui.WidgetBase
	Color color.RGBA
}

// NewSeparator creates a separator with the given dimensions.
func NewSeparator(width, height uint16, color color.RGBA) *Separator {
	return &Separator{
		WidgetBase: ui.NewWidgetBase(width, height),
		Color:      color,
	}
}

func (s *Separator) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
		return
	}
	x, y := ctx.DisplayPos()
	if s.Width >= s.Height {
		midY := y + int16(s.Height)/2
		ui.HLine(d, x, midY, int16(s.Width), s.Color)
	} else {
		midX := x + int16(s.Width)/2
		ui.VLine(d, midX, y, int16(s.Height), s.Color)
	}
}
