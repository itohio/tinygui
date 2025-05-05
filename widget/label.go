package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

type Label struct {
	ui.WidgetBase
	font  tinyfont.Fonter
	text  func() string
	color color.RGBA
}

func NewLabel(w, h uint16, font tinyfont.Fonter, text func() string, color color.RGBA) *Label {
	return &Label{
		WidgetBase: ui.NewWidgetBase(w, h),
		font:       font,
		text:       text,
		color:      color,
	}
}

func (l *Label) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	tinyfont.WriteLine(ctx.D(), l.font, x, y+int16(l.Height), l.text(), l.color)
}

func NewLabelArray(w, h uint16, font tinyfont.Fonter, color color.RGBA, text ...func() string) []ui.Widget {
	widgets := make([]ui.Widget, len(text))
	for i, fn := range text {
		widgets[i] = NewLabel(w, h, font, fn, color)
	}
	return widgets
}
