package app

import (
	"image/color"
	"time"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Boids (Reynolds: separation, alignment, cohesion) on a 64×48 LED matrix.
// Positions and velocities are fixed point, 1/256 of a cell. Each boid lights
// its cell in a colour for its heading; the cells it leaves fade out, so the
// flock draws trails. Neighbours are found by a plain scan with an early
// reject on |dx| and |dy|: at most 128 boids, so a spatial grid isn't needed
// (the step time is shown on the page).
const (
	boidsW, boidsH = 64, 48
	boidsFix       = 256
	boidsMax       = 128
	boidsPitch     = 3
	boidsR         = 6 * boidsFix // neighbour radius
	boidsMargin    = 4 * boidsFix // walls push back inside this margin
	boidsVMax      = 128
	boidsVMin      = 56
	boidsTrail     = 3 // trail palette indices 1..3, heads from 4
)

var (
	boidsCounts = []string{"32", "64", "128"}
	boidsN      = [3]int{32, 64, 128}
)

var boidsPalette = []color.RGBA{
	{14, 14, 22, 255},
	{22, 34, 46, 255}, {32, 60, 78, 255}, {50, 96, 120, 255}, // trail, fading
	// heads, one hue per heading octant
	{255, 90, 90, 255}, {255, 170, 60, 255}, {240, 240, 80, 255}, {120, 240, 90, 255},
	{70, 230, 200, 255}, {80, 160, 255, 255}, {160, 110, 255, 255}, {250, 100, 220, 255},
}

type boids struct {
	x, y, vx, vy [boidsMax]int32
	nvx, nvy     [boidsMax]int32
	cells        [boidsW * boidsH]uint8
	count        int // index into boidsN
	pose         boxPose
	stepUS       int64
	grid, full   *widget.CellGrid
	fullOn       bool
	txt          textBuf
	log          func(string)
}

func newBoids() *boids {
	b := &boids{count: 1}
	b.reset()
	b.grid = widget.NewCellGrid(boidsW, boidsH, boidsPitch, 1, widget.CellRound, b.cells[:], boidsPalette, colBG)
	b.full = widget.NewCellGrid(boidsW, boidsH, 5, 1, widget.CellRound, b.cells[:], boidsPalette, colBG)
	return b
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// limit keeps the speed between boidsVMin and boidsVMax, measured with the
// octagonal approximation max + 3/8 min (no square root).
func limit(vx, vy int32) (int32, int32) {
	ax, ay := abs32(vx), abs32(vy)
	mag := max(ax, ay) + min(ax, ay)>>2 + min(ax, ay)>>3
	switch {
	case mag > boidsVMax:
		return vx * boidsVMax / mag, vy * boidsVMax / mag
	case mag < boidsVMin && mag > 0:
		return vx * boidsVMin / mag, vy * boidsVMin / mag
	case mag == 0:
		return boidsVMin, 0
	}
	return vx, vy
}

func (b *boids) step() {
	n := boidsN[b.count]
	const r2 = (boidsR >> 4) * (boidsR >> 4)               // in (1/16 cell)²
	const sep2 = (2 * boidsFix >> 4) * (2 * boidsFix >> 4) // closer than 2 cells
	for i := range n {
		xi, yi := b.x[i], b.y[i]
		var cnt, sx, sy, svx, svy, px, py int32
		for j := range n {
			if j == i {
				continue
			}
			dx, dy := b.x[j]-xi, b.y[j]-yi
			if dx > boidsR || dx < -boidsR || dy > boidsR || dy < -boidsR {
				continue
			}
			// /16, not >>4: the shift rounds negative offsets away from zero,
			// so neighbours up and left counted as closer and separation
			// pushed the whole flock into the top-left corner
			d2 := (dx/16)*(dx/16) + (dy/16)*(dy/16)
			if d2 > r2 {
				continue
			}
			cnt++
			sx += dx
			sy += dy
			svx += b.vx[j]
			svy += b.vy[j]
			if d2 < sep2 {
				px -= dx
				py -= dy
			}
		}
		vx, vy := b.vx[i], b.vy[i]
		// steering divides rather than shifts: >> on a negative value rounds
		// towards -inf and drifted the whole flock up and left
		if cnt > 0 {
			vx += (svx/cnt - vx) / 8 // alignment
			vy += (svy/cnt - vy) / 8
			vx += sx / cnt / 64 // cohesion: towards the neighbours' centre
			vy += sy / cnt / 64
		}
		vx += px / 8 // separation
		vy += py / 8
		b.nvx[i], b.nvy[i] = limit(vx, vy)
	}
	// move in the world, then meet the box's walls: seen from the box, a
	// boid near a wall is pushed inwards, one past it (the wall moved, or it
	// flew) is put back and bounces (the minimum speed would otherwise undo
	// the push for a boid flying straight at a wall). Only boids a wall
	// touches are converted back, so the others keep their world heading.
	for i := range n {
		b.vx[i], b.vy[i] = b.nvx[i], b.nvy[i]
		b.x[i] += b.vx[i]
		b.y[i] += b.vy[i]
		b.walls(i)
	}
}

// walls applies the box's walls to boid i.
func (b *boids) walls(i int) {
	const push = 16
	const hiX, hiY = boidsW*boidsFix - 1, boidsH*boidsFix - 1
	p := &b.pose
	bx, by := p.toBox(b.x[i], b.y[i])
	near := bx < boidsMargin || bx > hiX-boidsMargin || by < boidsMargin || by > hiY-boidsMargin
	if !near {
		return
	}
	vx, vy := p.vecToBox(b.vx[i], b.vy[i])
	switch {
	case bx < boidsMargin:
		vx += push
	case bx > hiX-boidsMargin:
		vx -= push
	}
	switch {
	case by < boidsMargin:
		vy += push
	case by > hiY-boidsMargin:
		vy -= push
	}
	nx, nvx := bounce(bx, vx, hiX)
	ny, nvy := bounce(by, vy, hiY)
	nx, ny = min(max(nx, 0), hiX), min(max(ny, 0), hiY) // a fast wall can throw it far
	if nx != bx || ny != by {
		b.x[i], b.y[i] = p.toWorld(nx, ny)
	}
	b.vx[i], b.vy[i] = limit(p.vecToWorld(nvx, nvy))
}

func bounce(p, v, hi int32) (int32, int32) {
	switch {
	case p < 0:
		return -p, abs32(v)
	case p > hi:
		return 2*hi - p, -abs32(v)
	}
	return p, v
}

// octant of a heading: 0 east, counting clockwise on screen (y down).
func octant(vx, vy int32) uint8 {
	ax, ay := abs32(vx), abs32(vy)
	// steep / shallow / diagonal by the tan(22.5°) ≈ 0.414 ≈ 106/256 ratio
	var o uint8
	switch {
	case ay*256 < ax*106:
		o = 0 // horizontal
	case ax*256 < ay*106:
		o = 2 // vertical
	default:
		o = 1 // diagonal
	}
	// map to the 8 directions by signs
	switch {
	case vx >= 0 && vy >= 0:
		return o // 0 E, 1 SE, 2 S
	case vx < 0 && vy >= 0:
		return 4 - o // 4 W, 3 SW, 2 S
	case vx < 0 && vy < 0:
		return (4 + o) & 7 // 4 W, 5 NW, 6 N
	default:
		return (8 - o) & 7 // 0 E, 7 NE, 6 N
	}
}

func (b *boids) render() {
	// range over a slice: ranging over the array field copies it, and TinyGo
	// puts the 3 KB copy on the heap every tick (findings F-35)
	for i, v := range b.cells[:] { // heads become trail tops, trails fade
		switch {
		case v > boidsTrail:
			b.cells[i] = boidsTrail
		case v > 0:
			b.cells[i] = v - 1
		}
	}
	const hiX, hiY = boidsW*boidsFix - 1, boidsH*boidsFix - 1
	for i := range boidsN[b.count] {
		bx, by := b.pose.toBox(b.x[i], b.y[i])
		bx, by = min(max(bx, 0), hiX), min(max(by, 0), hiY)
		c := int(by>>8)*boidsW + int(bx>>8)
		b.cells[c] = boidsTrail + 1 + octant(b.pose.vecToBox(b.vx[i], b.vy[i]))
	}
}

func (b *boids) tick() {
	t0 := time.Now()
	b.pose.advance()
	b.step()
	b.render()
	b.stepUS = time.Since(t0).Microseconds()
}

func (b *boids) fullGrid() *widget.CellGrid {
	if b.fullOn {
		return b.full
	}
	return nil
}

// fullPress: UP/DOWN turn the box 22.5° (the walls sweep through the
// flock), A shakes it, B back to the page, C leaves and changes page.
func (b *boids) fullPress(btn Button) bool {
	switch btn {
	case BtnUp:
		b.pose.target += boxTurnStep
		b.log("box turns")
	case BtnDown:
		b.pose.target -= boxTurnStep
		b.log("box turns")
	case BtnA:
		b.pose.shake()
	case BtnB:
		b.fullOn = false
	case BtnC:
		b.fullOn = false
		return false
	}
	return true
}

// scatter sends every boid away from the box's centre at top speed.
func (b *boids) scatter() {
	for i := range boidsN[b.count] {
		b.vx[i], b.vy[i] = limit(b.x[i]-b.pose.cx, b.y[i]-b.pose.cy)
	}
}

func (a *App) pageBoids() *page {
	b := newBoids()
	b.log = a.log
	count := widget.NewInteractiveLabelChoice(52, 11, boidsCounts,
		widget.WithLabelChoiceIndex(&b.count), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	scatter := widget.NewToggle(44, 12, fontBody, colFG, "GO", "GO", colTrack, colTrack,
		func() bool { return false }, func(bool) { b.scatter(); a.log("boids scatter") })
	spin := widget.NewToggle(44, 12, fontBody, colFG, "ON", "OFF", colGood, colTrack,
		func() bool { return b.pose.spin }, func(v bool) { b.pose.spin = v })
	full := widget.NewToggle(44, 12, fontBody, colFG, "SHOW", "SHOW", colTrack, colTrack,
		func() bool { return false }, func(bool) { b.fullOn = true; a.log("boids full screen") })
	// explicit height: findings F-33
	p := a.newPageWith("Boids", layout.VFlow(4, boidsH*boidsPitch),
		still(b.grid),
		newField("Boids", 42, 104, count),
		newField("Scatter", 42, 104, scatter),
		newField("Spin", 42, 104, spin),
		newField("Full", 42, 104, full),
		still(a.text(104, func() string { return b.txt.reset().str("step ").int(int(b.stepUS)).str(" us").String() }, colMuted)),
	)
	p.demo = b
	return p
}

func (b *boids) preview() []uint8      { return b.cells[:] }
func (b *boids) palette() []color.RGBA { return boidsPalette }
func (b *boids) showFull()             { b.fullOn = true }

// reset scatters every boid at a random place and heading in a straight box.
func (b *boids) reset() {
	b.pose = newBoxPose()
	for i := range boidsMax {
		b.x[i] = int32(rand32() % (boidsW * boidsFix))
		b.y[i] = int32(rand32() % (boidsH * boidsFix))
		b.vx[i] = int32(rand32()%(2*boidsVMax)) - boidsVMax
		b.vy[i] = int32(rand32()%(2*boidsVMax)) - boidsVMax
	}
	clear(b.cells[:])
}
