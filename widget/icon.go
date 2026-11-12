package widget

import (
	ui "github.com/itohio/tinygui"
)

// Icon renders a PNG image. The PNG is decoded once into an RGB565 buffer of
// the icon's size (w*h*2 bytes, allocated on first draw) and re-decoded only
// when the image string changes; every draw is a single DrawRGBBitmap blit.
type Icon struct {
	ui.WidgetBase
	image func() string

	pix     []uint16
	shown   string // image the buffer holds
	valid   bool   // buffer holds a successfully decoded shown
	decodes int    // number of decodes (diagnostics, tests)
}

func (w *Icon) Draw(ctx ui.Context) {
	bmp, ok := ctx.D().(ui.BitmapDisplayer)
	if !ok {
		return
	}
	img := w.Image()
	if img == "" {
		return
	}
	if w.pix == nil || img != w.shown {
		w.decode(img)
	}
	if !w.valid {
		return
	}
	x, y := ctx.DisplayPos()
	_ = bmp.DrawRGBBitmap(x, y, w.pix, int16(w.Width), int16(w.Height))
}

func (w *Icon) decode(img string) {
	if w.pix == nil {
		w.pix = make([]uint16, int(w.Width)*int(w.Height))
	}
	clear(w.pix)
	w.shown = img
	w.decodes++
	w.valid = ui.DecodePng(w.pix, int16(w.Width), int16(w.Height), img) == nil
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

// SetImageProvider swaps the callback used to fetch the image during drawing.
func (w *Icon) SetImageProvider(fn func() string) {
	if fn != nil {
		w.image = fn
	}
}

// NewIconArray returns a slice of icons constructed from multiple callbacks.
func NewIconArray(w, h uint16, image ...func() string) []ui.Widget {
	widgets := make([]ui.Widget, len(image))
	for i, fn := range image {
		widgets[i] = NewIcon(w, h, fn)
	}
	return widgets
}
