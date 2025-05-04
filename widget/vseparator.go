package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

type VerticalSeparator struct {
	ui.WidgetBase
	color color.RGBA
}

func NewVSeparator(w, h uint16, color color.RGBA) *VerticalSeparator {
	return &VerticalSeparator{
		WidgetBase: ui.NewWidgetBase(w, h),
		color:      color,
	}
}

func (w *VerticalSeparator) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()

	d := ctx.D()
	if fast, ok := d.(ui.LineDisplayer); ok {
		fast.DrawFastVLine(x+int16(w.Width/2), y+1, y+int16(w.Height-1), w.color)
	} else {
		ui.VLine(d, x+int16(w.Width/2), y+1, int16(w.Height)-2, w.color)
	}
}
