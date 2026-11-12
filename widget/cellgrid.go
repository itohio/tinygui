package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// CellShape is how a CellGrid draws one cell.
type CellShape uint8

const (
	// CellSquare fills the cell's dot as a square.
	CellSquare CellShape = iota
	// CellRound cuts the dot's corners so it reads as a round LED.
	CellRound
)

// CellGrid draws a Cols×Rows grid of palette-indexed cells: an LED matrix,
// a Game of Life board, a heat map. The cells are a caller-owned []uint8
// (one palette index per cell, row-major) that the caller updates in place;
// the grid reads it on every Draw and allocates nothing.
//
// Each cell occupies Pitch×Pitch pixels: a dot of Pitch-Gap pixels and a
// Gap-pixel border in the Off colour.
type CellGrid struct {
	ui.WidgetBase
	Cols, Rows int
	Cells      []uint8
	Palette    []color.RGBA
	Pitch, Gap uint8
	Shape      CellShape
	Off        color.RGBA // colour of the gaps (and of out-of-palette cells)
}

// NewCellGrid creates a grid of cols×rows cells drawn pitch pixels apart.
// cells must hold cols*rows indices into palette.
func NewCellGrid(cols, rows int, pitch, gap uint8, shape CellShape, cells []uint8, palette []color.RGBA, off color.RGBA) *CellGrid {
	return &CellGrid{
		WidgetBase: ui.NewWidgetBase(uint16(cols)*uint16(pitch), uint16(rows)*uint16(pitch)),
		Cols:       cols,
		Rows:       rows,
		Cells:      cells,
		Palette:    palette,
		Pitch:      pitch,
		Gap:        gap,
		Shape:      shape,
		Off:        off,
	}
}

func (g *CellGrid) Draw(ctx ui.Context) {
	d := ctx.D()
	x0, y0 := ctx.DisplayPos()
	if len(g.Cells) < g.Cols*g.Rows {
		return
	}
	p := int16(g.Pitch)
	dot := p - int16(g.Gap)
	if dot <= 0 {
		return
	}
	if g.Gap > 0 {
		fillRect(d, x0, y0, int16(g.Width), int16(g.Height), g.Off)
	}
	round := g.Shape == CellRound && dot >= 3
	// rows to draw: all, or only those overlapping the display's clip
	r0, r1 := 0, g.Rows
	if c, ok := d.(ui.Clipper); ok {
		_, cy, _, ch := c.Clip()
		r0 = max(0, int((cy-y0)/p))
		r1 = min(g.Rows, int((cy+ch-y0+p-1)/p))
		if cy+ch <= y0 {
			r1 = 0
		}
	}
	i := r0 * g.Cols
	y := y0 + int16(r0)*p
	for range max(0, r1-r0) {
		x := x0
		for range g.Cols {
			c := g.Off
			if idx := int(g.Cells[i]); idx < len(g.Palette) {
				c = g.Palette[idx]
			}
			if round {
				// a dot with its four corner pixels left in the Off colour
				fillRect(d, x+1, y, dot-2, dot, c)
				fillRect(d, x, y+1, 1, dot-2, c)
				fillRect(d, x+dot-1, y+1, 1, dot-2, c)
			} else {
				fillRect(d, x, y, dot, dot, c)
			}
			x += p
			i++
		}
		y += p
	}
}
