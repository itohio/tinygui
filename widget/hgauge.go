package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

type HorizontalGauge struct {
	ui.WidgetBase
	color color.RGBA
	value func() uint16
}

func NewHGauge(w, h uint16, value func() uint16, color color.RGBA) *HorizontalGauge {
	return &HorizontalGauge{
		WidgetBase: ui.NewWidgetBase(w, h),
		color:      color,
	}
}

func (w *HorizontalGauge) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	d := ctx.D()
	value := w.value()
	width := int16((uint32(w.Width-2) * uint32(value)) >> 16)
	if fast, ok := d.(ui.RectangleDisplayer); ok {
		fast.FillRectangle(x+1, y+1, width, int16(w.Height-2), w.color)
	} else {
		for dx := range w.Height - 2 {
			ui.HLine(d, x+1, y+int16(dx), width, w.color)
		}
	}
}

func NewHGaugeArray(w, h uint16, color color.RGBA, value ...func() uint16) []ui.Widget {
	widgets := make([]ui.Widget, len(value))
	for i, fn := range value {
		widgets[i] = NewHGauge(w, h, fn, color)
	}
	return widgets
}
