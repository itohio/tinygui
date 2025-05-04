package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

type HorizontalSeparator struct {
	ui.WidgetBase
	color color.RGBA
}

func NewHSeparator(w, h uint16, color color.RGBA) *HorizontalSeparator {
	return &HorizontalSeparator{
		WidgetBase: ui.NewWidgetBase(w, h),
		color:      color,
	}
}

func (w *HorizontalSeparator) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	d := ctx.D()
	if fast, ok := d.(ui.LineDisplayer); ok {
		fast.DrawFastHLine(x+1, x+int16(w.Width)-1, y+int16(w.Height/2), w.color)
	} else {
		ui.HLine(d, x+1, y+int16(w.Height/2), int16(w.Width)-2, w.color)
	}
}
