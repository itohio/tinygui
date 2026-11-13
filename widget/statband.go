package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
)

// StatBandStyle is how stat bands look (shared appearance): the six strips
// of a calibration page share one, so they change together.
type StatBandStyle struct {
	BG        color.RGBA    // behind the trace; zero value: not painted
	Zone      color.RGBA    // the tolerance zone, centre ± threshold (see WithBandZone); zero value: not painted
	Threshold color.RGBA    // dashed 1-px lines at the zone's edges; zero value: none
	Drift     color.RGBA    // the mean's 1-px line
	States    [3]color.RGBA // the band's colour by column state: 0 settled, 1 warn, 2 bad
	BandAlpha uint8         // the band's opacity over Zone/BG (0..255; ux: ~40 % = 102; 0 means 102)
}

// StatBandOption is an optional knob of NewStatBand (convention: knobs with
// a sensible default are functional options).
type StatBandOption func(*statBandConfig)

// statBandConfig holds the knobs the options set; zero values are the
// defaults.
type statBandConfig struct {
	threshold float32 // the zone's half-height in data units; ≤ 0: no zone
}

// WithBandZone shades centre ± threshold (data units) as the tolerance
// zone, with dashed threshold lines on its top and bottom rows: the fixed
// target a settled band visibly shrinks into. Default: no zone; a
// threshold ≤ 0 keeps it off.
func WithBandZone(threshold float32) StatBandOption {
	return func(c *statBandConfig) { c.threshold = threshold }
}

// statCol is one pushed column, already in rows: 3 bytes, so a 160-column
// history is 480 bytes.
type statCol struct {
	mid   uint8 // the mean's row, 0 = top
	half  uint8 // the band's half-height in rows
	state uint8 // 0 settled, 1 warn, 2 bad
}

// StatBand is a strip showing a quantity's mean as a line and its spread as
// a band around the mean, over time (a calibration page stacks one per
// sensor axis). It is a plain non-interactive widget, so it works in every
// container and layout. Data is mapped to rows once per Push; Draw reads the
// ring buffer with integer math only and allocates nothing.
type StatBand struct {
	ui.WidgetBase
	style     StatBandStyle
	fullScale float32
	zoneQ     uint32 // the zone's half-height as a fraction of fullScale, in 1/65536; 0: no zone

	cols []statCol // ring buffer, width columns
	head int       // the slot the next Push writes
	n    int       // columns pushed, up to len(cols)

	// the band's colours, blended once here: [over BG, over Zone][state],
	// and the newest column's brighter ones
	band, bright [2][3]color.RGBA
}

// NewStatBand is a strip showing a quantity's mean as a line and its spread
// as a band around the mean, over time: one column per Push, newest at the
// right edge, oldest scrolling off the left. The centre row is the target
// (dev 0); ±fullScale (data units) reaches the top and bottom edges, beyond
// clamps. History is width columns, allocated here. Rows are stored as
// bytes, so heights past 256 rows clamp the band to the top 256.
func NewStatBand(width, height uint16, fullScale float32, style StatBandStyle, opts ...StatBandOption) *StatBand {
	b := &StatBand{
		WidgetBase: ui.NewWidgetBase(width, height),
		style:      style,
		fullScale:  fullScale,
		cols:       make([]statCol, width),
	}
	var cfg statBandConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	if cfg.threshold > 0 && fullScale > 0 {
		r := cfg.threshold / fullScale
		if r > 2 { // past the edges anyway
			r = 2
		}
		b.zoneQ = max(1, uint32(r*65536+0.5))
	}
	alpha := style.BandAlpha
	if alpha == 0 {
		alpha = 102
	}
	a := int32(alpha) + int32(alpha>>7) // 0..255 → 0..256
	br := a + (coverFull-a)/3           // the newest column: a shade brighter
	for i, under := range [2]color.RGBA{style.BG, style.Zone} {
		for s, c := range style.States {
			b.band[i][s] = opaque(mixRGBA(under, c, a))
			b.bright[i][s] = opaque(mixRGBA(under, c, br))
		}
	}
	return b
}

// opaque is c with full alpha: a band blended over an unpainted (zero) BG
// is blended over black, and stays opaque.
func opaque(c color.RGBA) color.RGBA { c.A = 255; return c }

// rows is v (data units) in rows at the current height, rounded and clamped
// to ±lim: one float32 multiply-divide, done at Push only.
func (b *StatBand) rows(v float32, lim int32) int32 {
	if !(b.fullScale > 0) {
		return 0
	}
	r := v * float32(b.Height/2) / b.fullScale
	switch {
	case !(r <= float32(lim)): // also NaN
		return lim
	case r < -float32(lim):
		return -lim
	case r < 0:
		return -int32(-r + 0.5)
	}
	return int32(r + 0.5)
}

// Push adds one column: dev is the mean minus its target, spread the band's
// half-width (both in the same units as fullScale), state 0/1/2 picks the
// band colour (larger counts as 2). The mapping to rows happens here, once
// per push, never in Draw; a later size change applies to later columns.
func (b *StatBand) Push(dev, spread float32, state uint8) {
	if len(b.cols) == 0 {
		return
	}
	h := int32(min(b.Height, 256))
	mid := int32(b.Height/2) - b.rows(dev, int32(b.Height))
	mid = min(max(mid, 0), max(h-1, 0))
	half := int32(0)
	if spread > 0 {
		half = b.rows(spread, 255)
	}
	b.cols[b.head] = statCol{mid: uint8(mid), half: uint8(half), state: min(state, 2)}
	b.head++
	if b.head == len(b.cols) {
		b.head = 0
	}
	b.n = min(b.n+1, len(b.cols))
}

// Reset clears the history (no columns drawn until the next Push), e.g.
// when a new calibration starts.
func (b *StatBand) Reset() { b.head, b.n = 0, 0 }

// zone is the zone's rows z0..z1 (inclusive) at the current height; z0 > z1
// without a zone. Integer only.
func (b *StatBand) zone() (z0, z1 int16) {
	h := int16(b.Height)
	if b.zoneQ == 0 {
		return h, h - 1
	}
	c := int32(b.Height / 2)
	t := int32((uint64(b.zoneQ)*uint64(b.Height/2) + 1<<15) >> 16)
	return int16(max(c-t, 0)), int16(min(c+t, int32(h)-1))
}

// Draw paints the strip inside the display's clip (a framebuffer band draws
// only its rows): zone and background, each column's band blended over
// them, the dashed threshold lines, then the mean's line on top. Columns
// are vertical runs (one fill each), read from the ring buffer across the
// wrap, oldest at the left.
func (b *StatBand) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || b.Width == 0 || b.Height == 0 {
		return
	}
	x0, y0 := ctx.DisplayPos()
	w, h := int16(b.Width), int16(b.Height)
	r0, r1 := int16(0), h // the rows to draw
	if c, ok := d.(ui.Clipper); ok {
		_, cy, _, ch := c.Clip()
		r0, r1 = max(r0, cy-y0), min(r1, cy+ch-y0)
	}
	if r0 >= r1 {
		return
	}
	z0, z1 := b.zone()
	st := &b.style
	// background and zone: three wide runs
	if !isZeroColor(st.BG) {
		vfill(d, x0, y0, w, 0, z0-1, r0, r1, st.BG)
		vfill(d, x0, y0, w, z1+1, h-1, r0, r1, st.BG)
	}
	if z0 <= z1 && !isZeroColor(st.Zone) {
		vfill(d, x0, y0, w, z0, z1, r0, r1, st.Zone)
	}

	n := min(b.n, int(w), len(b.cols))
	first := w - int16(n) // the leftmost drawn column's x
	start := b.head - n   // its slot
	if start < 0 {
		start += len(b.cols)
	}
	// bands, clipped to the widget, split where they cross the zone
	slot := start
	for x := first; x < w; x++ {
		col := b.cols[slot]
		colours := &b.band
		if x == w-1 {
			colours = &b.bright
		}
		mid, half := int16(col.mid), int16(col.half)
		t, u := max(mid-half, 0), min(mid+half, h-1)
		if half > 0 {
			cx := x0 + x
			vfill(d, cx, y0, 1, t, min(u, z0-1), r0, r1, colours[0][col.state])
			vfill(d, cx, y0, 1, max(t, z0), min(u, z1), r0, r1, colours[1][col.state])
			vfill(d, cx, y0, 1, max(t, z1+1), u, r0, r1, colours[0][col.state])
		}
		if slot++; slot == len(b.cols) {
			slot = 0
		}
	}
	// dashed threshold lines, 2 on 2 off, anchored at the right edge
	if z0 <= z1 && !isZeroColor(st.Threshold) {
		for x := w - 2; x > -2; x -= 4 {
			l := max(x, 0)
			vfill(d, x0+l, y0, x+2-l, z0, z0, r0, r1, st.Threshold)
			vfill(d, x0+l, y0, x+2-l, z1, z1, r0, r1, st.Threshold)
		}
	}
	// the mean: a 1-px line, a column joined to the previous one by a
	// vertical step so a fast drift stays connected
	slot, prev := start, int16(0)
	for x := first; x < w; x++ {
		mid := min(int16(b.cols[slot].mid), h-1)
		t, u := mid, mid
		if x > first {
			p := prev
			if mid > p+1 {
				t = p + 1
			} else if mid < p-1 {
				u = p - 1
			}
		}
		vfill(d, x0+x, y0, 1, t, u, r0, r1, st.Drift)
		prev = mid
		if slot++; slot == len(b.cols) {
			slot = 0
		}
	}
}

// vfill fills rows top..bot (inclusive, relative to y0) of a cw-wide run at
// x, cut to the clipped rows r0..r1-1; empty runs draw nothing.
func vfill(d drivers.Displayer, x, y0, cw, top, bot, r0, r1 int16, c color.RGBA) {
	top, bot = max(top, r0), min(bot, r1-1)
	if top > bot || cw <= 0 {
		return
	}
	fillRect(d, x, y0+top, cw, bot-top+1, c)
}
