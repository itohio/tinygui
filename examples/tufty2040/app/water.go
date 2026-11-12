package app

import (
	"image/color"
	"time"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Water: a falling-liquid cellular automaton on a 64×64 grid, drawn as an
// LED matrix on widget.CellGrid. Each step every water cell tries to move
// along gravity, then diagonally, then sideways; the order of the scan
// follows gravity so a cell moves at most once per step. The CellGrid shows
// a separate shade buffer: the surface is foam, deeper water darker.
const (
	waterN      = 64
	waterPitch  = 3
	waterShades = 7
	// steps per frame: one step moves a cell by one, so a few per frame make
	// the liquid flow at a believable speed (a step is cheap, see stepUS)
	waterStepsPerFrame = 3
)

var waterPalette = []color.RGBA{
	{12, 16, 24, 255}, // empty: unlit LED
	{20, 60, 140, 255}, {25, 80, 165, 255}, {30, 105, 190, 255}, {35, 130, 210, 255},
	{40, 160, 225, 255}, {70, 195, 240, 255},
	{215, 245, 255, 255}, // foam (surface)
}

type water struct {
	liquid
	shade  [waterN * waterN]uint8
	cells  int
	stepUS int64
	grid   *widget.CellGrid
	// full-screen LED views of the same grid: 64×48 at 5 px (320×240) or
	// 64×60 at 4 px (256×240); rows are a window chosen by gravity
	full    [2]*widget.CellGrid
	fullOn  bool
	fullIdx int
	txt     textBuf
	txt2    textBuf
	log     func(string)
}

func newWater() *water {
	w := &water{liquid: newLiquid(waterN, waterN)}
	w.grid = widget.NewCellGrid(waterN, waterN, waterPitch, 1, widget.CellRound, w.shade[:], waterPalette, colBG)
	w.full[0] = widget.NewCellGrid(waterN, 48, 5, 1, widget.CellRound, w.shade[:waterN*48], waterPalette, colBG)
	w.full[1] = widget.NewCellGrid(waterN, 60, 4, 1, widget.CellRound, w.shade[:waterN*60], waterPalette, colBG)
	w.fill()
	return w
}

// step advances the liquid and shades it for the grid.
func (w *water) step() {
	t0 := time.Now()
	w.liquid.step()
	w.shadeCells()
	w.stepUS = time.Since(t0).Microseconds()
}

// fill resets to a basin filled 40 % with water.
func (w *water) fill() {
	clear(w.state[:])
	for y := waterN * 6 / 10; y < waterN; y++ {
		for x := range waterN {
			w.state[y*waterN+x] = newCell()
		}
	}
	w.shadeCells()
}

var waterFullNames = []string{"5 px", "4 px"}

// fullGrid returns the full-screen grid with its rows pointed at the window
// where the water collects. Call from Update, not Draw.
func (w *water) fullGrid() *widget.CellGrid {
	if !w.fullOn {
		return nil
	}
	g := w.full[w.fullIdx]
	g.Cells = w.window(g.Rows)
	return g
}

// window is rows of the 64 where gravity collects the water: the bottom when
// it pulls down, the top when up, the middle when sideways.
func (w *water) window(rows int) []uint8 {
	off := (waterN - rows) / 2
	switch gravityVec[w.dir][1] {
	case 1:
		off = waterN - rows
	case -1:
		off = 0
	}
	return w.shade[off*waterN : (off+rows)*waterN]
}

func (w *water) preview() []uint8      { return w.window(tileRows) }
func (w *water) palette() []color.RGBA { return waterPalette }
func (w *water) showFull()             { w.fullOn = true }

// waterDepthShade maps depth below the surface (along gravity) to a palette
// index: 0 is the surface (foam), deeper is darker. A table, not a division.
var waterDepthShade = [16]uint8{7, 6, 6, 6, 5, 5, 5, 4, 4, 4, 3, 3, 3, 2, 2, 1}

// shadeCells sets the palette index of every cell. It scans from the side
// gravity comes from, so the cell against gravity already holds its depth
// (stored as depth+1 in shade during the scan): one read per cell.
func (w *water) shadeCells() {
	g := gravityVec[w.dir]
	gx, gy := g[0], g[1]
	y0, y1, dy := 0, waterN, 1
	if gy < 0 {
		y0, y1, dy = waterN-1, -1, -1
	}
	x0, x1, dx := 0, waterN, 1
	if gx < 0 {
		x0, x1, dx = waterN-1, -1, -1
	}
	n := 0
	for y := y0; y != y1; y += dy {
		uy := y - gy
		for x := x0; x != x1; x += dx {
			i := y*waterN + x
			if w.state[i]&liquidBit == 0 {
				w.shade[i] = 0
				continue
			}
			n++
			d := uint8(1)
			if ux := x - gx; ux >= 0 && uy >= 0 && ux < waterN && uy < waterN {
				if u := w.shade[uy*waterN+ux]; u > 0 {
					d = min(u+1, 16)
				}
			}
			w.shade[i] = d
		}
	}
	for i, d := range w.shade[:] { // a slice: ranging the array copies it to the heap (F-35)
		if d > 0 {
			w.shade[i] = waterDepthShade[d-1]
		}
	}
	w.cells = n
}

func (a *App) pageWater() *page {
	w := newWater()
	w.log = a.log
	grav := widget.NewInteractiveLabelChoice(52, 11, gravityNames,
		widget.WithLabelChoiceIndex(&w.dir), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG),
		widget.WithLabelChoiceChange(func(i int, s string) { a.log("gravity " + s) }))
	pour := widget.NewToggle(44, 12, fontBody, colFG, "ON", "OFF", colGood, colTrack,
		func() bool { return w.pour }, func(v bool) { w.pour = v })
	reset := widget.NewToggle(44, 12, fontBody, colFG, "FILL", "FILL", colTrack, colTrack,
		func() bool { return false }, func(bool) { w.fill(); a.log("water refilled") })
	full := widget.NewToggle(44, 12, fontBody, colFG, "SHOW", "SHOW", colTrack, colTrack,
		func() bool { return false }, func(bool) { w.fullOn = true; a.log("water full screen") })
	size := widget.NewInteractiveLabelChoice(52, 11, waterFullNames,
		widget.WithLabelChoiceIndex(&w.fullIdx), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	const side = waterN * waterPitch // 192
	// explicit height: findings F-33
	p := a.newPageWith("Water", layout.VFlow(4, side),
		still(w.grid),
		newField("Gravity", 42, 104, grav),
		newField("Pour", 42, 104, pour),
		newField("Reset", 42, 104, reset),
		newField("Full", 42, 104, full),
		newField("LEDs", 42, 104, size),
		still(a.text(104, func() string { return w.txt.reset().str("cells ").int(w.cells).String() }, colMuted)),
		still(a.text(104, func() string { return w.txt2.reset().str("step ").int(int(w.stepUS)).str(" us").String() }, colMuted)),
	)
	p.demo = w
	return p
}

func (w *water) tick() {
	for range waterStepsPerFrame {
		w.step()
	}
}

// fullPress: UP/DOWN rotate gravity, A pours, B back to the page, C leaves
// and changes page (returns false so the App carries on).
func (w *water) fullPress(b Button) bool {
	switch b {
	case BtnUp, BtnDown:
		if b == BtnUp {
			w.turn(1)
		} else {
			w.turn(-1)
		}
		w.log("gravity " + gravityNames[w.dir])
	case BtnA:
		w.pour = !w.pour
	case BtnB:
		w.fullOn = false
	case BtnC:
		w.fullOn = false
		return false
	}
	return true
}

func (w *water) reset() { w.fill() }
