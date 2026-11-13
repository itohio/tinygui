package widget

import (
	"image/color"
	"math"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
)

// The arc-raster core: the per-pixel work of every round widget (the Ring
// now; segmented rings, arc gauges, pies, spinners, multi-rings later). The
// RP2040 has no FPU, so nothing per pixel uses floating point or
// trigonometry: arcRaster.set turns the geometry into integers once
// (float32 sine/cosine of the span edges, when the value or the size
// changes), and arcRaster.draw classifies pixels with squared radii and
// cross products, smoothing every edge over one pixel, painting runs of
// one colour.
//
// Coordinates are in half pixels from the box's top-left corner, so pixel
// centres are odd integers and a d-pixel circle centred in a d×d square has
// its centre at (d, d) and radius d: no rounding for even or odd sizes.
// Angles are in turn units (fullTurn per turn, so they wrap in uint16),
// clockwise from 12 o'clock.

// fullTurn is a whole turn in the core's fixed-point angle units.
const fullTurn = 1 << 16

// degTurn converts degrees (any value) to turn units.
func degTurn(deg int16) uint16 {
	return uint16(int32(deg) % 360 * fullTurn / 360)
}

// arcSpan is the angular extent of an arc: from start, clockwise, by sweep
// (0: nothing, fullTurn or more: the whole circle).
type arcSpan struct {
	start uint16
	sweep uint32
}

// arcRaster draws an annulus (or a pie, with inner radius 0) whose track is
// an arcSpan and whose filled part is another arcSpan inside it, with butt
// or round caps on both spans' ends. Pixels outside the track are left to
// BG; inside it they are Track or Fill, mixed across the edges.
type arcRaster struct {
	cx, cy       int32 // centre, half pixels from the box's corner
	outer, inner disc
	track, fill  sector
}

// set computes the integer geometry: centre (cx, cy) and radii outer >
// inner ≥ 0, all in half pixels; the track and fill spans; round caps
// (half discs as wide as the ring, centred on its mid radius, at the ends
// of a span that is neither empty nor whole). Called when the geometry or
// the value changes, not per draw.
func (a *arcRaster) set(cx, cy, outer, inner int32, track, fill arcSpan, round bool) {
	inner = min(max(inner, 0), outer)
	a.cx, a.cy = cx, cy
	a.outer, a.inner = newDisc(outer), newDisc(inner)
	mid, capR := (outer+inner)/2, (outer-inner)/2
	if !round {
		capR = 0
	}
	a.track.set(track, mid, capR)
	a.fill.set(fill, mid, capR)
}

// draw paints the box of w×h pixels at (x0, y0) on d, row by row inside the
// display's clip (a framebuffer band draws only its rows). bg zero: pixels
// outside the arc are left as they are and edges toward them are hard (a
// pixel is painted when its centre is in the arc); else they are painted bg
// and the edges are blended into it. p carries the row runs (a field of the
// widget, so nothing escapes to the heap). It returns the rows it covered.
func (a *arcRaster) draw(p *spanPainter, d drivers.Displayer, x0, y0, w, h int16, track, fill, bg color.RGBA) (row0, row1 int16) {
	row0, row1 = 0, h
	if c, ok := d.(ui.Clipper); ok {
		_, cy, _, ch := c.Clip()
		row0, row1 = max(row0, cy-y0), min(row1, cy+ch-y0)
	}
	hasBG := !isZeroColor(bg)
	p.d = d
	for row := row0; row < row1; row++ {
		p.begin(y0 + row)
		dy := int32(2*row+1) - a.cy
		l, r, hl, hr := a.columns(dy, int32(w))
		p.push(x0, int16(l), bg, hasBG) // left of the circle
		for i := l; i <= r; i++ {
			if i == hl && hl <= hr { // wholly inside the inner circle
				p.push(x0+int16(hl), int16(hr-hl+1), bg, hasBG)
				i = hr
				continue
			}
			c, on := a.shade(int32(2*i+1)-a.cx, dy, track, fill, bg, hasBG)
			p.push(x0+int16(i), 1, c, on)
		}
		right := int16(max(r+1, l)) // r < l: an empty row, all bg
		p.push(x0+right, w-right, bg, hasBG)
		p.flush()
	}
	return row0, row1
}

// columns is which box columns (0..w-1) of the row dy half pixels from the
// centre need shading: l..r (l > r: none) touch the outer circle, and hl..hr
// (hl > hr: none) among them lie wholly inside the inner one. Two integer
// square roots per row.
func (a *arcRaster) columns(dy, w int32) (l, r, hl, hr int32) {
	hl, hr = 0, -1
	o := a.outer.r2 + 2*a.outer.r - dy*dy // outer coverage > 0 while dx² < o
	if a.outer.r <= 0 || o <= 0 {
		return 0, -1, hl, hr
	}
	x := isqrt(o - 1)
	l, r = max(0, (a.cx-x)>>1), min(w-1, (a.cx-1+x)>>1)
	if a.inner.r > 0 {
		if in := a.inner.r2 - 2*a.inner.r - dy*dy; in >= 0 { // inner coverage full while dx² ≤ in
			xi := isqrt(in)
			hl, hr = max(l, (a.cx-xi)>>1), min(r, (a.cx-1+xi)>>1)
		}
	}
	return l, r, hl, hr
}

// shade is the colour of the pixel at (dx, dy) half pixels from the centre
// and whether it is painted: Track mixed toward Fill by the fill's
// coverage, then toward bg by the arc's coverage (ring × track span) at its
// edges.
func (a *arcRaster) shade(dx, dy int32, track, fill, bg color.RGBA, hasBG bool) (color.RGBA, bool) {
	d2 := dx*dx + dy*dy
	cov := min(a.outer.coverage(d2), coverFull-a.inner.coverage(d2))
	if cov > 0 && !a.track.full {
		cov = min(cov, a.track.coverage(dx, dy))
	}
	if cov == 0 {
		return bg, hasBG
	}
	c := mixRGBA(track, fill, a.fill.coverage(dx, dy))
	if cov < coverFull {
		if hasBG {
			return mixRGBA(bg, c, cov), true
		}
		return c, cov >= coverFull/2
	}
	return c, true
}

// coverFull is full coverage: coverages are 0..256 (256ths of a pixel).
const coverFull = 256

// mixRGBA blends from a into b by t/256 (t = 0: a, 256: b), integer only.
func mixRGBA(a, b color.RGBA, t int32) color.RGBA {
	switch t {
	case 0:
		return a
	case coverFull:
		return b
	}
	u := coverFull - t
	return color.RGBA{
		R: uint8((int32(a.R)*u + int32(b.R)*t) >> 8),
		G: uint8((int32(a.G)*u + int32(b.G)*t) >> 8),
		B: uint8((int32(a.B)*u + int32(b.B)*t) >> 8),
		A: uint8((int32(a.A)*u + int32(b.A)*t) >> 8),
	}
}

func clampCover(c int32) int32 { return min(max(c, 0), coverFull) }

// disc is a circle's coverage test: r is its radius in half pixels.
type disc struct {
	r, r2 int32
	k     int32 // (64<<16)/r: the coverage slope across the edge
}

func newDisc(r int32) disc {
	if r <= 0 {
		return disc{}
	}
	return disc{r: r, r2: r * r, k: (64 << 16) / r}
}

// coverage is how much of the pixel at squared distance d2 (half pixels)
// from the centre the disc covers, 0..256: full inside, none outside, a
// linear ramp over the pixel straddling the edge. That pixel's distance
// inside the edge is (r²-d2)/(4r) pixels (r-√d2 to first order): no square
// root.
func (c *disc) coverage(d2 int32) int32 {
	if c.r <= 0 {
		return 0
	}
	e := c.r2 - d2
	switch {
	case e >= 2*c.r:
		return coverFull
	case e <= -2*c.r:
		return 0
	}
	return coverFull/2 + (e*c.k)>>16
}

// sector is an arcSpan's angular test: the pixels clockwise from direction
// a to direction b (Q14 unit vectors, screen axes: y down), plus round caps.
// A sweep up to half a turn is the intersection of the two half planes, a
// wider one their union; empty and whole spans are flags.
type sector struct {
	ax, ay, bx, by int32
	wide           bool
	empty, full    bool
	caps           bool
	cap            [2]struct{ x, y int32 } // cap centres, half pixels from the centre
	capDisc        disc
}

// set computes the sector of span, with caps of radius capR (half pixels,
// 0: butt ends) centred at radius mid on its edges. Two float32 sine and
// cosine pairs.
func (s *sector) set(span arcSpan, mid, capR int32) {
	*s = sector{empty: span.sweep == 0, full: span.sweep >= fullTurn}
	if s.empty || s.full {
		return
	}
	s.ax, s.ay = turnVector(span.start)
	s.bx, s.by = turnVector(span.start + uint16(span.sweep))
	s.wide = span.sweep > fullTurn/2
	if capR > 0 {
		s.caps = true
		s.cap[0].x, s.cap[0].y = (s.ax*mid+1<<13)>>14, (s.ay*mid+1<<13)>>14
		s.cap[1].x, s.cap[1].y = (s.bx*mid+1<<13)>>14, (s.by*mid+1<<13)>>14
		s.capDisc = newDisc(capR)
	}
}

// coverage is how much of the pixel at (dx, dy) half pixels from the centre
// the sector covers, 0..256. The cross product of a Q14 unit
// vector and (dx, dy) is the signed distance to that edge in 1/32768 px, so
// >>7 gives 256ths of a pixel.
func (s *sector) coverage(dx, dy int32) int32 {
	switch {
	case s.empty:
		return 0
	case s.full:
		return coverFull
	}
	ca := clampCover(coverFull/2 + (s.ax*dy-s.ay*dx)>>7) // clockwise of a: cross(a, p) > 0
	cb := clampCover(coverFull/2 - (s.bx*dy-s.by*dx)>>7) // short of b: cross(b, p) < 0
	c := min(ca, cb)
	if s.wide {
		c = max(ca, cb)
	}
	if s.caps && c < coverFull {
		for i := range s.cap {
			ex, ey := dx-s.cap[i].x, dy-s.cap[i].y
			c = max(c, s.capDisc.coverage(ex*ex+ey*ey))
		}
	}
	return c
}

// turnVector is the Q14 unit vector pointing at angle turn (clockwise from
// 12 o'clock) in screen axes (y down): (sin, -cos).
func turnVector(turn uint16) (x, y int32) {
	s, c := sinCosQ14(turn)
	return s, -c
}

// sinCosQ14 is the sine and cosine of turn (fullTurn = 360°) in Q14, from
// float32 Taylor polynomials on a quarter turn (error under 4e-6, well below
// one Q14 step): math.Sincos is float64, soft-float on the RP2040.
func sinCosQ14(turn uint16) (s, c int32) {
	const quarter = math.Pi / 2 / (fullTurn / 4)
	x := float32(turn&(fullTurn/4-1)) * quarter
	x2 := x * x
	sn := x * (1 - x2*(1.0/6)*(1-x2*(1.0/20)*(1-x2*(1.0/42)*(1-x2*(1.0/72)))))
	cs := 1 - x2*(1.0/2)*(1-x2*(1.0/12)*(1-x2*(1.0/30)*(1-x2*(1.0/56)*(1-x2*(1.0/90)))))
	is, ic := int32(sn*(1<<14)+0.5), int32(cs*(1<<14)+0.5)
	switch turn >> 14 {
	case 0:
		return is, ic
	case 1:
		return ic, -is
	case 2:
		return -is, -ic
	}
	return -ic, is
}

// isqrt is ⌊√n⌋ (0 for n ≤ 0), bit by bit: no float, no division.
func isqrt(n int32) int32 {
	if n <= 0 {
		return 0
	}
	var r int32
	bit := int32(1) << 30
	for bit > n {
		bit >>= 2
	}
	for bit != 0 {
		if n >= r+bit {
			n -= r + bit
			r = r>>1 + bit
		} else {
			r >>= 1
		}
		bit >>= 2
	}
	return r
}

// spanPainter merges a row's pixels into runs of one colour, so a row costs
// a few fillRect calls instead of a display call per pixel. Runs that don't
// paint (no BG) only end the current run.
type spanPainter struct {
	d     drivers.Displayer
	y     int16
	x, n  int16
	c     color.RGBA
	paint bool
}

// begin starts row y.
func (p *spanPainter) begin(y int16) { p.y, p.n = y, 0 }

// push adds n pixels from x in colour c (paint false: left as they are).
func (p *spanPainter) push(x, n int16, c color.RGBA, paint bool) {
	if n <= 0 {
		return
	}
	if p.n > 0 && x == p.x+p.n && paint == p.paint && (!paint || c == p.c) {
		p.n += n
		return
	}
	p.flush()
	p.x, p.n, p.c, p.paint = x, n, c, paint
}

// flush paints the pending run.
func (p *spanPainter) flush() {
	if p.n > 0 && p.paint {
		fillRect(p.d, p.x, p.y, p.n, 1, p.c)
	}
	p.n = 0
}
