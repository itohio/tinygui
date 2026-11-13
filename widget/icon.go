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

// iconSheet holds a set of icons decoded once, at construction, into one
// RGB565 buffer (n·w·h pixels): switching icons on a press is an index
// change, drawing one a single blit — no PNG decode, no allocation after
// setup (verify TG-1-postcommit R7: re-decoding on every press made 32–48
// allocations).
type iconSheet struct {
	w, h int16
	pix  []uint16
	ok   []bool // icon i decoded
}

func newIconSheet(w, h uint16, images []string) iconSheet {
	n := int(w) * int(h)
	s := iconSheet{w: int16(w), h: int16(h), pix: make([]uint16, n*len(images)), ok: make([]bool, len(images))}
	for i, img := range images {
		s.ok[i] = img != "" && ui.DecodePng(s.pix[i*n:(i+1)*n], s.w, s.h, img) == nil
	}
	return s
}

// draw blits icon i at the context's position (nothing for a bad index or
// an icon that didn't decode, or a display without bitmaps).
func (s *iconSheet) draw(ctx ui.Context, i int) {
	bmp, isBmp := ctx.D().(ui.BitmapDisplayer)
	if !isBmp || i < 0 || i >= len(s.ok) || !s.ok[i] {
		return
	}
	n := int(s.w) * int(s.h)
	x, y := ctx.DisplayPos()
	_ = bmp.DrawRGBBitmap(x, y, s.pix[i*n:(i+1)*n], s.w, s.h)
}
