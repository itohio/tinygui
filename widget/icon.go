package widget

import (
	ui "github.com/itohio/tinygui"
)

// Icon renders a PNG string using TinyGUI's DrawPng helper.
type Icon struct {
	ui.WidgetBase
	image func() string
}

func (w *Icon) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	if bmp, ok := ctx.D().(ui.BitmapDisplayer); ok {
		ui.DrawPng(bmp, x, y, w.Image())
	}
}

// SetImage updates the PNG payload rendered by the icon.
func (w *Icon) SetImage(image string) {
	w.image = func() string { return image }
}

// Image returns the underlying image payload.
func (w *Icon) Image() string {
	return w.image()
}

// NewIcon constructs an icon of fixed size.
func NewIcon(w, h uint16, image func() string) *Icon {
	if image == nil {
		image = func() string { return "" }
	}
	return &Icon{
		WidgetBase: ui.NewWidgetBase(w, h),
		image:      image,
	}
}

// SetTextProvider swaps the callback used to fetch text during drawing.
func (w *Icon) SetImageProvider(fn func() string) {
	if fn != nil {
		w.image = fn
	}
}

// NewIconArray returns a slice of labels constructed from multiple callbacks.
func NewIconArray(w, h uint16, image ...func() string) []ui.Widget {
	widgets := make([]ui.Widget, len(image))
	for i, fn := range image {
		widgets[i] = NewIcon(w, h, fn)
	}
	return widgets
}
