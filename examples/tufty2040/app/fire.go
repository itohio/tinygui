package app

import (
	"image/color"
	"time"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Fire burns a liquid fuel (owner: "a little bit of liquid that piles up in
// the corner or spreads over the edge ... the liquid never consumed though,
// but it sloshes and seeds the fire"). The fuel is the water's liquid
// engine: it pools, spreads and follows gravity, and is never used up.
// Every fuel cell open to the air (the cell against gravity is empty) seeds
// flames; flames rise against gravity as heat spreading from the surface,
// cooling and flickering (the "Doom fire" spread, turned with gravity).
// Drawn on a 64×48 LED grid: the flame palette over a dim fuel colour.
const (
	fireW, fireH = 64, 48
	fireMax      = 36 // hottest flame (palette index)
	fireFuel     = fireMax + 1
	firePitch    = 3
	firePool     = 120 // fuel cells in a fresh pool
	fireSplash   = 40  // fuel cells a splash adds
	fireMaxFuel  = 400 // splashes stop adding fuel above this
)

var (
	fireHeights = []string{"low", "mid", "tall"}
	// sideways flicker for two random bits
	fireDrift = [4]int8{-1, 0, 0, 1}
	// cooling per cell travelled, for two random bits, per height; the mean
	// sets how far flames reach (fireMax / mean): 2.25 → 16, 1.25 → 29,
	// 0.75 → 48 cells, the spread makes the tongues
	fireCool = [3][4]uint8{{1, 2, 3, 3}, {0, 1, 2, 2}, {0, 0, 1, 2}}
)

// firePalette is the classic Doom fire palette (index 0 an unlit LED) plus
// the fuel's dim amber.
var firePalette = func() []color.RGBA {
	hex := [fireMax + 2]uint32{
		0x120e0e, 0x1f0707, 0x2f0f07, 0x470f07, 0x571707, 0x671f07, 0x771f07, 0x8f2707, 0x9f2f07,
		0xaf3f07, 0xbf4707, 0xc74707, 0xdf4f07, 0xdf5707, 0xdf5707, 0xd75f07, 0xd75f07, 0xd7670f,
		0xcf6f0f, 0xcf770f, 0xcf7f0f, 0xcf8717, 0xc78717, 0xc78f17, 0xc7971f, 0xbf9f1f, 0xbf9f1f,
		0xbfa727, 0xbfa727, 0xbfaf2f, 0xb7af2f, 0xb7b72f, 0xb7b737, 0xcfcf6f, 0xdfdf9f, 0xefefc7,
		0xffffff,
		0x7a4c14, // fuel: dim amber, visible next to the flames
	}
	p := make([]color.RGBA, len(hex))
	for i, h := range hex {
		p[i] = color.RGBA{uint8(h >> 16), uint8(h >> 8), uint8(h), 255}
	}
	return p
}()

type fire struct {
	liquid                          // the fuel
	heat, next [fireW * fireH]uint8 // flame heat 0..fireMax, and the next step's
	cells      [fireW * fireH]uint8 // what the grid shows
	height     int
	stepUS     int64
	grid, full *widget.CellGrid
	fullOn     bool
	txt, txt2  textBuf
	log        func(string)
}

func newFire() *fire {
	f := &fire{liquid: newLiquid(fireW, fireH), height: 1}
	f.grid = widget.NewCellGrid(fireW, fireH, firePitch, 1, widget.CellRound, f.cells[:], firePalette, colBG)
	f.full = widget.NewCellGrid(fireW, fireH, 5, 1, widget.CellRound, f.cells[:], firePalette, colBG)
	f.reset()
	return f
}

// reset pours a fresh pool of fuel into the bottom-left corner.
func (f *fire) reset() {
	clear(f.state)
	clear(f.heat[:])
	f.dir = 0
	n := 0
	for y := fireH - 1; n < firePool; y-- {
		for x := 0; x < 24 && n < firePool; x++ {
			f.state[y*fireW+x] = newCell()
			n++
		}
	}
}

// splash drops a blob of fuel in from the side opposite gravity (up to
// fireMaxFuel in all).
func (f *fire) splash() {
	if f.count() > fireMaxFuel-fireSplash {
		return
	}
	g := gravityVec[f.dir]
	cx, cy := fireW/2-g[0]*(fireW/2-4), fireH/2-g[1]*(fireH/2-4)
	n := 0
	for dy := -3; dy <= 3 && n < fireSplash; dy++ {
		for dx := -4; dx <= 4 && n < fireSplash; dx++ {
			x, y := cx+dx, cy+dy
			if f.at(x, y) {
				f.state[y*fireW+x] = newCell()
				n++
			}
		}
	}
}

// exposed: the cell against gravity is open air (a wall is not air).
func (f *fire) exposed(x, y int) bool {
	g := gravityVec[f.dir]
	ux, uy := x-g[0], y-g[1]
	return ux >= 0 && uy >= 0 && ux < fireW && uy < fireH && f.state[uy*fireW+ux]&liquidBit == 0
}

func (f *fire) step() {
	t0 := time.Now()
	f.liquid.step()
	g := gravityVec[f.dir]
	gx, gy := g[0], g[1]
	px, py := -gy, gx

	// flames: each cell takes the heat of the cell below it (towards
	// gravity), flickered sideways, minus cooling; the fuel's surface is the
	// source
	cool := &fireCool[f.height]
	for y := range fireH {
		for x := range fireW {
			r := rand32()
			d := int(fireDrift[r&3])
			sx, sy := x+gx+px*d, y+gy+py*d
			var h uint8
			if sx >= 0 && sy >= 0 && sx < fireW && sy < fireH {
				h = f.heat[sy*fireW+sx]
			}
			if c := cool[r>>2&3]; h > c {
				h -= c
			} else {
				h = 0
			}
			i := y*fireW + x
			s := f.state[i]
			if s&liquidBit != 0 {
				h = 0 // no flames inside the pool
			}
			f.next[i] = h
		}
	}
	// the fuel's surface lights the air above it, so the fuel stays visible
	// under its flames
	for i, s := range f.state {
		if s&liquidBit == 0 {
			continue
		}
		x, y := i%fireW-gx, i/fireW-gy
		if x >= 0 && y >= 0 && x < fireW && y < fireH && f.state[y*fireW+x]&liquidBit == 0 {
			// a flickering source: full heat 3 times in 4, else weaker,
			// so the flames break into tongues
			h := uint8(fireMax)
			if rand32()&3 == 0 {
				h = fireMax - 10
			}
			f.next[y*fireW+x] = h
		}
	}
	copy(f.heat[:], f.next[:]) // not f.heat = f.next: TinyGo heap-allocates that copy (F-35)

	for i, h := range f.heat[:] {
		switch {
		case h > 0:
			f.cells[i] = h
		case f.state[i]&liquidBit != 0:
			f.cells[i] = fireFuel
		default:
			f.cells[i] = 0
		}
	}
	f.stepUS = time.Since(t0).Microseconds()
}

func (f *fire) tick() { f.step() }

func (f *fire) fullGrid() *widget.CellGrid {
	if f.fullOn {
		return f.full
	}
	return nil
}

// fullPress: UP/DOWN turn gravity (the fuel flows, the flames lean), A
// splashes fuel in, B back to the page, C leaves and changes page.
func (f *fire) fullPress(b Button) bool {
	switch b {
	case BtnUp:
		f.turn(1)
		f.log("gravity " + gravityNames[f.dir])
	case BtnDown:
		f.turn(-1)
		f.log("gravity " + gravityNames[f.dir])
	case BtnA:
		f.splash()
	case BtnB:
		f.fullOn = false
	case BtnC:
		f.fullOn = false
		return false
	}
	return true
}

func (f *fire) preview() []uint8      { return f.cells[:] }
func (f *fire) palette() []color.RGBA { return firePalette }
func (f *fire) showFull()             { f.fullOn = true }

func (a *App) pageFire() *page {
	f := newFire()
	f.log = a.log
	grav := widget.NewInteractiveLabelChoice(52, 11, gravityNames,
		widget.WithLabelChoiceIndex(&f.dir), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG),
		widget.WithLabelChoiceChange(func(_ int, s string) { a.log("gravity " + s) }))
	pour := widget.NewToggle(44, 12, fontBody, colFG, "ON", "OFF", colGood, colTrack,
		func() bool { return f.pour }, func(v bool) { f.pour = v })
	splash := widget.NewToggle(44, 12, fontBody, colFG, "DROP", "DROP", colTrack, colTrack,
		func() bool { return false }, func(bool) { f.splash(); a.log("fire splash") })
	height := widget.NewInteractiveLabelChoice(52, 11, fireHeights,
		widget.WithLabelChoiceIndex(&f.height), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	full := widget.NewToggle(44, 12, fontBody, colFG, "SHOW", "SHOW", colTrack, colTrack,
		func() bool { return false }, func(bool) { f.fullOn = true; a.log("fire full screen") })
	// explicit height: findings F-33
	p := a.newPageWith("Fire", layout.VFlow(4, fireH*firePitch),
		still(f.grid),
		newField("Gravity", 42, 104, grav),
		newField("Fuel", 42, 104, pour),
		newField("Splash", 42, 104, splash),
		newField("Flames", 42, 104, height),
		newField("Full", 42, 104, full),
		still(a.text(104, func() string {
			return f.txt.reset().str("fuel ").int(f.count()).String()
		}, colMuted)),
		still(a.text(104, func() string { return f.txt2.reset().str("step ").int(int(f.stepUS)).str(" us").String() }, colMuted)),
	)
	p.demo = f
	return p
}
