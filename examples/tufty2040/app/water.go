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
	waterBit    = 1
	waterMoved  = 2
	waterShades = 7
	// steps per frame: one step moves a cell by one, so a few per frame make
	// the liquid flow at a believable speed (a step is cheap, see stepUS)
	waterStepsPerFrame = 3
)

var waterDirs = []string{"down", "dn-left", "left", "up-left", "up", "up-right", "right", "dn-right"}

// gravity vector per direction (screen coordinates, y down)
var waterGrav = [8][2]int{{0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}}

var waterPalette = []color.RGBA{
	{12, 16, 24, 255}, // empty: unlit LED
	{20, 60, 140, 255}, {25, 80, 165, 255}, {30, 105, 190, 255}, {35, 130, 210, 255},
	{40, 160, 225, 255}, {70, 195, 240, 255},
	{215, 245, 255, 255}, // foam (surface)
}

type water struct {
	state  [waterN * waterN]uint8
	shade  [waterN * waterN]uint8
	dir    int
	pour   bool
	count  int
	stepUS int64
	grid   *widget.CellGrid
	// full-screen LED views of the same grid: 64×48 at 5 px (320×240) or
	// 64×60 at 4 px (256×240); rows are a window chosen by gravity
	full    [2]*widget.CellGrid
	fullOn  bool
	fullIdx int
	txt     textBuf
	txt2    textBuf
}

func newWater() *water {
	w := &water{}
	w.grid = widget.NewCellGrid(waterN, waterN, waterPitch, 1, widget.CellRound, w.shade[:], waterPalette, colBG)
	w.full[0] = widget.NewCellGrid(waterN, 48, 5, 1, widget.CellRound, w.shade[:waterN*48], waterPalette, colBG)
	w.full[1] = widget.NewCellGrid(waterN, 60, 4, 1, widget.CellRound, w.shade[:waterN*60], waterPalette, colBG)
	w.fill()
	return w
}

// fill resets to a basin filled 40 % with water.
func (w *water) fill() {
	clear(w.state[:])
	for y := waterN * 6 / 10; y < waterN; y++ {
		for x := range waterN {
			w.state[y*waterN+x] = waterBit
		}
	}
	w.shadeCells()
}

func (w *water) at(x, y int) bool {
	return x >= 0 && y >= 0 && x < waterN && y < waterN && w.state[y*waterN+x]&waterBit == 0
}

func (w *water) move(x, y, nx, ny int) {
	w.state[y*waterN+x] = 0
	w.state[ny*waterN+nx] = waterBit | waterMoved
}

func (w *water) step() {
	t0 := time.Now()
	g := waterGrav[w.dir]
	gx, gy := g[0], g[1]
	px, py := -gy, gx // perpendicular
	if w.pour {
		// a two-cell stream entering opposite to gravity, at the centre
		cx, cy := waterN/2-gx*(waterN/2-1), waterN/2-gy*(waterN/2-1)
		for _, o := range [2]int{0, 1} {
			x, y := cx+px*o, cy+py*o
			if w.at(x, y) {
				w.state[y*waterN+x] = waterBit
			}
		}
	}
	// scan from the side gravity points to, so falling cells move once
	y0, y1, dy := 0, waterN, 1
	if gy > 0 {
		y0, y1, dy = waterN-1, -1, -1
	}
	x0, x1, dx := 0, waterN, 1
	if gx > 0 {
		x0, x1, dx = waterN-1, -1, -1
	}
	flip := int(rand32() & 1)
	for y := y0; y != y1; y += dy {
		for x := x0; x != x1; x += dx {
			s := w.state[y*waterN+x]
			if s != waterBit { // empty, or already moved this step
				continue
			}
			if w.at(x+gx, y+gy) {
				w.move(x, y, x+gx, y+gy)
				continue
			}
			s1, s2 := 1, -1
			if (x+y+flip)&1 == 1 {
				s1, s2 = -1, 1
			}
			switch {
			case w.at(x+gx+px*s1, y+gy+py*s1):
				w.move(x, y, x+gx+px*s1, y+gy+py*s1)
			case w.at(x+gx+px*s2, y+gy+py*s2):
				w.move(x, y, x+gx+px*s2, y+gy+py*s2)
			case w.at(x+px*s1, y+py*s1):
				w.move(x, y, x+px*s1, y+py*s1)
			case w.at(x+px*s2, y+py*s2):
				w.move(x, y, x+px*s2, y+py*s2)
			}
		}
	}
	for i := range w.state {
		w.state[i] &= waterBit
	}
	w.shadeCells()
	w.stepUS = time.Since(t0).Microseconds()
}

var waterFullNames = []string{"5 px", "4 px"}

// fullView returns the full-screen grid with its rows pointed at the part of
// the 64 rows where the water collects: the bottom when gravity pulls down,
// the top when up, the middle when sideways. Call from Update, not Draw.
func (w *water) fullView() *widget.CellGrid {
	g := w.full[w.fullIdx]
	off := (waterN - g.Rows) / 2
	switch waterGrav[w.dir][1] {
	case 1:
		off = waterN - g.Rows
	case -1:
		off = 0
	}
	g.Cells = w.shade[off*waterN : (off+g.Rows)*waterN]
	return g
}

// turn rotates gravity one step (the full-screen view's UP/DOWN).
func (w *water) turn(by int) { w.dir = (w.dir + by) & 7 }

// waterDepthShade maps depth below the surface (along gravity) to a palette
// index: 0 is the surface (foam), deeper is darker. A table, not a division.
var waterDepthShade = [16]uint8{7, 6, 6, 6, 5, 5, 5, 4, 4, 4, 3, 3, 3, 2, 2, 1}

// shadeCells sets the palette index of every cell. It scans from the side
// gravity comes from, so the cell against gravity already holds its depth
// (stored as depth+1 in shade during the scan): one read per cell.
func (w *water) shadeCells() {
	g := waterGrav[w.dir]
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
			if w.state[i]&waterBit == 0 {
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
	for i, d := range w.shade {
		if d > 0 {
			w.shade[i] = waterDepthShade[d-1]
		}
	}
	w.count = n
}

func (a *App) pageWater() *page {
	w := a.water
	grav := widget.NewInteractiveLabelChoice(52, 11, waterDirs,
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
	return a.newPageWith("Water", layout.VFlow(4, side),
		still(w.grid),
		newField("Gravity", 42, 104, grav),
		newField("Pour", 42, 104, pour),
		newField("Reset", 42, 104, reset),
		newField("Full", 42, 104, full),
		newField("LEDs", 42, 104, size),
		still(a.text(104, func() string { return w.txt.reset().str("cells ").int(w.count).String() }, colMuted)),
		still(a.text(104, func() string { return w.txt2.reset().str("step ").int(int(w.stepUS)).str(" us").String() }, colMuted)),
	)
}
