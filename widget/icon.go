package widget

import (
	ui "github.com/itohio/tinygui"
)

type Icon struct {
	ui.WidgetBase
	image string
}

func NewIcon(w, h uint16, image string) *Icon {
	return &Icon{
		WidgetBase: ui.NewWidgetBase(w, h),
		image:      image,
	}
}

func (w *Icon) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	if bmp, ok := ctx.D().(ui.BitmapDisplayer); ok {
		ui.DrawPng(bmp, x, y, w.image)
	}
}
