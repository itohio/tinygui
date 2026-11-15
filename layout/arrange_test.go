package layout

import (
	"testing"

	"github.com/itohio/tinygui/internal/check"
)

type sizes [][2]uint16

func (s sizes) Len() int                      { return len(s) }
func (s sizes) SizeAt(i int) (uint16, uint16) { return s[i][0], s[i][1] }

func arrange(a Arranger, w, h uint16, s sizes) []Rect {
	out := make([]Rect, len(s))
	a.Arrange(w, h, s, out)
	return out
}

func pos(r []Rect) [][2]int16 {
	p := make([][2]int16, len(r))
	for i, v := range r {
		p[i] = [2]int16{v.X, v.Y}
	}
	return p
}

func TestRowPacksAndPlacesByMainAndCross(t *testing.T) {
	s := sizes{{10, 4}, {20, 8}}
	// 30 px of children + gap 2 in 100: free 68
	check.Equal(t, [][2]int16{{0, 0}, {12, 0}}, pos(arrange(Row(2, Start, Start), 100, 10, s)))
	check.Equal(t, [][2]int16{{34, 3}, {46, 1}}, pos(arrange(Row(2, Center, Center), 100, 10, s)))
	check.Equal(t, [][2]int16{{68, 6}, {80, 2}}, pos(arrange(Row(2, End, End), 100, 10, s)))
	check.Equal(t, [][2]int16{{0, 0}, {80, 0}}, pos(arrange(Row(2, SpaceBetween, Start), 100, 10, s)))
}

func TestColumnIsARowOnTheOtherAxis(t *testing.T) {
	s := sizes{{4, 10}, {8, 20}}
	check.Equal(t, [][2]int16{{3, 34}, {1, 46}}, pos(arrange(Column(2, Center, Center), 10, 100, s)))
}

// Weighted cells: a weight-0 child keeps its size, the others share the rest
// in proportion (here 1:3 of 100-10-2*2 = 86 → 21 and 65, the last taking the
// remainder) and sit in their cell by Main.
func TestRowWeightsSplitTheFreeSpace(t *testing.T) {
	s := sizes{{10, 4}, {6, 4}, {6, 4}}
	r := arrange(Row(2, Center, Start, 0, 1, 3), 100, 4, s)
	check.Equal(t, int16(0), r[0].X)
	// cell 1: x 12, width 21 → child centred at 12 + (21-6)/2 = 19
	check.Equal(t, int16(19), r[1].X)
	// cell 2: x 12+21+2 = 35, width 65 → 35 + (65-6)/2 = 64
	check.Equal(t, int16(64), r[2].X)
}

func TestWrapBreaksRowsAndCentresThem(t *testing.T) {
	// widths 40, 40, 40 in 100 with gap 4: rows [40 40] (84) and [40]
	s := sizes{{40, 10}, {40, 20}, {40, 10}}
	w := &Wrap{Gap: 4, Main: Center, Cross: Center, Block: Center}
	// block height 20 + 4 + 10 = 34 in 50 → starts at 8
	check.Equal(t, [][2]int16{{8, 13}, {52, 8}, {30, 32}}, pos(arrange(w, 100, 50, s)))
}

func TestWrapKeepsAnOversizeChildOnItsOwnRow(t *testing.T) {
	s := sizes{{150, 10}, {10, 10}}
	check.Equal(t, [][2]int16{{0, 0}, {0, 10}}, pos(arrange(&Wrap{}, 100, 50, s)))
}

func TestTableEqualColumnsAlignedInCells(t *testing.T) {
	// 2 columns in 100 with gap 4: cells 48 wide
	s := sizes{{20, 10}, {40, 6}, {10, 10}}
	tb := &Table{Cols: 2, GapX: 4, GapY: 2, H: Center, V: End}
	check.Equal(t, [][2]int16{{14, 0}, {56, 4}, {19, 12}}, pos(arrange(tb, 100, 100, s)))
}

func TestAnchorStacksChildren(t *testing.T) {
	s := sizes{{20, 10}, {4, 4}}
	check.Equal(t, [][2]int16{{40, 45}, {48, 48}}, pos(arrange(&Anchor{H: Center, V: Center}, 100, 100, s)))
	check.Equal(t, [][2]int16{{78, 2}, {94, 2}}, pos(arrange(&Anchor{H: End, V: Start, DX: -2, DY: 2}, 100, 100, s)))
}

// Containers pass themselves (a pointer) as Sizes, so arranging per frame
// costs no allocation.
func TestArrangersDoNotAllocate(t *testing.T) {
	s := &sizes{{10, 4}, {20, 8}, {30, 6}}
	out := make([]Rect, s.Len())
	for _, a := range []Arranger{Row(2, Center, Center, 1, 2, 0), &Wrap{Gap: 2}, &Table{Cols: 2}, &Anchor{}} {
		if n := testing.AllocsPerRun(10, func() { a.Arrange(100, 50, s, out) }); n != 0 {
			t.Fatalf("%T allocates %.0f", a, n)
		}
	}
}

// Large weights don't overflow the share (found in review: shared*weight was
// int16; 200:55 in 320 put the second cell at x = -6).
func TestRowLargeWeightsDoNotOverflow(t *testing.T) {
	s := sizes{{0, 1}, {0, 1}}
	r := arrange(Row(0, Start, Start, 200, 55), 320, 1, s)
	check.Equal(t, int16(0), r[0].X)
	check.Equal(t, int16(250), r[1].X) // 320·200/255 = 250; the second cell spans 250..320
}

// Table measures the box Arrange fills (a container of zero size takes it):
// columns as wide as the widest child, rows as tall as their tallest.
func TestTableMeasure(t *testing.T) {
	s := sizes{{10, 4}, {30, 6}, {20, 8}, {5, 5}, {12, 3}}
	tb := &Table{Cols: 2, GapX: 2, GapY: 1}
	w, h := tb.Measure(s)
	check.Equal(t, uint16(2*30+2), w)
	check.Equal(t, uint16(6+1+8+1+3), h)
	out := make([]Rect, len(s))
	tb.Arrange(w, h, s, out)
	for i, r := range out {
		check.LessOrEqual(t, r.X+int16(r.W), int16(w), "child %d inside the measured width", i)
		check.LessOrEqual(t, r.Y+int16(r.H), int16(h), "child %d inside the measured height", i)
	}
}
