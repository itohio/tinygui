package app

import (
	"image/color"
	"time"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Fire: the "Doom fire" (heat spreads up from a burning bottom row, cooling
// and drifting at random) on a 64×48 grid, drawn as an LED matrix: 3 px in
// the page, 5 px round LEDs on the whole screen.
const (
	fireW, fireH = 64, 48
	fireMax      = 36
	firePitch    = 3
)

var (
	fireWinds   = []string{"left", "calm", "right"}
	fireHeights = []string{"low", "mid", "tall"}
	// horizontal drift for two random bits, per wind
	fireDrift = [3][4]int8{{-2, -1, -1, 0}, {-1, 0, 0, 1}, {0, 1, 1, 2}}
	// cooling per row for two random bits, per height; the mean sets how
	// high flames reach (fireMax / mean rows): 2.25 → 16, 1.25 → 29,
	// 0.75 → 48, the spread makes the tongues
	fireCool = [3][4]uint8{{1, 2, 3, 3}, {0, 1, 2, 2}, {0, 0, 1, 2}}
)

// firePalette is the classic Doom fire palette; index 0 is an unlit LED.
var firePalette = func() []color.RGBA {
	hex := [fireMax + 1]uint32{
		0x120e0e, 0x1f0707, 0x2f0f07, 0x470f07, 0x571707, 0x671f07, 0x771f07, 0x8f2707, 0x9f2f07,
		0xaf3f07, 0xbf4707, 0xc74707, 0xdf4f07, 0xdf5707, 0xdf5707, 0xd75f07, 0xd75f07, 0xd7670f,
		0xcf6f0f, 0xcf770f, 0xcf7f0f, 0xcf8717, 0xc78717, 0xc78f17, 0xc7971f, 0xbf9f1f, 0xbf9f1f,
		0xbfa727, 0xbfa727, 0xbfaf2f, 0xb7af2f, 0xb7b72f, 0xb7b737, 0xcfcf6f, 0xdfdf9f, 0xefefc7,
		0xffffff,
	}
	p := make([]color.RGBA, len(hex))
	for i, h := range hex {
		p[i] = color.RGBA{uint8(h >> 16), uint8(h >> 8), uint8(h), 255}
	}
	return p
}()

type fire struct {
	cells   [fireW * fireH]uint8
	burning bool
	wind    int
	height  int
	stepUS  int64
	grid    *widget.CellGrid
	full    *widget.CellGrid
	fullOn  bool
	txt     textBuf
	log     func(string)
}

func newFire() *fire {
	f := &fire{burning: true, wind: 1, height: 1}
	f.grid = widget.NewCellGrid(fireW, fireH, firePitch, 1, widget.CellRound, f.cells[:], firePalette, colBG)
	f.full = widget.NewCellGrid(fireW, fireH, 5, 1, widget.CellRound, f.cells[:], firePalette, colBG)
	return f
}

func (f *fire) step() {
	t0 := time.Now()
	src := uint8(0)
	if f.burning {
		src = fireMax
	}
	last := (fireH - 1) * fireW
	for x := range fireW {
		f.cells[last+x] = src
	}
	drift, cool := &fireDrift[f.wind], &fireCool[f.height]
	for y := 1; y < fireH; y++ {
		row, up := y*fireW, (y-1)*fireW
		for x := range fireW {
			v := f.cells[row+x]
			r := rand32()
			nx := x + int(drift[r>>1&3])
			if nx < 0 {
				nx = 0
			} else if nx >= fireW {
				nx = fireW - 1
			}
			if v > 0 {
				cool := cool[r>>3&3]
				if v > cool {
					v -= cool
				} else {
					v = 0
				}
			}
			f.cells[up+nx] = v
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

// fullPress: UP/DOWN change the wind, A lights or puts out the fire, B back
// to the page, C leaves and changes page.
func (f *fire) fullPress(b Button) bool {
	switch b {
	case BtnUp:
		f.wind = min(f.wind+1, len(fireWinds)-1)
		f.log("wind " + fireWinds[f.wind])
	case BtnDown:
		f.wind = max(f.wind-1, 0)
		f.log("wind " + fireWinds[f.wind])
	case BtnA:
		f.burning = !f.burning
	case BtnB:
		f.fullOn = false
	case BtnC:
		f.fullOn = false
		return false
	}
	return true
}

func (a *App) pageFire() *page {
	f := newFire()
	f.log = a.log
	burn := widget.NewToggle(44, 12, fontBody, colFG, "ON", "OFF", colGood, colTrack,
		func() bool { return f.burning }, func(v bool) { f.burning = v })
	wind := widget.NewInteractiveLabelChoice(52, 11, fireWinds,
		widget.WithLabelChoiceIndex(&f.wind), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	height := widget.NewInteractiveLabelChoice(52, 11, fireHeights,
		widget.WithLabelChoiceIndex(&f.height), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	full := widget.NewToggle(44, 12, fontBody, colFG, "SHOW", "SHOW", colTrack, colTrack,
		func() bool { return false }, func(bool) { f.fullOn = true; a.log("fire full screen") })
	// explicit height: findings F-33
	p := a.newPageWith("Fire", layout.VFlow(4, fireH*firePitch),
		still(f.grid),
		newField("Fire", 42, 104, burn),
		newField("Wind", 42, 104, wind),
		newField("Height", 42, 104, height),
		newField("Full", 42, 104, full),
		still(a.text(104, func() string { return f.txt.reset().str("step ").int(int(f.stepUS)).str(" us").String() }, colMuted)),
	)
	p.demo = f
	return p
}

func (f *fire) preview() []uint8      { return f.cells[:] }
func (f *fire) palette() []color.RGBA { return firePalette }
func (f *fire) showFull()             { f.fullOn = true }
