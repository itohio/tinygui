package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
)

var (
	sbBG    = color.RGBA{10, 14, 20, 255}
	sbZone  = color.RGBA{14, 42, 34, 255}
	sbThr   = color.RGBA{45, 212, 191, 255}
	sbDrift = color.RGBA{230, 237, 243, 255}
	sbOK    = color.RGBA{52, 211, 153, 255}
	sbWarn  = color.RGBA{251, 191, 36, 255}
	sbBad   = color.RGBA{248, 113, 113, 255}

	sbStyle = StatBandStyle{BG: sbBG, Zone: sbZone, Threshold: sbThr, Drift: sbDrift,
		States: [3]color.RGBA{sbOK, sbWarn, sbBad}, BandAlpha: 128}
)

// blended is state colour c over under at BandAlpha 128 (129/256), opaque;
// bright is the newest column's shade, a third of the way on to c.
func blended(under, c color.RGBA) color.RGBA { return opaque(mixRGBA(under, c, 129)) }
func bright(under, c color.RGBA) color.RGBA {
	return opaque(mixRGBA(under, c, 129+(256-129)/3))
}

// drawBand draws b on a canvas of its size pre-filled with rUnset.
func drawBand(b *StatBand) *canvas {
	c := newCanvas(int16(b.Width), int16(b.Height))
	for i := range c.px {
		c.px[i] = rUnset
	}
	ctx := ui.NewContext(c, b.Width, b.Height, 0, 0)
	b.Draw(&ctx)
	return c
}

// column is column x of c, top to bottom.
func column(c *canvas, x int16) []color.RGBA {
	col := make([]color.RGBA, c.h)
	for y := range c.h {
		col[y] = c.at(x, y)
	}
	return col
}

// fill pushes the same column width times (a flat history, so the mean's
// line has no steps).
func fill(b *StatBand, dev, spread float32, state uint8) {
	for range b.Width {
		b.Push(dev, spread, state)
	}
}

// 8×21 with fullScale 10: the centre row is 10 and one data unit is one
// row, so dev d sits at row 10-d.
func newTestBand(opts ...StatBandOption) *StatBand { return NewStatBand(8, 21, 10, sbStyle, opts...) }

// A pushed column's band covers exactly the rows of dev ± spread, the mean
// is one pixel at dev's row on top of it, the rest is BG.
func TestStatBandColumnCoversDevPlusMinusSpread(t *testing.T) {
	b := newTestBand()
	fill(b, 3, 2, 1)
	col := column(drawBand(b), 3)
	for y, got := range col {
		want := sbBG
		switch {
		case y == 7:
			want = sbDrift
		case y >= 5 && y <= 9:
			want = blended(sbBG, sbWarn)
		}
		check.Equal(t, want, got, "row %d", y)
	}
}

// dev 0 is the centre row; ±fullScale reaches the top and bottom edges and
// beyond clamps there — for odd and even heights.
func TestStatBandCentreAndEdges(t *testing.T) {
	for _, c := range []struct {
		h   uint16
		dev float32
		row int
	}{
		{21, 0, 10}, {21, 10, 0}, {21, 25, 0}, {21, -10, 20}, {21, -25, 20}, {21, 4, 6},
		{26, 0, 13}, {26, 10, 0}, {26, 1e9, 0}, {26, -10, 25}, {26, -1e9, 25},
	} {
		b := NewStatBand(4, c.h, 10, sbStyle)
		fill(b, c.dev, 0, 0)
		col := column(drawBand(b), 1)
		for y, got := range col {
			check.Equal(t, y == c.row, got == sbDrift, "h %d dev %v: row %d is %v", c.h, c.dev, y, got)
		}
	}
}

// Spread 0 (or negative) draws just the mean's pixel: no band.
func TestStatBandZeroSpreadIsJustTheMean(t *testing.T) {
	for _, spread := range []float32{0, -3, 0.4} { // 0.4 rows rounds to none
		b := newTestBand()
		fill(b, -2, spread, 2)
		col := column(drawBand(b), 2)
		for y, got := range col {
			want := sbBG
			if y == 12 {
				want = sbDrift
			}
			check.Equal(t, want, got, "spread %v, row %d", spread, y)
		}
	}
}

// meanRow is the row of the mean's pixel in column x, -1 if there is none.
func meanRow(c *canvas, x int16) int {
	r := -1
	for y, p := range column(c, x) {
		if p == sbDrift {
			r = y
		}
	}
	return r
}

// Columns appear at the right edge and scroll left; columns not yet pushed
// show only the background; after width pushes the oldest drops off the
// left, read across the ring buffer's wrap in order.
func TestStatBandScrollsNewestRight(t *testing.T) {
	b := NewStatBand(4, 21, 10, sbStyle)
	rows := func() []int {
		c := drawBand(b)
		return []int{meanRow(c, 0), meanRow(c, 1), meanRow(c, 2), meanRow(c, 3)}
	}
	check.Equal(t, []int{-1, -1, -1, -1}, rows(), "empty")
	b.Push(1, 0, 0)
	b.Push(2, 0, 0)
	check.Equal(t, []int{-1, -1, 9, 8}, rows(), "two pushes")
	for d := float32(3); d <= 6; d++ {
		b.Push(d, 0, 0)
	}
	check.Equal(t, []int{7, 6, 5, 4}, rows(), "six pushes into four columns: 1 and 2 dropped")
	for d := float32(7); d <= 9; d++ {
		b.Push(d, 0, 0)
	}
	check.Equal(t, []int{4, 3, 2, 1}, rows(), "nine pushes")
}

// The state picks the band's colour, blended over the Zone inside the
// tolerance zone and over BG outside it; the newest column is a shade
// brighter. Zone ±3 is rows 7..13; with 8 columns the dashes are on at
// columns 2, 3, 6, 7 (anchored at the right), off at 1.
func TestStatBandStateColourOverZoneAndBG(t *testing.T) {
	for s, sc := range sbStyle.States {
		b := newTestBand(WithBandZone(3))
		fill(b, 0, 6, uint8(s))
		c := drawBand(b)
		for y, got := range column(c, 1) {
			want := blended(sbBG, sc)
			switch {
			case y == 10:
				want = sbDrift
			case y >= 7 && y <= 13:
				want = blended(sbZone, sc)
			case y < 4 || y > 16: // past the band (rows 4..16)
				want = sbBG
			}
			check.Equal(t, want, got, "state %d, row %d", s, y)
		}
		check.Equal(t, sbThr, c.at(2, 7), "state %d: the dash is drawn over the band", s)
		check.Equal(t, bright(sbZone, sc), c.at(7, 9), "state %d: the newest column in the zone", s)
		check.Equal(t, bright(sbBG, sc), c.at(7, 15), "state %d: the newest column outside the zone", s)
	}
	b := newTestBand()
	fill(b, 0, 2, 9) // out of range: bad
	check.Equal(t, blended(sbBG, sbBad), drawBand(b).at(0, 9), "state 9 draws as bad")
}

// The zone spans exactly ±threshold: its edge rows are where a mean of
// ±threshold sits, with the dashed lines on them; before any push the strip
// is just zone and BG.
func TestStatBandZoneSpansThreshold(t *testing.T) {
	b := newTestBand(WithBandZone(3))
	c := drawBand(b)
	for x := range int16(8) {
		on := x == 2 || x == 3 || x == 6 || x == 7
		for y, got := range column(c, x) {
			want := sbBG
			switch {
			case (y == 7 || y == 13) && on:
				want = sbThr
			case y >= 7 && y <= 13:
				want = sbZone
			}
			check.Equal(t, want, got, "column %d, row %d", x, y)
		}
	}
	// ux's size and scale: 26 rows, fullScale 3 × threshold
	for _, dev := range []float32{1, -1} {
		b := NewStatBand(162, 26, 3, sbStyle, WithBandZone(1))
		fill(b, dev, 0, 0)
		c := drawBand(b)
		r := meanRow(c, 0)
		check.Equal(t, sbZone, c.at(1, int16(r+int(dev))), "dev %v: the row inside the edge is zone", dev)
		check.Equal(t, sbBG, c.at(1, int16(r-int(dev))), "dev %v: the row outside the edge is BG", dev)
	}
	check.Equal(t, uint32(0), newTestBand(WithBandZone(0)).zoneQ, "threshold 0: no zone")
}

// Reset clears the history: the next draw is a fresh strip's.
func TestStatBandReset(t *testing.T) {
	b := newTestBand(WithBandZone(3))
	fill(b, 2, 3, 1)
	b.Reset()
	check.Equal(t, drawBand(newTestBand(WithBandZone(3))).px, drawBand(b).px)
	b.Push(-1, 0, 0)
	check.Equal(t, 11, meanRow(drawBand(b), 7), "pushing again after Reset")
	check.Equal(t, -1, meanRow(drawBand(b), 6), "only one column")
}

// With a clipping display (one band of a framebuffer) only the clip's rows
// are drawn, and those as in a full draw.
func TestStatBandDrawsOnlyClippedRows(t *testing.T) {
	b := newTestBand(WithBandZone(3))
	for i := range 8 {
		b.Push(float32(i-4), float32(i%3), uint8(i%3))
	}
	full := drawBand(b)
	c := &clippedCanvas{canvas: newCanvas(8, 21), cy: 5, ch: 4}
	for i := range c.px {
		c.px[i] = rUnset
	}
	ctx := ui.NewContext(c, 8, 21, 0, 0)
	b.Draw(&ctx)
	for y := range int16(21) {
		for x := range int16(8) {
			want := rUnset
			if y >= 5 && y < 9 {
				want = full.at(x, y)
			}
			check.Equal(t, want, c.at(x, y), "(%d, %d)", x, y)
		}
	}
	c = &clippedCanvas{canvas: newCanvas(8, 21), cy: 30, ch: 10}
	ctx = ui.NewContext(c, 8, 21, 0, 0)
	b.Draw(&ctx)
	check.Zero(t, c.fills, "a clip below the strip draws nothing")
}

// With a zero BG nothing is painted outside the zone except the band and
// the mean (the band is blended over black there).
func TestStatBandZeroBGIsNotPainted(t *testing.T) {
	st := sbStyle
	st.BG = color.RGBA{}
	b := NewStatBand(8, 21, 10, st)
	fill(b, 0, 2, 0)
	col := column(drawBand(b), 0)
	check.Equal(t, rUnset, col[0])
	check.Equal(t, blended(color.RGBA{}, sbOK), col[9])
	check.Equal(t, sbDrift, col[10])
}
