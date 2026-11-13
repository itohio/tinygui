package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
)

var (
	scBG    = color.RGBA{0, 0, 0, 255}
	scTrace = color.RGBA{0, 255, 0, 255}
	scGrid  = color.RGBA{30, 30, 30, 255}
)

func newTestScope(w, h uint16) (*Scope, *canvas, ui.ContextImpl) {
	s := NewScope(w, h, scBG, scGrid, scGrid, scTrace, color.RGBA{255, 0, 0, 255})
	s.Lo, s.Hi = 0, uint16(h-1) // one count per row: row = h-1-v
	c := newCanvas(int16(w), int16(h))
	return s, c, ui.NewContext(c, w, h, 0, 0)
}

// traceRows lists the rows of column x that hold the trace colour.
func traceRows(c *canvas, x int) []int {
	var r []int
	for y := range int(c.h) {
		if c.px[y*int(c.w)+x] == scTrace {
			r = append(r, y)
		}
	}
	return r
}

// More samples than columns: each column spans the min…max of its samples
// (a 1-sample spike between pixels still shows), joined to the column before.
func TestScopeColumnShowsMinToMaxOfItsSamples(t *testing.T) {
	s, c, ctx := newTestScope(4, 20)
	s.Samples = []uint16{5, 5, 5, 5, 5, 15, 5, 5, 2, 2, 2, 2, 2, 2, 2, 2} // 4 samples per column
	s.Draw(&ctx)
	check.Equal(t, []int{14}, traceRows(c, 0))                                   // 5 → row 19-5
	check.Equal(t, []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, traceRows(c, 1)) // spike to 15 → row 4
	check.Equal(t, []int{14, 15, 16, 17}, traceRows(c, 2))                       // from 5 down to 2 → rows 14..17
	check.Equal(t, []int{17}, traceRows(c, 3))
}

// Fewer samples than columns: each sample spreads over its columns with a
// vertical step where it changes; values outside Lo…Hi stay on the edges.
func TestScopeStretchesFewSamplesAndClamps(t *testing.T) {
	s, c, ctx := newTestScope(6, 10)
	s.Samples = []uint16{2, 7, 50} // 2 columns each; 50 is above Hi (9)
	s.Draw(&ctx)
	check.Equal(t, []int{7}, traceRows(c, 1))
	check.Equal(t, []int{2, 3, 4, 5, 6, 7}, traceRows(c, 2)) // step 2 → 7
	check.Equal(t, []int{0, 1, 2}, traceRows(c, 4))          // step 7 → clamped to the top row
	check.Equal(t, []int{0}, traceRows(c, 5))
}

// Drawing allocates nothing (it runs every frame, once per band).
func TestScopeDrawsWithoutAllocating(t *testing.T) {
	s, _, ctx := newTestScope(300, 150)
	s.Samples = make([]uint16, 2048)
	for i := range s.Samples {
		s.Samples[i] = uint16(i % 150)
	}
	s.Trigger, s.TrigAt = 75, 100
	check.Zero(t, testing.AllocsPerRun(20, func() { s.Draw(&ctx) }))
}

var (
	scAxis   = color.RGBA{90, 90, 90, 255}
	scMarker = color.RGBA{255, 0, 0, 255}
)

// The grid: DivX-1 vertical and DivY-1 horizontal lines at even spacing,
// the centre ones in the axis colour, the rest in the grid colour.
func TestScopeGridAndCentreAxes(t *testing.T) {
	s := NewScope(20, 16, scBG, scGrid, scAxis, scTrace, scMarker)
	s.DivX, s.DivY = 4, 4
	c := newCanvas(20, 16)
	ctx := ui.NewContext(c, 20, 16, 0, 0)
	s.Draw(&ctx)
	for _, p := range []struct {
		x, y int16
		want color.RGBA
	}{
		{5, 1, scGrid}, {10, 1, scAxis}, {15, 1, scGrid}, // vertical lines (row 1: no horizontal line)
		{1, 4, scGrid}, {1, 8, scAxis}, {1, 12, scGrid}, // horizontal lines
		{1, 1, scBG}, {6, 5, scBG}, {19, 15, scBG},
	} {
		check.Equal(t, p.want, c.at(p.x, p.y), "pixel (%d, %d)", p.x, p.y)
	}
}

// The trigger level is a 5×3 tick on the left edge centred on its row, the
// trigger point a 3×5 tick on the top edge centred on its column; at the
// edges they stay inside the widget (review R6: they drew 1-2 px outside).
func TestScopeTriggerTicks(t *testing.T) {
	const w, h = 40, 20
	for _, c := range []struct {
		name                 string
		trigger, trigAt      int
		wantLevel, wantPoint [4]int16 // x, y, w, h
	}{
		{"middle", 10, 32, [4]int16{0, 8, 5, 3}, [4]int16{19, 0, 3, 5}},
		{"level at Hi, point at the start", h - 1, 0, [4]int16{0, 0, 5, 3}, [4]int16{0, 0, 3, 5}},
		{"level at Lo, point at the end", 0, 63, [4]int16{0, h - 3, 5, 3}, [4]int16{w - 3, 0, 3, 5}},
	} {
		s := NewScope(w, h, scBG, scGrid, scAxis, scTrace, scMarker)
		s.Lo, s.Hi = 0, h-1
		s.Samples = make([]uint16, 64)
		s.Trigger, s.TrigAt = c.trigger, c.trigAt
		d := &rectLog{}
		ctx := ui.NewContext(d, w, h, 0, 0)
		s.Draw(&ctx)
		for _, r := range d.rects {
			if r[0] < 0 || r[1] < 0 || r[0]+r[2] > w || r[1]+r[3] > h {
				t.Fatalf("%s: a rectangle %v outside the %d×%d widget", c.name, r, w, h)
			}
		}
		n := len(d.rects)
		check.Equal(t, c.wantLevel, d.rects[n-2], c.name+": level tick")
		check.Equal(t, c.wantPoint, d.rects[n-1], c.name+": point tick")
	}
}

// Hi <= Lo has no range: the trace runs along the middle row.
func TestScopeWithoutARangeDrawsTheMiddle(t *testing.T) {
	for _, lohi := range [][2]uint16{{5, 5}, {9, 2}} {
		s, c, ctx := newTestScope(4, 11)
		s.Lo, s.Hi = lohi[0], lohi[1]
		s.Samples = []uint16{0, 3, 10, 7}
		s.Draw(&ctx)
		for x := range 4 {
			check.Equal(t, []int{5}, traceRows(c, x), "Lo %d Hi %d column %d", lohi[0], lohi[1], x)
		}
	}
}

// Each column joins on from the previous column's last sample, not its
// first: the trace is continuous in time.
func TestScopeJoinsFromTheColumnsLastSample(t *testing.T) {
	s, c, ctx := newTestScope(2, 20)
	s.Samples = []uint16{5, 5, 5, 10, 2, 2, 2, 2} // column 0 ends at 10 (row 9), column 1 is all 2 (row 17)
	s.Draw(&ctx)
	check.Equal(t, []int{9, 10, 11, 12, 13, 14}, traceRows(c, 0))
	check.Equal(t, []int{9, 10, 11, 12, 13, 14, 15, 16, 17}, traceRows(c, 1))
}
