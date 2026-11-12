package app

import "testing"

func (w *water) wet() int {
	n := 0
	for _, s := range w.state[:] {
		if s&liquidBit != 0 {
			n++
		}
	}
	return n
}

// heights is the water column height in every x (gravity down).
func (w *water) heights() (h [waterN]int) {
	for x := range waterN {
		for y := range waterN {
			if w.state[y*waterN+x]&liquidBit != 0 {
				h[x]++
			}
		}
	}
	return h
}

// A column of water dropped in the middle of an empty basin must spread
// into a level layer (a liquid, not a sand heap), evenly to both sides,
// without losing or creating water. One second on the badge is ~45 steps.
func TestWaterLevelsOutLikeALiquid(t *testing.T) {
	w := newWater()
	clear(w.state[:])
	for y := 16; y < waterN; y++ {
		for x := 28; x < 36; x++ {
			w.state[y*waterN+x] = liquidBit
		}
	}
	n := w.wet()   // 8 × 48 = 384 → 6 rows when level
	for range 90 { // 2 s
		w.step()
	}
	if got := w.wet(); got != n {
		t.Fatalf("water not conserved: %d cells, want %d", got, n)
	}
	h := w.heights()
	lo, hi := 99, 0
	for _, v := range h {
		lo, hi = min(lo, v), max(hi, v)
	}
	if hi-lo > 1 {
		t.Errorf("surface not level after 2 s: heights %d..%d: %v", lo, hi, h)
	}
	left, right := 0, 0
	for x := range waterN / 2 {
		left += h[x]
		right += h[waterN-1-x]
	}
	if d := left - right; d > n/20 || d < -n/20 {
		t.Errorf("spread unevenly: left %d, right %d", left, right)
	}
}

// Turning gravity to the left, the basin's water must become a wall against
// the left side with a vertical surface within 3 s (135 steps at 45 steps a
// second). Moving a cell one place a step, it took 392 steps.
func TestWaterFollowsGravityQuickly(t *testing.T) {
	w := newWater() // bottom 40 % full
	n := w.wet()
	w.dir = 2 // left
	for range 135 {
		w.step()
	}
	if got := w.wet(); got != n {
		t.Fatalf("water not conserved: %d cells, want %d", got, n)
	}
	lo, hi := 99, 0
	for y := range waterN {
		width := 0
		for x := range waterN {
			if w.state[y*waterN+x]&liquidBit != 0 {
				width++
			}
		}
		lo, hi = min(lo, width), max(hi, width)
	}
	if hi-lo > 1 {
		t.Errorf("surface not vertical after 3 s: row widths %d..%d", lo, hi)
	}
}

// Every one of the 8 gravity directions moves the water: settled against
// the opposite side, it must travel towards the new "down" within 1 s (45
// steps). The owner saw water "not reacting to gravity change".
func TestWaterReactsToEveryGravityDirection(t *testing.T) {
	com := func(w *water) (float64, float64) {
		var sx, sy, n float64
		for y := range waterN {
			for x := range waterN {
				if w.state[y*waterN+x]&liquidBit != 0 {
					sx, sy, n = sx+float64(x), sy+float64(y), n+1
				}
			}
		}
		return sx / n, sy / n
	}
	for d := range 8 {
		w := newWater()
		w.dir = (d + 4) % 8 // settle on the opposite side first
		for range 135 {
			w.step()
		}
		x0, y0 := com(w)
		w.dir = d
		for range 45 {
			w.step()
		}
		x1, y1 := com(w)
		g := gravityVec[d]
		if moved := (x1-x0)*float64(g[0]) + (y1-y0)*float64(g[1]); moved < 8 {
			t.Errorf("gravity %s: the water moved only %.1f cells towards it in 1 s", gravityNames[d], moved)
		}
	}
}
