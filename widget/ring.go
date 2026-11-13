package widget

import (
	"image/color"
	"unsafe"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
)

// RingStyle is how rings look: the colours and font several rings share
// (convention: shared appearance is a Style struct, passed whole).
type RingStyle struct {
	Track, Fill color.RGBA      // the unfilled part of the ring, the filled arc
	BG          color.RGBA      // painted behind the ring inside its box (corners and hole), and what the ring's edges are smoothed into; zero value: not painted, and the outer and inner edges are hard
	Font        tinyfont.Fonter // the centre label's font (nil: no label is drawn)
	Label       color.RGBA      // the centre label's colour
}

// RingCap is how the filled arc ends.
type RingCap uint8

const (
	// RingCapButt ends the arc square, along the radius (the default).
	RingCapButt RingCap = iota
	// RingCapRound ends the arc in a half disc as wide as the ring, so the
	// arc looks like a rounded stroke.
	RingCapRound
)

// RingOption is an optional knob of NewRing (convention: knobs with a
// sensible default are functional options).
type RingOption func(*ringConfig)

// ringConfig holds the knobs the options set; zero values are the defaults.
type ringConfig struct {
	thickness uint16 // 0: max(2, diameter/8)
	start     int16  // degrees clockwise from 12 o'clock
	ccw       bool
	cap       RingCap
	percent   bool
	text      func() string
}

// WithThickness sets the ring's width in pixels, along the radius. The
// default, max(2, diameter/8), suits a progress ring; 0 keeps it, and a
// width past the radius makes a full disc.
func WithThickness(px uint16) RingOption { return func(c *ringConfig) { c.thickness = px } }

// WithStartAngle sets where 0 % sits, in degrees: 0 is 12 o'clock and
// positive angles go clockwise (90 is 3 o'clock). The default is 0.
func WithStartAngle(deg int16) RingOption { return func(c *ringConfig) { c.start = deg } }

// WithCounterClockwise fills the ring counter-clockwise from the start
// angle; the default is clockwise.
func WithCounterClockwise() RingOption { return func(c *ringConfig) { c.ccw = true } }

// WithCap sets how the filled arc ends; the default is RingCapButt. Caps
// show only while the ring is partly filled (above 0 %, below 100 %).
func WithCap(c RingCap) RingOption { return func(cfg *ringConfig) { cfg.cap = c } }

// WithPercentLabel draws the value's percentage of min..max ("0%".."100%",
// rounded down: it never shows 100 % before the ring is full) in the ring's
// centre. It is formatted into a buffer inside the widget, so drawing it
// allocates nothing. Needs RingStyle.Font.
func WithPercentLabel() RingOption { return func(c *ringConfig) { c.percent = true } }

// WithLabel draws text's result in the ring's centre on every draw, as
// returned; for a GC-free app text must not allocate (return a constant or a
// view of a buffer the caller keeps). It overrides WithPercentLabel. Needs
// RingStyle.Font.
func WithLabel(text func() string) RingOption { return func(c *ringConfig) { c.text = text } }

// Ring is a circular progress bar: the value's part of min..max fills an arc
// of the ring from the start angle, the rest shows the track. It is a plain
// non-interactive widget, so it works in every container and layout.
//
// It draws through the arc-raster core (arcraster.go) with a whole-turn
// track: integer math per pixel, edges smoothed over one pixel, the arc's
// geometry recomputed only when the value or the size changes. Draw
// allocates nothing.
type Ring[T Number] struct {
	ui.WidgetBase
	value    *T
	min, max T
	style    RingStyle
	cfg      ringConfig

	// cache, recomputed when the size or the value's fraction changes
	cached bool
	w, h   uint16
	frac   uint32 // the value's fraction of min..max, 0..fullTurn
	dia    int16  // the ring's diameter
	arc    arcRaster

	paint spanPainter
	pct   [4]byte // the percent label: up to "100%"
}

// NewRing is a ring progress bar: value between min and max (clamped) fills an
// arc from the start angle. The ring's diameter is min(width, height), centred in its box.
func NewRing[T Number](width, height uint16, value *T, min, max T, style RingStyle, opts ...RingOption) *Ring[T] {
	if min > max {
		min, max = max, min
	}
	r := &Ring[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		value:      value,
		min:        min,
		max:        max,
		style:      style,
	}
	for _, o := range opts {
		if o != nil {
			o(&r.cfg)
		}
	}
	return r
}

// fraction is the value's part of min..max in 1/65536 (fullTurn),
// truncated: it is fullTurn only when the value is at max (min == max
// counts as full, as in Gauge). One float32 division per draw.
func (r *Ring[T]) fraction() uint32 {
	if r.max <= r.min {
		return fullTurn
	}
	v := clamp(r.min, r.max, *r.value)
	f := (float32(v) - float32(r.min)) / (float32(r.max) - float32(r.min))
	switch {
	case f <= 0:
		return 0
	case f >= 1:
		return fullTurn
	}
	return min(uint32(f*fullTurn), fullTurn-1)
}

// update recomputes the cached geometry when the size or the fraction
// changed since the last draw.
func (r *Ring[T]) update() {
	frac := r.fraction()
	if r.cached && r.w == r.Width && r.h == r.Height && frac == r.frac {
		return
	}
	r.cached, r.w, r.h, r.frac = true, r.Width, r.Height, frac
	dia := min(r.w, r.h)
	t := r.cfg.thickness
	if t == 0 {
		t = max(2, dia/8)
	}
	r.dia = int16(dia)
	// half pixels: the square's centre, and radii dia and dia-2t
	cx, cy := int32(r.w-dia)/2*2+int32(dia), int32(r.h-dia)/2*2+int32(dia)
	start := degTurn(r.cfg.start)
	fill := arcSpan{start: start, sweep: frac}
	if r.cfg.ccw {
		fill.start = start - uint16(frac)
	}
	r.arc.set(cx, cy, int32(dia), int32(dia)-2*int32(t), arcSpan{sweep: fullTurn}, fill, r.cfg.cap == RingCapRound)
}

// Draw paints the ring inside the display's clip (a framebuffer band draws
// only its rows), then the centre label.
func (r *Ring[T]) Draw(ctx ui.Context) {
	if r == nil || r.value == nil {
		return
	}
	d := ctx.D()
	if d == nil || r.Width == 0 || r.Height == 0 {
		return
	}
	r.update()
	x0, y0 := ctx.DisplayPos()
	w, h := int16(r.Width), int16(r.Height)
	row0, row1 := r.arc.draw(&r.paint, d, x0, y0, w, h, r.style.Track, r.style.Fill, r.style.BG)
	r.drawLabel(d, x0+(w-r.dia)/2, y0+(h-r.dia)/2, r.dia, y0+row0, y0+row1)
}

// drawLabel draws the centre label, if any, with its ink box centred on the
// ring's centre (the ring's square is at x, y, dia wide), unless it lies
// outside the clipped rows clipY0..clipY1.
func (r *Ring[T]) drawLabel(d drivers.Displayer, x, y, dia, clipY0, clipY1 int16) {
	f := r.style.Font
	if f == nil {
		return
	}
	var s string
	switch {
	case r.cfg.text != nil:
		s = r.cfg.text()
	case r.cfg.percent:
		s = r.percent()
	}
	if s == "" {
		return
	}
	x0, y0, x1, y1 := inkBox(f, s)
	bx := x + (dia-x0-x1)/2 // pen start and baseline
	by := y + (dia-y0-y1)/2
	if by+y1 <= clipY0 || by+y0 >= clipY1 {
		return
	}
	ui.DrawText(d, f, bx, by, s, r.style.Label)
}

// percent formats the percentage, rounded down, into r.pct and returns a
// view of it (valid until the next draw), allocating nothing. It is worked
// out from the value, not from the arc's 1/65536 fraction (whose truncation
// would turn 5 % into 4 %), multiplying by 100 before dividing so whole
// percentages stay exact in float32. "100%" shows only with a full ring.
func (r *Ring[T]) percent() string {
	var pct uint32
	switch {
	case r.frac >= fullTurn:
		pct = 100
	case r.frac > 0:
		v := clamp(r.min, r.max, *r.value)
		p := (float32(v) - float32(r.min)) * 100 / (float32(r.max) - float32(r.min))
		pct = min(uint32(max(p, 0)), 99)
	}
	n := 0
	if pct >= 100 {
		r.pct[n] = '1'
		n++
	}
	if pct >= 10 {
		r.pct[n] = byte('0' + pct/10%10)
		n++
	}
	r.pct[n] = byte('0' + pct%10)
	r.pct[n+1] = '%'
	return unsafe.String(&r.pct[0], n+2)
}

// inkBox is the box the glyphs of s cover, relative to the pen start and
// the baseline (y grows down; x1, y1 exclusive). Centring the ink rather
// than the advance box keeps "100%" visually centred whatever the font's
// side bearings and ascent.
func inkBox(f tinyfont.Fonter, s string) (x0, y0, x1, y1 int16) {
	pen := int16(0)
	first := true
	for _, c := range s {
		g := f.GetGlyph(c).Info()
		if g.Width > 0 && g.Height > 0 {
			gx, gy := pen+int16(g.XOffset), int16(g.YOffset)
			gx1, gy1 := gx+int16(g.Width), gy+int16(g.Height)
			if first {
				x0, y0, x1, y1, first = gx, gy, gx1, gy1, false
			} else {
				x0, y0, x1, y1 = min(x0, gx), min(y0, gy), max(x1, gx1), max(y1, gy1)
			}
		}
		pen += int16(g.XAdvance)
	}
	return x0, y0, x1, y1
}
