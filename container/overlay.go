package container

import ui "github.com/itohio/tinygui"

// MaxOverlays is how many layers an Overlay can stack (a dialog opened from
// a dialog, a dropdown inside a dialog, ...).
const MaxOverlays = 4

// Overlay draws a base widget and then up to MaxOverlays layers over it:
// popups, dialogs, dropdown lists. The top layer is modal: it receives every
// command until it closes. Layers are drawn over whatever is under them and
// nothing restores it when they close; the base redraws it on the next frame
// (no damage tracking).
type Overlay struct {
	ui.WidgetBase
	Base   ui.Widget
	ctx    ui.ContextImpl
	layers [MaxOverlays]overlayLayer
	n      int
}

type overlayLayer struct {
	w   ui.Widget
	x   int16
	y   int16
	ctx ui.ContextImpl
}

// NewOverlay wraps base; the overlay takes base's size.
func NewOverlay(base ui.Widget) *Overlay {
	w, h := base.Size()
	o := &Overlay{WidgetBase: ui.NewWidgetBase(w, h), Base: base}
	base.SetParent(o)
	return o
}

// Open puts w on top at display position x, y. It returns false when all
// layers are in use.
func (o *Overlay) Open(w ui.Widget, x, y int16) bool {
	if o.n == MaxOverlays {
		return false
	}
	o.layers[o.n] = overlayLayer{w: w, x: x, y: y}
	o.n++
	w.SetParent(o)
	return true
}

// OpenCentered opens w in the middle of the overlay.
func (o *Overlay) OpenCentered(w ui.Widget) bool {
	ow, oh := o.Size()
	ww, wh := w.Size()
	return o.Open(w, (int16(ow)-int16(ww))/2, (int16(oh)-int16(wh))/2)
}

// Close removes w and every layer opened after it.
func (o *Overlay) Close(w ui.Widget) {
	for i := range o.n {
		if o.layers[i].w == w {
			for j := i; j < o.n; j++ {
				o.layers[j] = overlayLayer{}
			}
			o.n = i
			return
		}
	}
}

// ModalFor reports whether a Navigator rooted at root is blocked (ui.Modal):
// a layer is open and root is not inside the top layer — it is the base or
// in a layer under the top. A Navigator rooted in the top layer (a panel
// opened as a layer, navigated by its own Navigator) keeps working.
func (o *Overlay) ModalFor(root ui.Widget) bool {
	if o.n == 0 {
		return false
	}
	top := o.layers[o.n-1].w
	for w := root; w != nil && w != ui.Widget(o); w = w.Parent() {
		if w == top {
			return false
		}
	}
	return true
}

// Top is the layer receiving commands, or nil.
func (o *Overlay) Top() ui.Widget {
	if o.n == 0 {
		return nil
	}
	return o.layers[o.n-1].w
}

func (o *Overlay) Draw(ctx ui.Context) {
	w, h := o.Size()
	o.Base.Draw(ui.ChildContext(ctx, &o.ctx, o.Base, w, h))
	o.DrawLayers(ctx)
}

// DrawLayers draws only the open layers, for apps that paint more over the
// base themselves (a status bar) before the overlays.
func (o *Overlay) DrawLayers(ctx ui.Context) {
	d := ctx.D()
	for i := range o.n {
		l := &o.layers[i]
		lw, lh := l.w.Size()
		l.ctx = ui.NewContext(d, lw, lh, l.x, l.y)
		l.w.Draw(&l.ctx)
	}
}

// Interact sends cmd to the top layer, or to the base when none is open.
func (o *Overlay) Interact(cmd ui.UserCommand) bool {
	if top := o.Top(); top != nil {
		top.Interact(cmd)
		return true // modal: the base never sees commands while a layer is open
	}
	return o.Base.Interact(cmd)
}
