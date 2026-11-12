package app

import (
	"image/color"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Game of Life on a widget.CellGrid drawn as an LED matrix. Cell values are
// ages: 0 dead, 1 newborn … lifeMaxAge old; the palette fades with age.
const (
	lifeCols   = 64
	lifeRows   = 48
	lifePitch  = 3
	lifeMaxAge = 7
)

var lifePatterns = []string{"Random", "Glider gun", "Acorn", "R-pentomino"}

var lifePalette = []color.RGBA{
	{18, 22, 30, 255},    // dead: an unlit LED
	{230, 255, 255, 255}, // newborn
	{120, 240, 255, 255},
	{40, 210, 240, 255},
	{30, 170, 220, 255},
	{40, 130, 200, 255},
	{60, 95, 170, 255},
	{70, 70, 140, 255}, // old
}

type life struct {
	cells, next [lifeCols * lifeRows]uint8
	xm, xp      [lifeCols]int // wrapped column neighbours, no modulo in the step
	running     bool
	pattern     int
	gen         int
	grid        *widget.CellGrid
	full        *widget.CellGrid // 5 px LEDs on the whole screen
	fullOn      bool
	txt         textBuf
	log         func(string)
}

func newLife() *life {
	l := &life{running: true}
	for x := range lifeCols {
		l.xm[x] = (x + lifeCols - 1) % lifeCols
		l.xp[x] = (x + 1) % lifeCols
	}
	l.grid = widget.NewCellGrid(lifeCols, lifeRows, lifePitch, 1, widget.CellRound, l.cells[:], lifePalette, colBG)
	l.full = widget.NewCellGrid(lifeCols, lifeRows, 5, 1, widget.CellRound, l.cells[:], lifePalette, colBG)
	l.seed()
	return l
}

func (l *life) seed() {
	clear(l.cells[:])
	l.gen = 0
	set := func(x, y int) { l.cells[(y%lifeRows)*lifeCols+x%lifeCols] = 1 }
	switch lifePatterns[l.pattern] {
	case "Random":
		for i := range l.cells {
			if rand32()%4 == 0 {
				l.cells[i] = 1
			}
		}
	case "Glider gun": // Gosper glider gun, 36×9
		for _, p := range [][2]int{{24, 0}, {22, 1}, {24, 1}, {12, 2}, {13, 2}, {20, 2}, {21, 2}, {34, 2}, {35, 2},
			{11, 3}, {15, 3}, {20, 3}, {21, 3}, {34, 3}, {35, 3}, {0, 4}, {1, 4}, {10, 4}, {16, 4}, {20, 4}, {21, 4},
			{0, 5}, {1, 5}, {10, 5}, {14, 5}, {16, 5}, {17, 5}, {22, 5}, {24, 5}, {10, 6}, {16, 6}, {24, 6},
			{11, 7}, {15, 7}, {12, 8}, {13, 8}} {
			set(p[0]+2, p[1]+2)
		}
	case "Acorn":
		for _, p := range [][2]int{{1, 0}, {3, 1}, {0, 2}, {1, 2}, {4, 2}, {5, 2}, {6, 2}} {
			set(p[0]+lifeCols/2-3, p[1]+lifeRows/2-1)
		}
	case "R-pentomino":
		for _, p := range [][2]int{{1, 0}, {2, 0}, {0, 1}, {1, 1}, {1, 2}} {
			set(p[0]+lifeCols/2-1, p[1]+lifeRows/2-1)
		}
	}
}

// step advances one generation on a torus.
func (l *life) step() {
	c, n := &l.cells, &l.next
	for y := range lifeRows {
		up := (y + lifeRows - 1) % lifeRows * lifeCols
		mid := y * lifeCols
		dn := (y + 1) % lifeRows * lifeCols
		for x := range lifeCols {
			xm, xp := l.xm[x], l.xp[x]
			cnt := alive(c[up+xm]) + alive(c[up+x]) + alive(c[up+xp]) +
				alive(c[mid+xm]) + alive(c[mid+xp]) +
				alive(c[dn+xm]) + alive(c[dn+x]) + alive(c[dn+xp])
			age := c[mid+x]
			switch {
			case age > 0 && (cnt == 2 || cnt == 3):
				n[mid+x] = min(age+1, lifeMaxAge)
			case age == 0 && cnt == 3:
				n[mid+x] = 1
			default:
				n[mid+x] = 0
			}
		}
	}
	copy(l.cells[:], l.next[:]) // not l.cells = l.next: TinyGo builds that 3 KB copy on the heap (F-35)
	l.gen++
}

func alive(v uint8) uint8 {
	if v > 0 {
		return 1
	}
	return 0
}

var rngState uint32 = 2463534242

func rand32() uint32 {
	rngState ^= rngState << 13
	rngState ^= rngState >> 17
	rngState ^= rngState << 5
	return rngState
}

func (a *App) pageLife() *page {
	l := newLife()
	l.log = a.log
	run := widget.NewToggle(44, 12, fontBody, colFG, "RUN", "PAUSE", colGood, colTrack,
		func() bool { return l.running }, func(v bool) { l.running = v })
	pat := widget.NewInteractiveLabelChoice(70, 11, lifePatterns,
		widget.WithLabelChoiceIndex(&l.pattern), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG),
		widget.WithLabelChoiceChange(func(int, string) { l.seed(); a.log("life " + lifePatterns[l.pattern]) }))
	full := widget.NewToggle(44, 12, fontBody, colFG, "SHOW", "SHOW", colTrack, colTrack,
		func() bool { return false }, func(bool) { l.fullOn = true; a.log("life full screen") })
	gen := a.text(104, func() string { return l.txt.reset().str("gen ").int(l.gen).String() }, colMuted)
	// explicit height: findings F-33
	p := a.newPageWith("Life", layout.VFlow(4, lifeRows*lifePitch),
		still(l.grid),
		newField("Run", 34, 104, run),
		newField("", 0, 104, pat),
		newField("Full", 34, 104, full),
		still(gen),
	)
	p.demo = l
	return p
}

func (l *life) tick() {
	if l.running {
		l.step()
	}
}

func (l *life) fullGrid() *widget.CellGrid {
	if l.fullOn {
		return l.full
	}
	return nil
}

// fullPress: UP/DOWN next / previous pattern (reseeds), A run / pause, B
// back to the page, C leaves and changes page.
func (l *life) fullPress(b Button) bool {
	switch b {
	case BtnUp, BtnDown:
		d := 1
		if b == BtnDown {
			d = len(lifePatterns) - 1
		}
		l.pattern = (l.pattern + d) % len(lifePatterns)
		l.seed()
		l.log("life " + lifePatterns[l.pattern])
	case BtnA:
		l.running = !l.running
	case BtnB:
		l.fullOn = false
	case BtnC:
		l.fullOn = false
		return false
	}
	return true
}

func (l *life) preview() []uint8      { return l.cells[:] }
func (l *life) palette() []color.RGBA { return lifePalette }
func (l *life) showFull()             { l.fullOn = true }

func (l *life) reset() { l.seed() }
