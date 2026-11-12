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
