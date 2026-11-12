package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

// canvas is a tiny RGBA framebuffer with the rectangle fast path.
type canvas struct {
	w, h int16
	px   []color.RGBA
}

func newCanvas(w, h int16) *canvas { return &canvas{w: w, h: h, px: make([]color.RGBA, int(w)*int(h))} }

func (c *canvas) Size() (int16, int16) { return c.w, c.h }
func (c *canvas) Display() error       { return nil }
func (c *canvas) SetPixel(x, y int16, col color.RGBA) {
	if x >= 0 && y >= 0 && x < c.w && y < c.h {
		c.px[int(y)*int(c.w)+int(x)] = col
	}
}
func (c *canvas) FillRectangle(x, y, w, h int16, col color.RGBA) error {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.SetPixel(i, j, col)
		}
	}
	return nil
}
func (c *canvas) FillRectangleWithBuffer(int16, int16, int16, int16, []color.RGBA) error { return nil }
func (c *canvas) FillScreen(col color.RGBA)                                              {}
func (c *canvas) at(x, y int16) color.RGBA                                               { return c.px[int(y)*int(c.w)+int(x)] }

var (
	off  = color.RGBA{1, 1, 1, 255}
	dim  = color.RGBA{30, 0, 0, 255}
	lit  = color.RGBA{250, 40, 40, 255}
	back = color.RGBA{9, 9, 9, 255}
)

// Round 2×2 grid, pitch 5, gap 1: each cell is a 4×4 dot without its corner
// pixels, gaps in the Off colour, cells coloured by palette index.
func TestCellGridRoundDots(t *testing.T) {
	cells := []uint8{0, 1, 1, 7} // 7 is outside the palette → Off
	g := NewCellGrid(2, 2, 5, 1, CellRound, cells, []color.RGBA{dim, lit}, off)
	w, h := g.Size()
	require.Equal(t, [2]uint16{10, 10}, [2]uint16{w, h})

	c := newCanvas(14, 14)
	for i := range c.px {
		c.px[i] = back
	}
	ctx := ui.NewContext(c, 14, 14, 2, 2)
	g.Draw(&ctx)

	// cell (0,0) at (2,2): dim dot, corners and gap Off
	require.Equal(t, dim, c.at(3, 3))
	require.Equal(t, dim, c.at(2, 3)) // left edge, not a corner
	require.Equal(t, off, c.at(2, 2)) // corner
	require.Equal(t, off, c.at(5, 5)) // corner
	require.Equal(t, off, c.at(6, 3)) // gap column
	// cell (1,0) at (7,2): lit
	require.Equal(t, lit, c.at(8, 3))
	// cell (1,1) at (7,7): index out of palette → Off
	require.Equal(t, off, c.at(8, 8))
	// outside the grid untouched
	require.Equal(t, back, c.at(12, 12))
	require.Equal(t, back, c.at(1, 1))

	// cells are read on every draw
	cells[0] = 1
	g.Draw(&ctx)
	require.Equal(t, lit, c.at(3, 3))

	if n := testing.AllocsPerRun(20, func() { g.Draw(&ctx) }); n != 0 {
		t.Fatalf("Draw allocates %.0f times", n)
	}
}

func TestCellGridSquareNoGap(t *testing.T) {
	g := NewCellGrid(2, 1, 3, 0, CellSquare, []uint8{1, 0}, []color.RGBA{dim, lit}, off)
	c := newCanvas(6, 3)
	ctx := ui.NewContext(c, 6, 3, 0, 0)
	g.Draw(&ctx)
	require.Equal(t, lit, c.at(0, 0)) // square: corners filled
	require.Equal(t, lit, c.at(2, 2))
	require.Equal(t, dim, c.at(3, 0)) // second cell starts at x = 3, no gap
}

// clippedCanvas reports a clip rectangle and counts fill calls.
type clippedCanvas struct {
	*canvas
	cy, ch int16
	fills  int
}

func (c *clippedCanvas) Clip() (int16, int16, int16, int16) { return 0, c.cy, c.w, c.ch }
func (c *clippedCanvas) FillRectangle(x, y, w, h int16, col color.RGBA) error {
	c.fills++
	return c.canvas.FillRectangle(x, y, w, h, col)
}

// With a clipping display (e.g. one band of a framebuffer) the grid only
// draws cell rows that overlap the clip rectangle, and those correctly.
func TestCellGridSkipsRowsOutsideClip(t *testing.T) {
	cells := make([]uint8, 4*10)
	for i := range cells {
		cells[i] = 1
	}
	g := NewCellGrid(4, 10, 5, 0, CellSquare, cells, []color.RGBA{dim, lit}, off)
	c := &clippedCanvas{canvas: newCanvas(20, 50), cy: 12, ch: 10} // rows 12..21 → cell rows 2..4
	ctx := ui.NewContext(c, 20, 50, 0, 0)
	g.Draw(&ctx)
	require.Equal(t, 3*4, c.fills, "3 cell rows × 4 cells")
	require.Equal(t, lit, c.at(0, 12))
	require.Equal(t, lit, c.at(19, 21))
	require.Equal(t, color.RGBA{}, c.at(0, 30), "rows outside the clip are not drawn")
}

// canvas565 stores RGB565 and takes bitmaps, so CellGrid uses its strip path.
type canvas565 struct {
	*canvas
	bits []uint16
}

func (c *canvas565) SetPixel(x, y int16, col color.RGBA) {
	if x >= 0 && y >= 0 && x < c.w && y < c.h {
		c.bits[int(y)*int(c.w)+int(x)] = rgb565(col)
	}
}
func (c *canvas565) FillRectangle(x, y, w, h int16, col color.RGBA) error {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.SetPixel(i, j, col)
		}
	}
	return nil
}
func (c *canvas565) DrawRGBBitmap(x, y int16, data []uint16, w, h int16) error {
	for j := range h {
		for i := range w {
			if x+i >= 0 && y+j >= 0 && x+i < c.w && y+j < c.h {
				c.bits[int(y+j)*int(c.w)+int(x+i)] = data[int(j)*int(w)+int(i)]
			}
		}
	}
	return nil
}
func (c *canvas565) DrawRGBBitmap8(int16, int16, []uint8, int16, int16) error { return nil }

// The RGB565 strip path draws exactly the pixels of the rectangle path, for
// round and square cells, with and without gaps (round with gap 0: the
// corners are Off on both paths, verify B-F1).
func TestCellGridStripsMatchRectangles(t *testing.T) {
	pal := []color.RGBA{dim, lit, {10, 200, 90, 255}}
	cells := []uint8{0, 1, 2, 9, 2, 1, 1, 0, 2, 0, 0, 1}
	for _, tc := range []struct {
		pitch, gap uint8
		shape      CellShape
	}{{5, 1, CellRound}, {4, 1, CellRound}, {3, 1, CellRound}, {4, 0, CellRound}, {5, 2, CellSquare}, {3, 0, CellSquare}} {
		g := NewCellGrid(4, 3, tc.pitch, tc.gap, tc.shape, cells, pal, off)
		ref := newCanvas(30, 20)
		fast := &canvas565{canvas: newCanvas(30, 20), bits: make([]uint16, 30*20)}
		for i := range ref.px {
			ref.px[i] = back
			fast.bits[i] = rgb565(back)
		}
		rctx := ui.NewContext(ref, 30, 20, 2, 3)
		fctx := ui.NewContext(fast, 30, 20, 2, 3)
		g.Draw(&rctx)
		g.Draw(&fctx)
		for i := range ref.px {
			require.Equalf(t, rgb565(ref.px[i]), fast.bits[i], "pitch %d gap %d shape %d pixel (%d,%d)",
				tc.pitch, tc.gap, tc.shape, i%30, i/30)
		}
		if n := testing.AllocsPerRun(10, func() { g.Draw(&fctx) }); n != 0 {
			t.Fatalf("strip Draw allocates %.0f times", n)
		}
	}
}
