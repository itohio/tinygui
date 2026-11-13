package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// Scope draws an oscilloscope screen: a DivX×DivY grid and a trace of
// caller-owned samples spread across the width. With more samples than
// pixel columns each column shows the min…max of its samples (nothing is
// lost between pixels, like a scope's peak-detect mode); with fewer, the
// samples are joined by vertical steps. The trigger level is a tick on the
// left edge and the trigger point a tick on the top edge (inside the widget
// at its edges). The caller
// updates Samples (and the range) in place; Draw reads them and allocates
// nothing.
type Scope struct {
	ui.WidgetBase
	Samples    []uint16 // the trace, oldest first
	Lo, Hi     uint16   // vertical range: Lo at the bottom edge, Hi at the top
	DivX, DivY int      // grid divisions
	Trigger    int      // trigger level in sample units, < 0: none
	TrigAt     int      // index into Samples of the trigger point, < 0: none
	ScopeStyle          // colours; promoted: s.BG, s.Trace, …
}

// ScopeStyle is how a Scope looks.
type ScopeStyle struct {
	BG, Grid color.RGBA
	Axis     color.RGBA // the centre lines of the grid
	Trace    color.RGBA
	Marker   color.RGBA // trigger ticks
}

// NewScope makes a w×h scope with a 10×8 grid and no trigger marks.
func NewScope(w, h uint16, style ScopeStyle) *Scope {
	return &Scope{WidgetBase: ui.NewWidgetBase(w, h), Hi: 0xffff, DivX: 10, DivY: 8,
		Trigger: -1, TrigAt: -1, ScopeStyle: style}
}

// y is the row of value v (clamped to the widget).
func (s *Scope) y(v int) int16 {
	h := int(s.Height) - 1
	lo, hi := int(s.Lo), int(s.Hi)
	if hi <= lo {
		return int16(h / 2)
	}
	r := h - (v-lo)*h/(hi-lo)
	return int16(min(max(r, 0), h))
}

func (s *Scope) Draw(ctx ui.Context) {
	d := ctx.D()
	x0, y0 := ctx.DisplayPos()
	w, h := int16(s.Width), int16(s.Height)
	fillRect(d, x0, y0, w, h, s.BG)
	for i := 1; i < s.DivX; i++ {
		c := s.Grid
		if 2*i == s.DivX {
			c = s.Axis
		}
		fillRect(d, x0+int16(int(w)*i/s.DivX), y0, 1, h, c)
	}
	for i := 1; i < s.DivY; i++ {
		c := s.Grid
		if 2*i == s.DivY {
			c = s.Axis
		}
		fillRect(d, x0, y0+int16(int(h)*i/s.DivY), w, 1, c)
	}
	n := len(s.Samples)
	if n > 0 {
		prev := s.y(int(s.Samples[0]))
		for x := range int(w) {
			i0 := x * n / int(w)
			i1 := max(i0+1, (x+1)*n/int(w))
			top, bot := prev, prev // join the previous column's last sample
			for _, v := range s.Samples[i0:min(i1, n)] {
				r := s.y(int(v))
				top, bot = min(top, r), max(bot, r)
			}
			fillRect(d, x0+int16(x), y0+top, 1, bot-top+1, s.Trace)
			prev = s.y(int(s.Samples[min(i1, n)-1]))
		}
	}
	// the ticks are 3 px wide across their position, kept inside the widget
	// at the edges (a level at Hi or Lo, a trigger point at either end)
	if s.Trigger >= 0 {
		fillRect(d, x0, y0+min(max(s.y(s.Trigger)-1, 0), h-3), 5, 3, s.Marker)
	}
	if s.TrigAt >= 0 && n > 0 {
		tx := int16(min(s.TrigAt, n-1) * int(w) / n)
		fillRect(d, x0+min(max(tx-1, 0), w-3), y0, 3, 5, s.Marker)
	}
}
