package app

import (
	"image/color"
	"time"

	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Pong on a 64×48 LED matrix. Positions are fixed point, 1/256 of a cell
// (no FPU on the M0+). In the page both paddles are CPU players; full
// screen the left paddle is yours (UP/DOWN).
const (
	pongW, pongH = 64, 48
	pongFix      = 256
	pongPadLen   = 8
	pongLeftX    = 2 // paddle columns
	pongRightX   = pongW - 3
	pongWin      = 9
	pongTicks    = 2 // physics ticks per frame
	pongServeVX  = 128
	pongMaxVX    = 320
	pongPitch    = 3
)

var pongLevels = []string{"easy", "normal", "hard"}

// CPU paddle speed per level, 1/256 cell per tick
var pongCPUSpeed = [3]int32{44, 72, 112}

// palette indices
const (
	pongOff = iota
	pongNet
	pongScore
	pongPaddle
	pongTrail1
	pongTrail2
	pongTrail3
	pongBall
)

var pongPalette = []color.RGBA{
	{14, 14, 22, 255}, {40, 44, 70, 255}, {90, 120, 160, 255}, {235, 240, 255, 255},
	{70, 30, 0, 255}, {140, 70, 10, 255}, {210, 120, 20, 255}, {255, 210, 70, 255},
}

// 3×5 digits, rows top to bottom, 3 bits per row (bit 2 = left column)
var pongDigits = [10][5]uint8{
	{7, 5, 5, 5, 7}, {2, 6, 2, 2, 7}, {7, 1, 7, 4, 7}, {7, 1, 7, 1, 7}, {5, 5, 7, 1, 1},
	{7, 4, 7, 1, 7}, {7, 4, 7, 5, 7}, {7, 1, 1, 1, 1}, {7, 5, 7, 5, 7}, {7, 5, 7, 1, 7},
}

type pong struct {
	cells          [pongW * pongH]uint8
	bx, by, vx, vy int32
	pad            [2]int32 // paddle tops (left, right)
	target         int32    // player's paddle target (top)
	score          [2]int
	trail          [3][2]int32 // previous ball cells, newest first
	level          int
	player         bool // left paddle follows target instead of the CPU
	paused         bool
	stepUS         int64
	grid, full     *widget.CellGrid
	fullOn         bool
	txt            textBuf
	log            func(string)
}

func newPong() *pong {
	p := &pong{level: 1}
	p.pad[0] = (pongH - pongPadLen) / 2 * pongFix
	p.pad[1] = p.pad[0]
	p.target = p.pad[0]
	p.serve(1)
	p.grid = widget.NewCellGrid(pongW, pongH, pongPitch, 1, widget.CellRound, p.cells[:], pongPalette, colBG)
	p.full = widget.NewCellGrid(pongW, pongH, 5, 1, widget.CellRound, p.cells[:], pongPalette, colBG)
	p.render()
	return p
}

// serve puts the ball in the centre moving towards side dir (-1 left, +1 right).
func (p *pong) serve(dir int32) {
	p.bx, p.by = pongW/2*pongFix, pongH/2*pongFix
	p.vx = dir * pongServeVX
	p.vy = int32(rand32()&63) + 24
	if rand32()&1 == 0 {
		p.vy = -p.vy
	}
	for i := range p.trail {
		p.trail[i] = [2]int32{p.bx >> 8, p.by >> 8}
	}
}

// step advances the physics one tick.
func (p *pong) step() {
	for side := range 2 {
		var speed, goal int32
		if side == 0 && p.player {
			speed, goal = 160, p.target
		} else {
			speed = pongCPUSpeed[p.level]
			goal = (pongH - pongPadLen) / 2 * pongFix // drift to the centre
			if (side == 0) == (p.vx < 0) {            // ball coming: follow it
				goal = p.by - pongPadLen/2*pongFix
			}
		}
		d := min(max(goal-p.pad[side], -speed), speed)
		p.pad[side] = min(max(p.pad[side]+d, 0), (pongH-pongPadLen)*pongFix)
	}

	prev := p.bx
	p.bx += p.vx
	p.by += p.vy
	const bottom = (pongH - 1) * pongFix
	if p.by < 0 {
		p.by, p.vy = -p.by, -p.vy
	} else if p.by > bottom {
		p.by, p.vy = 2*bottom-p.by, -p.vy
	}
	// a paddle is hit when the ball crosses into its column this tick (a
	// crossing test: at top speed the ball moves more than a cell per tick)
	const leftFace, rightFace = (pongLeftX + 1) * pongFix, pongRightX * pongFix
	switch {
	case p.vx < 0 && prev >= leftFace && p.bx < leftFace:
		p.hit(0, leftFace)
	case p.vx > 0 && prev < rightFace && p.bx >= rightFace:
		p.hit(1, rightFace-1)
	case p.bx < 0:
		p.point(1)
	case p.bx >= pongW*pongFix:
		p.point(0)
	}
}

// hit bounces the ball off paddle side if it is in reach, angling it by
// where it struck the paddle and speeding it up by 1/16.
func (p *pong) hit(side int, face int32) {
	off := p.by + pongFix/2 - p.pad[side] // ball centre from the paddle top
	if off < 0 || off >= pongPadLen*pongFix {
		return // missed: the ball flies on and scores
	}
	p.bx = face
	vx := -p.vx
	vx += vx >> 4
	p.vx = min(max(vx, -pongMaxVX), pongMaxVX)
	p.vy = (off - pongPadLen/2*pongFix) >> 3
}

// point scores for side (0 left, 1 right); after pongWin points both reset.
func (p *pong) point(side int) {
	p.score[side]++
	if p.score[side] > pongWin {
		p.score = [2]int{}
	}
	p.serve(int32(1 - 2*side)) // towards the side that missed
}

// render draws the court into the cells.
func (p *pong) render() {
	clear(p.cells[:])
	for y := 0; y < pongH; y += 2 {
		p.cells[y*pongW+pongW/2] = pongNet
	}
	p.digit(pongW/2-6, 1, p.score[0])
	p.digit(pongW/2+4, 1, p.score[1])
	for side, x := range [2]int{pongLeftX, pongRightX} {
		top := int(p.pad[side] >> 8)
		for y := top; y < top+pongPadLen; y++ {
			p.cells[y*pongW+x] = pongPaddle
		}
	}
	for i := len(p.trail) - 1; i >= 0; i-- {
		t := p.trail[i]
		p.set(int(t[0]), int(t[1]), pongTrail3-uint8(i))
	}
	p.set(int(p.bx>>8), int(p.by>>8), pongBall)
}

func (p *pong) set(x, y int, v uint8) {
	if x >= 0 && y >= 0 && x < pongW && y < pongH {
		p.cells[y*pongW+x] = v
	}
}

func (p *pong) digit(x, y, d int) {
	for r, bits := range pongDigits[d] {
		for c := range 3 {
			if bits&(4>>c) != 0 {
				p.cells[(y+r)*pongW+x+c] = pongScore
			}
		}
	}
}

func (p *pong) tick() {
	if p.paused {
		return
	}
	t0 := time.Now()
	for range pongTicks {
		cell := [2]int32{p.bx >> 8, p.by >> 8}
		p.step()
		if cell != p.trail[0] {
			copy(p.trail[1:], p.trail[:len(p.trail)-1])
			p.trail[0] = cell
		}
	}
	p.render()
	p.stepUS = time.Since(t0).Microseconds()
}

func (p *pong) fullGrid() *widget.CellGrid {
	if p.fullOn {
		return p.full
	}
	return nil
}

// fullPress: UP/DOWN move your paddle (6 cells per press or repeat), A
// pauses, B back to the page, C leaves and changes page.
func (p *pong) fullPress(b Button) bool {
	switch b {
	case BtnUp:
		p.target = max(p.target-6*pongFix, 0)
	case BtnDown:
		p.target = min(p.target+6*pongFix, (pongH-pongPadLen)*pongFix)
	case BtnA:
		p.paused = !p.paused
	case BtnB:
		p.leave()
	case BtnC:
		p.leave()
		return false
	}
	return true
}

func (p *pong) leave() {
	p.fullOn, p.player, p.paused = false, false, false
}

func (a *App) pagePong() *page {
	p := newPong()
	p.log = a.log
	level := widget.NewInteractiveLabelChoice(52, 11, pongLevels,
		widget.WithLabelChoiceIndex(&p.level), widget.WithLabelChoiceFont(fontBody), widget.WithLabelChoiceColor(colFG))
	play := widget.NewToggle(44, 12, fontBody, colFG, "PLAY", "PLAY", colTrack, colTrack,
		func() bool { return false }, func(bool) { p.showFull(); a.log("pong: you vs cpu") })
	// explicit height: findings F-33
	pg := a.newPageWith("Pong", layout.VFlow(4, pongH*pongPitch),
		still(p.grid),
		newField("CPU", 42, 104, level),
		newField("You", 42, 104, play),
		a.caption(104, "UP/DOWN: paddle"),
		a.caption(104, "A pause, B back"),
		still(a.text(104, func() string { return p.txt.reset().str("step ").int(int(p.stepUS)).str(" us").String() }, colMuted)),
	)
	pg.demo = p
	return pg
}

func (p *pong) preview() []uint8      { return p.cells[:] }
func (p *pong) palette() []color.RGBA { return pongPalette }

// showFull starts a game against the CPU on the whole screen.
func (p *pong) showFull() {
	p.fullOn, p.player, p.score = true, true, [2]int{}
	p.target = p.pad[0]
}

func (p *pong) reset() { p.score = [2]int{}; p.serve(1) }
