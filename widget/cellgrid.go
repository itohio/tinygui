package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
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

	lines []uint16 // RGB565 scanlines of the fast path, reused
	pal   []uint16 // Palette in RGB565, rebuilt each Draw (the caller may edit it)
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
	if b, ok := d.(ui.BitmapDisplayer); ok {
		g.drawStrips(b, d, x0, y0, r0, r1, dot, round)
		return
	}
	if g.Gap > 0 || round { // gaps and round cells' corners show the Off colour
		fillRect(d, x0, y0, int16(g.Width), int16(g.Height), g.Off)
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

// drawStrips is the fast path for displays that take RGB565 bitmaps: for
// each row of cells it builds two scanlines (the dot's top/bottom edge and
// its middle) once and copies them to every pixel row of the cell row,
// instead of filling up to three small rectangles per cell.
func (g *CellGrid) drawStrips(b ui.BitmapDisplayer, d drivers.Displayer, x0, y0 int16, r0, r1 int, dot int16, round bool) {
	p := int16(g.Pitch)
	w := g.Cols * int(p)
	if cap(g.lines) < 2*w {
		g.lines = make([]uint16, 2*w)
	}
	edge, mid := g.lines[:w], g.lines[w:2*w]
	off := rgb565(g.Off)
	n := min(len(g.Palette), 256)
	if cap(g.pal) < n {
		g.pal = make([]uint16, n)
	}
	pal := g.pal[:n]
	for k := range n {
		pal[k] = rgb565(g.Palette[k])
	}
	i := r0 * g.Cols
	y := y0 + int16(r0)*p
	for range max(0, r1-r0) {
		x := 0
		for range g.Cols {
			c := off
			if idx := int(g.Cells[i]); idx < n {
				c = pal[idx]
			}
			for k := range p {
				v := off
				if k < dot {
					v = c
				}
				mid[x+int(k)] = v
				if round && (k == 0 || k == dot-1) {
					v = off
				}
				edge[x+int(k)] = v
			}
			x += int(p)
			i++
		}
		for k := range dot {
			line := mid
			if round && (k == 0 || k == dot-1) {
				line = edge
			}
			_ = b.DrawRGBBitmap(x0, y+k, line, int16(w), 1)
		}
		if dot < p {
			fillRect(d, x0, y+dot, int16(w), p-dot, g.Off)
		}
		y += p
	}
}

// rgb565 packs a colour the way DrawRGBBitmap takes it (5-6-5, red high).
func rgb565(c color.RGBA) uint16 {
	return uint16(c.R&0xF8)<<8 | uint16(c.G&0xFC)<<3 | uint16(c.B)>>3
}
