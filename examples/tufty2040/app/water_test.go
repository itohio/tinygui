package app

import "testing"

func (w *water) cells() int {
	n := 0
	for _, s := range w.state[:] {
		if s&waterBit != 0 {
			n++
		}
	}
	return n
}

// heights is the water column height in every x (gravity down).
func (w *water) heights() (h [waterN]int) {
	for x := range waterN {
		for y := range waterN {
			if w.state[y*waterN+x]&waterBit != 0 {
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
			w.state[y*waterN+x] = waterBit
		}
	}
	n := w.cells() // 8 × 48 = 384 → 6 rows when level
	for range 90 { // 2 s
		w.step()
	}
	if got := w.cells(); got != n {
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
	n := w.cells()
	w.dir = 2 // left
	for range 135 {
		w.step()
	}
	if got := w.cells(); got != n {
		t.Fatalf("water not conserved: %d cells, want %d", got, n)
	}
	lo, hi := 99, 0
	for y := range waterN {
		width := 0
		for x := range waterN {
			if w.state[y*waterN+x]&waterBit != 0 {
				width++
			}
		}
		lo, hi = min(lo, width), max(hi, width)
	}
	if hi-lo > 1 {
		t.Errorf("surface not vertical after 3 s: row widths %d..%d", lo, hi)
	}
}
