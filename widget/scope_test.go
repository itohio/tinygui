package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, []int{14}, traceRows(c, 0))                                   // 5 → row 19-5
	require.Equal(t, []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, traceRows(c, 1)) // spike to 15 → row 4
	require.Equal(t, []int{14, 15, 16, 17}, traceRows(c, 2))                       // from 5 down to 2 → rows 14..17
	require.Equal(t, []int{17}, traceRows(c, 3))
}

// Fewer samples than columns: each sample spreads over its columns with a
// vertical step where it changes; values outside Lo…Hi stay on the edges.
func TestScopeStretchesFewSamplesAndClamps(t *testing.T) {
	s, c, ctx := newTestScope(6, 10)
	s.Samples = []uint16{2, 7, 50} // 2 columns each; 50 is above Hi (9)
	s.Draw(&ctx)
	require.Equal(t, []int{7}, traceRows(c, 1))
	require.Equal(t, []int{2, 3, 4, 5, 6, 7}, traceRows(c, 2)) // step 2 → 7
	require.Equal(t, []int{0, 1, 2}, traceRows(c, 4))          // step 7 → clamped to the top row
	require.Equal(t, []int{0}, traceRows(c, 5))
}

// Drawing allocates nothing (it runs every frame, once per band).
func TestScopeDrawsWithoutAllocating(t *testing.T) {
	s, _, ctx := newTestScope(300, 150)
	s.Samples = make([]uint16, 2048)
	for i := range s.Samples {
		s.Samples[i] = uint16(i % 150)
	}
	s.Trigger, s.TrigAt = 75, 100
	require.Zero(t, testing.AllocsPerRun(20, func() { s.Draw(&ctx) }))
}
